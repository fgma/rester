package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionRevision = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number"`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("rester v0.1.0 -- " + versionRevision)
	},
}
