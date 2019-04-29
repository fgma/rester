package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reposCmd)
}

var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "List configured repositories",
	Long:  `List configured repositories`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		fmt.Fprintln(w, "name\tURL")
		fmt.Fprintln(w, "----\t---")

		for _, repo := range config.Repositories {
			fmt.Fprintf(w, "%s\t%s\n", repo.Name, repo.URL)
		}

		w.Flush()

	},
}
