package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check configured repositories",
	Long:  `Check configured repositories`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {

		ensureRepositoriesExist(args)

		if len(args) == 0 {
			for _, repo := range config.Repositories {
				runCheck(repo.Name)
			}
		} else {
			for _, repoName := range args {
				runCheck(repoName)
			}
		}

	},
}

func runCheck(repoName string) {

	repo := config.GetRepositoryByName(repoName)

	if repo == nil {
		fmt.Fprintf(os.Stderr, "Repository %s is not a configured backup\n", repoName)
		os.Exit(1)
	}

	if err := restic.RunCheck(*repo); err != nil {
		fmt.Fprintf(os.Stderr, "Check %s failed to run: %s\n", repoName, err.Error())
	}
}

func ensureRepositoriesExist(repositories []string) {
	for _, repoName := range repositories {
		isExistingRepo := false
		for _, b := range config.Repositories {
			if b.Name == repoName {
				isExistingRepo = true
				break
			}
		}

		if !isExistingRepo {
			fmt.Fprintf(os.Stderr, "%s is not a configured repository\n", repoName)
			os.Exit(1)
		}
	}
}
