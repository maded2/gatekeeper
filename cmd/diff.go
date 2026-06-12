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
	diffBase    string
	diffTarget  string
	diffFormat  string
	diffOutput  string
)

// diffCmd evaluates quality changes between two branches.
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Check quality of changes between two branches",
	Long:  `Evaluates the quality impact of differences between a base and target branch.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if diffBase == "" || diffTarget == "" {
			printError("Error: --base and --target are required\n\nUsage: gatekeeper diff --base=main --target=feature")
			return nil
		}

		// Find config
		cfgPath := "gatekeeper.json"
		cfg, err := config.Load(cfgPath)
		if err != nil {
			printError(fmt.Sprintf("Error: %v\n\nRun 'gatekeeper init' to create a default configuration.", err))
			return nil
		}

		// Get changed files
		changed, err := git.GetChangedFiles(".", diffBase, diffTarget)
		if err != nil {
			printError(fmt.Sprintf("Error getting diff: %v", err))
			return nil
		}

		if len(changed) == 0 {
			fmt.Println("No changes detected between branches.")
			os.Exit(evaluator.ExitPass)
			return nil
		}

		// Check for trivial changes
		allTrivial := true
		for _, f := range changed {
			diff, _ := git.GetFileDiff(".", diffBase, diffTarget, f)
			if !evaluator.IsTrivialChange(f, diff) {
				allTrivial = false
				break
			}
		}

		if allTrivial {
			fmt.Println("Changes are trivial (non-code or whitespace-only) — skipping analysis.")
			os.Exit(evaluator.ExitPass)
			return nil
		}

		// Evaluate changed files
		result := evaluator.CheckDiff(*cfg, changed)

		// Output
		switch diffFormat {
		case "json":
			if diffOutput != "" {
				if err := reporter.WriteJSON(result, diffOutput); err != nil {
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

		// Exit code
		evalResult := evaluator.Evaluate(*cfg, result.Total)
		os.Exit(evalResult.ExitCode)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().StringVarP(&diffBase, "base", "b", "", "base branch (required)")
	diffCmd.Flags().StringVarP(&diffTarget, "target", "t", "", "target branch (required)")
	diffCmd.Flags().StringVarP(&diffFormat, "format", "f", "pretty", "output format: pretty, json, markdown")
	diffCmd.Flags().StringVarP(&diffOutput, "output", "o", "", "write output to file")
}
