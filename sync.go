package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

// Configuration structures
type Config struct {
	GitHub struct {
		RepoName              string `json:"repo_name"`
		CommitMessageTemplate string `json:"commit_message_template"`
	} `json:"github"`
	CategoryMappings   map[string]string `json:"category_mappings"`
	LanguageExtensions map[string]string `json:"language_extensions"`
}

// LeetCode API response structures
type Submission struct {
	ID         int    `json:"id"`
	Lang       string `json:"lang"`
	StatusCode int    `json:"status_code"` // 10 = Accepted (0 also treated as accepted)
	Title      string `json:"title"`
	TitleSlug  string `json:"title_slug"`
	Timestamp  int64  `json:"timestamp"`
}

// Not currently used, kept for future extension
type SubmissionDetail struct {
	Code      string `json:"code"`
	Lang      string `json:"lang"`
	Question  string `json:"question"`
	StatusMsg string `json:"status_msg"`
}

type TopicTag struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProblemDetail struct {
	QuestionID         string     `json:"questionId"`
	QuestionFrontendID string     `json:"questionFrontendId"`
	Title              string     `json:"title"`
	TitleSlug          string     `json:"titleSlug"`
	Difficulty         string     `json:"difficulty"`
	TopicTags          []TopicTag `json:"topicTags"`
	Content            string     `json:"content"`
}

type GraphQLResponse struct {
	Data struct {
		Question ProblemDetail `json:"question"`
	} `json:"data"`
}

// Sync database structure
type SyncDatabase struct {
	Synced     map[string]SyncEntry    `json:"synced"`
	Failed     map[string]*FailedEntry `json:"failed,omitempty"`
	LastSynced time.Time               `json:"last_synced"`
}

type SyncEntry struct {
	SubmissionID int       `json:"submission_id"`
	ProblemID    string    `json:"problem_id"`
	Title        string    `json:"title"`
	TitleSlug    string    `json:"title_slug"`
	Difficulty   string    `json:"difficulty"`
	Category     string    `json:"category"`
	Timestamp    time.Time `json:"timestamp"`
}

type FailedEntry struct {
	SubmissionID int       `json:"submission_id"`
	Title        string    `json:"title"`
	TitleSlug    string    `json:"title_slug"`
	LastError    string    `json:"last_error"`
	RetryCount   int       `json:"retry_count"`
	LastTried    time.Time `json:"last_tried"`
}

// Global variables
var (
	config       Config
	syncDB       SyncDatabase
	sessionToken string
	csrfToken    string
	debugMode    bool
)

const (
	syncDBFile          = ".leetcode_sync_db.json"
	submissionsEndpoint = "https://leetcode.com/api/submissions/"
	graphqlEndpoint     = "https://leetcode.com/graphql"

	maxRetriesPerSubmission = 5
	staleFailedDays         = 30 // remove failed entries older than this with max retries reached
)

func main() {
	fmt.Println("🚀 LeetCode → GitHub Sync Daemon")
	fmt.Println("==================================")
	fmt.Println("📅 Schedule: Continuous (every 10 minutes)")
	fmt.Println("⏱️  Frequency: Every 10 minutes")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found. Make sure LEETCODE_SESSION is set.")
	}

	sessionToken = os.Getenv("LEETCODE_SESSION")
	if sessionToken == "" {
		fmt.Println("❌ Error: LEETCODE_SESSION environment variable not set")
		os.Exit(1)
	}

	csrfToken = os.Getenv("CSRF_TOKEN")
	if csrfToken == "" {
		fmt.Println("❌ Error: CSRF_TOKEN environment variable not set")
		fmt.Println("💡 Get it from browser cookies (Application > Cookies > leetcode.com > csrftoken)")
		os.Exit(1)
	}

	// Validate session immediately
	fmt.Println("🔑 Validating session...")
	if err := validateSession(); err != nil {
		fmt.Printf("❌ Session validation failed: %v\n", err)
		fmt.Println("⚠️  Your LEETCODE_SESSION cookie might be expired. Please update it in .env")
		os.Exit(1)
	}
	fmt.Println("✅ Session is valid!")

	debugMode = os.Getenv("DEBUG") == "true"

	// Load configuration
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Load sync database
	if err := loadSyncDB(); err != nil {
		fmt.Printf("⚠️  Creating new sync database: %v\n", err)
		syncDB = SyncDatabase{
			Synced: make(map[string]SyncEntry),
			Failed: make(map[string]*FailedEntry),
		}
	}

	// Check for pending changes that failed to push previously
	fmt.Println("🔍 Checking for unpushed changes...")
	if err := checkAndPushPendingChanges(); err != nil {
		fmt.Printf("⚠️  Warning: Could not push pending changes: %v\n", err)
	}

	// Setup signal handling for manual trigger (kill -USR1 <pid> inside container)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)

	// Run sync immediately on startup
	fmt.Println("🚀 Starting initial sync...")
	performSync()

	// Daemon Loop - always sync every 10 minutes
	for {
		fmt.Printf("\n[%s] ⏰ Time to sync!\n", time.Now().Format("15:04:05"))
		performSync()

		fmt.Printf("[%s] 💤 Sleeping for 10 minutes (or until manual trigger)...\n", time.Now().Format("15:04:05"))

		// Wait for 10 minutes OR a manual trigger signal
		select {
		case <-time.After(10 * time.Minute):
			// continue loop
		case <-sigChan:
			fmt.Println("\n⚡ Manual sync triggered!")
			performSync()
		}
	}
}

