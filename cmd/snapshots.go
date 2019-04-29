package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(snapshotsCmd)
}

var snapshotsCmd = &cobra.Command{
	Use:   "snapshots",
	Short: "List snapshots",
	Long:  `List snapshots specified for repositories specified on the commandline`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			for _, repository := range config.Repositories {
				printSnapshotsForRepository(repository.Name)
			}
		} else {
			for _, repositoryName := range args {
				printSnapshotsForRepository(repositoryName)
			}
		}

	},
}

func printSnapshotsForRepository(repositoryName string) {
	repository := config.GetRepositoryByName(repositoryName)

	if repository == nil {
		fmt.Fprintf(os.Stderr, "Repository %s is not a configured repository\n", repositoryName)
		os.Exit(1)
	}

	if err := restic.PrintSnapshots(*repository); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get snapshots for repository %s\n%s\n", repositoryName, err)
		os.Exit(1)
	}
}
