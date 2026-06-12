# Product & Technical Specification: "Gatekeeper" — Automated Quality Score Tool

**Document Version:** 1.2.0  
**Status:** Approved for Implementation / RFC Closed  
**Target Audience:** Platform Engineering, DevOps, Engineering Leads, Core Developers  
**Core Objective:** Define the CLI, API, architectural requirements, and configuration frameworks for **Gatekeeper**, a lightweight, high-performance binary tool that evaluates git commits, pull requests, or full codebases against the organizational Quality Score Standard using an option-compatible LLM execution core.

---

## 1. System Overview & Archetypes

**Gatekeeper** is an objective, single-binary engine designed to unify static analysis, structural code metrics, security logs, and Farley Index attributes using a combination of fast programmatic logic and advanced LLM-driven reasoning.

The tool provides an execution runtime wrapper for three critical enterprise personas:
1. **The Developer (Local CLI Mode):** Low-latency (sub-second) loop context triggered via the command line or local git pre-commit hooks.
2. **The Code Reviewer / Dev Lead (Differential Mode):** Delta code metrics and architectural regression logs generated directly as a markdown payload into active Pull Requests.
3. **The CI/CD Engine (Orchestration/Gating Mode):** A headless, deterministic execution loop that gates branches from moving through integration environments using standard system exit codes.

---

## 2. Command Line Interface (CLI) Specification

### 2.1 Primary Command Architecture
```bash
# Evaluate the current directory state (Full Workspace Check)
gatekeeper check --path=./

# Evaluate target differences between the local branch and a base branch (PR / Merge Check)
gatekeeper diff --base=main --target=feature/auth-refactor

# Evaluate an explicitly isolated range of git commits
gatekeeper commit-range --range=HEAD~3..HEAD

```

### 2.2 Output Format Strategies

```bash
# Output localized terminal tables with rich formatting (Default for local devs)
gatekeeper check --format=pretty

# Output structural machine JSON for pipeline logging or database ingestion
gatekeeper check --format=json --output=report.json

```

---

## 3. Functional Requirements Matrix

### 3.1 Input Ingestion & Context Scope

* **Workspace Ingestion:** Reads path configurations, abstracts git state loops, respects `.gitignore` rules, and parses a mandatory, centralized structural file named `gatekeeper.json` located at the root of the source directory.
* **Git Commit Delta Parsing:** Isolates structural lines and modified syntax structures using localized git operations, focusing the evaluation on the active delta envelope to keep processing optimized.

### 3.2 Data Aggregation Engine (Pillar Computation)

Combines programmatic parser outputs with dynamic metrics reports to compute the global **Quality Score** out of a strict **100 points** based on the following algorithm:

$$\text{Quality Score} = S_{\text{Static}} + A_{\text{Architecture}} + F_{\text{Verification}} + \text{Sec}_{\text{Integrity}}$$

* **Static Code Health ($S_{\text{Static}}$) — Max 20 Points:** Tracks formatting compliance, dead code elimination, and linter warnings.
* **Engineering Architecture ($A_{\text{Architecture}}$) — Max 25 Points:** Computes the net percentage of codebase functions maintaining a Cyclomatic Complexity score of $\le 10$, alongside a Cognitive nesting depth limit of $\le 3$.
* **Dynamic Verification Layer ($F_{\text{Verification}}$) — Max 35 Points:** Ingests test outputs and normalizes the 0–10 scale of your calculated **Farley Index ($FI$)**:

$$\text{Base Points} = \text{Farley Score (FI)} \times 3.5$$


* *The Mutation Testing Guardrail:* If mutation code coverage drops below 80%, a mandatory **0.8x multiplier** is automatically slapped onto this entire pillar score.


* **Security & Supply Chain Integrity ($\text{Sec}_{\text{Integrity}}$) — Max 20 Points:** Auditors look for zero active High/Critical SAST flaws and zero unresolved package vulnerabilities (CVEs).

### 3.3 The Recommendation Engine

The engine processes results and formats human-remediations parsed natively from structured outputs. It delivers localized findings detailing file names, code offsets, and mitigation steps ordered by high-priority architectural impact.

---

## 4. LLM-Driven Engine & Abstraction Layer

Gatekeeper maps abstract metrics evaluation—such as Intent-Driven Readability and complex code deduplication rules—to an LLM abstraction matrix that requires zero specific platform lock-in.

### 4.1 OpenAI-Compatible Abstraction

The system core routes requests through a unified API routing module that translates incoming requests into OpenAI-compatible endpoints. This allows the engine to coordinate seamlessly with cloud-hosted platforms or air-gapped on-premise inference engines (vLLM, Ollama, DeepSeek, or localized corporate gateways).

### 4.2 Prompt Orchestration & JSON Gating