func performSync() {
	fmt.Println("📥 Fetching recent submissions...")
	submissions, err := fetchRecentSubmissions()
	if err != nil {
		fmt.Printf("❌ Error fetching submissions: %v\n", err)
		return
	}
	fmt.Printf("   ✓ Fetched %d total submissions from API\n", len(submissions))

	// Log first few submissions for debugging
	if len(submissions) > 0 {
		fmt.Println("\n📋 Latest 10 submissions:")
		limit := 10
		if len(submissions) < limit {
			limit = len(submissions)
		}
		for i := 0; i < limit; i++ {
			sub := submissions[i]
			statusText := "Unknown"
			switch sub.StatusCode {
			case 0:
				statusText = "Accepted (Status 0) ✅"
			case 10:
				statusText = "Accepted (Status 10) ✅"
			case 11:
				statusText = "Wrong Answer ❌"
			case 12:
				statusText = "Memory Limit Exceeded"
			case 13:
				statusText = "Output Limit Exceeded"
			case 14:
				statusText = "Time Limit Exceeded ⏱️"
			case 15:
				statusText = "Runtime Error 💥"
			case 16:
				statusText = "Internal Error"
			case 20:
				statusText = "Compile Error 🔨"
			default:
				statusText = fmt.Sprintf("Status %d", sub.StatusCode)
			}
			fmt.Printf("   %d. [%s] %s (ID: %d, Lang: %s)\n",
				i+1, statusText, sub.Title, sub.ID, sub.Lang)
		}
		fmt.Println()
	}

	// Filter accepted submissions
	acceptedSubmissions := filterAcceptedSubmissions(submissions)
	fmt.Printf("   ✓ Found %d accepted submissions (after filtering)\n", len(acceptedSubmissions))

	if len(acceptedSubmissions) == 0 {
		fmt.Println("✨ No accepted submissions found in recent history!")
		return
	}

	// Check sync database
	fmt.Printf("\n🔍 Checking sync database...\n")
	fmt.Printf("   Database has %d previously synced submissions\n", len(syncDB.Synced))
	if len(syncDB.Failed) > 0 {
		fmt.Printf("   Database has %d failed submission(s) pending retry\n", len(syncDB.Failed))
	}
	if !syncDB.LastSynced.IsZero() {
		fmt.Printf("   Last synced at: %s\n", syncDB.LastSynced.Format(time.RFC3339))
	}

	// Process new / failed submissions
	newCount := 0
	skippedCount := 0
	failedThisRun := 0
	retriedFixed := 0

	for i, sub := range acceptedSubmissions {
		submissionKey := fmt.Sprintf("%d", sub.ID)

		// Already synced
		if _, exists := syncDB.Synced[submissionKey]; exists {
			skippedCount++
			fmt.Printf("   ⏭️  [%d/%d] Skipping (already synced): %s (ID: %d)\n",
				i+1, len(acceptedSubmissions), sub.Title, sub.ID)
			continue
		}

		// Check if previously failed
		failedEntry, hadFailed := syncDB.Failed[submissionKey]
		if hadFailed && failedEntry.RetryCount >= maxRetriesPerSubmission {
			fmt.Printf("   ⏭️  [%d/%d] Skipping (max retries reached): %s (ID: %d). Last error: %s\n",
				i+1, len(acceptedSubmissions), sub.Title, sub.ID, failedEntry.LastError)
			continue
		}

		if hadFailed {
			fmt.Printf("\n🔄 [%d/%d] Retrying FAILED submission: %s\n", i+1, len(acceptedSubmissions), sub.Title)
		} else {
			fmt.Printf("\n🔄 [%d/%d] Processing NEW submission: %s\n", i+1, len(acceptedSubmissions), sub.Title)
		}
		fmt.Printf("   Submission ID: %d\n", sub.ID)
		fmt.Printf("   Language: %s\n", sub.Lang)
		fmt.Printf("   Title Slug: %s\n", sub.TitleSlug)

		if err := processSubmission(sub); err != nil {
			fmt.Printf("   ❌ Error processing: %v\n", err)
			failedThisRun++

			now := time.Now()
			if !hadFailed {
				failedEntry = &FailedEntry{
					SubmissionID: sub.ID,
					Title:        sub.Title,
					TitleSlug:    sub.TitleSlug,
				}
			}
			failedEntry.RetryCount++
			failedEntry.LastError = err.Error()
			failedEntry.LastTried = now
			syncDB.Failed[submissionKey] = failedEntry
			continue
		}

		// Success
		if hadFailed {
			retriedFixed++
			delete(syncDB.Failed, submissionKey)
		}
		newCount++
		fmt.Printf("   ✅ Successfully processed!\n")
	}

	// Cleanup stale failed entries
	if removed := cleanupFailedEntries(); removed > 0 {
		fmt.Printf("\n🧹 Cleaned up %d stale failed entrie(s) from sync DB\n", removed)
	}

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   Total accepted considered: %d\n", len(acceptedSubmissions))
	fmt.Printf("   Already synced (skipped): %d\n", skippedCount)
	fmt.Printf("   Newly synced this run: %d\n", newCount)
	fmt.Printf("   Failed this run: %d\n", failedThisRun)
	if retriedFixed > 0 {
		fmt.Printf("   Retried & fixed this run: %d\n", retriedFixed)
	}
	if len(syncDB.Failed) > 0 {
		fmt.Printf("   Total pending failed submissions: %d\n", len(syncDB.Failed))
	}

	// Update master README (always update to ensure latest stats/links)
	fmt.Println("\n📝 Updating master README...")
	if err := updateMasterREADME(); err != nil {
		fmt.Printf("⚠️  Error updating README: %v\n", err)
	} else {
		fmt.Println("   ✓ README updated")
	}

	if newCount == 0 && failedThisRun == 0 && retriedFixed == 0 {
		fmt.Println("\n✨ No changes to sync (no new or recovered submissions).")

		// Check for any changes (e.g. README updates)
		if err := checkAndPushPendingChanges(); err != nil {
			fmt.Printf("⚠️  Warning: Could not push pending changes: %v\n", err)
		}
		return
	}

	// Save sync database
	fmt.Println("\n💾 Saving sync database...")
	syncDB.LastSynced = time.Now()
	if err := saveSyncDB(); err != nil {
		fmt.Printf("❌ Error saving sync database: %v\n", err)
	} else {
		fmt.Println("   ✓ Database saved")
	}

	// Git operations
	fmt.Println("\n📤 Pushing to GitHub...")
	if err := gitAddCommitPush(newCount + retriedFixed); err != nil {
		fmt.Printf("❌ Error with git operations: %v\n", err)
	} else {
		fmt.Println("   ✅ Successfully pushed to GitHub!")
	}

	fmt.Printf("\n🎉 Sync cycle complete! Processed %d successful submission(s)\n", newCount+retriedFixed)
}

