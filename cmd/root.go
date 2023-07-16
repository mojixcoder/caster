package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd is the root command.
var rootCmd = &cobra.Command{
	Use:   "caster",
	Short: "Caster is a cache database.",
	Run:   root,
}

// root is the function that root command runs.
func root(cmd *cobra.Command, args []string) {
	fmt.Println("Run `caster --help` to see more information.")
}

// init adds other commands and flags.
func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringP("config", "c", "/etc/caster", "Path to config.yaml file.")
	viper.BindPFlag("config", runCmd.Flags().Lookup("config"))
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("executing command failed, reason: %s", err.Error())
		os.Exit(1)
	}
}
