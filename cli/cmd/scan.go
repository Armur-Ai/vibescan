package cmd

import (
	"armur-cli/internal/api"
	"armur-cli/internal/config"
	"armur-cli/internal/history"
	"armur-cli/internal/utils"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var scanInteractiveCmd = &cobra.Command{
	Use:   "scan-i",
	Short: "Enter interactive scanning mode (alias: vibescan scan --interactive)",
	Long:  "Launches a terminal UI to configure and start a scan interactively.",
	Deprecated: "Use 'vibescan scan --interactive' instead.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			color.Red("Error loading configuration: %v", err)
			os.Exit(1)
		}

		var targetType string
		var targetPath string
		var language string
		var scanMode string
		var output string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What would you like to scan?").
					Options(
						huh.NewOption("Repository URL", "repo"),
						huh.NewOption("Local file or directory", "local"),
					).
					Value(&targetType),
				huh.NewInput().
					Title("Enter the path or URL to scan").
					Prompt(">").
					Value(&targetPath),
				huh.NewSelect[string]().
					Title("Scan mode").
					Options(
						huh.NewOption("Simple", "simple"),
						huh.NewOption("Advanced", "advanced"),
					).
					Value(&scanMode),
				huh.NewSelect[string]().
					Title("Output format").
					Options(
						huh.NewOption("Text", "text"),
						huh.NewOption("JSON", "json"),
					).
					Value(&output),
			),
		)

		if err := form.Run(); err != nil {
			color.Red("Prompt error: %v", err)
			os.Exit(1)
		}

		if strings.HasPrefix(targetPath, "http://") || strings.HasPrefix(targetPath, "https://") {
			if err := huh.NewInput().
				Title("Specify programming language for repo").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("language is required for repositories")
					}
					return nil
				}).
				Value(&language).
				Run(); err != nil {
				color.Red("Language input error: %v", err)
				os.Exit(1)
			}
		}

		apiClient := api.NewClient(cfg.API.URL).WithAPIKey(cfg.APIKey)
		var taskID string

		fmt.Println(color.CyanString("Initiating scan..."))
		if targetType == "repo" {
			taskID, err = apiClient.ScanRepository(targetPath, language, scanMode == "advanced")
		} else {
			taskID, err = apiClient.ScanFile(targetPath, scanMode == "advanced")
		}

		if err != nil {
			color.Red("Failed to start scan: %v", err)
			os.Exit(1)
		}

		fmt.Println(color.GreenString("Scan started. Task ID: %s", taskID))

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Waiting for scan to complete..."
		s.Start()

		var result map[string]interface{}
		var status string
		for {
			status, result, err = apiClient.GetTaskStatus(taskID)
			if err != nil {
				s.Stop()
				color.Red("Error fetching task status: %v", err)
				os.Exit(1)
			}
			if status == "success" {
				s.Stop()
				break
			}
			if status == "failed" {
				s.Stop()
				color.Red("Scan failed.")
				os.Exit(1)
			}
			time.Sleep(2 * time.Second)
		}

		fmt.Println()
		if output == "json" {
			utils.PrintResultsJSON(result)
		} else {
			printFormattedResults(result)
		}
	},
}

