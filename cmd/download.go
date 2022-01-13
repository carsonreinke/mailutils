package cmd

import (
	"github.com/spf13/cobra"
	"github.com/carsonreinke/mailutils/internal"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "A brief description of your command",
	Long: ``,

	Run: run,
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func run(_ *cobra.Command, _ []string) {
	err := internal.NewDownloader(configuration).Download()
	cobra.CheckErr(err)
}
