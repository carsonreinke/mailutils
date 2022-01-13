package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/carsonreinke/mailutils/internal"
)

var (
	cfgFile string
	configuration = &internal.Configuration{}

	rootCmd = &cobra.Command{
		Use: "mailutils",
		Short: "A brief description of your application",
		Long: ``,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mailutils.yaml)")

	// rootCmd.PersistentFlags().StringVar(&configuration.StoragePath, "storage-path", "", "")
	// rootCmd.MarkPersistentFlagRequired("storage-path")

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		cfgFile = filepath.Join(home, ".mailutils.yaml")
	}

	// Parse config file
	contents, err := os.ReadFile(cfgFile)
	cobra.CheckErr(err)

	if err := yaml.Unmarshal(contents, configuration); err != nil {
		cobra.CheckErr(err)
	}
}

