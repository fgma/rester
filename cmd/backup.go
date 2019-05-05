package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(backupCmd)
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Run backups",
	Long:  `Run backups specified on the commandline or all if no backup is specified`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		runForBackupConfigurations(args, runBackup)
	},
}

func runBackup(backupName string, repositoryName string) (int, error) {

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

	if err := restic.RunBackup(*backup, *repository); err != nil {
		fmt.Fprintf(os.Stderr, "Backup %s failed to run: %s\n", backupName, err.Error())
	}

	return 0, nil
}
