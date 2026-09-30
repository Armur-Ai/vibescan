# vibescan roadmap

**The open-source security scanner for AI-generated code.**
Find vulnerabilities in Cursor, Claude Code, Copilot and Windsurf generated applications before they reach production.

This document replaces the 59-sprint `IMPROVEMENTS.md` plan (archived at
[`docs/archive/IMPROVEMENTS-legacy.md`](docs/archive/IMPROVEMENTS-legacy.md)). That plan was written for
"Armur, a general security agent" and optimised for breadth. This one is written for one bet:
**own the vibe-coding security category**, and cut everything that does not serve it.

Last reviewed: 2026-09-30.

---

## Contents

1. [The bet](#1-the-bet)
2. [Where we actually are](#2-where-we-actually-are)
3. [Who we are building for](#3-who-we-are-building-for)
4. [The landscape and our wedge](#4-the-landscape-and-our-wedge)
5. [The Vibe Top 10](#5-the-vibe-top-10)
6. [Product pillars](#6-product-pillars)
7. [Phases](#7-phases)
8. [Scanner expansion plan](#8-scanner-expansion-plan)
9. [What we are cutting](#9-what-we-are-cutting)
10. [Go-to-market](#10-go-to-market)
11. [Risks](#11-risks)
12. [Open decisions](#12-open-decisions)
13. [The first ten tasks](#13-the-first-ten-tasks)

---

## 1. The bet

AI assistants now write a large share of committed code, and that code is not getting safer:

- Veracode's 2026 GenAI Code Security Report puts the security pass rate of AI-generated code at
  **56%**, unchanged from the year before, across 100+ models. Syntax pass rate is near 100%.
  ([Veracode](https://www.veracode.com/news/llms-are-getting-smarter-but-not-safer-veracode-2026-genai-code-security-report-finds-ai-generated-code-security-has-stalled-at-56%25-pass-rate/))
- CodeRabbit's analysis of 470 open-source pull requests found AI co-authored PRs carry about
  **1.7x more issues** than human-only PRs, with XSS up to 2.74x.
  ([CodeRabbit](https://www.coderabbit.ai/blog/state-of-ai-vs-human-code-generation-report))
- A USENIX Security 2025 study of 16 code models found **19.7% of recommended packages did not exist**
  (205,474 unique hallucinated names), which is the raw material for slopsquatting.
  ([USENIX](https://www.usenix.org/conference/usenixsecurity25/presentation/spracklen))
- CVE-2025-48757: **170 of 1,645** scanned Lovable-generated apps exposed data through missing
  Supabase row-level security.

The failure modes are specific and repetitive. Generic SAST was built for human-written enterprise
code and misses most of them: it does not know that a Supabase table has no RLS policy, that an
`npm install` line names a package that has never existed, or that `.cursorrules` contains
invisible instructions.

**The bet:** the winner of this category is the tool a developer can run in ten seconds, that
understands the stack AI tools actually generate, and that feeds its findings straight back to the
agent that wrote the code. It is open source, local-first, and has a named taxonomy everyone else
ends up citing.

---

## 2. Where we actually are

This audit was run against `main` at `52efb12` on 2026-09-30, by building and running everything
on a clean machine. It is blunt on purpose: every later phase depends on fixing this first.

Three things were fixed alongside this document so that the README's quick start is true: the
Docker image builds again, `docker compose up` starts a working engine, and the CLI can complete a
repository scan and a single-file scan. They are marked *fixed* below. Everything else stands.

### What is real

| Area | State |
|---|---|
| Go server (Gin + Asynq + Redis), parallel tool runner with per-tool timeouts | Builds, tests pass |
| 37 tool runners wired into the scan pipeline | Code exists; each needs its binary on `PATH` |
| SARIF output, severity normalisation, dedup, fingerprinting | Implemented with tests |
| CLI (Cobra) with Bubbletea menu, wizard, results browser | Builds; menu launches |
| MCP server implementation (`internal/mcp`, 655 lines, 5 tools) | Implemented, not reachable (see below) |
| Agent, AI, sandbox/DAST, exploit, attack-path, compliance, intel packages (~2,900 lines) | Implemented, not reachable |

### What the README promised and what happens

| Promise | What happens today |
|---|---|
| `brew install Armur-Ai/tap/vibescan` | The tap has no vibescan formula. The repo has no GitHub releases, so nothing can be installed from it. |
| `npm install -g @vibescan/cli` | 404. The package was never published. |
| `pip install vibescan` | **Installs someone else's package.** `vibescan` on PyPI is an unrelated third-party project. We were telling users to install code we do not control. |
| `curl ... install.sh \| sh` | Downloads from GitHub releases. There are none. |
| "No flags, no config" | Scanning needs the `armur-server` binary, a Redis instance, and every tool installed by hand. On a clean dev Mac `vibescan doctor` reports 2 of 25 tools. |
| `docker-compose up` | *Fixed.* The image did not build (Go 1.23 base against a 1.25 `go.mod`, a Python dependency conflict, a pinned Trivy download that no longer exists) and the compose file ran `go run` in an image with no Go toolchain. |
| `vibescan scan <repo>` | *Fixed.* The CLI waited forever: the progress stream only ever reports `queued`, because nothing writes per-tool progress. The CLI now polls task status as well. The progress feed itself is still empty. |
| `vibescan scan <file>` | *Fixed.* The CLI sent JSON to an endpoint that expects a multipart upload. |
| `vibescan scan <directory>` | Not supported. `/scan/local` is never called by the CLI, and it enqueues the wrong scan type (a repo clone). |
| `--advanced` | Runs a separate set of tools instead of adding to the standard scan, and the summary card does not count its findings ("Total findings: 0" on a scan with 200+). |
| Scan output | The `FILE` column is empty, messages are cut at 40 characters, and the total disagrees with the rows shown. |
| `vibescan mcp` | Prints setup instructions and exits. The real server in `internal/mcp` is never started. Tool names are still `armur_*`. |
| DAST, exploit simulation, attack paths, AI PR review | Nothing in the server or CLI imports these packages. |
| "30+ tools" | 37 runners are wired, but the Docker image installs about 17, and it installs `trufflehog3` (an unrelated older Python project) rather than TruffleHog (now removed from the image, so deep scans skip secret detection). Gitleaks, Grype, govulncheck, npm/pip audit, LLM, PII and crypto checks have wrappers that nothing calls. ShellCheck is advertised and does not exist. |
| JavaScript / TypeScript | The main language of vibe-coded apps gets two tools: Semgrep and ESLint. |
| `.vibescan.yml` | The CLI writes `.vibescan.yml`. The server reads `.armur.yml`. |

### Bugs to fix before anything else

1. **Scans strip the files other scanners need.** When a language is passed (the CLI requires
   `-l` for repo URLs), `RemoveNonRelevantFiles` deletes every file that is not that language's
   source: `package.json`, lockfiles, `.env`, Dockerfiles, SQL migrations. SCA, secrets and IaC
   tools then run on what is left.
2. **Latent data-loss path.** `RunScanTaskLocal` calls the same function on the user's own
   directory (`internal/tasks/tasks.go:193`). It is only unreached because `/scan/local` enqueues
   the wrong scan type. Fixing the routing without fixing this would delete users' files.
3. **Semgrep runs with `--config=auto`.** That needs network access, sends metrics, and pulls
   registry rules whose licence restricts use inside a competing product. See [Risks](#11-risks).
4. **Leftover branding.** `action.yml`, pre-commit hooks, MCP tool names, docs and the changelog
   still say `armur`; the Action installs from `install.armur.ai`.
5. **The banner names tools we do not run** (Nuclei, Snyk).

### The honest summary

vibescan is a well-tested orchestration server with a good-looking CLI shell and a lot of
unconnected feature code. It is not yet a product someone can install and run. The old roadmap
shows 996 boxes checked; most of the later ones are scaffolds. **Phase 0 exists to close that gap.**

---

## 3. Who we are building for

| Persona | What they want | How they find us | What they will not tolerate |
|---|---|---|---|
| **Solo builder** shipping with Cursor, Claude Code, Lovable, Bolt, v0 | "Is my app safe to launch?" in one command, in plain English, with a fix they can paste back into their agent | Reddit, X, Hacker News, YouTube, their agent suggesting it | Installing Docker and Redis. Jargon. 400 low-severity lint findings. |
| **Startup engineer** on a team that adopted coding agents | A gate in CI and in the agent loop so AI-written PRs do not ship the usual mistakes | GitHub Marketplace, blog posts, colleagues | False positives that block merges. Slow scans. |
| **Security lead** at a company where AI now writes half the code | Evidence of which AI-written code was checked, and a policy they can enforce | The benchmark report, conference talks | Closed-source black boxes. No SARIF. |

Phases 0 to 2 are for the solo builder. The startup engineer is served from Phase 2. The security
lead is served by the benchmark and provenance work in Phases 2 and 4.

---

## 4. The landscape and our wedge

The category already exists and is filling up fast.

| Group | Examples | Strength | Gap we exploit |
|---|---|---|---|
| Hosted URL scanners | Vibe App Scanner, CheckVibe, VibeCheck, amihackable.dev | Zero install, non-technical friendly | Closed source; mostly probe the deployed app; cannot see source-only flaws; no agent loop |
| Small open-source scanners | vibe-audit, VibeGuard, ChakraView | Right positioning | Tens of regex patterns; no taint analysis, no dependency intelligence |
| Platforms | Aikido, Snyk, Semgrep | Depth, enterprise trust | Heavy; priced and designed for security teams; vibe coding is a feature, not the product |
| Agent-surface scanners | Snyk agent-scan (formerly mcp-scan), Cisco MCP Scanner, rulesentry | Cover MCP, skills, rules files | Do not scan the application code at all |
| Built-in platform checks | Lovable security scan, Supabase advisors | In the workflow | Single platform; check that features exist, not that they are correct |

**Our wedge:** nobody combines all four of these in one open-source binary.

1. Real static analysis depth (taint-capable engine plus the best existing scanners).
2. Rules written for the vibe stack: Next.js, Supabase, Firebase, Vercel, Stripe, Clerk, LLM SDKs.
3. Dependency intelligence aimed at hallucinated and malicious packages.
4. The agent's own attack surface: rules files, MCP configs, skills, hooks.

And then closes the loop by handing the fix back to the agent.

---

## 5. The Vibe Top 10

A named taxonomy is the most valuable asset in a new category. It gives every finding a stable ID,
gives journalists and other tools something to cite, and organises our rule writing. This is the
first draft; it should be validated against the benchmark corpus in Phase 4 and versioned yearly.

| ID | Name | Typical AI-generated form |
|---|---|---|
| **V01** | Secrets shipped to the client | Supabase service-role key in a React component; `NEXT_PUBLIC_` or `VITE_` prefix on a secret; API keys in the JS bundle or a committed `.env` |
| **V02** | Missing or client-only authorization | API routes and server actions with no session check; auth enforced only by hiding a button; IDOR on `/api/users/[id]` |
| **V03** | Open data layer | Supabase tables without RLS, or policies of `using (true)`; Firebase rules `allow read, write: if true`; public storage buckets |
| **V04** | Hallucinated and risky dependencies | Packages that do not exist, were registered last week, are one edit away from a popular name, or are flagged malicious |
| **V05** | String-built injection | SQL via template literals; `exec` with user input; SSRF through user-supplied URLs; `dangerouslySetInnerHTML` |
| **V06** | Placeholder security | `JWT_SECRET = "secret"`; `admin/admin`; `verify=False`; `cors({ origin: "*" })`; debug mode on; `// TODO: add auth` |
| **V07** | Unverified webhooks and payment logic | Stripe webhook without signature verification; price or plan trusted from the client |
| **V08** | Missing abuse controls | No rate limiting on auth or LLM endpoints; unbounded uploads; no input validation |
| **V09** | Unsafe LLM integration | User input concatenated into system prompts; model output passed to `eval`, SQL or HTML; provider keys called from the browser |
| **V10** | Poisoned agent context | Hidden Unicode in `.cursorrules`, `CLAUDE.md`, `AGENTS.md`; unpinned `npx -y` MCP servers; blanket agent permissions; secrets in MCP config |

Every rule we ship maps to one of these plus a CWE. Reports lead with the Vibe Top 10 view.

---

## 6. Product pillars

1. **Ten seconds to first finding.** One static binary. No server, no Redis, no Docker for the
   default path. `npx vibescan` works on a fresh machine.
2. **Vibe-native detection.** Our own rule pack and stack analyzers for the Vibe Top 10, on top of
   the best open-source scanners.
3. **In the agent loop.** MCP server, editor hooks, and fix prompts, so the agent that wrote the
   bug fixes it before the human sees it.
4. **Signal over volume.** A default profile that shows only security findings worth acting on.
   Complexity, docstring and style tools are off unless asked for.
5. **Proof in public.** A reproducible benchmark of what each AI tool generates and what vibescan
   catches, published on a schedule.

---

## 7. Phases

Phases are ordered by dependency, not by date. Each has an exit test that can be checked by a
stranger on a clean machine.

### Phase 0: Make the promise true (v0.1)

Nothing else matters until a stranger can install vibescan and get a real result.

**Local engine**
- [ ] Merge CLI and server into one Go module and one binary. The scan pipeline is called as a
      library; `vibescan serve` remains for the API use case.
- [ ] Remove Redis and Asynq from the default path. In-process worker pool.
- [ ] Delete `RemoveNonRelevantFiles` from every scan path. Select files per tool with include
      globs instead. Never mutate the scan target. Add a test that scans a fixture and asserts the
      directory is byte-identical afterwards.
- [ ] Auto-detect languages and frameworks per directory; `-l` becomes optional everywhere.
- [ ] Fix local directory scanning end to end, including monorepos.
- [ ] Write real per-tool progress so the CLI and dashboard show what is running.
- [ ] Make `--advanced` a superset of the standard scan, and count its findings in the summary.

**Tool provisioning**
- [ ] `vibescan tools install`: download pinned, checksummed releases of each scanner into
      `~/.vibescan/tools`. No `pip install`, no `go install` at scan time.
- [ ] Fallback `--docker` mode that runs the same pipeline in one published image.
- [x] Make the Dockerfile build again.
- [ ] Pin every tool version in the image, add the real TruffleHog, and build the image in CI on
      every PR so it cannot rot again.
- [ ] A scan with zero external tools still returns results from the built-in checks (Phase 1).

**Truth in packaging**
- [ ] First tagged release through goreleaser; verify brew, npm and the install script on clean
      macOS, Linux and Windows runners in CI.
- [ ] Resolve the PyPI name collision (see [Open decisions](#12-open-decisions)).
- [ ] Wire `vibescan mcp` to the real server; rename tools to `vibescan_*`.
- [ ] One config file name: `.vibescan.yml`, with `.armur.yml` read as a fallback for one release.
- [ ] Purge `armur` from `action.yml`, pre-commit hooks, docs and changelog.
- [ ] Hide or remove CLI commands whose backends are not wired (`review`, `explain`, `fix`) until
      they work.
- [ ] Stop using Semgrep `--config=auto`; ship a pinned, licence-clean rule set, offline by default.

**Exit test:** on a clean macOS and a clean Ubuntu runner, `npx vibescan scan .` against
`examples/vibe-coded-app` finishes in under 60 seconds, reports the planted findings, leaves the
directory untouched, and makes no network calls other than tool downloads and vulnerability
database updates.

### Phase 1: Vibe-native detection (v0.2 to v0.3)

This is the moat. Each item is a built-in Go analyzer or a rule pack, tested against fixtures with
both vulnerable and safe variants.

**Secrets in the wrong place (V01)**
- [ ] Wire Gitleaks; add TruffleHog or Kingfisher for live validation of found keys.
- [ ] Client-exposure analyzer: secret-shaped values behind `NEXT_PUBLIC_`, `VITE_`, `EXPO_PUBLIC_`,
      `REACT_APP_`; Supabase service-role JWTs (decode and check the `role` claim); provider keys
      reachable from files marked `"use client"` or imported into browser bundles.
- [ ] Committed `.env` files and `.env` missing from `.gitignore`.

**Open data layer (V03)**
- [ ] Supabase migration analyzer: parse SQL migrations; flag tables without
      `enable row level security`, policies with `using (true)` or `with check (true)`, policies
      granted to `anon`, `security definer` functions without `search_path`, public buckets.
- [ ] Firebase rules analyzer for Firestore, Realtime Database and Storage: `if true`, missing
      `request.auth` checks, wildcard paths.
- [ ] Optional live mode using Supabase's Splinter lints when a database URL is supplied.

**Missing authorization (V02)**
- [ ] Next.js route inventory: enumerate `app/**/route.ts`, `pages/api/**`, and server actions;
      flag handlers that read or write data with no session or token check on the path.
- [ ] Middleware matcher gaps: routes that fall outside the `matcher` protecting the rest.
- [ ] The same inventory for Express, FastAPI, Flask, Hono and Supabase Edge Functions.
- [ ] IDOR heuristic: handlers that take an ID from the request and query by it without an owner
      predicate.

**Dependency intelligence (V04)**
- [ ] `vibescan deps`: for every dependency in manifests, lockfiles and import statements, check
      the registry. Flag names that do not exist, were published within the last 30 days, have
      negligible downloads, or sit within edit distance of a popular package.
- [ ] Malicious package detection via OSV `MAL-` advisories and GuardDog heuristics.
- [ ] Wire `npm audit`, `pnpm audit`, `pip-audit`, govulncheck and Grype (wrappers exist).
- [ ] Unused dependencies (knip, depcheck): AI adds packages it never uses.
- [ ] Offline mode with a bundled snapshot of popular package names.

**Placeholder security, webhooks, abuse controls (V06, V07, V08)**
- [ ] Rule pack for hardcoded JWT and session secrets, default credentials, disabled TLS
      verification, wildcard CORS with credentials, debug flags, security `TODO` markers.
- [ ] Stripe, Clerk, GitHub and Svix webhook handlers without signature verification.
- [ ] Prices, plans or roles read from the request body.
- [ ] Auth and LLM endpoints with no rate limiter on the route.

**Unsafe LLM integration (V09)**
- [ ] Replace the regex patterns in `llmsecurity.go` with taint rules: user input to prompt, model
      output to `eval`, SQL, shell or HTML sinks. Cover the OpenAI, Anthropic, Vercel AI SDK and
      LangChain call shapes.
- [ ] Provider SDK clients constructed in browser code.

**Reporting**
- [ ] Default profile `vibe`: security findings only. Complexity, docstring, duplication and style
      tools move to `--profile quality`.
- [ ] Rebuild the text report: file and line, the full message, one row per finding, a total that
      matches.
- [ ] Plain-English output: what it is, why an attacker cares, the fix. Vibe Top 10 ID plus CWE on
      every finding.
- [ ] A single launch-readiness score with the three things to fix first.

**Exit test:** on the fixture corpus, each Vibe Top 10 class has at least one rule with a
vulnerable and a safe fixture, and the default profile reports zero findings on the safe variants.

### Phase 2: In the agent loop (v0.4)

Where we stop being "another scanner" and become part of how AI-written code gets made.

- [ ] **MCP server that earns its place.** Tools: scan path, scan diff, check a dependency before
      installing it, explain a finding, list open findings. Works with Claude Code, Cursor,
      Windsurf, Copilot and Codex. One-line install for each.
- [ ] **Editor hooks.** `vibescan hooks install` writes a Claude Code `PostToolUse` hook that scans
      files after each edit and a `Stop` hook that blocks completion on new critical findings; the
      equivalent for Cursor hooks.
- [ ] **Pre-install guard.** A hook on shell commands that checks `npm install`, `pip install` and
      `npx` targets against `vibescan deps` before they run.
- [ ] **Fix prompts.** `vibescan fix --prompt` emits a paste-ready prompt with file, line, the
      vulnerable code, the required property of the fix, and how to verify it. No API key needed;
      the user's own agent does the work.
- [ ] **Guard rules.** `vibescan guard` writes a security section into `CLAUDE.md`, `AGENTS.md`,
      `.cursor/rules` and `copilot-instructions.md`, tailored to the detected stack and to the
      findings this repo keeps producing.
- [ ] **Agent surface scan (V10).** `vibescan agent-scan`: hidden and bidirectional Unicode in
      rules files and skills; instruction-injection phrases; MCP servers launched with unpinned
      `npx -y` or `uvx`; secrets in MCP config; blanket permission allow-lists; hooks that run
      remote code.
- [ ] **AI provenance.** Identify AI-authored commits from trailers and agent metadata;
      `vibescan scan --ai-authored` scans only that code; reports show findings by author type.
- [ ] Wire `vibescan explain` and `vibescan fix` to the existing `internal/ai` providers.

**Exit test:** in a recorded session, an agent writes a Supabase table without RLS, the hook
fires, the agent receives the finding, and fixes it without the human typing anything.

### Phase 3: Scanner breadth (v0.5)

Add and replace underlying tools per the [expansion plan](#8-scanner-expansion-plan). Every
addition needs a pinned version, a provisioning entry, a normaliser into the `Finding` type, a
fixture that proves it fires, and a Vibe Top 10 mapping.

**Exit test:** the tool matrix in the README is generated from the registry in code, and CI fails
if a listed tool has no passing fixture.

### Phase 4: Proof and distribution (v0.6 to v1.0)

- [ ] **VibeBench.** A fixed set of app prompts ("SaaS with auth and Stripe", "AI chat app",
      "internal admin dashboard") run through each major AI coding tool. Commit the outputs, scan
      them, publish the numbers and the method. This is also our precision and recall harness.
- [ ] **State of Vibe-Coded Security.** A short report from VibeBench, refreshed quarterly.
- [ ] **GitHub Action** on the Marketplace with SARIF upload and a PR comment that leads with the
      three findings that matter.
- [ ] **PR review** wired to `internal/agent`, diff-aware, commenting inline.
- [ ] **Web scan at vibescan.dev.** Paste a public repo URL, get a report. The top of the funnel
      for people who will never open a terminal.
- [ ] **Badge.** "Scanned by vibescan" with grade, linking to the public report.
- [ ] **Extension** for VS Code, Cursor and Windsurf (one codebase).
- [ ] Docs site that matches the product, plus `CONTRIBUTING.md` with a rule-writing guide.

**Exit test:** v1.0 ships when VibeBench shows recall above an agreed bar for every Vibe Top 10
class and the false-positive rate on the safe corpus is below an agreed bar. Set both bars from the
first benchmark run, not before.

### Phase 5: Runtime verification (after v1.0)

- [ ] `vibescan probe <url>` for apps you own: exposed `.env` and `.git`, source maps, keys in
      bundles, Supabase tables readable with the anon key, open Firebase databases, headers.
      Requires proof of ownership before any active check.
- [ ] Wire the existing sandbox, DAST and exploit packages to confirm static findings, so reports
      can say "confirmed exploitable" instead of "possible".
- [ ] Attack-path view built on confirmed findings.

### Later, demand-driven

Hosted dashboard and team features, policy enforcement, SBOM and compliance exports. These fund
the project; they do not define it.

---

## 8. Scanner expansion plan

### Fix or retire what we already wrap

| Tool | Action | Why |
|---|---|---|
| Semgrep `--config=auto` | Replace with a pinned rule set, evaluate Opengrep as the default engine | Offline, no metrics, licence-clean, cross-function taint |
| `trufflehog3` | Replace with TruffleHog (Go) or Kingfisher | Wrong project is installed today |
| tfsec | Retire in favour of Trivy config scanning | Upstream folded it into Trivy |
| Terrascan | Verify upstream status, likely retire | Not actively maintained |
| golint | Retire | Deprecated upstream for years |
| gocyclo, radon, pydocstyle, pylint, jscpd | Move to `--profile quality` | Noise for the default audience |
| Slither, Mythril | Keep, move to an opt-in `web3` profile | Not the vibe stack |
| Gitleaks, Grype, govulncheck, npm and pip audit | Wire into the pipeline | Wrappers already exist |

### Add

Ordered by value to the vibe stack. Tier 1 lands in Phases 1 and 2; the rest in Phase 3.

| Tier | Tool | Category | What it gives us |
|---|---|---|---|
| 1 | [Opengrep](https://github.com/opengrep/opengrep) | SAST engine | Semgrep-compatible rules with cross-function taint; the engine for our own rule pack |
| 1 | [Gitleaks](https://github.com/gitleaks/gitleaks) | Secrets | Fast, single binary, git history |
| 1 | [Kingfisher](https://github.com/mongodb/kingfisher) or [TruffleHog](https://github.com/trufflesecurity/trufflehog) | Secrets | Live validation: "this key works" is the finding people act on |
| 1 | [OSV-Scanner](https://github.com/google/osv-scanner) v2 | SCA | Lockfile scanning, malicious-package advisories, guided remediation |
| 1 | [GuardDog](https://github.com/DataDog/guarddog) | Supply chain | Heuristics for malicious npm, PyPI and Go packages |
| 1 | [Snyk agent-scan](https://github.com/snyk/agent-scan) | Agent surface | MCP, skills and agent config scanning, alongside our own rules-file checks |
| 1 | ESLint security stack | JS/TS SAST | `eslint-plugin-security`, `no-unsanitized`, `eslint-plugin-react` danger rules, `@next/eslint-plugin-next` |
| 2 | [ast-grep](https://github.com/ast-grep/ast-grep) | Structural rules | Very fast tree-sitter matching for framework-shape rules |
| 2 | [Ruff](https://github.com/astral-sh/ruff) (`S` rules) | Python SAST | Bandit's checks at a fraction of the runtime |
| 2 | [njsscan](https://github.com/ajinabraham/njsscan) | Node SAST | Node-specific sinks Semgrep's generic rules miss |
| 2 | [zizmor](https://github.com/zizmorcore/zizmor), [actionlint](https://github.com/rhysd/actionlint) | CI | AI-written GitHub workflows: injection, broad permissions, unpinned actions |
| 2 | [Squawk](https://github.com/sbdchd/squawk) | Database | Postgres migration linting next to our RLS analyzer |
| 2 | [Splinter](https://github.com/supabase/splinter) | Database | Supabase's own lints, in live mode |
| 2 | [Syft](https://github.com/anchore/syft) | SBOM | One SBOM feeding Grype, OSV and the dependency checks |
| 2 | [retire.js](https://github.com/RetireJS/retire.js) | Frontend SCA | Vendored and CDN-loaded libraries that no lockfile lists |
| 2 | [knip](https://github.com/webpro-nl/knip) | Dependencies | Unused packages and exports |
| 3 | [mobsfscan](https://github.com/MobSF/mobsfscan) | Mobile | Expo and React Native apps |
| 3 | [Vacuum](https://github.com/daveshanley/vacuum) or Spectral with the OWASP rules | API | OpenAPI specs |
| 3 | [Dockle](https://github.com/goodwithtech/dockle), [Kubescape](https://github.com/kubescape/kubescape) | Containers, K8s | Image and cluster posture |
| 3 | [poutine](https://github.com/boostsecurityio/poutine) | CI | Build-pipeline supply chain |
| 3 | [ShellCheck](https://github.com/koalaman/shellcheck) | Shell | Advertised already; actually add it |
| 3 | [Bearer](https://github.com/Bearer/bearer) | Data-flow SAST | Sensitive-data flows; licence needs review before bundling |
| 3 | [promptfoo](https://github.com/promptfoo/promptfoo), [garak](https://github.com/NVIDIA/garak) | LLM red-teaming | Dynamic testing of the app's own LLM features; belongs with Phase 5 |

### Build ourselves

No existing open-source tool does these well, which is why they are the moat:

- Supabase migration and policy analyzer
- Firebase rules analyzer
- Next.js, Express, FastAPI route and authorization inventory
- Client-exposure analyzer for secrets
- Hallucinated-dependency checker
- Rules-file and agent-config scanner
- The Vibe Top 10 rule pack

---

## 9. What we are cutting

The old plan had 59 sprints. These are removed from the roadmap and from the README until there
is demand:

- Distributed workers, multi-tenant API, RBAC and SSO
- Binary analysis, fuzzing, threat-model generation, test generation
- PCI-DSS, HIPAA and NIST compliance mapping (OWASP and CWE stay)
- Rules marketplace, gamification, Slack and Teams bots
- Per-language breadth for its own sake: PHP, Ruby, C/C++, C#, Swift, Java stay supported where
  wrappers work, but get no new investment. **TypeScript, JavaScript and Python get all of it**,
  with Go next.
- Exploit simulation and attack paths as headline features. They return in Phase 5 as
  verification of static findings.

The code for these stays in the tree for now. It leaves the marketing.

---

## 10. Go-to-market

**One sentence everywhere:** "Find vulnerabilities in Cursor, Claude Code, Copilot and Windsurf
generated applications before they reach production."

**Launch sequence**
1. Phase 0 done: quiet release, fix what breaks for the first fifty users.
2. Phase 1 done: public launch. Show HN, r/cursor, r/ClaudeAI, r/vibecoding, X. The demo is a
   thirty-second recording: generate an app, scan it, paste the fix prompt, rescan clean.
3. Phase 2 done: second launch aimed at agent users: "vibescan inside Claude Code and Cursor".
4. Phase 4: the benchmark report is the third launch and the one that gets press.

**Compounding loops**
- The Vibe Top 10 as a standalone page others link to.
- The badge on scanned repos.
- Agents recommending vibescan because `guard` put it in the repo's rules file.
- A rule-contribution path simple enough that a finding on X becomes a merged rule in a day.
- One "we scanned N apps built with X" post per major AI tool.

**What to measure**
- Time from install to first finding on a clean machine
- Weekly scans (opt-in, anonymous) and weekly active repos in CI
- Benchmark recall and false-positive rate per Vibe Top 10 class
- Stars and contributors as a lagging signal only

---

## 11. Risks

| Risk | Why it matters | Mitigation |
|---|---|---|
| **Name collision** | `vibescan` on PyPI belongs to someone else; several "vibe" scanners already exist | Decide the package names now; reserve npm, the VS Code publisher and the domain; never document an install path we do not own |
| **Rule licensing** | Semgrep's registry rules carry a licence that restricts use in competing products | Own rule pack plus permissively licensed community rules; engine via Opengrep |
| **Bundled tool licences** | Some candidates (Bearer, CodeQL) restrict redistribution or commercial use | Licence review is part of the definition of done for every tool added |
| **Crowded category** | New scanners appear monthly | Depth that regex scanners cannot copy quickly: stack analyzers, taint rules, benchmark, agent loop |
| **Platforms absorb the category** | Cursor, Lovable, GitHub ship built-in checks | Be the cross-tool, open, local option they cannot be; integrate rather than compete |
| **False positives** | One noisy launch loses the solo builder for good | Safe fixtures for every rule; the default profile is security-only; benchmark gates releases |
| **Probing third-party apps** | Active checks on apps the user does not own are a legal problem | Ownership proof before any active probe; passive checks only otherwise |
| **Credibility** | The previous README claimed features and install paths that did not work | Status table in the README; nothing is advertised before its exit test passes |

---

## 12. Open decisions

These need an owner's call before Phase 0 finishes.

1. **Package names.** PyPI `vibescan` is taken. Options: drop pip entirely (npm, brew and curl
   cover the audience), or publish as `vibescan-cli`. Recommendation: drop pip.
2. **Default engine.** Opengrep versus Semgrep CE. Recommendation: Opengrep, with Semgrep as a
   fallback if already installed.
3. **Module and repo naming.** The Go module is still `armur-codescanner` and the working folder
   `Armur-Code-Scanner`. Recommendation: rename the module to `github.com/Armur-Ai/vibescan` as
   part of the single-binary merge.
4. **Versioning.** The CLI prints `v0.1.0` and nothing has been released. Recommendation: tag
   `v0.1.0` at the Phase 0 exit test and stay below 1.0 until the benchmark gates pass.
5. **Telemetry.** Opt-in anonymous scan counts are the only way to measure the funnel.
   Recommendation: opt-in, documented, off by default.
6. **Hosted scan.** Whether vibescan.dev offers a free web scan in Phase 4 or earlier. It is the
   strongest top-of-funnel move and also an operating cost. The domain currently returns 503.

---

## 13. The first ten tasks

In order. Each is a separate PR.

1. Remove `pip install vibescan` from every doc and package README. *(Done in the README.)*
2. Delete `RemoveNonRelevantFiles` from scan paths; add the "target is untouched" test.
3. Add a Docker image build to CI. *(The Dockerfile itself is fixed.)*
4. Merge `cli/` and the server into one module; `vibescan scan .` runs the pipeline in-process.
5. Replace `--config=auto` with a pinned local rule directory.
6. Wire Gitleaks, OSV-Scanner, npm and pip audit into the default JS/TS and Python pipelines.
7. `vibescan tools install` with pinned, checksummed downloads.
8. Wire `vibescan mcp` to `internal/mcp`; rename tools to `vibescan_*`.
9. Rename leftovers: `action.yml`, pre-commit hooks, config file name, docs.
10. Tag `v0.1.0`, publish brew and npm, run the Phase 0 exit test in CI, then re-record the demos
    with `vhs demo/scan.tape` and `vhs demo/menu.tape`.
