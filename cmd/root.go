package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fgma/rester/internal"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var cfgFile string
var cfgXdgDefault = ".config/"
var cfgFileDefault = "rester/config.json"

var config internal.Config
var restic internal.Restic

var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: os.Args[0] + " is a wrapper around restic",
	Long:  `A wrapper around restic for configuring and running backups`,
	Args:  cobra.NoArgs,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(
		&cfgFile, "config", "c", "",
		fmt.Sprintf("config file (default is $HOME/%s)", cfgXdgDefault+cfgFileDefault),
	)
}

func initConfig() {

	if cfgFile == "" {

		if configHome, isDefined := os.LookupEnv("XDG_CONFIG_HOME"); isDefined {
			cfgFile = filepath.Join(configHome, cfgXdgDefault, cfgFileDefault)
		} else {

			homedir, err := homedir.Dir()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get homedir: %s\n", err)
				os.Exit(1)
			}

			cfgFile = filepath.Join(homedir, cfgXdgDefault, cfgFileDefault)
		}

	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(cfgFile)
		if err == nil {
			mode := info.Mode()
			if mode&0x7 != 0 {
				fmt.Fprintf(os.Stderr,
					"Config file permissions allow access for other than user or group. "+
						"This is insecure. Please restrict file permissions.\n",
				)
				os.Exit(1)
			}
		}
	}

	var err error
	if config, err = internal.Load(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: "+err.Error())
		os.Exit(1)
	}

	restic = internal.NewRestic(config.ResticExecutable)

	if !restic.IsResticAvailable() {
		fmt.Fprintf(os.Stderr, "Restic command is not available")
		os.Exit(1)
	}
}

func runForBackupConfigurations(
	configurationsToRun []string,
	handler func(backupName string, repoName string) (returnCode int, err error),
) {

	type Configuration struct {
		backupName string
		repoName   string
	}

	var configsToRun []Configuration

	if len(configurationsToRun) == 0 {
		// if args are empty run all configurations
		for _, backup := range config.Backups {
			for _, repo := range backup.Repositories {
				configsToRun = append(configsToRun, Configuration{backup.Name, repo})
			}
		}
	} else {
		// otherwise run given configurations
		for _, configurationName := range configurationsToRun {

			split := strings.Split(configurationName, "/")

			if len(split) == 1 {
				// no specific repository given, run against all repositories
				backupName := split[0]
				backup := config.GetBackupByName(backupName)

				if backup == nil {
					fmt.Fprintf(os.Stderr, "Backup %s is not a configured backup\n", backupName)
					os.Exit(1)
				}

				for _, repo := range backup.Repositories {
					configsToRun = append(configsToRun, Configuration{backupName, repo})
				}

			} else if len(split) == 2 {
				// specific repository given, just run against this repository
				backupName := split[0]
				repoName := split[1]
				backup := config.GetBackupByName(backupName)
				repo := config.GetRepositoryByName(repoName)

				if backup == nil {
					fmt.Fprintf(os.Stderr, "Backup %s is not a configured backup\n", backupName)
					os.Exit(1)
				}

				if repo == nil {
					fmt.Fprintf(os.Stderr, "Repository %s is not a configured repository\n", repoName)
					os.Exit(1)
				}

				if !internal.Contains(backup.Repositories, repo.Name) {
					fmt.Fprintf(os.Stderr, "Repository %s is not a configured for backup %s\n", repo.Name, backup.Name)
					os.Exit(1)
				}

				configsToRun = append(configsToRun, Configuration{backup.Name, repo.Name})

			} else {
				fmt.Fprintf(os.Stderr, "Configuration %s is invalid\n", configurationName)
				os.Exit(1)
			}

		}
	}

	finalExitCode := 0

	for _, cfg := range configsToRun {
		exitCode, err := handler(cfg.backupName, cfg.repoName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		}
		if exitCode > finalExitCode {
			finalExitCode = exitCode
		}
	}

	os.Exit(finalExitCode)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
