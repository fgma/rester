package cmd

import (
	"fmt"
	"os"

	"github.com/riywo/loginshell"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(shellCmd)
}

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start interative shell prepared with restic environment variables",
	Long:  `Start interative shell prepared with restic environment variables`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		repositoryName := args[0]
		repository := config.GetRepositoryByName(repositoryName)

		if repository == nil {
			fmt.Fprintf(os.Stderr, "Repository %s is not a configured repository\n", repositoryName)
			os.Exit(1)
		}

		shell, err := loginshell.Shell()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get shell: %s\n", err)
			os.Exit(1)
		}

		c := restic.PrepareResticEnvironmentCommand(shell, repository.URL, repository.Password, repository.Environment, 0, 0, []string{})
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err = c.Run()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to execute shell: %s\n", err)
			os.Exit(1)
		}

	},
}
