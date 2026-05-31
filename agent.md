# Gatekeeper — Agent Context

You are working on **Gatekeeper**, an LLM-powered automated code quality gate written in **Golang** that evaluates code diffs and enforces a quality threshold.

## Tech Stack

- **Primary language:** Golang
- **LLM client framework:** [CloudWeGo Eino](https://github.com/cloudwego/eino) — the LLM client and workflow orchestration layer
- **Runtime:** Go (cross-platform: macOS, Linux, CI containers)

## Project Overview

- **Location:** `/Users/eddie/work/gatekeeper`
- **Artifact:** `problem-analysis.md` (the complete problem specification)
- **Status:** Analysis phase — no code exists yet.

## The Problem

Software teams lack an automated, consistent, and objective gate that evaluates every pull request against meaningful quality criteria and enforces a hard rejection when quality falls below a defensible threshold. Manual code reviews are inconsistent, fatigued, and often superficial. Linting tools catch syntax but not architectural soundness. CI catches too late and creates a push-and-pray dynamic.

## What Gatekeeper Is

Gatekeeper is an **LLM-based evaluation agent** that:
- Evaluates **code diffs** (not entire codebases) against quality criteria
- Produces a **Farley Score** (1–10) aggregating code quality, test coverage adequacy, and deployability
- **Rejects** submissions below a configurable threshold in CI (authoritative mode)
- **Warns** with actionable feedback when run locally (advisory mode)
- Uses the **same criteria and scoring** in both environments — only the consequence differs

## What Gatekeeper Is NOT

- Not a replacement for human code reviewers (architectural/design review)
- Not a full-codebase evaluator (diffs only)
- Not a code generator or fixer
- Not a test coverage generator (consumes existing coverage data)
- Not a security scanner (SAST/DAST is separate)
- Not a performance profiler

## Core Evaluation Pillars

1. **Code Quality** — structural problems, anti-patterns, maintainability
2. **Test Coverage Adequacy** — whether changed code paths are tested (not just numeric coverage)
3. **Deployability** — readiness for production deployment

## Key Architectural Constraints

| Constraint | Value |
|---|---|
| Local latency | < 10 seconds for typical diffs (< 500 lines) |
| CI latency | < 60 seconds for typical diffs |
| False rejection rate | ≤ 5% |
| False approval rate | ≤ 10% |
| Default threshold | 6/10 (Farley Score) |
| Evaluation scope | Diff / change set only |
| Two modes | Advisory (local) and Authoritative (CI) |
| Portability | Identical results on macOS, Linux, CI containers |
| Config | Repository-checked-in, shared between local and CI |

## Key Terminology

- **Farley Score:** 1–10 quality metric aggregating quality, coverage, deployability
- **Advisory mode (local):** Warns but does not block; near-instant feedback
- **Authoritative mode (CI):** Hard rejects below threshold; blocks merge
- **Diff evaluation:** Scope is the change set, not the full codebase
- **Threshold:** Default 6/10 — below = reject in CI, warn in local
- **Shift-left:** Developer runs gate locally before pushing to catch issues early

## Problem Analysis Reference

The full problem specification is in [`problem-analysis.md`](problem-analysis.md). Key sections:

- **Section 1:** Problem statement and pain points
- **Section 2:** Stakeholder analysis (developers, reviewers, managers, DevOps, QA, security)
- **Section 3:** User scenarios (pre-commit, reviewer, CI happy path, rejection, borderline)
- **Section 4:** Functional requirements (FR-1 through FR-16)
- **Section 5:** Non-functional requirements (latency, accuracy, cost, security, portability)
- **Section 6:** Business rules (hard rejection, override, branch protection, config ownership)
- **Section 8:** Assumptions requiring validation (A1–A11)
- **Section 9:** Risks and mitigations

## Design Principles

- **Same diff, same score:** Identical evaluation given the same diff and coverage data, whether local or CI
- **Actionable feedback:** Every rejection includes specific, line-level reasons
- **Config as source of truth:** Shared configuration file ensures local/CI consistency
- **Graceful degradation:** Fallback when LLM API is unavailable — don't permanently block deployments
- **Cost awareness:** Local usage multiplies evaluation frequency; budgets must account for 3–5 local checks per PR
- **Security first:** Code diffs sent to external APIs require data handling scrutiny and optional self-hosted models

## Important Unknowns (flag these when designing solutions)

1. Borderline score behavior (5.8?) — hard reject or mandatory human review?
2. Override/exception process — who can override, audit trail?
3. Local vs. CI disagreement — resolution path, which is authoritative?
4. Coverage data generation locally — required, optional, or on-the-fly?
5. API credential distribution for local use — per-developer or shared keys?
6. Multi-language handling — language-specific criteria?
7. Large diffs (>500 lines) — split, summarize, or bypass?

## Rules for Implementation

- **Golang is the main language** — all production code should be idiomatic Go
- Use **CloudWeGo Eino** for LLM client interactions, model routing, and workflow orchestration — do not use ad-hoc HTTP calls or other LLM libraries
- When implementing features, always design for both **local (advisory)** and **CI (authoritative)** modes
- Always produce the same score for the same diff regardless of execution context
- Configuration must be shareable and checked into the repository
- Always handle the fallback path when the LLM API is unavailable
- Always provide specific, actionable feedback — never vague rejections
- Local mode never blocks; it warns. CI mode can hard-reject.
- Keep evaluation scoped to diffs, never full codebases
- Consider cost implications — developers may run 3–5 local checks per PR
