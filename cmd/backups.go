package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(backupsCmd)
}

var backupsCmd = &cobra.Command{
	Use:   "backups",
	Short: "Show configured backups",
	Long:  `Show configured backups`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		fmt.Fprintln(w, "name\tdata\trepositories")
		fmt.Fprintln(w, "----\t----\t----------")

		for _, backup := range config.Backups {
			data := ""
			if backup.DataStdinCommand == "" {
				data = strings.Join(backup.Data, ",")
			} else {
				data = backup.DataStdinCommand
			}
			fmt.Fprintf(
				w, "%s\t%s\t%s\n",
				backup.Name, data, strings.Join(backup.Repositories, ","),
			)
		}

		w.Flush()
	},
}
