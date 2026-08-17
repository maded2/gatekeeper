package evaluator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"gatekeeper/internal/config"
	"gatekeeper/internal/llm"
	"gatekeeper/internal/scanner"
	"gatekeeper/pkg/score"
)

// maxLLMTransmissionBytes caps how much code is sent to the LLM per
// evaluation, keeping requests within provider context windows (spec §5.3).
const maxLLMTransmissionBytes = 24 * 1024

// EnhanceWithLLM applies LLM pillar adjustments and remediations on top of a
// rule-based score. When no LLM is configured or any failure occurs (auth,
// network, air-gap policy), the score is returned unchanged — the run never
// fails just because the LLM is unavailable (Story G-2).
func EnhanceWithLLM(ctx context.Context, cfg config.GatekeeperConfig, s score.Score, files []string) score.Score {
	if !llm.IsConfigured(cfg) {
		return s
	}

	llmCfg, err := llm.FromConfig(cfg)
	if err != nil || llmCfg == nil {
		slog.Warn("llm configuration invalid; using rule-based fallback", slog.Any("err", err))
		return s
	}

	code := composeLLMPayload(files)
	if code == "" {
		return s
	}

	privacy := config.PrivacyConfig{}
	if cfg.Gatekeeper.Privacy != nil {
		privacy = *cfg.Gatekeeper.Privacy
	}

	eval, err := llm.NewClient(llmCfg).Evaluate(ctx, privacy, code)
	if err != nil {
		if errors.Is(err, llm.ErrAirGapped) {
			slog.Debug("llm evaluation skipped: air-gapped mode")
		} else {
			slog.Warn("llm evaluation failed; using rule-based fallback", slog.Any("err", err))
		}
		return s
	}

	s.Pillars[score.PillarStatic] = applyDeduction(s.Pillars[score.PillarStatic], eval.PillarAdjustments.StaticHealthDeduction)
	s.Pillars[score.PillarArchitecture] = applyDeduction(s.Pillars[score.PillarArchitecture], eval.PillarAdjustments.ArchitectureDeduction)

	for _, r := range eval.Remediations {
		s.Findings = append(s.Findings, score.Finding{
			Priority:    r.Priority,
			Pillar:      r.Pillar,
			Location:    r.Location,
			Description: r.Finding,
			Remediation: r.ActionableFix,
		})
	}

	s.Total = 0
	for _, v := range s.Pillars {
		s.Total += v
	}
	s.LLMEnhanced = true
	return s
}

// EnhanceWorkspaceWithLLM is EnhanceWithLLM for whole-workspace checks: it
// discovers the source files to transmit on its own.
func EnhanceWorkspaceWithLLM(ctx context.Context, cfg config.GatekeeperConfig, s score.Score, root string) score.Score {
	sc := scanner.New(cfg.Exclusions.Paths)
	files, _ := sc.Scan(root)

	var sourceFiles []string
	for _, f := range files {
		if isSourceFile(f) {
			sourceFiles = append(sourceFiles, f)
		}
	}
	return EnhanceWithLLM(ctx, cfg, s, sourceFiles)
}

// applyDeduction subtracts a (possibly malformed) LLM deduction from a pillar
// score, clamping the result at 0.
func applyDeduction(value, deduction float64) float64 {
	if deduction < 0 {
		deduction = 0
	}
	v := value - deduction
	if v < 0 {
		return 0
	}
	return v
}

// composeLLMPayload concatenates the contents of the given source files with
// path headers, capped at maxLLMTransmissionBytes total.
func composeLLMPayload(files []string) string {
	var b strings.Builder
	for _, f := range files {
		if !isSourceFile(f) {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		header := fmt.Sprintf("=== %s ===\n", f)
		if b.Len()+len(header)+len(data) > maxLLMTransmissionBytes {
			b.WriteString("\n[truncated: remaining files omitted to fit provider context window]\n")
			break
		}
		b.WriteString(header)
		b.Write(data)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}