func loadConfig() error {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &config)
}

func loadSyncDB() error {
	data, err := os.ReadFile(syncDBFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &syncDB); err != nil {
		return err
	}
	if syncDB.Synced == nil {
		syncDB.Synced = make(map[string]SyncEntry)
	}
	if syncDB.Failed == nil {
		syncDB.Failed = make(map[string]*FailedEntry)
	}
	return nil
}

func saveSyncDB() error {
	data, err := json.MarshalIndent(syncDB, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(syncDBFile, data, 0644)
}

func validateSession() error {
	// Query global data to check if session is valid
	query := `{"query": "query globalData { userStatus { isSignedIn } }"}`

	req, err := http.NewRequest("POST", graphqlEndpoint, strings.NewReader(query))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("LEETCODE_SESSION=%s", sessionToken))
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			UserStatus struct {
				IsSignedIn bool `json:"isSignedIn"`
			} `json:"userStatus"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Data.UserStatus.IsSignedIn {
		return fmt.Errorf("session is invalid or expired (isSignedIn: false)")
	}

	return nil
}

func fetchRecentSubmissions() ([]Submission, error) {
	maxSubmissions := 10
	if val := os.Getenv("MAX_SUBMISSIONS_TO_CHECK"); val != "" {
		fmt.Sscanf(val, "%d", &maxSubmissions)
	}

	url := fmt.Sprintf("%s?offset=0&limit=%d", submissionsEndpoint, maxSubmissions)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cookie", fmt.Sprintf("LEETCODE_SESSION=%s", sessionToken))
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		SubmissionsData []Submission `json:"submissions_dump"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.SubmissionsData, nil
}

