package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the root command for gatekeeper.
var rootCmd = &cobra.Command{
	Use:   "gatekeeper",
	Short: "Gatekeeper — Automated Quality Score Tool",
	Long: `Gatekeeper evaluates git commits, pull requests, or full codebases
against an organizational Quality Score Standard using static analysis
and LLM-driven reasoning.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// printError prints an error message to stderr and exits with code 1.
func printError(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
