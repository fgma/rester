package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkAgeCmd)
}

var checkAgeCmd = &cobra.Command{
	Use:   "check-age",
	Short: "Check age of the given backups",
	Long:  `Check age of the given backups`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {

		ensureBackupsExist(args)

		var error error = nil
		var checkExitCode int = 0

		if len(args) == 0 {
			for _, backup := range config.Backups {
				for _, repo := range backup.Repositories {
					exitCode, err := runCheckAge(backup.Name, repo)
					if error == nil {
						error = err
					}
					if exitCode > checkExitCode {
						checkExitCode = exitCode
					}
				}
			}
		} else {
			for _, backupName := range args {
				backup := config.GetBackupByName(backupName)

				if backup == nil {
					fmt.Fprintf(os.Stderr, "Backup %s is not a configured backup\n", backupName)
					os.Exit(1)
				}

				for _, repo := range backup.Repositories {
					exitCode, err := runCheckAge(backup.Name, repo)
					if error == nil {
						error = err
					}
					if exitCode > checkExitCode {
						checkExitCode = exitCode
					}
				}
			}
		}

		os.Exit(checkExitCode)
	},
}

func runCheckAge(backupName string, repositoryName string) (int, error) {

	backup := config.GetBackupByName(backupName)

	if backup == nil {
		fmt.Fprintf(os.Stderr, "Backup %s is not a configured backup\n", backupName)
		os.Exit(1)
	}

	repository := config.GetRepositoryByName(repositoryName)

	if repository == nil {
		fmt.Fprintf(os.Stderr, "Repository %s is not a configured repository\n", repositoryName)
		os.Exit(1)
	}

	limitWarn, limitError, err := restic.CheckAge(*backup, *repository)
	exitCode := 0

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking age for backup %s to repository %s\n", backup.Name, repository.Name)
		exitCode = 1
	} else if limitError {
		fmt.Fprintf(os.Stderr, "Error limit reached for backup %s to repository %s\n", backup.Name, repository.Name)
		exitCode = 3
	} else if limitWarn {
		fmt.Fprintf(os.Stdout, "Warning limit reached for backup %s to repository %s\n", backup.Name, repository.Name)
		exitCode = 2
	}

	return exitCode, err
}
