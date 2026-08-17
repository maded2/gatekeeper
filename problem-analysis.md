# Problem Analysis — Using Google Gemini (OAuth) as the LLM Provider

**Date:** 2026-07-15
**Request:** "Add ability to use Google Gemini OAuth API as LLM"
**Status:** Analysis complete — several ambiguities must be resolved before scope can be defined (see §10).

---

## 1. Problem Statement & Context

Gatekeeper's differentiator per spec (`gatekeeper.md`) is combining static analysis with **LLM-driven reasoning** to produce pillar adjustments and actionable remediations. The LLM is configured in `gatekeeper.json` and, per project convention (AGENTS.md), targets any OpenAI-compatible endpoint via the Eino framework.

### What already exists (verified in code)

- A **generic OAuth2 authorization-code browser flow** (`internal/llm/oauth.go`): opens a browser, runs a local callback server, exchanges the code, caches the token to disk (`~/.cache/gatekeeper/oauth_token.json` per README), and transparently refreshes expired tokens. Unit-tested in isolation.
- **Config schema support** for it: `auth_type: "oauth_browser"` plus `oauth_*` fields in `internal/config/config.go`, with validation requiring token URL, auth URL, client ID/secret env vars, redirect URL, and scopes.
- **README documentation** presenting a Google Gemini example (personal Google account, `accounts.google.com` auth URL, `cloud-platform` scope, `generativelanguage.googleapis.com/v1beta` base URL).
- Secret scrubbing (`ScrubSecrets`) and a retry contract (2 retries, then rule-based fallback per Story G-2).

### The actual gap (verified in code)

**Nothing calls an LLM.** `internal/llm` is imported by no other package; `GetAPIKey` and `ScrubSecrets` have zero production callers; there is no chat client, no prompt execution, no structured-output handling, and no Eino dependency in `go.mod`. The evaluator today is entirely rule-based. Consequences:

1. The README **promises a capability the product does not deliver** — a user following the documented Gemini config would complete a browser login and then… nothing would ever use the token.
2. The OAuth flow has **never been exercised end-to-end against Google's real endpoints** (only unit tests with mocked servers). Whether it completes, which scopes work, and whether the resulting token is accepted by the configured base URL are all unverified.
3. "Add Gemini OAuth as LLM" therefore likely means two different things to different people: **(a)** make Gemini actually usable end-to-end as the evaluation LLM (requires the entire missing call path), or **(b)** add/fix Gemini-specific auth support (which largely already exists). This must be clarified with the user.

### Operational context that shapes the problem

Gatekeeper has three modes: local CLI, differential (PR), and **CI/CD gating**. CI runners are headless — no browser, no interactive user. Any auth design that only works via an interactive browser flow leaves the primary gating mode unable to use Gemini at all. Additionally, spec §5.3 mandates air-gap controls and secret scrubbing, which directly constrain how a public Google endpoint can be used.

---

## 2. Stakeholder Analysis

| Stakeholder | Perspective / Concerns |
|---|---|
| **Developer (local CLI)** | Wants one-time browser sign-in, then zero-friction runs. Hates re-login prompts, opaque token errors, and anything that slows the <300ms diff loop. |
| **CI/CD pipeline** | No human, no browser. Needs non-interactive authentication or it cannot use Gemini at all. Depends on deterministic exit codes (0/1/2); an auth failure must not silently flip a gate's meaning. |
| **Security Officer** | Air-gap policy (`allow_public_cloud_transmission: false`) may forbid any transmission to Google. A refresh token persisted in plaintext on disk is a **new class of long-lived secret**. Least-privilege scopes matter (the documented `cloud-platform` scope is very broad). |
| **Org / platform admin** | Owns the Google Cloud project and OAuth client registration. Concerned about quota, billing, model governance, and whether end users must each register their own OAuth client (a significant onboarding burden if so). |
| **Gatekeeper maintainers** | Must preserve: deterministic exit codes, performance budgets, slog-only structured logging with no secret leakage, the Eino mandate, and docs that match behavior. |

---

## 3. User Needs & Pain Points

- **Broken promise:** The README documents Gemini OAuth; following it yields a working login and a useless token. Users will perceive this as a bug or a half-finished feature.
- **Token lifecycle friction:** Access tokens live ~1 hour. Without transparent refresh, users would re-login daily. Refresh failures (revoked consent, network) currently fall back to a full browser flow — fine locally, fatal in CI.
- **Headless gap:** A developer who can use Gemini on their laptop cannot in the pipeline that actually gates releases — the mode where quality gating has its highest value.
- **Failure semantics ambiguity:** When auth fails or the LLM is unreachable, does Gatekeeper (a) fall back to rule-based scoring (Story G-2 behavior), or (b) hard-error with exit 1? A user who explicitly configured Gemini may not want silent degradation that hides a broken setup; a pipeline may not want a false gate failure from an expired token.
- **Onboarding burden:** If each user must create their own Google Cloud OAuth client (client ID + secret env vars), the setup cost is high for a "just check my code" tool. Users need to understand what this means before adopting it.
- **Determinism anxiety:** Gating decisions should not flip run-to-run because of LLM nondeterminism or quota exhaustion mid-pipeline.

