package config

import "fmt"

// ValidationError describes a specific configuration problem.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("invalid config field %q: %s", e.Field, e.Message)
}

// Validate checks the configuration for errors.
// Returns nil if valid, or a slice of ValidationErrors.
func Validate(cfg Configuration) []error {
	var errors []error

	// Validate threshold range
	if cfg.Threshold < 1.0 || cfg.Threshold > 10.0 {
		errors = append(errors, ValidationError{
			Field:   "threshold",
			Message: fmt.Sprintf("must be between 1 and 10, got %f", cfg.Threshold),
		})
	}

	// Validate pillar weights
	validatePillarWeights(&cfg.PillarWeights, &errors)

	return errors
}

func validatePillarWeights(w *PillarWeights, errors *[]error) {
	if w.CodeQuality < 0 || w.CodeQuality > 1 {
		*errors = append(*errors, ValidationError{
			Field:   "pillar_weights.code_quality",
			Message: fmt.Sprintf("must be between 0 and 1, got %f", w.CodeQuality),
		})
	}
	if w.TestCoverage < 0 || w.TestCoverage > 1 {
		*errors = append(*errors, ValidationError{
			Field:   "pillar_weights.test_coverage",
			Message: fmt.Sprintf("must be between 0 and 1, got %f", w.TestCoverage),
		})
	}
	if w.Deployability < 0 || w.Deployability > 1 {
		*errors = append(*errors, ValidationError{
			Field:   "pillar_weights.deployability",
			Message: fmt.Sprintf("must be between 0 and 1, got %f", w.Deployability),
		})
	}

	sum := w.CodeQuality + w.TestCoverage + w.Deployability
	if sum < 0.99 || sum > 1.01 {
		*errors = append(*errors, ValidationError{
			Field:   "pillar_weights",
			Message: fmt.Sprintf("weights must sum to 1.0, got %f", sum),
		})
	}
}
