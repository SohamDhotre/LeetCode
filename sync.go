package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
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
	StatusCode int    `json:"status_code"`
	Title      string `json:"title"`
	TitleSlug  string `json:"title_slug"`
	Timestamp  int64  `json:"timestamp"`
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

// Global variables
var (
	config       Config
	syncDB       SyncDatabase
	sessionToken string
	csrfToken    string
	debugMode    bool
)

const (
	syncDBFile              = ".leetcode_sync_db.json"
	submissionsEndpoint     = "https://leetcode.com/api/submissions/"
	graphqlEndpoint         = "https://leetcode.com/graphql"
	maxRetriesPerSubmission = 5
	staleFailedDays         = 30
)

func main() {
	fmt.Println("🚀 LeetCode → GitHub Sync Daemon")
	fmt.Println("==================================")
	fmt.Println("📅 Schedule: Continuous (every 10 minutes)")
	fmt.Println("⏱️  Frequency: Every 10 minutes")

	// Load environment variables
	sessionToken = os.Getenv("LEETCODE_SESSION")
	csrfToken = os.Getenv("CSRF_TOKEN")

	// Validate session immediately
	fmt.Println("🔑 Validating session...")
	if err := validateSession(); err != nil {
		fmt.Printf("❌ Session validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Session is valid!")

	debugMode = os.Getenv("DEBUG") == "true"

	if err := loadConfig(); err != nil {
		fmt.Printf("❌ Error loading config: %v\n", err)
		os.Exit(1)
	}

	if err := loadSyncDB(); err != nil {
		fmt.Printf("⚠️  Creating new sync database: %v\n", err)
		syncDB = SyncDatabase{
			Synced: make(map[string]SyncEntry),
			Failed: make(map[string]*FailedEntry),
		}
	}

	fmt.Println("🔍 Checking for unpushed changes...")
	checkAndPushPendingChanges()

	fmt.Println("🚀 Starting initial sync...")
	performSync()

	for {
		fmt.Printf("\n[%s] ⏰ Time to sync!\n", time.Now().Format("15:04:05"))
		performSync()
		fmt.Printf("[%s] 💤 Sleeping for 10 minutes...\n", time.Now().Format("15:04:05"))
		time.Sleep(10 * time.Minute)
	}
}

func performSync() {
	fmt.Println("📥 Fetching recent submissions...")
	submissions, err := fetchRecentSubmissions()
	if err != nil {
		fmt.Printf("❌ Error fetching submissions: %v\n", err)
		return
	}
	fmt.Printf("   ✓ Fetched %d total submissions\n", len(submissions))

	accepted := filterAcceptedSubmissions(submissions)
	fmt.Printf("   ✓ Found %d accepted submissions\n", len(accepted))

	newCount := 0
	failedThisRun := 0

	for _, sub := range accepted {
		key := fmt.Sprintf("%d", sub.ID)

		fmt.Printf("\n🔄 Processing: %s (ID: %d)\n", sub.Title, sub.ID)

		if err := processSubmission(sub); err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			failedThisRun++
			syncDB.Failed[key] = &FailedEntry{SubmissionID: sub.ID, Title: sub.Title, TitleSlug: sub.TitleSlug}
			continue
		}

		newCount++
		fmt.Println("   ✅ Success!")
	}

	fmt.Println("\n📊 Summary:")
	fmt.Printf("   Newly synced: %d | Failed: %d\n", newCount, failedThisRun)

	saveSyncDB()

	fmt.Println("📤 Pushing to GitHub...")
	if err := gitAddCommitPush(newCount); err != nil {
		fmt.Printf("⚠️ Git push warning: %v\n", err)
	}
}

func processSubmission(sub Submission) error {
	details, err := fetchProblemDetail(sub.TitleSlug)
	if err != nil {
		return err
	}

	code, err := fetchSubmissionCode(sub.ID)
	if err != nil {
		return err
	}

	// FIX: Use sub.Lang directly
	extension := getFileExtension(sub.Lang) // FIX ✔✔✔

	folder := fmt.Sprintf("%s.%s", details.QuestionFrontendID, sanitizeFilename(sub.TitleSlug))
	path := filepath.Join(determineCategory(details.TopicTags), details.Difficulty, folder)

	os.MkdirAll(path, 0755)

	file := filepath.Join(path, fmt.Sprintf("%s.%s.%s",
		details.QuestionFrontendID,
		sanitizeFilename(sub.TitleSlug),
		extension))

	os.WriteFile(file, []byte(code), 0644)
	fmt.Println("   ✓ Wrote:", file)

	readmePath := filepath.Join(path, "README.md")
	createProblemREADME(readmePath, details, sub.Lang)

	syncDB.Synced[fmt.Sprintf("%d", sub.ID)] = SyncEntry{
		SubmissionID: sub.ID,
		ProblemID:    details.QuestionFrontendID,
		Title:        details.Title,
		TitleSlug:    details.TitleSlug,
		Difficulty:   details.Difficulty,
		Category:     determineCategory(details.TopicTags),
		Timestamp:    time.Now(),
	}

	return nil
}

func fetchProblemDetail(titleSlug string) (ProblemDetail, error) {
	query := fmt.Sprintf(`{
		"query": "query getQuestionDetail($titleSlug: String!) { question(titleSlug: $titleSlug) { questionId questionFrontendId title titleSlug difficulty topicTags { name slug } content } }",
		"variables": {"titleSlug": "%s"}
	}`, titleSlug)

	req, _ := http.NewRequest("POST", graphqlEndpoint, strings.NewReader(query))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "LEETCODE_SESSION="+sessionToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ProblemDetail{}, err
	}
	defer resp.Body.Close()

	var result GraphQLResponse
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Data.Question, nil
}

func fetchSubmissionCode(id int) (string, error) {
	query := fmt.Sprintf(`{
		"query": "query submissionDetails($submissionId: Int!) { submissionDetails(submissionId: $submissionId) { code lang } }",
		"variables": {"submissionId": %d}
	}`, id)

	req, _ := http.NewRequest("POST", graphqlEndpoint, strings.NewReader(query))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("LEETCODE_SESSION=%s; csrftoken=%s", sessionToken, csrfToken))
	req.Header.Set("x-csrftoken", csrfToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			SubmissionDetails struct {
				Code string `json:"code"`
				Lang string `json:"lang"`
			} `json:"submissionDetails"`
		} `json:"data"`
	}

	json.Unmarshal(body, &result)
	return result.Data.SubmissionDetails.Code, nil
}

func fetchRecentSubmissions() ([]Submission, error) {
	url := submissionsEndpoint + "?offset=0&limit=10"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", "LEETCODE_SESSION="+sessionToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		SubmissionsData []Submission `json:"submissions_dump"`
	}
	json.NewDecoder(resp.Body).Decode(&data)
	return data.SubmissionsData, nil
}

func determineCategory(tags []TopicTag) string {
	if len(tags) == 0 {
		return "Miscellaneous"
	}
	if c, ok := config.CategoryMappings[tags[0].Slug]; ok {
		return c
	}
	return capitalize(tags[0].Name)
}

func getFileExtension(lang string) string {
	lang = strings.ToLower(lang)
	if ext, ok := config.LanguageExtensions[lang]; ok {
		return ext
	}
	return "txt"
}

func sanitizeFilename(s string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9-]+`)
	return strings.Trim(reg.ReplaceAllString(s, "-"), "-")
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
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