---

## 4. Functional Requirements (user terms)

| ID | Requirement |
|---|---|
| FR1 | A user can configure `gatekeeper.json` so that Gemini is the LLM behind evaluation, and running `check`, `diff`, or `commit-range` produces LLM-influenced pillar adjustments and remediations (the spec's structured output shape). |
| FR2 | First-time authentication completes via an interactive browser sign-in; the resulting token is cached locally so subsequent runs are non-interactive. |
| FR3 | Expired tokens refresh transparently with no user action; when refresh is impossible, the user gets a clear, actionable message (re-login instructions), not a crash or a cryptic error. |
| FR4 | *(Scope TBD — see Q2)* Headless environments can authenticate without a browser, or the product explicitly documents that Gemini cannot be used headlessly and degrades predictably. |
| FR5 | Auth failures, quota errors (429), and LLM unavailability degrade according to one clearly-defined policy: either rule-based fallback (G-2) or hard error (exit 1) — chosen deliberately, documented, and consistent across all three operational modes. |
| FR6 | When `privacy.allow_public_cloud_transmission` is `false`, using a public Google endpoint behaves per a defined policy (block with clear message, or warn-and-continue) rather than silently violating the air-gap setting. |
| FR7 | All code content sent to Gemini passes through secret scrubbing before transmission; only changed functions/classes are transmitted (minimal transmission), never full files. |
| FR8 | Gemini responses conform to the structured output contract (`pillar_adjustments` + `remediations` JSON) so scores and remediations remain well-formed regardless of provider. |
| FR9 | Model identity is user-configurable; a nonexistent or deprecated model name produces a clear, actionable error (not a generic timeout). |
| FR10 | The documented Gemini configuration in the README either works end-to-end as written or the documentation is corrected to match reality. |

---

## 5. Non-Functional Requirements

- **Performance:** diff checks on < 10 files must stay under 300ms (AGENTS.md rule 4). In the common case (valid cached token), authentication must add no network round-trip. Refresh adds at least one round-trip — its budget against the performance target is currently undefined.
- **Determinism:** temperature 0 must be honored by Gemini for stable gating; exit-code semantics must be identical whether or not the LLM participated.
- **Security:** access tokens, refresh tokens, and client secrets must never appear in logs, terminal output, reports, JSON output, or error messages. The on-disk token cache is a long-lived credential: its location, file permissions, and sharing semantics (per-user vs shared) need an explicit stance.
- **Portability:** must work on Linux, macOS, Windows; headless behavior must be defined rather than accidental.
- **Reliability:** the existing contract of 2 retries then rule-based fallback (Story G-2) applies to LLM call failures; whether it also covers auth failures is an open decision (FR5).
- **Observability:** structured slog entries for auth state transitions (loaded from cache, refreshed, re-authenticated, failed) with zero secret material.

---

## 6. Business Rules & Constraints

1. **Exit code contract is immutable** (spec §5.2): 0 = pass, 1 = runtime error, 2 = fail. CI pipelines depend on it; any new auth failure mode must map to one of these without ambiguity.
2. **Air-gap and scrubbing are spec-mandated** (§5.3) — a Gemini integration cannot bypass them.
3. **Project conventions (AGENTS.md):** all LLM orchestration via Eino; slog only; no secrets in logs; minimal transmission; tests for every new package; commit per story.
4. **Google-side constraints (external, to be validated — see §8):** which authentication mechanisms Google officially supports for the chosen Gemini endpoint; required scopes; whether a client secret exists for the user's OAuth client type (the current config *requires* one, but some Google client types are public and have no secret); quota and billing rules for OAuth-authenticated calls.
5. **Documentation consistency:** README currently asserts Gemini OAuth works; product behavior and docs must converge in whichever direction the decisions go.
6. **Protocol compatibility:** the product's LLM layer speaks an OpenAI-compatible wire protocol (per spec §6 and AGENTS.md). The documented Gemini base URL is a native Gemini endpoint, which is a different wire protocol. Whether the configured URL can satisfy the protocol the product speaks is unverified (see A1/A3).

---

## 7. Success Criteria & Metrics

- **Setup:** a developer with a Google account goes from fresh clone to a working Gemini-backed evaluation in a small, documented number of steps — without creating an API key.
- **Day-to-day:** the second and later runs are fully non-interactive and complete within existing performance budgets.
- **Token expiry:** an expired token refreshes with no user action; a revoked token produces a clear re-login path and never crashes the binary.
- **CI:** a pipeline run authenticates without a browser (or degrades per the chosen policy) and its exit code reflects score, not auth state.
- **Security audit:** zero occurrences of token/secret material in any log line, report, or error message; the token cache file is created with restrictive permissions.
- **Resilience:** with Gemini unreachable, Gatekeeper still produces a rule-based score and a well-defined exit code (per FR5 policy).
- **Docs:** every Gemini-related claim in the README has been verified against real behavior.

---

## 8. Assumptions Requiring Validation

| # | Assumption | Why it matters |
|---|---|---|
| A1 | A personal-Google-account OAuth token (browser flow) is accepted by the Gemini endpoint at the configured base URL with the configured scope | The entire documented setup may not work; no live call path exists to have caught this. |
| A2 | The user's Google Cloud OAuth client has a usable **client secret** | Current config validation *requires* `oauth_client_secret_env_var`; public/installed-app clients have no secret, which would make the documented flow impossible for them. |
| A3 | The Gemini endpoint to be used speaks the same wire protocol as the product's OpenAI-compatible LLM layer | Native Gemini and OpenAI-compatible request/response shapes differ; mismatch breaks FR1/FR8 regardless of auth success. |
| A4 | Quota/billing behavior for OAuth-authenticated calls is acceptable (free-tier limits, 429 semantics) | Quota exhaustion mid-pipeline would make gates flaky — a determinism violation. |
| A5 | `gemini-pro` (the README's example model name) is a valid current model | Google's model naming churns fast; the docs may already be stale. |
| A6 | temperature 0 on Gemini yields sufficiently stable output for gating decisions | Otherwise scores could flip run-to-run. |
| A7 | Refresh tokens remain valid across days/weeks without re-consent | If Google rotates or invalidates them, "transparent refresh" degrades into periodic forced re-logins. |
| A8 | One cached token per user is the right sharing model on shared machines | A shared cache file creates conflicts and a cross-user credential exposure question. |

---

## 9. Risks & Unknowns

- **Terms-of-service risk:** programmatic API access authenticated with personal-account user credentials may be outside Google's intended usage patterns; worst case is quota throttling or account-level consequences for users. The officially sanctioned server-side mechanisms (e.g., service-account / server-to-server auth) are a different setup story — which one is in scope is an open decision, not a technical detail.
- **New long-lived secret on disk:** the refresh token cache is a credential that outlives any single session. Theft or accidental commit/share of that file grants access to the user's Gemini quota. Its protection level is currently unspecified beyond 0600 writes.
- **CI silent breakage:** if a refresh token expires on a runner with no human present, every subsequent gate run fails auth. Depending on the FR5 policy this becomes either a wall of exit-1 errors or silently rule-only scoring that hides the broken configuration.
- **Air-gap policy conflict:** a Security Officer who set `allow_public_cloud_transmission: false` may view *any* Google transmission as a policy violation; the product needs an explicit stance (FR6) rather than implicit behavior.
- **Scope over-broadness:** the documented `cloud-platform` scope grants far more access than evaluating code requires — a least-privilege concern for security review.
- **Model deprecation:** configured model names can stop existing; without clear errors (FR9), users would debug timeouts instead of a renamed model.
- **Unverified end-to-end flow:** the OAuth implementation has only ever run against mock servers in tests. Real Google endpoints may surface redirects, consent screens, or token-response shapes that were not anticipated.
- **Unknown:** whether the intended audience registers their own OAuth clients (high friction) or the organization ships pre-registered client IDs (centralized secret management question).

---

## 10. Questions to Clarify With the User

1. **Scope:** is the goal *end-to-end Gemini-backed evaluation* (the LLM actually influences scores and remediations — which requires building the currently missing call path), or only *auth support* for Gemini (which mostly exists)?
2. **Headless/CI:** must Gemini work in CI runners without a browser? If yes, which non-interactive mechanism is acceptable to the user (e.g., pre-provisioned token, service-account-style credentials, or "not supported — degrade per policy")?
3. **Endpoint family:** Gemini Developer API (personal account) or Vertex AI (organization project)? These differ in auth model, project requirements, and billing.
4. **Auth failure semantics:** when auth fails or the LLM is unreachable — rule-based fallback (G-2) or hard error (exit 1)? Should this differ between local and CI modes?
5. **Air-gap interaction:** with `allow_public_cloud_transmission: false` and Gemini configured, what should happen — hard block, warning, or silent fallback?
6. **Client registration:** who creates the Google OAuth client — each end user, or an org-level pre-registered client ID shipped in config?
7. **Token cache policy:** acceptable location and sharing model for the on-disk token cache in team/CI environments?

---

*This document analyzes WHAT needs to be solved and WHY. It deliberately contains no implementation proposals; decisions on HOW belong to the design phase after §10 is resolved.*