The engine configures inference parameters using **JSON Mode / Structured Outputs**. The prompt controller packages the targeted file code diff into a prompt envelope alongside rule profiles, forcing the model to respond in a strict JSON object shape corresponding exactly to internal aggregation definitions.

The engine requires the model to reply exclusively inside this structural blueprint:

```json
{
  "pillar_adjustments": {
    "static_health_deduction": 1.5,
    "architecture_deduction": 0.0
  },
  "remediations": [
    {
      "priority": "HIGH",
      "pillar": "static",
      "location": "src/auth.py:lines 12-24",
      "finding": "Variable naming violates intent-driven readability standard. Identifiers 'a', 'b', and 'temp_cache' hide business context.",
      "actionable_fix": "Rename 'a' to 'token_payload', 'b' to 'signature_secret', and convert the logic into self-documenting method parameters."
    }
  ]
}

```

---

## 5. Technical, Performance & Security Requirements

### 5.1 Speed Boundaries

* **Delta Changes:** Commits modifying fewer than 10 files must calculate and exit within **300ms** to avoid interrupting standard engineering routines.
* **Full Workspaces:** Complete codebases up to 100,000 lines must execute within **10 seconds** using multi-threaded local parsing engines.

### 5.2 Deterministic System Outputs (Exit Signaling)

The system binary relies strictly on predictable system exits to signal pipeline architectures:

* **`Exit Code 0`:** Quality Score $\ge 75$ (Success, clearance given).
* **`Exit Code 1`:** Core technical runtime system error (Unreadable schemas, missing paths).
* **`Exit Code 2`:** Quality Score $< 75$ (Gating validation breach, pipeline block triggered).

### 5.3 Performance & Security Safeguards

* **Pre-Filter Cache Gate:** If a git commit delta consists purely of structural whitespace, documentation blocks, or non-functional data, the network LLM client call is completely skipped to preserve sub-second speed loops.
* **Structural Slicing:** The tool isolates and transmits only code segments directly tied to the changed files, filtering out global project footprints to stay within provider context windows and limit processing costs.
* **Air-Gapped Controls:** Incorporates data scrubbing mechanics to erase patterns matching authentication secrets, keys, or internal environment configurations prior to leaving the local terminal loop.

---

## 6. Configuration Management Schema (`gatekeeper.json`)

The entire operational surface, rule enforcement settings, exclusion scopes, and LLM providers are configured cleanly via a structured JSON schema configuration file located at the root directory layer of the project.

### 6.1 Production Configuration Template (`gatekeeper.json`)

```json
{
  "gatekeeper": {
    "target_threshold": 75.0,
    "fail_on_critical_security": true,
    "llm": {
      "provider": "openai-compatible",
      "base_url": "[https://api.openai.com/v1](https://api.openai.com/v1)",
      "model_name": "gpt-4o-mini",
      "api_key_env_var": "GATEKEEPER_API_KEY",
      "timeout_ms": 4000,
      "temperature": 0.0
    },
    "privacy": {
      "allow_public_cloud_transmission": false,
      "data_scrubbing": true
    }
  },
  "pillars": {
    "static": {
      "max_cyclomatic_complexity": 10,
      "allow_experimental_modules": false
    },
    "verification": {
      "require_mutation_testing": true,
      "mutation_coverage_floor": 80.0,
      "historical_pipeline_telemetry_url": "[https://internal.devops.corp/api/telemetry](https://internal.devops.corp/api/telemetry)"
    }
  },
  "exclusions": {
    "paths": [
      "**/vendor/**",
      "**/dist/**",
      "**/generated/*.go"
    ]
  }
}

```

---

## 7. Codebase Operational Tiers & Deployment Clearance

The calculated aggregate Quality Score maps directly to automated release gates:

| Total Quality Score | Operational Status | Release Automation Pipeline Action |
| --- | --- | --- |
| **90 – 100** | **Elite Tier** | Unrestricted deployment. Fully automated continuous delivery enabled. |
| **75 – 89** | **Target Standard** | Healthy codebase. Approved for release; minor technical debt must be tracked. |
| **60 – 74** | **Technical Debt Alert** | Degrading health. Pipeline flags a warning. Pull Requests that lower the score are automatically blocked. |
| **0 – 59** | **Critical Remediate Status** | **Pipeline Blocked.** Automated release engine disabled. Feature delivery must halt to refactor code complexity and failing Farley properties. |

---

## 8. Implementation Milestones

1. **Milestone 1:** Standardize `gatekeeper.json` configuration schema validations and core terminal reporting blocks.
2. **Milestone 2:** Implement local Git differential line parsing loops.
3. **Milestone 3:** Deploy the OpenAI-compatible abstraction layers and structured prompt runners using JSON Mode.
4. **Milestone 4:** Package the automation architecture as custom CI/CD pipeline plugins and local pre-commit hooks.

```

```
