# Problem Analysis: LLM Code Quality & Gatekeeping Agent ("Gatekeeper")

## 1. Problem Statement & Context

### 1.1 The Core Problem

Software teams practicing continuous integration face a recurring tension: **how to maintain consistent, high-quality code at the speed of continuous delivery.** As teams grow and merge frequency increases, human code reviewers become a bottleneck. They are inconsistent across reviewers, subject to fatigue, and often focus on superficial formatting rather than structural quality, test coverage adequacy, or adherence to sound engineering principles.

The result is that **substandard code enters the main branch**, accumulating technical debt, increasing regression risk, and eroding deployability. Teams lack a **scalable, objective, and consistent gate** that can evaluate every single merge request against meaningful quality criteria — and enforce a hard rejection when quality falls below a defensible threshold.

### 1.2 The "Why Now" — Pain Points Driving This Need

| Pain Point | Impact |
|---|---|
| **Inconsistent manual reviews** | Quality evaluation depends on who is available to review, leading to unpredictable outcomes |
| **Reviewer fatigue** | High-volume teams overwhelm reviewers; shortcuts are taken, rubber-stamp approvals become common |
| **Superficial linting** | Automated lint tools only catch syntax/style — they can't assess architectural soundness, test adequacy, or deployment readiness |
| **Coverage theater** | Teams achieve numeric coverage targets with meaningless tests that don't actually guard against regressions |
| **No objective "reject" threshold** | Without a clearly defined, consistently applied quality floor, low-quality code slips through under schedule pressure |
| **Slow feedback loops** | Delayed reviews mean developers have moved on, making rework costly and frustrating |
| **CI-only feedback is too late** | Developers only discover quality issues after pushing and waiting for pipeline execution. Issues caught at CI stage require context-switching back to code that may no longer be top-of-mind, wasting time and breaking flow |
| **Local/CI quality mismatch** | When quality checks exist only in CI, developers lack a way to self-assess before pushing. This creates a "push-and-pray" dynamic where the first quality signal arrives minutes (or hours) after the work is done |

### 1.3 Context & Environment

The agent must operate in **two distinct contexts**:

1. **Local environment (developer workstation / reviewer machine):** Run on-demand by individual developers before committing or pushing code, and by code reviewers during the review process. In this context, the evaluation is advisory — it informs and warns but does not enforce. Feedback must be near-instantaneous (seconds, not minutes) to preserve developer flow.

2. **CI/CD pipeline (automated gate):** Triggered automatically on pull request or merge request events against protected branches. In this context, the evaluation is authoritative — a score below threshold results in a hard rejection that blocks the merge. The same evaluation criteria must apply, producing consistent results regardless of where it runs.

In both contexts, the agent evaluates the **diff** (the change set) rather than the entire codebase, making its judgment scoped and incremental. The agent is not intended to replace human reviewers entirely, but to serve as an **automated first-pass gate** that catches clear quality issues before human time is wasted — whether that happens on the developer's machine before push, or in CI before merge.

## 2. Stakeholder Analysis

| Stakeholder | Perspective & Concerns |
|---|---|
| **Individual Contributors (Developers)** | Want fast, fair, transparent feedback. Fear false rejections blocking their work. Need actionable reasons for rejection to fix issues quickly. **Want to self-check quality locally before pushing — "shift-left" to catch issues in their own environment, within seconds, without waiting for CI.** |
| **Code Reviewers / Senior Engineers** | Want to focus on high-value review (design, architecture) rather than catching basic mistakes. Need trust that the automated gate won't create a false sense of security. **Want to run the same quality check locally during review to validate or challenge CI results, and to assess quality of PRs assigned to them without relying solely on pipeline output.** |
| **Engineering Managers / Tech Leads** | Want consistent quality standards enforced without adding headcount. Need metrics and visibility into code quality trends. Concerned about developer friction and velocity impact. |
| **DevOps / Platform Engineers** | Responsible for pipeline reliability, latency, and cost. Concerned about flaky gates, API outages, token costs, and CI minutes consumed. **Concerned about consistency between local and CI evaluations — divergent results undermine trust in both contexts. Must manage API key distribution and cost attribution across local and CI usage.** |
| **QA / Test Engineers** | Want assurance that test coverage analysis is meaningful, not just numeric. Worried about false confidence from weak tests passing the gate. |
| **Security / Compliance Teams** | Concerned about sensitive data in diffs being sent to external LLM APIs. Need data handling, retention, and privacy guarantees. |
| **Release Managers** | Need confidence that what's on main is deployable. Want clear go/no-go signals for release readiness. |

