package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gatekeeper/internal/config"

	"github.com/spf13/cobra"
)

// initCmd creates a default gatekeeper.json configuration file.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a gatekeeper.json configuration file",
	Long:  `Creates a default gatekeeper.json configuration file in the current directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		path := filepath.Join(dir, "gatekeeper.json")

		// Check if file already exists
		if _, err := os.Stat(path); err == nil {
			fmt.Fprintf(os.Stderr, "gatekeeper.json already exists at %s\n", path)
			return nil
		}

		if err := config.GenerateDefault(path); err != nil {
			return fmt.Errorf("generate default config: %w", err)
		}

		fmt.Printf("Created gatekeeper.json at %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