var scanCmd = &cobra.Command{
	Use:   "scan [target]",
	Short: "Scan a repository or file",
	Long: `Scan a Git repository (by providing the URL) or a local file/directory (by providing the path)
for security vulnerabilities using the Armur Code Scanner service.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		interactive, _ := cmd.Flags().GetBool("interactive")

		// If --interactive or no args, launch interactive mode
		if interactive || len(args) == 0 {
			scanInteractiveCmd.Run(cmd, args)
			return
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			color.Red("Error loading configuration: %v", err)
			os.Exit(1)
		}

		// --api-key flag overrides config file
		if flagKey, _ := cmd.Root().PersistentFlags().GetString("api-key"); flagKey != "" {
			cfg.APIKey = flagKey
		}

		// Auto-start server if not running (unless --no-server is set)
		noServer, _ := cmd.Flags().GetBool("no-server")
		cfg.API.URL = ensureServer(cfg.API.URL, noServer)

		target := args[0]
		language, _ := cmd.Flags().GetString("language")
		isAdvanced, _ := cmd.Flags().GetBool("advanced")
		outputFormat, _ := cmd.Flags().GetString("output")

		// --watch mode
		watchMode, _ := cmd.Flags().GetBool("watch")
		if watchMode {
			runWatchMode(cfg, target, language, isAdvanced, outputFormat)
			return
		}

		apiClient := api.NewClient(cfg.API.URL).WithAPIKey(cfg.APIKey)

		if (strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://")) && language == "" {
			color.Red("Error: Language must be specified when scanning a repository. Use --language or -l flag.")
			os.Exit(1)
		}

		var taskID string
		var scanErr error

		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
			fmt.Println(color.CyanString("Initiating scan for repository: %s", target))
			taskID, scanErr = apiClient.ScanRepository(target, language, isAdvanced)
		} else {
			fmt.Println(color.CyanString("Initiating scan for local target: %s", target))
			taskID, scanErr = apiClient.ScanFile(target, isAdvanced)
		}

		if scanErr != nil {
			color.Red("Error initiating scan: %v", scanErr)
			os.Exit(1)
		}

		fmt.Println(color.GreenString("Scan initiated successfully. Task ID: %s", taskID))

		scanResult := waitForScan(apiClient, taskID)

		// Save to history
		counts := countSeverities(scanResult)
		saveToHistory(taskID, target, language, isAdvanced, scanResult, counts)

		// Apply severity filter
		minSeverity, _ := cmd.Flags().GetString("min-severity")
		failOnSeverity, _ := cmd.Flags().GetString("fail-on-severity")

		if outputFormat == "json" {
			utils.PrintResultsJSON(scanResult)
		} else {
			printFormattedResults(scanResult)
		}

		// Print severity summary card
		printSummaryCard(counts)

		// --fail-on-severity: exit non-zero if findings at or above the threshold exist
		if failOnSeverity != "" {
			threshold := severityLevel(failOnSeverity)
			for sev, count := range counts {
				if count > 0 && severityLevel(sev) >= threshold {
					os.Exit(1)
				}
			}
		}

		// --min-severity: already used in display filtering (checked within print functions)
		_ = minSeverity
	},
}

// waitForScan waits for a scan to complete, streaming progress when available,
// and falling back to simple polling. Returns the scan results.
func waitForScan(apiClient *api.APIClient, taskID string) map[string]interface{} {
	// Try SSE streaming first
	updates := make(chan api.ProgressUpdate, 16)
	go apiClient.StreamProgress(taskID, updates)

	// Give SSE a moment to connect
	streamActive := false
	select {
	case update, ok := <-updates:
		if ok {
			streamActive = true
			displayProgress(update)
			if update.Status == "completed" || update.Status == "failed" {
				return fetchFinalResult(apiClient, taskID, update.Status)
			}
		}
	case <-time.After(3 * time.Second):
		// SSE not available, fall back to polling
	}

	if streamActive {
		// Continue reading SSE updates. The stream can stay on "queued" for the
		// whole scan when the server has no per-tool progress to report, so the
		// task status is polled alongside it.
		poll := time.NewTicker(2 * time.Second)
		defer poll.Stop()
	stream:
		for {
			select {
			case update, ok := <-updates:
				if !ok {
					break stream
				}
				displayProgress(update)
				if update.Status == "completed" || update.Status == "failed" {
					return fetchFinalResult(apiClient, taskID, update.Status)
				}
			case <-poll.C:
				status, result, err := apiClient.GetTaskStatus(taskID)
				if err != nil {
					continue
				}
				if status == "success" {
					fmt.Println()
					return result
				} else if status == "failed" {
					return fetchFinalResult(apiClient, taskID, status)
				}
			}
		}
	}

	// Fallback: simple polling with spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Waiting for scan to complete..."
	s.Start()

	for {
		status, result, err := apiClient.GetTaskStatus(taskID)
		if err != nil {
			s.Stop()
			color.Red("Error getting task status: %v", err)
			os.Exit(1)
		}

		if status == "success" {
			s.Stop()
			return result
		} else if status == "failed" {
			s.Stop()
			color.Red("Scan failed.")
			os.Exit(1)
		}
		time.Sleep(2 * time.Second)
	}
}

// displayProgress renders the current progress to the terminal.
func displayProgress(update api.ProgressUpdate) {
	// Clear previous output and redraw
	fmt.Print("\033[2K\r")

	running := []string{}
	completed := 0
	totalFindings := 0

	for _, t := range update.Tools {
		switch t.Status {
		case "running":
			elapsed := ""
			if t.StartedAt > 0 {
				dur := time.Since(time.Unix(t.StartedAt, 0)).Truncate(time.Second)
				elapsed = fmt.Sprintf(" (%s)", dur)
			}
			running = append(running, t.Tool+elapsed)
		case "completed":
			completed++
			totalFindings += t.Findings
		case "failed":
			completed++
		}
	}

	progress := fmt.Sprintf("[%d/%d]", completed, update.TotalTools)
	findingStr := fmt.Sprintf("%d findings", totalFindings)

	if len(running) > 0 {
		toolList := strings.Join(running, ", ")
		fmt.Printf("%s Running: %s | %s",
			color.CyanString(progress),
			color.YellowString(toolList),
			color.GreenString(findingStr),
		)
	} else {
		fmt.Printf("%s %s | %s",
			color.CyanString(progress),
			color.GreenString("Waiting..."),
			color.GreenString(findingStr),
		)
	}
}

// fetchFinalResult retrieves the final scan result after completion.
func fetchFinalResult(apiClient *api.APIClient, taskID, status string) map[string]interface{} {
	fmt.Println() // newline after progress line

	if status == "failed" {
		color.Red("Scan failed.")
		os.Exit(1)
	}

	_, result, err := apiClient.GetTaskStatus(taskID)
	if err != nil {
		color.Red("Error getting final results: %v", err)
		os.Exit(1)
	}
	return result
}

func printFormattedResults(results map[string]interface{}) {
	categories := []string{
		"security_issues",
		"antipatterns_bugs",
		"complex_functions",
		"dockstring_absent",
	}

	for _, category := range categories {
		if categoryData, ok := results[category]; ok {
			fmt.Printf("\n%s\n", color.CyanString("Category: %s", category))
			printCategoryData(category, categoryData)
		}
	}
	fmt.Println(color.GreenString("\nScan completed. Results displayed."))
}

func printCategoryData(category string, data interface{}) {
	descriptions := map[string]string{
		"security_issues":   "Security vulnerabilities and potential threats",
		"antipatterns_bugs": "Code patterns that are considered bad practices or likely to cause bugs",
		"complex_functions": "Functions with high cyclomatic complexity",
		"dockstring_absent": "Missing documentation strings",
	}

	fmt.Printf("Description: %s\n\n", descriptions[category])

	switch category {
	case "security_issues":
		printSecurityIssues(data)
	case "antipatterns_bugs":
		printAntipatternsBugs(data)
	case "complex_functions":
		printComplexFunctions(data)
	case "dockstring_absent":
		printDocstringIssues(data)
	}
}

func printSecurityIssues(data any) {
	issues, ok := data.([]any)
	if !ok || len(issues) == 0 {
		color.Yellow("No security issues found.\n")
		return
	}

	fmt.Printf("%-20s %-10s %-40s %-15s %-20s\n", "FILE", "LINE", "MESSAGE", "SEVERITY", "CWE")
	fmt.Println(strings.Repeat("-", 105))

	for _, issue := range issues {
		if issueMap, ok := issue.(map[string]interface{}); ok {
			files, ok := issueMap["files"].([]interface{})
			if !ok {
				continue
			}

			cwe := getStringValue(issueMap, "cwe", "N/A")
			for _, file := range files {
				if fileMap, ok := file.(map[string]interface{}); ok {
					issues, ok := fileMap["issues"].([]interface{})
					if !ok {
						continue
					}

					for _, issueDetail := range issues {
						if detail, ok := issueDetail.(map[string]interface{}); ok {
							fmt.Printf("%-20s %-10s %-40s %-15s %-20s\n",
								truncateString(getStringValue(detail, "path", "-"), 20),
								getStringValue(detail, "line", "-"),
								truncateString(getStringValue(detail, "message", "-"), 40),
								getSeverityColored(getStringValue(detail, "severity", "INFO")),
								cwe,
							)
						}
					}
				}
			}
		}
	}
	fmt.Println()
}

func printAntipatternsBugs(data interface{}) {
	issues, ok := data.([]interface{})
	if !ok || len(issues) == 0 {
		color.Yellow("No antipatterns or bugs found.\n")
		return
	}

	fmt.Printf("%-30s %-10s %-50s\n", "FILE", "LINE", "MESSAGE")
	fmt.Println(strings.Repeat("-", 90))

	for _, issue := range issues {
		if issueMap, ok := issue.(map[string]interface{}); ok {
			files, ok := issueMap["files"].([]interface{})
			if !ok {
				continue
			}

			for _, file := range files {
				if fileMap, ok := file.(map[string]interface{}); ok {
					path := getStringValue(fileMap, "path", "-")
					issues, ok := fileMap["issues"].([]interface{})
					if !ok {
						continue
					}

					for _, issueDetail := range issues {
						if detail, ok := issueDetail.(map[string]interface{}); ok {
							fmt.Printf("%-30s %-10s %-50s\n",
								truncateString(path, 30),
								getStringValue(detail, "line", "-"),
								truncateString(getStringValue(detail, "message", "-"), 50),
							)
						}
					}
				}
			}
		}
	}
	fmt.Println()
}

func printComplexFunctions(data interface{}) {
	issues, ok := data.([]interface{})
	if !ok || len(issues) == 0 {
		color.Yellow("No complex functions found.\n")
		return
	}

	fmt.Printf("%-40s %-10s %-20s\n", "FILE", "LINE", "COMPLEXITY")
	fmt.Println(strings.Repeat("-", 70))

	for _, issue := range issues {
		if issueMap, ok := issue.(map[string]interface{}); ok {
			path := getStringValue(issueMap, "path", "-")
			issues, ok := issueMap["issues"].([]interface{})
			if !ok {
				continue
			}

			for _, issueDetail := range issues {
				if detail, ok := issueDetail.(map[string]interface{}); ok {
					fmt.Printf("%-40s %-10s %-20s\n",
						truncateString(path, 40),
						getStringValue(detail, "line", "-"),
						getStringValue(detail, "complexity", "-"),
					)
				}
			}
		}
	}
	fmt.Println()
}

func printDocstringIssues(data interface{}) {
	issues, ok := data.([]interface{})
	if !ok || len(issues) == 0 {
		color.Yellow("No docstring issues found.\n")
		return
	}

	fmt.Printf("%-40s %-10s %-40s\n", "FILE", "LINE", "MESSAGE")
	fmt.Println(strings.Repeat("-", 90))

	for _, issue := range issues {
		if issueMap, ok := issue.(map[string]interface{}); ok {
			path := getStringValue(issueMap, "path", "-")
			fmt.Printf("%-40s %-10s %-40s\n",
				truncateString(path, 40),
				getStringValue(issueMap, "line", "-"),
				truncateString(getStringValue(issueMap, "message", "Missing docstring"), 40),
			)
		}
	}
	fmt.Println()
}

func getStringValue(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok && val != nil {
		switch v := val.(type) {
		case string:
			return v
		case float64:
			return fmt.Sprintf("%d", int(v))
		case int:
			return fmt.Sprintf("%d", v)
		default:
			return defaultValue
		}
	}
	return defaultValue
}

func getSeverityColored(severity string) string {
	switch strings.ToUpper(severity) {
	case "HIGH":
		return color.RedString(severity)
	case "MEDIUM":
		return color.YellowString(severity)
	case "LOW":
		return color.GreenString(severity)
	case "INFO":
		return color.CyanString(severity)
	default:
		return severity
	}
}

// saveToHistory persists the scan result to the local SQLite history database.
func saveToHistory(taskID, target, language string, isAdvanced bool, results map[string]interface{}, counts map[string]int) {
	db, err := history.Open()
	if err != nil {
		return // non-fatal — don't break the scan flow
	}
	defer db.Close()

	scanType := "simple"
	if isAdvanced {
		scanType = "advanced"
	}

	resultsJSON, _ := json.Marshal(results)

	db.Save(&history.ScanRecord{
		TaskID:    taskID,
		Target:    target,
		Language:  language,
		ScanType:  scanType,
		Status:    "success",
		Critical:  counts["CRITICAL"],
		High:      counts["HIGH"],
		Medium:    counts["MEDIUM"],
		Low:       counts["LOW"],
		Info:      counts["INFO"],
		Results:   string(resultsJSON),
		CreatedAt: time.Now(),
	})
}

// severityLevel returns a numeric severity level for comparison.
func severityLevel(sev string) int {
	switch strings.ToUpper(strings.TrimSpace(sev)) {
	case "CRITICAL":
		return 5
	case "HIGH":
		return 4
	case "MEDIUM":
		return 3
	case "LOW":
		return 2
	case "INFO":
		return 1
	default:
		return 0
	}
}

// countSeverities walks the scan results and counts findings by severity.
func countSeverities(results map[string]interface{}) map[string]int {
	counts := map[string]int{
		"CRITICAL": 0,
		"HIGH":     0,
		"MEDIUM":   0,
		"LOW":      0,
		"INFO":     0,
	}

	for _, categoryData := range results {
		issues, ok := categoryData.([]interface{})
		if !ok {
			continue
		}
		for _, issue := range issues {
			issueMap, ok := issue.(map[string]interface{})
			if !ok {
				continue
			}
			// Direct severity field
			if sev, ok := issueMap["severity"].(string); ok {
				key := strings.ToUpper(sev)
				if _, exists := counts[key]; exists {
					counts[key]++
				}
			}
			// Nested files -> issues -> severity
			if files, ok := issueMap["files"].([]interface{}); ok {
				for _, file := range files {
					if fileMap, ok := file.(map[string]interface{}); ok {
						if fileIssues, ok := fileMap["issues"].([]interface{}); ok {
							for _, fi := range fileIssues {
								if detail, ok := fi.(map[string]interface{}); ok {
									if sev, ok := detail["severity"].(string); ok {
										key := strings.ToUpper(sev)
										if _, exists := counts[key]; exists {
											counts[key]++
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return counts
}

// printSummaryCard displays a box-drawing severity summary card.
func printSummaryCard(counts map[string]int) {
	total := 0
	for _, c := range counts {
		total += c
	}

	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────┐")
	fmt.Println("│                   Scan Complete                     │")
	fmt.Println("├──────────┬──────────┬──────────┬──────────┬─────────┤")
	fmt.Printf("│ %s │ %s │ %s │ %s │ %s │\n",
		color.New(color.FgHiRed, color.Bold).Sprintf("%-8s", "Critical"),
		color.New(color.FgRed).Sprintf("%-8s", "High"),
		color.New(color.FgYellow).Sprintf("%-8s", "Medium"),
		color.New(color.FgGreen).Sprintf("%-8s", "Low"),
		color.New(color.FgCyan).Sprintf("%-7s", "Info"),
	)
	fmt.Printf("│ %-8d │ %-8d │ %-8d │ %-8d │ %-7d │\n",
		counts["CRITICAL"], counts["HIGH"], counts["MEDIUM"], counts["LOW"], counts["INFO"],
	)
	fmt.Println("└──────────┴──────────┴──────────┴──────────┴─────────┘")
	fmt.Printf("Total findings: %d\n", total)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(scanInteractiveCmd) // deprecated alias
	scanCmd.Flags().StringP("language", "l", "", "Specify the programming language")
	scanCmd.Flags().BoolP("simple", "s", true, "Perform a simple scan")
	scanCmd.Flags().BoolP("advanced", "a", false, "Perform an advanced scan")
	scanCmd.Flags().StringP("output", "o", "text", "Output format (text, json)")
	scanCmd.Flags().StringP("diff", "d", "", "Only scan files changed since this git ref (e.g. HEAD~1, main)")
	scanCmd.Flags().Bool("fail-on-findings", false, "Exit with code 1 if any findings are reported")
	scanCmd.Flags().StringP("min-severity", "S", "", "Minimum severity to display (info, low, medium, high, critical)")
	scanCmd.Flags().String("fail-on-severity", "", "Exit with code 1 if findings at or above this severity exist (info, low, medium, high, critical)")
	scanCmd.Flags().Bool("staged-only", false, "Scan only git-staged files (for pre-commit hooks)")
	scanCmd.Flags().StringP("format", "f", "text", "Output format (text, json, sarif)")
	scanCmd.Flags().Bool("no-server", false, "Skip auto-starting the embedded server")
	scanCmd.Flags().BoolP("interactive", "i", false, "Launch interactive scanning mode")
	scanCmd.Flags().BoolP("watch", "w", false, "Watch for file changes and re-scan automatically")
}
