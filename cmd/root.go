package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
