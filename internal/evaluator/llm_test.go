package evaluator_test

// Acceptance tests for LLM enhancement of rule-based scores.
//
// Story: As a developer with a configured LLM provider, when an evaluation
// runs, then LLM pillar adjustments and remediations are applied to the
// score and report; when the LLM is unavailable, the rule-based score stands
// unchanged and the run does not fail (Story G-2).

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gatekeeper/internal/config"
	"gatekeeper/internal/evaluator"
	"gatekeeper/pkg/score"
)

// fullScore builds a score with every pillar at its maximum (total 100).
func fullScore() score.Score {
	s := score.NewScore()
	for name, max := range score.PillarMaxPoints {
		s.Pillars[name] = max
	}
	s.Total = 100
	return s
}

// llmTestConfig returns a config pointing the LLM at the mock server.
func llmTestConfig(t *testing.T, baseURL string) config.GatekeeperConfig {
	t.Helper()
	os.Setenv("TEST_LLM_KEY", "test-llm-key")
	t.Cleanup(func() { os.Unsetenv("TEST_LLM_KEY") })

	allow := true
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		BaseURL:      baseURL,
		ModelName:    "gemini-2.0-flash",
		AuthType:     config.AuthAPIKey,
		APIKeyEnvVar: "TEST_LLM_KEY",
		TimeoutMS:    5000,
		Temperature:  0,
	}
	cfg.Gatekeeper.Privacy = &config.PrivacyConfig{AllowPublicCloudTransmission: &allow}
	return cfg
}

func evaluationResponse(staticDeduction, archDeduction float64) string {
	b, _ := json.Marshal(map[string]interface{}{
		"pillar_adjustments": map[string]float64{
			"static_health_deduction": staticDeduction,
			"architecture_deduction":  archDeduction,
		},
		"remediations": []map[string]string{
			{
				"priority":       "HIGH",
				"pillar":         "static",
				"location":       "main.go:lines 3-9",
				"finding":        "Poor function naming",
				"actionable_fix": "Rename to descriptive names",
			},
		},
	})
	return string(b)
}

func newMockLLM(t *testing.T, content string, status int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if status != http.StatusOK {
			http.Error(w, "boom", status)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"created": 0,
			"model":   "mock-model",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": content}, "finish_reason": "stop"},
			},
			"usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func writeSourceFile(t *testing.T, content string) []string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return []string{path}
}

// Given an LLM is configured and responds with pillar adjustments and a remediation,
// when a rule-based score is enhanced,
// then the pillar scores are reduced by the deductions (clamped at 0), the
// remediation appears as a finding, the total is recomputed, and the score is
// marked as LLM-enhanced.
func TestEnhanceWithLLM_AppliesAdjustmentsAndRemediations(t *testing.T) {
	mock := newMockLLM(t, evaluationResponse(5.0, 2.5), http.StatusOK)
	cfg := llmTestConfig(t, mock.URL)
	files := writeSourceFile(t, "package main\n\nfunc main() {}\n")

	s := fullScore()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out := evaluator.EnhanceWithLLM(ctx, cfg, s, files)

	if out.Pillars[score.PillarStatic] != 20-5.0 {
		t.Errorf("static pillar = %v, want %v", out.Pillars[score.PillarStatic], 15.0)
	}
	if out.Pillars[score.PillarArchitecture] != 25-2.5 {
		t.Errorf("architecture pillar = %v, want %v", out.Pillars[score.PillarArchitecture], 22.5)
	}
	if out.Total != 100-5.0-2.5 {
		t.Errorf("total = %v, want %v", out.Total, 92.5)
	}
	if !out.LLMEnhanced {
		t.Error("score should be marked as LLM-enhanced")
	}
	if len(out.Findings) != 1 {
		t.Fatalf("expected 1 remediation finding, got %d", len(out.Findings))
	}
	f := out.Findings[0]
	if f.Priority != "HIGH" || f.Pillar != score.PillarStatic || f.Description != "Poor function naming" || f.Remediation != "Rename to descriptive names" {
		t.Errorf("unexpected finding: %+v", f)
	}
}

