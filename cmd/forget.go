package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(forgetCmd)
}

var forgetCmd = &cobra.Command{
	Use:   "forget",
	Short: "Forget backups in repositories according to policy",
	Long:  `Forget backups in repositories according to policy`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {

		ensureRepositoriesExist(args)

		if len(args) == 0 {
			for _, repo := range config.Repositories {
				runForget(repo.Name)
			}
		} else {
			for _, repoName := range args {
				runForget(repoName)
			}
		}

	},
}

func runForget(repoName string) {

	repo := config.GetRepositoryByName(repoName)

	if repo == nil {
		fmt.Fprintf(os.Stderr, "Repository %s is not a configured backup\n", repoName)
		os.Exit(1)
	}

	if err := restic.RunForget(*repo); err != nil {
		fmt.Fprintf(os.Stderr, "Forget %s failed to run: %s\n", repoName, err.Error())
	}
}
