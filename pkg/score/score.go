// Package score defines the public Quality Score types and pillar definitions.
package score

import "time"

// Pillar names used throughout the system.
const (
	PillarStatic       = "static"
	PillarArchitecture = "architecture"
	PillarVerification = "verification"
	PillarSecurity     = "security"
)

// PillarMaxPoints defines the maximum points per pillar.
var PillarMaxPoints = map[string]float64{
	PillarStatic:       20,
	PillarArchitecture: 25,
	PillarVerification: 35,
	PillarSecurity:     20,
}

// Score represents the computed quality score with pillar breakdown.
type Score struct {
	Total        float64            `json:"total"`
	Pillars      map[string]float64 `json:"pillars"`
	Findings     []Finding          `json:"findings"`
	Timestamp    time.Time          `json:"timestamp"`
}

// Finding represents a single quality issue with location and remediation.
type Finding struct {
	Priority       string `json:"priority"`
	Pillar         string `json:"pillar"`
	Location       string `json:"location"`
	LineStart      int    `json:"line_start"`
	LineEnd        int    `json:"line_end"`
	Description    string `json:"description"`
	Remediation    string `json:"remediation"`
	Severity       string `json:"severity,omitempty"`
}

// NewScore creates a Score with all pillars initialized to 0.
func NewScore() Score {
	return Score{
		Pillars:   make(map[string]float64),
		Timestamp: time.Now(),
	}
}
