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

		ensureBackupsExist(args)

		if len(args) == 0 {
			for _, backup := range config.Backups {
				runBackup(backup.Name)
			}
		} else {
			for _, backupName := range args {
				runBackup(backupName)
			}
		}

	},
}

func runBackup(backupName string) {

	backup := config.GetBackupByName(backupName)

	if backup == nil {
		fmt.Fprintf(os.Stderr, "Backup %s is not a configured backup\n", backupName)
		os.Exit(1)
	}

	repository := config.GetRepositoryByName(backup.Repository)

	if repository == nil {
		fmt.Fprintf(os.Stderr, "Repository %s is not a configured repository\n", backup.Repository)
		os.Exit(1)
	}

	if err := restic.RunBackup(*backup, *repository); err != nil {
		fmt.Fprintf(os.Stderr, "Backup %s failed to run: %s\n", backupName, err.Error())
	}
}

func ensureBackupsExist(backups []string) {
	for _, backupName := range backups {
		isExistingBackup := false
		for _, b := range config.Backups {
			if b.Name == backupName {
				isExistingBackup = true
				break
			}
		}

		if !isExistingBackup {
			fmt.Fprintf(os.Stderr, "%s is not a configured backup\n", backupName)
			os.Exit(1)
		}
	}
}
