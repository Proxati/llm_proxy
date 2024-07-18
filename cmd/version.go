package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proxati/llm_proxy/v2/version"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version of this app to standard output.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
