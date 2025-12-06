package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// ==========================
// Config & Data Structures
// ==========================

type Config struct {
	GitHub struct {
		RepoName              string `json:"repo_name"`
		CommitMessageTemplate string `json:"commit_message_template"`
	} `json:"github"`
	CategoryMappings   map[string]string `json:"category_mappings"`
	LanguageExtensions map[string]string `json:"language_extensions"`
}

// Shape matches /api/submissions/ JSON (we only keep fields we care about)
type Submission struct {
	ID            int    `json:"id"`
	QuestionID    int    `json:"question_id"`
	Lang          string `json:"lang"`
	LangName      string `json:"lang_name"`
	Time          string `json:"time"`
	Timestamp     int64  `json:"timestamp"`
	Status        int    `json:"status"`         // 10 = Accepted
	StatusDisplay string `json:"status_display"` // "Accepted", "Wrong Answer", etc
	URL           string `json:"url"`            // "/submissions/detail/xxxx/"
	Title         string `json:"title"`          // "Reverse Nodes in k-Group"
	TitleSlug     string `json:"title_slug"`     // "reverse-nodes-in-k-group"
	Code          string `json:"code"`           // full solution code (this is what we want!)
	FrontendID    int    `json:"frontend_id"`    // LeetCode frontend id
}

// Not currently used, kept for possible future
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

// ==========================
// Globals & Constants
// ==========================

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
	staleFailedDays         = 30
)

// ==========================
// Main
// ==========================

func main() {
	fmt.Println("🚀 LeetCode → GitHub Sync Daemon")
	fmt.Println("==================================")
	fmt.Println("📅 Schedule: Continuous (every 10 minutes)")
	fmt.Println("⏱️  Frequency: Every 10 minutes")

	// Load .env (optional but recommended)
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found. Make sure LEETCODE_SESSION and CSRF_TOKEN are set.")
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

	// Validate session immediately using GraphQL
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

	// Run sync immediately on startup
	fmt.Println("🚀 Starting initial sync...")
	performSync()

	// Daemon loop - always sync every 10 minutes
	for {
		fmt.Printf("\n[%s] ⏰ Time to sync!\n", time.Now().Format("15:04:05"))
		performSync()

		fmt.Printf("[%s] 💤 Sleeping for 10 minutes (or until manual trigger)...\n", time.Now().Format("15:04:05"))
		time.Sleep(10 * time.Minute)
	}
}

