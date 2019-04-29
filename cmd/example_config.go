package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fgma/rester/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(exampleConfigCmd)
}

var exampleConfigCmd = &cobra.Command{
	Use:   "example-config",
	Short: "Print an example configuration as a template",
	Long:  `Print an example configuration as a template`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		exampleConfig := `{
  "repositories": [
    {
      "name": "backup-repo",
      "url": "/home/test/backups/backup-repo",
      "password": "a secure password",
      "check": {
        "read_data_percentage": 5
      }
    }
  ],
  "backups": [
    {
      "name": "some data",
      "repository": "backup-repo",
      "data": [
        "/home/testuser/pictures",
        "/home/testuser/data"
      ],
      "exclude": [ "*.tmp", "*.abc" ],
      "one_file_system": true,
      "tags": [ "home", "pictures" ],
      "handler": {
        "before": "notify_send -t 1000 Backup before",
        "after": "notify_send -t 1000 Backup after",
        "success": "notify_send -t 1000 Backup done",
        "failure": "notify_send Backup failed"
      }
    },
    {
      "name": "crontab",
      "repository": "backup-repo",
      "data_stdin_command": "crontab -l",
      "stdin_filename": "crontab.txt",
      "one_file_system": true,            
      "tags": [ "cron" ],
      "age": {
        "warn": "6h",
        "error": "12h"
      }
    }
  ]
}`
		reader := strings.NewReader(exampleConfig)

		_, err := internal.LoadFromReader(reader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse example config: %s", err)
		}

		fmt.Println(exampleConfig)
	},
}