func filterAcceptedSubmissions(submissions []Submission) []Submission {
	var accepted []Submission
	seen := make(map[string]bool)

	for _, sub := range submissions {
		// Status 10 = Accepted, Status 0 also appears to be Accepted in some cases
		if (sub.StatusCode == 10 || sub.StatusCode == 0) && !seen[sub.TitleSlug] {
			accepted = append(accepted, sub)
			seen[sub.TitleSlug] = true
		}
	}

	return accepted
}

func processSubmission(sub Submission) error {
	// Fetch problem details
	problemDetail, err := fetchProblemDetail(sub.TitleSlug)
	if err != nil {
		return fmt.Errorf("fetching problem details: %w", err)
	}

	// Fetch submission code
	code, err := fetchSubmissionCode(sub.ID)
	if err != nil {
		return fmt.Errorf("fetching submission code: %w", err)
	}

	// Determine category and file paths
	category := determineCategory(problemDetail.TopicTags)
	difficulty := problemDetail.Difficulty
	extension := getFileExtension(sub.Lang)

	// Create directory structure
	// Structure: Category/Difficulty/ID.TitleSlug/
	problemFolder := fmt.Sprintf("%s.%s", problemDetail.QuestionFrontendID, sanitizeFilename(sub.TitleSlug))
	dirPath := filepath.Join(category, difficulty, problemFolder)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Create filename (inside the problem folder)
	filename := fmt.Sprintf("%s.%s.%s",
		problemDetail.QuestionFrontendID,
		sanitizeFilename(sub.TitleSlug),
		extension)

	filePath := filepath.Join(dirPath, filename)

	// Write solution file
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return fmt.Errorf("writing solution file: %w", err)
	}
	fmt.Printf("  ✓ Created: %s\n", filePath)

	// Create problem README
	readmePath := filepath.Join(dirPath, "README.md")

	if err := createProblemREADME(readmePath, problemDetail, sub.Lang); err != nil {
		return fmt.Errorf("creating problem README: %w", err)
	}
	fmt.Printf("  ✓ Created: %s\n", readmePath)

	// Update sync database
	syncDB.Synced[fmt.Sprintf("%d", sub.ID)] = SyncEntry{
		SubmissionID: sub.ID,
		ProblemID:    problemDetail.QuestionFrontendID,
		Title:        problemDetail.Title,
		TitleSlug:    problemDetail.TitleSlug,
		Difficulty:   difficulty,
		Category:     category,
		Timestamp:    time.Now(),
	}

	return nil
}

func fetchProblemDetail(titleSlug string) (ProblemDetail, error) {
	query := fmt.Sprintf(`{
		"query": "query getQuestionDetail($titleSlug: String!) { question(titleSlug: $titleSlug) { questionId questionFrontendId title titleSlug difficulty topicTags { name slug } content } }",
		"variables": {"titleSlug": "%s"}
	}`, titleSlug)

	req, err := http.NewRequest("POST", graphqlEndpoint, strings.NewReader(query))
	if err != nil {
		return ProblemDetail{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("LEETCODE_SESSION=%s", sessionToken))
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ProblemDetail{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return ProblemDetail{}, fmt.Errorf("problem detail query returned status %d: %s", resp.StatusCode, string(body))
	}

	var result GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ProblemDetail{}, err
	}

	return result.Data.Question, nil
}

