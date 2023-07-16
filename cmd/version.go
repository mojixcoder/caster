package cmd

import (
	"fmt"

	"github.com/mojixcoder/caster/internal/app"
	"github.com/spf13/cobra"
)

// versionCmd is the root command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version of Caster.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Caster %s\n", app.Version)
	},
}
