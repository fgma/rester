package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(ageCmd)
}

var ageCmd = &cobra.Command{
	Use:   "age",
	Short: "Show age of each backup",
	Long:  `Show age of each backup`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		fmt.Fprintln(w, "name\tdata\trepository\tage")
		fmt.Fprintln(w, "----\t----\t----------\t---")

		now := time.Now()

		for _, backup := range config.Backups {
			data := ""
			if backup.DataStdinCommand == "" {
				data = strings.Join(backup.Data, ",")
			} else {
				data = backup.DataStdinCommand
			}

			for _, repo := range backup.Repositories {
				repository := config.GetRepositoryByName(repo)

				if repository == nil {
					fmt.Fprintf(os.Stderr, "Repository %s is not a configured repository\n", repo)
					os.Exit(1)
				}

				lastBackupTimestamp, err := restic.GetLastBackupTimestamp(backup, *repository)

				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to get age for backup %s: %s\n", backup.Name, err)
					os.Exit(1)
				}

				age := "-"
				if (lastBackupTimestamp != time.Time{}) {
					age = now.Sub(lastBackupTimestamp).String()
				}

				fmt.Fprintf(
					w, "%s\t%s\t%s\t%s\n",
					backup.Name, data, repository.Name, age,
				)
			}
		}

		w.Flush()
	},
}