func fetchSubmissionCode(submissionID int) (string, error) {
	// GraphQL query for submission details (updated to new schema)
	query := fmt.Sprintf(`{
		"query": "query submissionDetails($submissionId: Int!) { submissionDetails(submissionId: $submissionId) { code runtimeDisplay memoryDisplay lang { slug name } } }",
		"variables": {"submissionId": %d}
	}`, submissionID)

	req, err := http.NewRequest("POST", graphqlEndpoint, strings.NewReader(query))
	if err != nil {
		return "", err
	}

	// Set required headers for GraphQL API
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("LEETCODE_SESSION=%s; csrftoken=%s", sessionToken, csrfToken))
	req.Header.Set("x-csrftoken", csrfToken)
	req.Header.Set("Referer", "https://leetcode.com")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Debug: Log response status
	fmt.Printf("   [DEBUG] GraphQL API status: %d\n", resp.StatusCode)

	// Read response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		bodyStr := string(body)
		fmt.Printf("   [DEBUG] Response body: %s\n", bodyStr)
		if strings.Contains(bodyStr, "Cannot query field") || strings.Contains(bodyStr, "must have a sub selection") {
			fmt.Println("   [DEBUG] Possible GraphQL schema change detected. Please verify the submissionDetails query.")
		}
		return "", fmt.Errorf("GraphQL API returned status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			SubmissionDetails struct {
				Code           string `json:"code"`
				RuntimeDisplay string `json:"runtimeDisplay"`
				MemoryDisplay  string `json:"memoryDisplay"`
				Lang           struct {
					Slug string `json:"slug"`
					Name string `json:"name"`
				} `json:"lang"`
			} `json:"submissionDetails"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("   [DEBUG] Failed to parse response: %s\n", string(body))
		return "", err
	}

	code := result.Data.SubmissionDetails.Code

	// Debug: Log code length and metadata
	fmt.Printf("   [DEBUG] Code length: %d chars\n", len(code))
	if len(code) > 0 {
		fmt.Printf("   [DEBUG] Runtime: %s, Memory: %s, Lang: %s (%s)\n",
			result.Data.SubmissionDetails.RuntimeDisplay,
			result.Data.SubmissionDetails.MemoryDisplay,
			result.Data.SubmissionDetails.Lang.Slug,
			result.Data.SubmissionDetails.Lang.Name)
	}

	return code, nil
}

func determineCategory(tags []TopicTag) string {
	if len(tags) == 0 {
		return "Miscellaneous"
	}

	// Try to find a mapped category
	for _, tag := range tags {
		if category, exists := config.CategoryMappings[tag.Slug]; exists {
			return category
		}
	}

	// Use first tag as fallback
	return capitalizeFirst(tags[0].Name)
}

func getFileExtension(lang string) string {
	lang = strings.ToLower(lang)
	if ext, exists := config.LanguageExtensions[lang]; exists {
		return ext
	}
	return "txt"
}

func sanitizeFilename(name string) string {
	// Remove special characters and replace spaces/hyphens with hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-]+`)
	name = reg.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	return strings.ToLower(name)
}

func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func createProblemREADME(path string, problem ProblemDetail, lang string) error {
	tags := make([]string, len(problem.TopicTags))
	for i, tag := range problem.TopicTags {
		tags[i] = tag.Name
	}

	content := fmt.Sprintf(`# %s. %s

**Difficulty:** %s  
**Topics:** %s  
**Language:** %s

## Problem Link
[LeetCode Problem](https://leetcode.com/problems/%s/)

## Problem Description
%s
`,
		problem.QuestionFrontendID,
		problem.Title,
		problem.Difficulty,
		strings.Join(tags, ", "),
		lang,
		problem.TitleSlug,
		stripHTML(problem.Content))

	return os.WriteFile(path, []byte(content), 0644)
}

func stripHTML(html string) string {
	// Basic HTML stripping - removes tags
	reg := regexp.MustCompile(`<[^>]*>`)
	text := reg.ReplaceAllString(html, "")

	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&quot;", "\"")

	// Limit length for README
	if len(text) > 10000 {
		text = text[:10000] + "..."
	}

	return strings.TrimSpace(text)
}

