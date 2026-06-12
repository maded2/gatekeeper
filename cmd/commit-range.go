package cmd

import (
	"fmt"
	"os"

	"gatekeeper/internal/config"
	"gatekeeper/internal/evaluator"
	"gatekeeper/internal/git"
	"gatekeeper/internal/reporter"

	"github.com/spf13/cobra"
)

var (
	rangeValue  string
	rangeFormat string
	rangeOutput string
)

// commitRangeCmd evaluates quality of a specific commit range.
var commitRangeCmd = &cobra.Command{
	Use:   "commit-range",
	Short: "Check quality of a specific commit range",
	Long:  `Evaluates the quality of changes introduced in a specific range of commits.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if rangeValue == "" {
			printError("Error: --range is required\n\nUsage: gatekeeper commit-range --range=HEAD~3..HEAD")
			return nil
		}

		cfgPath := "gatekeeper.json"
		cfg, err := config.Load(cfgPath)
		if err != nil {
			printError(fmt.Sprintf("Error: %v\n\nRun 'gatekeeper init' to create a default configuration.", err))
			return nil
		}

		// Get commits in range
		commits, err := git.GetCommitsInRange(".", rangeValue)
		if err != nil {
			printError(fmt.Sprintf("Error getting commits: %v", err))
			return nil
		}

		if len(commits) == 0 {
			fmt.Println("No commits found in the specified range.")
			os.Exit(evaluator.ExitPass)
			return nil
		}

		// Get changed files
		changed, err := git.GetChangedFilesInRange(".", rangeValue)
		if err != nil {
			printError(fmt.Sprintf("Error getting changed files: %v", err))
			return nil
		}

		if len(changed) == 0 {
			fmt.Println("No file changes in the specified commit range.")
			os.Exit(evaluator.ExitPass)
			return nil
		}

		// Evaluate
		result := evaluator.CheckDiff(*cfg, changed)

		// Output
		switch rangeFormat {
		case "json":
			if rangeOutput != "" {
				if err := reporter.WriteJSON(result, rangeOutput); err != nil {
					printError(fmt.Sprintf("Error writing output: %v", err))
					return nil
				}
			} else {
				reporter.NewJSON(os.Stdout).Print(result)
			}
		case "markdown":
			reporter.NewMarkdown(os.Stdout).Print(result)
		default:
			reporter.NewPretty(os.Stdout).Print(result)
		}

		evalResult := evaluator.Evaluate(*cfg, result.Total)
		os.Exit(evalResult.ExitCode)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(commitRangeCmd)
	commitRangeCmd.Flags().StringVarP(&rangeValue, "range", "r", "", "commit range (e.g., HEAD~3..HEAD)")
	commitRangeCmd.Flags().StringVarP(&rangeFormat, "format", "f", "pretty", "output format: pretty, json, markdown")
	commitRangeCmd.Flags().StringVarP(&rangeOutput, "output", "o", "", "write output to file")
}
