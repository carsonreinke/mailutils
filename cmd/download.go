package cmd

import (
	"github.com/spf13/cobra"
	"github.com/carsonreinke/mailutils/internal"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "TODO",
	Long: `TODO`,

	Run: runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}

func runDownload(_ *cobra.Command, _ []string) {
	err := internal.NewDownloader(configuration).Download()
	cobra.CheckErr(err)
}