func updateMasterREADME() error {
	// Gather statistics
	stats := make(map[string]map[string]int) // category -> difficulty -> count
	var entries []SyncEntry

	for _, entry := range syncDB.Synced {
		entries = append(entries, entry)

		if stats[entry.Category] == nil {
			stats[entry.Category] = make(map[string]int)
		}
		stats[entry.Category][entry.Difficulty]++
	}

	// Sort entries by timestamp (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	// Calculate totals
	totalEasy, totalMedium, totalHard := 0, 0, 0
	for _, diffMap := range stats {
		totalEasy += diffMap["Easy"]
		totalMedium += diffMap["Medium"]
		totalHard += diffMap["Hard"]
	}
	totalProblems := totalEasy + totalMedium + totalHard

	// Build README content
	var sb strings.Builder
	sb.WriteString("# 🚀 LeetCode Solutions\n\n")
	sb.WriteString("A collection of my LeetCode solutions, automatically synced from LeetCode.\n\n")
	sb.WriteString("## 📊 Progress Statistics\n\n")
	sb.WriteString(fmt.Sprintf("**Total Problems Solved:** %d\n\n", totalProblems))
	sb.WriteString(fmt.Sprintf("- 🟢 Easy: %d\n", totalEasy))
	sb.WriteString(fmt.Sprintf("- 🟡 Medium: %d\n", totalMedium))
	sb.WriteString(fmt.Sprintf("- 🔴 Hard: %d\n\n", totalHard))

	// Category breakdown
	sb.WriteString("## 📂 Solutions by Category\n\n")

	categories := make([]string, 0, len(stats))
	for cat := range stats {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	for _, cat := range categories {
		diffMap := stats[cat]
		total := diffMap["Easy"] + diffMap["Medium"] + diffMap["Hard"]
		sb.WriteString(fmt.Sprintf("### %s (%d)\n", cat, total))
		sb.WriteString(fmt.Sprintf("- Easy: %d | Medium: %d | Hard: %d\n\n",
			diffMap["Easy"], diffMap["Medium"], diffMap["Hard"]))
	}

	// Recent submissions
	sb.WriteString("## 🕒 Recent Submissions\n\n")
	sb.WriteString("| # | Problem | Difficulty | Category | Date |\n")
	sb.WriteString("|---|---------|------------|----------|------|\n")

	recentCount := 10
	if len(entries) < recentCount {
		recentCount = len(entries)
	}

	for i := 0; i < recentCount; i++ {
		entry := entries[i]
		diffEmoji := getDifficultyEmoji(entry.Difficulty)

		// Create clickable link to the specific problem folder
		// Format: ./Category/Difficulty/ID.TitleSlug
		problemPath := fmt.Sprintf("./%s/%s/%s.%s",
			entry.Category, entry.Difficulty, entry.ProblemID, entry.TitleSlug)

		sb.WriteString(fmt.Sprintf("| %s | [%s](%s) | %s %s | %s | %s |\n",
			entry.ProblemID,
			entry.Title,
			problemPath,
			diffEmoji,
			entry.Difficulty,
			entry.Category,
			entry.Timestamp.Format("2006-01-02")))
	}

	sb.WriteString("\n---\n\n")
	sb.WriteString("*This repository is automatically updated using a custom sync tool.*\n")

	return os.WriteFile("README.md", []byte(sb.String()), 0644)
}

func getDifficultyEmoji(difficulty string) string {
	switch difficulty {
	case "Easy":
		return "🟢"
	case "Medium":
		return "🟡"
	case "Hard":
		return "🔴"
	default:
		return "⚪"
	}
}

func checkAndPushPendingChanges() error {
	// Check if there are any untracked or uncommitted files
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	hasChanges := len(strings.TrimSpace(string(output))) > 0

	if hasChanges {
		fmt.Println("   📦 Found unpushed changes from previous run")
		fmt.Println("   🔄 Attempting to push pending changes...")

		// Count the number of synced items to use in commit message
		count := len(syncDB.Synced)
		if count == 0 {
			count = 1 // At least 1 if there are changes
		}

		// Try to push the changes
		if err := gitAddCommitPush(count); err != nil {
			return fmt.Errorf("failed to push pending changes: %w", err)
		}

		fmt.Println("   ✅ Successfully pushed pending changes!")
	} else {
		fmt.Println("   ✓ No pending changes found")
	}

	return nil
}

func ensureGitSetup() error {
	// Check if .git directory exists
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		fmt.Println("   Initializing Git repository...")
		cmd := exec.Command("git", "init")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git init failed: %w\n%s", err, output)
		}
	}

	// Get GitHub token from environment
	githubToken := os.Getenv("GITHUB_TOKEN")
	var remoteURL string

	if githubToken != "" {
		// Use token-authenticated HTTPS URL
		remoteURL = fmt.Sprintf("https://%s@github.com/SohamDhotre/LeetCode.git", githubToken)
	} else {
		// Fall back to regular HTTPS (will fail without credentials)
		remoteURL = "https://github.com/SohamDhotre/LeetCode.git"
		fmt.Println("   ⚠️  Warning: GITHUB_TOKEN not set. Push may fail without authentication.")
	}

	// Check if remote origin exists
	cmd := exec.Command("git", "remote", "get-url", "origin")
	if err := cmd.Run(); err != nil {
		// Remote doesn't exist, add it
		fmt.Println("   Adding GitHub remote...")
		cmd = exec.Command("git", "remote", "add", "origin", remoteURL)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git remote add failed: %w\n%s", err, output)
		}
	} else {
		// Remote exists, update it to use the token
		fmt.Println("   Updating GitHub remote URL...")
		cmd = exec.Command("git", "remote", "set-url", "origin", remoteURL)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git remote set-url failed: %w\n%s", err, output)
		}
	}

	// Set default branch to main
	cmd = exec.Command("git", "branch", "-M", "main")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Ignore error if branch doesn't exist yet
		if !strings.Contains(string(output), "not found") {
			fmt.Printf("   Warning: git branch -M main: %s\n", output)
		}
	}

	return nil
}