// ==========================
// Core Sync Flow
// ==========================

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
			statusText := describeStatus(sub.Status)
			fmt.Printf("   %d. [%s] %s (ID: %d, Lang: %s)\n",
				i+1, statusText, sub.Title, sub.ID, sub.Lang)
		}
		fmt.Println()
	}

	// Filter accepted submissions (latest per problem)
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

	newCount := 0
	skippedCount := 0
	failedThisRun := 0
	retriedFixed := 0

	for i, sub := range acceptedSubmissions {
		submissionKey := fmt.Sprintf("%d", sub.ID)

		// Already synced
		// Check by problem slug instead of submission ID
		alreadySynced := false
		for _, entry := range syncDB.Synced {
   		if entry.TitleSlug == sub.TitleSlug {
        		alreadySynced = true
        		break
    		}
	}

	if alreadySynced {
    		skippedCount++
    		fmt.Printf("  ⏭️  [%d/%d] Skipping (already synced problem): %s (Slug: %s)\n",
        		i+1, len(acceptedSubmissions), sub.Title, sub.TitleSlug)
    		continue
	}


		// Previously failed?
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

	// Update master README (always update)
	fmt.Println("\n📝 Updating master README...")
	if err := updateMasterREADME(); err != nil {
		fmt.Printf("⚠️  Error updating README: %v\n", err)
	} else {
		fmt.Println("   ✓ README updated")
	}

	if newCount == 0 && failedThisRun == 0 && retriedFixed == 0 {
		fmt.Println("\n✨ No changes to sync (no new or recovered submissions).")

		// Still check for any pending git changes
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

// ==========================
// Config & DB Helpers
// ==========================

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

// ==========================
// LeetCode API Calls
// ==========================

func validateSession() error {
	query := `{"query": "query { userStatus { isSignedIn } }"}`

	req, err := http.NewRequest("POST", graphqlEndpoint, strings.NewReader(query))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("LEETCODE_SESSION=%s; csrftoken=%s", sessionToken, csrfToken))
	req.Header.Set("x-csrftoken", csrfToken)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
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

	req.Header.Set("Cookie", fmt.Sprintf("LEETCODE_SESSION=%s; csrftoken=%s", sessionToken, csrfToken))
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
		// LeetCode: status 10 = Accepted
		if (sub.Status == 10 || sub.Status == 0) && !seen[sub.TitleSlug] {
			accepted = append(accepted, sub)
			seen[sub.TitleSlug] = true
		}
	}

	return accepted
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
	req.Header.Set("Cookie", fmt.Sprintf("LEETCODE_SESSION=%s; csrftoken=%s", sessionToken, csrfToken))
	req.Header.Set("x-csrftoken", csrfToken)
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

// ==========================
// Problem Processing
// ==========================

func processSubmission(sub Submission) error {
	// Sanity check: if API didn't give code, we can't do much
	if strings.TrimSpace(sub.Code) == "" {
		return fmt.Errorf("submission %d has empty code (likely locked or older format)", sub.ID)
	}

	// Fetch problem details (difficulty, tags, description)
	problemDetail, err := fetchProblemDetail(sub.TitleSlug)
	if err != nil {
		return fmt.Errorf("fetching problem details: %w", err)
	}

	// Determine category and file paths
	category := determineCategory(problemDetail.TopicTags)
	difficulty := problemDetail.Difficulty
	extension := getFileExtension(sub.Lang)

	// Structure: Category/Difficulty/ID.TitleSlug/
	problemFolder := fmt.Sprintf("%s.%s", problemDetail.QuestionFrontendID, sanitizeFilename(sub.TitleSlug))
	dirPath := filepath.Join(category, difficulty, problemFolder)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Filename inside folder: ID.slug.ext
	filename := fmt.Sprintf("%s.%s.%s",
		problemDetail.QuestionFrontendID,
		sanitizeFilename(sub.TitleSlug),
		extension)

	filePath := filepath.Join(dirPath, filename)

	// Write solution file
	if err := os.WriteFile(filePath, []byte(sub.Code), 0644); err != nil {
		return fmt.Errorf("writing solution file: %w", err)
	}
	fmt.Printf("  ✓ Created: %s\n", filePath)

	// Problem README
	readmePath := filepath.Join(dirPath, "README.md")
	if err := createProblemREADME(readmePath, problemDetail, sub.Lang); err != nil {
		return fmt.Errorf("creating problem README: %w", err)
	}
	fmt.Printf("  ✓ Created: %s\n", readmePath)

	// Update sync DB
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

func determineCategory(tags []TopicTag) string {
	if len(tags) == 0 {
		return "Miscellaneous"
	}

	for _, tag := range tags {
		if category, exists := config.CategoryMappings[tag.Slug]; exists {
			return category
		}
	}

	return capitalizeFirst(tags[0].Name)
}

func getFileExtension(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if ext, exists := config.LanguageExtensions[lang]; exists {
		return ext
	}
	return "txt"
}

func sanitizeFilename(name string) string {
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
	reg := regexp.MustCompile(`<[^>]*>`)
	text := reg.ReplaceAllString(html, "")

	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&quot;", "\"")

	if len(text) > 10000 {
		text = text[:10000] + "..."
	}

	return strings.TrimSpace(text)
}

// ==========================
// README + Git Helpers
// ==========================

func updateMasterREADME() error {
	stats := make(map[string]map[string]int)
	var entries []SyncEntry

	for _, entry := range syncDB.Synced {
		entries = append(entries, entry)

		if stats[entry.Category] == nil {
			stats[entry.Category] = make(map[string]int)
		}
		stats[entry.Category][entry.Difficulty]++
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	totalEasy, totalMedium, totalHard := 0, 0, 0
	for _, diffMap := range stats {
		totalEasy += diffMap["Easy"]
		totalMedium += diffMap["Medium"]
		totalHard += diffMap["Hard"]
	}
	totalProblems := totalEasy + totalMedium + totalHard

	var sb strings.Builder
	sb.WriteString("# 🚀 LeetCode Solutions\n\n")
	sb.WriteString("A collection of my LeetCode solutions, automatically synced from LeetCode.\n\n")
	sb.WriteString("## 📊 Progress Statistics\n\n")
	sb.WriteString(fmt.Sprintf("**Total Problems Solved:** %d\n\n", totalProblems))
	sb.WriteString(fmt.Sprintf("- 🟢 Easy: %d\n", totalEasy))
	sb.WriteString(fmt.Sprintf("- 🟡 Medium: %d\n", totalMedium))
	sb.WriteString(fmt.Sprintf("- 🔴 Hard: %d\n\n", totalHard))

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

		// Proper GitHub path (spaces and slashes encoded)
		problemPath := fmt.Sprintf(
			"https://github.com/SohamDhotre/LeetCode/tree/main/%s/%s/%s.%s",
			url.PathEscape(entry.Category),
			url.PathEscape(entry.Difficulty),
			url.PathEscape(entry.ProblemID),
			url.PathEscape(entry.TitleSlug),
		)

		// Update markdown row: Title as a link, category plain text
		sb.WriteString(fmt.Sprintf("| %s | [%s](%s) | %s %s | %s | %s |\n",
			entry.ProblemID,
			entry.Title,
			problemPath,
			diffEmoji,
			entry.Difficulty,
			entry.Category,
			entry.Timestamp.Format("2006-01-02"),
		))

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
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	hasChanges := len(strings.TrimSpace(string(output))) > 0

	if hasChanges {
		fmt.Println("   📦 Found unpushed changes from previous run")
		fmt.Println("   🔄 Attempting to push pending changes...")

		count := len(syncDB.Synced)
		if count == 0 {
			count = 1
		}

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
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		fmt.Println("   Initializing Git repository...")
		cmd := exec.Command("git", "init")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git init failed: %w\n%s", err, output)
		}
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	var remoteURL string

	if githubToken != "" {
		remoteURL = fmt.Sprintf("https://%s@github.com/SohamDhotre/LeetCode.git", githubToken)
	} else {
		remoteURL = "https://github.com/SohamDhotre/LeetCode.git"
		fmt.Println("   ⚠️  Warning: GITHUB_TOKEN not set. Push may fail without authentication.")
	}

	cmd := exec.Command("git", "remote", "get-url", "origin")
	if err := cmd.Run(); err != nil {
		fmt.Println("   Adding GitHub remote...")
		cmd = exec.Command("git", "remote", "add", "origin", remoteURL)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git remote add failed: %w\n%s", err, output)
		}
	} else {
		fmt.Println("   Updating GitHub remote URL...")
		cmd = exec.Command("git", "remote", "set-url", "origin", remoteURL)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git remote set-url failed: %w\n%s", err, output)
		}
	}

	cmd = exec.Command("git", "branch", "-M", "main")
	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "not found") {
			fmt.Printf("   Warning: git branch -M main: %s\n", output)
		}
	}

	return nil
}

