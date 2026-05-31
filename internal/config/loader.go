package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const DefaultFilename = ".gatekeeper.yml"

// LoadFromPath reads a configuration file from the given path.
// Returns default configuration if the file does not exist.
func LoadFromPath(path string) (Configuration, error) {
	exists, err := fileExists(path)
	if err != nil {
		return Configuration{}, fmt.Errorf("checking config %q: %w", path, err)
	}
	if !exists {
		return Default(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Configuration{}, fmt.Errorf("reading config %q: %w", path, err)
	}
	return parseYAML(data), nil
}

// loadFromPathOrError reads config or returns os.ErrNotExist if missing.
// Used internally by LoadFromRoot to detect "not found".
func loadFromPathOrError(path string) (Configuration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Configuration{}, err
	}
	return parseYAML(data), nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// LoadFromRoot walks up from startDir to find .gatekeeper.yml.
// Returns default configuration if no file is found.
func LoadFromRoot(startDir string) (Configuration, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return Default(), nil
	}

	for {
		candidate := filepath.Join(dir, DefaultFilename)
		cfg, err := loadFromPathOrError(candidate)
		if err == nil {
			return cfg, nil
		}
		if !os.IsNotExist(err) {
			return Configuration{}, err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return Default(), nil
		}
		dir = parent
	}
}

// parseYAML implements a minimal YAML parser for our config format.
func parseYAML(data []byte) Configuration {
	cfg := Default()
	lines := strings.Split(string(data), "\n")
	currentSection := ""
	currentSubSection := ""

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Detect indentation level
		indent := len(rawLine) - len(strings.TrimLeft(rawLine, " \t"))

		// Parse key: value
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Strip quotes
		value = strings.Trim(value, "\"'")

		if indent == 0 {
			// Top-level key
			if value == "" {
				currentSection = key
				currentSubSection = ""
			} else {
				setTopLevel(&cfg, key, value)
				currentSection = ""
				currentSubSection = ""
			}
		} else if indent <= 4 {
			// First-level nested key
			if strings.HasPrefix(line, "- ") {
				// List item
				item := strings.Trim(strings.TrimPrefix(line, "- "), "\"'")
				appendList(&cfg, currentSection, item)
			} else {
				if value == "" {
					currentSubSection = key
				} else {
					setNested(&cfg, currentSection, currentSubSection, key, value)
				}
			}
		} else {
			// Deeper nested key
			setNested(&cfg, currentSection, currentSubSection, key, value)
		}
	}

	return cfg
}

func setTopLevel(cfg *Configuration, key, value string) {
	switch key {
	case "threshold":
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			cfg.Threshold = f
		}
	}
}

func appendList(cfg *Configuration, section string, item string) {
	switch section {
	case "exclude":
		cfg.ExcludePatterns = append(cfg.ExcludePatterns, item)
	}
}

func setNested(cfg *Configuration, section, subSection, key, value string) {
	switch section {
	case "pillar_weights":
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			switch key {
			case "code_quality":
				cfg.PillarWeights.CodeQuality = f
			case "test_coverage":
				cfg.PillarWeights.TestCoverage = f
			case "deployability":
				cfg.PillarWeights.Deployability = f
			}
		}
	}
}