func gitAddCommitPush(count int) error {
	// Ensure Git is properly set up
	if err := ensureGitSetup(); err != nil {
		return fmt.Errorf("git setup failed: %w", err)
	}

	// Configure git user from environment variables
	gitUserName := os.Getenv("GIT_USER_NAME")
	gitUserEmail := os.Getenv("GIT_USER_EMAIL")

	if gitUserName == "" {
		gitUserName = "LeetCode Sync Bot"
	}
	if gitUserEmail == "" {
		gitUserEmail = "leetcode-sync@example.com"
	}

	// Set git config
	configCmd := exec.Command("git", "config", "user.name", gitUserName)
	if err := configCmd.Run(); err != nil {
		return fmt.Errorf("git config user.name failed: %w", err)
	}

	configCmd = exec.Command("git", "config", "user.email", gitUserEmail)
	if err := configCmd.Run(); err != nil {
		return fmt.Errorf("git config user.email failed: %w", err)
	}

	// Git add
	cmd := exec.Command("git", "add", ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\n%s", err, output)
	}

	// Git commit
	commitMsg := fmt.Sprintf("Add %d new LeetCode solution(s)", count)
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Check if it's just "nothing to commit"
		if !strings.Contains(string(output), "nothing to commit") {
			return fmt.Errorf("git commit failed: %w\n%s", err, output)
		}
	}

	// Pull remote changes first (in case remote has content)
	fmt.Println("   🔄 Pulling remote changes...")
	cmd = exec.Command("git", "pull", "origin", "main", "--allow-unrelated-histories", "--no-edit")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Ignore error if remote doesn't exist yet
		if !strings.Contains(string(output), "couldn't find remote ref") {
			fmt.Printf("   ⚠️  Pull warning: %s\n", string(output))
		}
	}

	// Git push (with upstream tracking)
	cmd = exec.Command("git", "push", "-u", "origin", "main")
	if output, err := cmd.CombinedOutput(); err != nil {
		// If push fails due to conflicts, try force push
		if strings.Contains(string(output), "non-fast-forward") || strings.Contains(string(output), "rejected") {
			fmt.Println("   ⚠️  Normal push rejected, using force push...")
			cmd = exec.Command("git", "push", "-u", "origin", "main", "--force")
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("git push --force failed: %w\n%s", err, output)
			}
		} else {
			return fmt.Errorf("git push failed: %w\n%s", err, output)
		}
	}

	return nil
}

func cleanupFailedEntries() int {
	if syncDB.Failed == nil {
		return 0
	}
	removed := 0
	if staleFailedDays <= 0 {
		return 0
	}
	cutoff := time.Now().Add(-time.Duration(staleFailedDays) * 24 * time.Hour)

	for key, entry := range syncDB.Failed {
		if entry.RetryCount >= maxRetriesPerSubmission && entry.LastTried.Before(cutoff) {
			delete(syncDB.Failed, key)
			removed++
		}
	}
	return removed
}

func debugLog(msg string) {
	if debugMode {
		fmt.Printf("[DEBUG] %s\n", msg)
	}
}
