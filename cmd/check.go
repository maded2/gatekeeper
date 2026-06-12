package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gatekeeper/internal/config"
	"gatekeeper/internal/evaluator"
	"gatekeeper/internal/reporter"

	"github.com/spf13/cobra"
)

var (
	checkPath   string
	checkFormat string
	checkOutput string
)

// checkCmd evaluates all source files in the given path.
var checkCmd = &cobra.Command{
	Use:   "check [path]",
	Short: "Check quality of the entire workspace",
	Long:  `Evaluates all source files in the given path and outputs a quality score with findings.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine target path
		target := checkPath
		if target == "" {
			if len(args) > 0 {
				target = args[0]
			} else {
				target = "."
			}
		}

		// Resolve to absolute path
		target, err := filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
		}

		// Find config file
		cfgPath := filepath.Join(target, "gatekeeper.json")
		cfg, err := config.Load(cfgPath)
		if err != nil {
			printError(fmt.Sprintf("Error: %v\n\nRun 'gatekeeper init' to create a default configuration.", err))
			return nil // exit code handled by printError
		}

		// Validate config
		if err := config.Validate(cfg); err != nil {
			printError(fmt.Sprintf("Invalid configuration: %v", err))
			return nil
		}

		// Run workspace check
		result := evaluator.CheckWorkspace(*cfg, target)

		// Output results
		switch checkFormat {
		case "json":
			if checkOutput != "" {
				if err := reporter.WriteJSON(result, checkOutput); err != nil {
					printError(fmt.Sprintf("Error writing JSON output: %v", err))
					return nil
				}
				fmt.Printf("JSON report written to %s\n", checkOutput)
			} else {
				r := reporter.NewJSON(os.Stdout)
				if err := r.Print(result); err != nil {
					printError(fmt.Sprintf("Error formatting JSON: %v", err))
					return nil
				}
			}
		default:
			r := reporter.NewPretty(os.Stdout)
			if err := r.Print(result); err != nil {
				printError(fmt.Sprintf("Error formatting output: %v", err))
				return nil
			}
		}

		// Determine exit code
		evalResult := evaluator.Evaluate(*cfg, result.Total)
		os.Exit(evalResult.ExitCode)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVarP(&checkPath, "path", "p", "", "path to evaluate (default: current directory)")
	checkCmd.Flags().StringVarP(&checkFormat, "format", "f", "pretty", "output format: pretty, json")
	checkCmd.Flags().StringVarP(&checkOutput, "output", "o", "", "write output to file (for JSON format)")
}