## 3. User Needs & Pain Points

### 3.1 Primary User Needs

1. **Developers need immediate, objective feedback** on code quality — within seconds when running locally (during development), and within a minute when running in CI (during PR checks). The same tool, same criteria, same output format.
2. **Developers need to understand WHY a submission was rejected** — vague rejections are frustrating and unactionable. Feedback must be specific and tied to the diff.
3. **Teams need a consistent quality floor** — the definition of "good enough" must not change depending on who is approving or what time of day it is.
4. **Teams need confidence in the gate** — neither too permissive (missed problems) nor too strict (developer frustration and velocity loss).
5. **Engineering leadership needs visibility** — trends in rejection rates, common failure categories, and quality trajectory over time. **Also needs to know whether developers are using the local check proactively (adoption rate), and whether local pre-checks correlate with higher CI pass rates.**
6. **Developers and reviewers need the quality check to be portable** — it must produce identical results whether run locally on a Mac, a Linux workstation, or in a CI container. Consistency between environments is non-negotiable for trust.
7. **Developers need the local check to be lightweight** — running on a developer machine alongside an IDE, browser, and other tools, without consuming excessive CPU, memory, or network bandwidth.

### 3.2 User Scenarios

**Local Context (Developer Workstation)**

- **Pre-Commit Self-Check:** A developer finishes a feature branch. Before pushing, they run the gate locally against their uncommitted changes. The gate scores a 7.2 — above threshold. Confident the CI gate will pass, they push and open a PR. Total feedback time: 8 seconds.
- **Pre-Commit Rejection:** A developer runs the local check. The gate scores a 4.1 and reports: "Missing error handling on lines 45-52. No tests covering new function `calculateFees()`. Hardcoded credentials on line 23." The developer fixes these issues locally, reruns the check, gets an 8.0, then pushes. No CI rejection, no wasted pipeline time, no context-switch penalty.
- **Reviewer Pre-Review Check:** A code reviewer is assigned a large PR. Before diving into a detailed architectural review, they pull the branch locally and run the gate. It flags three low-quality patterns. The reviewer can focus their human review on whether the flagged patterns are acceptable or need fixing, rather than discovering them from scratch.
- **Reviewer Challenging CI:** CI rejects a PR with a score of 5.5. The reviewer disagrees — the code looks sound to them. They pull the branch, run the local check, and get a 7.0. The divergence reveals a CI-specific issue (stale coverage data, different diff base). The inconsistency is investigated and resolved, preserving trust in the system.

**CI/CD Pipeline Context**

- **Happy Path:** A developer opens a PR. The gate analyzes the diff within 30 seconds, finds quality satisfactory (with strong test coverage and clean code), and auto-approves. The human reviewer focuses on architecture and design feedback only.
- **Rejection Path:** A developer pushes a quick fix late on Friday — no tests, error handling omitted, hardcoded values. The gate immediately rejects the PR with specific feedback: "Missing exception handling on external API call (line 12). No covering tests found. Configuration values hardcoded (line 15)." The developer knows exactly what to fix.
- **Borderline Case:** A submission scores 5.8 on the Farley scale — just below the 6.0 threshold. The gate rejects it with clear improvement steps. The team must decide: is the threshold appropriate, or should borderline evaluations trigger a mandatory human review instead?
- **False Positive:** A legitimate refactoring (e.g., extracting a method) triggers a rejection because coverage metrics haven't caught up. The developer is blocked on something mechanically correct. How does the override/exception process work?

## 4. Functional Requirements (User Terms)

### 4.1 Core Evaluation Capabilities

