package cmd

import (
	"github.com/spf13/cobra"
	"github.com/carsonreinke/mailutils/internal"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "TODO",
	Long: `TODO`,

	Run: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func runSearch(_ *cobra.Command, args []string) {
	err := internal.NewSearcher(configuration).Search(args)
	cobra.CheckErr(err)
}