// Given the LLM fails on every attempt,
// when a rule-based score is enhanced,
// then the score stands completely unchanged and no error escapes (the run
// must not fail just because the LLM is down — Story G-2).
func TestEnhanceWithLLM_FallsBackToRuleBasedOnFailure(t *testing.T) {
	mock := newMockLLM(t, "", http.StatusInternalServerError)
	cfg := llmTestConfig(t, mock.URL)
	files := writeSourceFile(t, "package main\n\nfunc main() {}\n")

	s := fullScore()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out := evaluator.EnhanceWithLLM(ctx, cfg, s, files)

	if out.Total != 100 {
		t.Errorf("total = %v, want unchanged 100", out.Total)
	}
	for name, max := range score.PillarMaxPoints {
		if out.Pillars[name] != max {
			t.Errorf("pillar %s = %v, want unchanged %v", name, out.Pillars[name], max)
		}
	}
	if len(out.Findings) != 0 {
		t.Errorf("expected no findings from a failed LLM call, got %d", len(out.Findings))
	}
	if out.LLMEnhanced {
		t.Error("score must not be marked LLM-enhanced when the LLM failed")
	}
}

// Given no LLM is configured,
// when a score is enhanced,
// then nothing happens (no network call, no changes).
func TestEnhanceWithLLM_NotConfiguredIsNoOp(t *testing.T) {
	cfg := config.DefaultConfig() // no LLM section
	files := writeSourceFile(t, "package main\n\nfunc main() {}\n")

	s := fullScore()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out := evaluator.EnhanceWithLLM(ctx, cfg, s, files)

	if out.Total != 100 || out.LLMEnhanced || len(out.Findings) != 0 {
		t.Errorf("score changed without a configured LLM: total=%v enhanced=%v findings=%d",
			out.Total, out.LLMEnhanced, len(out.Findings))
	}
}

// Given the LLM returns deductions larger than a pillar's value,
// when adjustments are applied,
// then pillar scores are clamped at 0 (never negative).
func TestEnhanceWithLLM_ClampsDeductionsAtZero(t *testing.T) {
	mock := newMockLLM(t, evaluationResponse(999, -5), http.StatusOK)
	cfg := llmTestConfig(t, mock.URL)
	files := writeSourceFile(t, "package main\n\nfunc main() {}\n")

	s := fullScore()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out := evaluator.EnhanceWithLLM(ctx, cfg, s, files)

	if out.Pillars[score.PillarStatic] != 0 {
		t.Errorf("static pillar = %v, want clamped 0", out.Pillars[score.PillarStatic])
	}
	for name, v := range out.Pillars {
		if v < 0 {
			t.Errorf("pillar %s is negative (%v) — deductions must clamp at 0", name, v)
		}
	}
}

// Given air-gapped mode (allow_public_cloud_transmission = false),
// when a score is enhanced with an LLM configured,
// then no transmission occurs and the rule-based score stands unchanged.
func TestEnhanceWithLLM_AirGappedSkipsTransmission(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": evaluationResponse(5, 0)}, "finish_reason": "stop"},
			},
		})
	}))
	t.Cleanup(srv.Close)

	cfg := llmTestConfig(t, srv.URL)
	disallow := false
	cfg.Gatekeeper.Privacy = &config.PrivacyConfig{AllowPublicCloudTransmission: &disallow}
	files := writeSourceFile(t, "package main\n\nfunc main() {}\n")

	s := fullScore()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out := evaluator.EnhanceWithLLM(ctx, cfg, s, files)

	if requests != 0 {
		t.Errorf("expected 0 network requests in air-gapped mode, got %d", requests)
	}
	if out.Total != 100 || out.LLMEnhanced {
		t.Errorf("score must stand unchanged in air-gapped mode: total=%v enhanced=%v", out.Total, out.LLMEnhanced)
	}
}
