package cmd

import (
	"github.com/spf13/cobra"
	"github.com/carsonreinke/mailutils/internal"
)

// printCmd represents the print command
var printCmd = &cobra.Command{
	Use:   "print",
	Short: "TODO",
	Long: `TODO`,
	Run: runPrint,
}

func init() {
	rootCmd.AddCommand(printCmd)
}

func runPrint(_ *cobra.Command, args []string) {
	err := internal.NewPrinter(configuration).Print(args[0])
	cobra.CheckErr(err)
}