| ID | Requirement | User Value |
|---|---|---|
| **FR-1** | Evaluate code diffs for overall code quality — consistently across local and CI environments | Catch structural problems, anti-patterns, and maintainability issues before merge |
| **FR-2** | Assess test coverage adequacy specifically for the changed code paths | Prevent "test-free" code from entering the codebase; go beyond numeric coverage to assess whether tests actually test the changed behavior |
| **FR-3** | Assign a holistic quality score (Farley Score, 1–10) aggregating quality, coverage, and deployability | Provide a single, intuitively understood metric for go/no-go decisions |
| **FR-4** | Automatically reject submissions scoring below a configurable threshold (in CI context); warn/advise in local context | Enforce a quality floor; let developers self-correct before reaching CI |
| **FR-5** | Provide specific, actionable improvement recommendations on rejection | Enable developers to fix issues without guessing |
| **FR-6** | Filter out irrelevant files from evaluation (lock files, generated code, binary assets) | Focus evaluation on meaningful changes, reduce noise and cost |

### 4.2 Execution Context Requirements

| ID | Requirement |
|---|---|
| **FR-7** | Run as a local command-line tool on developer workstations, accepting uncommitted changes, staged changes, or a specified diff as input |
| **FR-8** | Run within CI/CD pipelines, triggered automatically on pull request / merge request events against protected branches |
| **FR-9** | Produce identical evaluation results (same score, same feedback) given the same diff and coverage data, regardless of whether run locally or in CI |
| **FR-10** | Operate in two behavioral modes: **advisory** (local — warns but does not block the developer's workflow) and **authoritative** (CI — rejects and blocks the merge when below threshold) |

### 4.3 Integration Requirements

| ID | Requirement |
|---|---|
| **FR-11** | Report results in a format consumable by CI/CD platforms (exit codes, step summaries, PR comments) |
| **FR-12** | Integrate with existing test coverage tooling to consume pre-computed metrics in both local and CI environments |
| **FR-13** | Report results locally in a human-readable terminal output, with optional structured output (JSON) for IDE/editor plugin integration |

### 4.4 Operational Requirements

| ID | Requirement |
|---|---|
| **FR-14** | Provide a fallback path when the gate cannot execute (e.g., API unavailable) — should not permanently block critical deployments |
| **FR-15** | Surface the raw data used in evaluation alongside the judgment, enabling reproducibility and auditability |
| **FR-16** | Support evaluation of working-directory changes (uncommitted), staged changes, committed changes against a base ref, and arbitrary diff files — enabling use at any point in the development workflow |

## 5. Non-Functional Requirements

| Category | Requirement | Rationale |
|---|---|---|
| **Latency (local)** | Evaluation must complete in under 10 seconds for typical diffs (under 500 lines changed) when run locally | Developers expect near-instant feedback to preserve flow; waiting >10s encourages skipping the local check |
| **Latency (CI)** | Evaluation must complete in under 60 seconds for typical diffs (under 500 lines changed) when run in CI | CI pipelines already take minutes; 60s is acceptable within that window |
| **Accuracy / Precision** | False rejection rate (rejecting good code) must be ≤ 5% | High false-positive rates cause developer frustration and erode trust |
| **Accuracy / Recall** | False approval rate (passing bad code) must be ≤ 10% | High false-negative rates defeat the purpose of the gate |
| **Cost** | Per-evaluation cost must be sustainable given the combined local + CI evaluation volume. Local usage multiplies evaluation frequency (developers may run 3-5 local checks per PR). Combined cost must remain within budget (e.g., ≤ $1.50 total per merged PR including all local and CI evaluations) | Local usage amplifies cost; if left unchecked, developers running dozens of local checks per day could make the system cost-prohibitive |
| **Reliability** | Gate availability ≥ 99.5% during working hours | A flaky gate that drops 1 in 20 evaluations creates unpredictability |
| **Security** | Code diffs must be handled with data security appropriate to the organization's posture | Code may contain proprietary IP; sending to third-party APIs requires data handling scrutiny |
| **Transparency** | Every rejection must include the specific reasoning and evidence | Developers must be able to understand and challenge the decision |
| **Configurability** | Thresholds, excluded file patterns, and evaluation weights must be tunable per repository/team. Configuration must be shareable (checked into the repository) so local and CI evaluations use identical settings | Different teams have different risk tolerances; shared config ensures local/CI consistency |
| **Portability** | The tool must run identically on macOS, Linux, and CI container environments (e.g., Ubuntu runners) | Developer workstations are heterogeneous; reviewers must get the same results as CI |
| **Scalability** | Must handle concurrent evaluations (multiple PRs in parallel) without degraded quality or timeout | Busy monorepos can have dozens of concurrent PRs |

## 6. Business Rules & Constraints

1. **Hard Rejection Rule:** Any submission scoring below the defined threshold (default: 6 on the Farley scale) is automatically rejected. This is the central business rule that gives the agent its "gatekeeper" identity.
2. **Override Mechanism (TBD):** It is unclear whether there is a human override path. If a senior engineer can override a rejection, under what circumstances? Is an audit trail required?
3. **Branch Protection:** The gate must protect specified branches (main, develop) but may be optional or advisory on feature branches.
4. **Coverage Data Ownership:** The gate consumes coverage data from CI tooling but does not generate it. The quality of the gate's coverage assessment is only as good as the coverage data it receives.
5. **Diff Size Limits:** Extremely large diffs (e.g., monorepo dependency updates touching thousands of files) may need special handling — either splitting, summarization, or bypass.
6. **Local vs. CI Behavioral Difference:** In the local (advisory) context, a sub-threshold score produces a warning with improvement suggestions but does not block the developer from committing or pushing. In the CI (authoritative) context, the same score produces a hard rejection that blocks the merge. The evaluation criteria and scoring must be identical in both contexts; only the consequence differs.
7. **Configuration as Source of Truth:** The evaluation configuration (thresholds, excluded patterns, evaluation criteria) must be stored in the repository (e.g., a config file) so that local and CI evaluations share the same rules without drift.

## 7. Success Criteria & Metrics

| Success Metric | Target | Measurement |
|---|---|---|
| **Defect escape rate reduction** | 30% fewer production defects originating from code merged after gate activation | Compare defect rates pre- and post-adoption |
| **Review cycle time reduction** | 40% reduction in median time-to-merge for approved PRs | PR lifecycle analytics |
| **Developer satisfaction** | ≥ 80% of developers rate the gate as "fair" and "helpful" in quarterly surveys | Developer NPS / survey |
| **Rejection actionability** | ≥ 90% of rejected PRs are resubmitted and approved within 2 iterations without escalating for manual override | PR lifecycle tracking |
| **Gate reliability** | ≥ 99.5% uptime during working hours, ≤ 0.5% of evaluations fail due to gate errors | CI pipeline monitoring |
| **Cost efficiency** | ≤ 2% increase in CI/CD pipeline cost attributable to the gate | Cloud/pipeline cost analysis |

## 8. Assumptions Requiring Validation

| Assumption | Risk if Wrong | Validation Method |
|---|---|---|
| **A1:** An LLM can reliably evaluate code quality from diffs alone (without full codebase context) at sufficient accuracy | High false-positive/negative rates | Pilot with historical PRs of known quality; compare against expert human evaluation |
| **A2:** The Farley Score (1–10) is an appropriate, consistent, and well-defined metric that can be reliably computed from a diff | Inconsistent scoring across evaluations | Inter-evaluation consistency testing; same diff evaluated multiple times |
| **A3:** Static coverage tooling provides sufficient signal for the gate to assess test adequacy | Gate misses untested code or flags tested code as uncovered | Cross-reference coverage data against manual test quality audits |
| **A4:** A hard threshold (6/10) creates the right balance between quality and velocity for the target team | Too strict: developer frustration. Too lenient: purpose defeated. | Tune threshold during pilot based on rejection rate and defect correlation |
| **A5:** Developers will trust and act on LLM-generated feedback | If trust is low, developers will game or bypass the system | Measure override requests, sentiment surveys, bypass attempts |
| **A6:** Sending code diffs to external LLM APIs is acceptable from a security and compliance standpoint | Data leakage, compliance violation | Security review, data handling agreement with API provider, option for self-hosted models |
| **A7:** Token costs for LLM evaluation are sustainable at the team's PR velocity | Budget overrun, gate becomes cost-prohibitive | Cost modeling based on average diff size and PR volume |
| **A8:** Diff filtering (removing lock files, generated code) is sufficient to keep evaluations within token limits and focused on meaningful changes | Important changes in large files are missed or evaluations become too expensive | Test with real-world diffs from the target repository |
| **A9:** Developers will voluntarily run local checks before pushing if the tool is fast and easy enough | Low local adoption means the CI gate remains the first quality signal, defeating the shift-left benefit | Track local check frequency; survey developers on friction points; correlate local check usage with CI pass rates |
| **A10:** Coverage data can be generated locally with comparable fidelity to CI-generated coverage data | Local and CI evaluations diverge because coverage data differs, eroding trust in both | Compare coverage reports from local vs. CI runs across a sample of PRs; identify and close gaps |
| **A11:** The same LLM API is accessible from both developer workstations and CI runners with consistent behavior | API access restrictions, different rate limits, or network policies prevent local execution or cause divergent results | Verify API access from representative developer environments and CI runners before rollout |

## 9. Risks & Unknowns

### 9.1 Identified Risks

| Risk | Severity | Mitigation Strategy |
|---|---|---|
| **False rejections erode trust** | High | Clear, specific feedback; override mechanism; threshold tuning; developer appeals process |
| **LLM hallucination in evaluation** | Medium | Structured output enforcement; deterministic coverage metrics fed as data, not derived by LLM; consistency monitoring |
| **API rate limits or outages block all merges** | High | Fallback to deterministic lint/coverage check; graceful degradation path |
| **Sensitive IP in code sent to external API** | High | Pre-filtering of sensitive patterns; option for self-hosted models; data processing agreement |
| **Teams game the system** | Medium | Developers learn what patterns the LLM flags and write to satisfy the gate rather than for genuine quality. Continuous refinement of evaluation criteria needed. |
| **Token cost overrun on large diffs** | Medium | Diff size limits; cost monitoring and alerts |
| **Cost amplification from local usage** | High | Developers running the check 5-10 times per day per person can multiply costs 5-10x versus CI-only usage. Without cost controls, a 20-developer team could generate thousands of evaluations daily. |
| **Evaluation inconsistency** | Medium | Same diff should produce similar scores across evaluations. Prompt stability and low temperature needed. |
| **Latency violates CI expectations** | Medium | 60-second SLA for typical diffs; need for async evaluation option if larger diffs take longer |

### 9.2 Unknowns

1. **What happens at the borderline (score = 5.8)?** Is it a hard reject, or should borderline cases trigger a mandatory human review while still flagging? The exact boundary behavior is undefined.
2. **How should the gate handle merge conflicts or stacked PRs?** If a PR's base branch has moved, does the gate re-evaluate against the new base?
3. **What is the process for updating evaluation criteria?** As teams grow and standards evolve, how do the prompts or rules get updated? Who owns them?
4. **How does the gate interact with required status checks?** If the gate is one of several required checks, does it run before or after other checks? In parallel or sequentially?
5. **Multi-language repositories?** Does the gate handle repositories with multiple programming languages? Are evaluation criteria language-specific?
6. **What about non-code changes?** Documentation-only PRs, configuration changes, infrastructure-as-code — does the gate evaluate these differently or skip them entirely?
7. **How do local checks handle lack of coverage data?** If a developer runs the gate locally without first generating a coverage report, does the tool require coverage data, skip coverage assessment, or generate it on-the-fly? What is the local experience when coverage tooling is not set up?
8. **How are API credentials distributed to developers for local use?** Does every developer get their own API key? Is there a shared key? How is usage tracked and attributed per developer vs. per CI run?
9. **What happens when local and CI evaluations disagree?** If a developer's local check passes but CI rejects (or vice versa), what is the resolution path? Which result is authoritative? How is the divergence debugged?

## 10. Problem Domain Boundaries

### In Scope
- Automated evaluation of code diffs — both locally (advisory) and in CI/CD pipelines (authoritative)
- Consistent evaluation across three pillars: code quality, test coverage, deployability (Farley Score)
- Hard automated rejection below quality threshold (in CI context)
- Advisory warnings with improvement suggestions (in local context)
- Shared, repository-checked-in configuration ensuring local/CI evaluation consistency
- Integration into CI/CD pipelines and local developer workflows

### Out of Scope (Explicitly)
- **Replacing human code reviewers entirely** — the gate is a first-pass filter, not a replacement for architectural and design review
- **Evaluating entire codebases** — it evaluates incremental changes (diffs), not the full repository
- **Generating or fixing code** — the gate assesses and rejects, it does not auto-fix
- **Generating test coverage data** — it consumes coverage metrics from existing tooling
- **Security vulnerability scanning** — while code quality assessment touches on safety, dedicated SAST/DAST scanning is a separate concern
- **Performance profiling or benchmarking** — the gate does not execute code or measure runtime characteristics
- **Blocking local commits or pushes** — the local mode is strictly advisory; it does not prevent the developer from proceeding
