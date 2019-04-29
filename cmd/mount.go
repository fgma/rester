package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	if runtime.GOOS != "windows" {
		rootCmd.AddCommand(mountCmd)
	}
}

var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount repostitory",
	Long:  `Mount the given repository using restic to the given mountpoint`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		repositoryName := args[0]
		mountPoint := args[1]

		repository := config.GetRepositoryByName(repositoryName)

		if repository == nil {
			fmt.Fprintf(os.Stderr, "Repository %s is not a configured repository\n", repositoryName)
			os.Exit(1)
		}

		if err := restic.Mount(*repository, mountPoint); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to mount repository %s\n%s\n", repositoryName, err)
			os.Exit(1)
		}

	},
}