func gitAddCommitPush(count int) error {
	if err := ensureGitSetup(); err != nil {
		return fmt.Errorf("git setup failed: %w", err)
	}

	gitUserName := os.Getenv("GIT_USER_NAME")
	gitUserEmail := os.Getenv("GIT_USER_EMAIL")

	if gitUserName == "" {
		gitUserName = "LeetCode Sync Bot"
	}
	if gitUserEmail == "" {
		gitUserEmail = "leetcode-sync@example.com"
	}

	configCmd := exec.Command("git", "config", "user.name", gitUserName)
	if err := configCmd.Run(); err != nil {
		return fmt.Errorf("git config user.name failed: %w", err)
	}

	configCmd = exec.Command("git", "config", "user.email", gitUserEmail)
	if err := configCmd.Run(); err != nil {
		return fmt.Errorf("git config user.email failed: %w", err)
	}

	cmd := exec.Command("git", "add", ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\n%s", err, output)
	}

	commitMsg := fmt.Sprintf("Add %d new LeetCode solution(s)", count)
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "nothing to commit") {
			return fmt.Errorf("git commit failed: %w\n%s", err, output)
		}
	}

	fmt.Println("   🔄 Pulling remote changes...")
	cmd = exec.Command("git", "pull", "origin", "main", "--allow-unrelated-histories", "--no-edit")
	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "couldn't find remote ref") {
			fmt.Printf("   ⚠️  Pull warning: %s\n", string(output))
		}
	}

	cmd = exec.Command("git", "push", "-u", "origin", "main")
	if output, err := cmd.CombinedOutput(); err != nil {
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

// ==========================
// Failed entries & Debug
// ==========================

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

func describeStatus(status int) string {
	switch status {
	case 10:
		return "Accepted (10) ✅"
	case 11:
		return "Wrong Answer ❌"
	case 12:
		return "Memory Limit Exceeded"
	case 13:
		return "Output Limit Exceeded"
	case 14:
		return "Time Limit Exceeded ⏱️"
	case 15:
		return "Runtime Error 💥"
	case 16:
		return "Internal Error"
	case 20:
		return "Compile Error 🔨"
	case 0:
		return "Accepted (0?) ✅"
	default:
		return fmt.Sprintf("Status %d", status)
	}
}
