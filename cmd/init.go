package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configured repositories using restic",
	Long:  `Initialize configured repositories using restic`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			for _, repository := range config.Repositories {
				initRepository(repository.Name)
			}
		} else {
			for _, repoName := range args {
				initRepository(repoName)
			}
		}

	},
}

func initRepository(repoName string) {
	repository := config.GetRepositoryByName(repoName)

	if repository == nil {
		fmt.Fprintf(os.Stderr, "Repository %s is not a configured repository\n", repoName)
		os.Exit(1)
	}

	if err := restic.Init(*repository); err != nil {
		fmt.Fprintf(os.Stderr, "Init %s failed to run: %s\n", repoName, err.Error())
	}
}
