package cmd

import (
	"github.com/carsonreinke/mailutils/internal"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "TODO",
	Long:  `TODO`,
	Run:   runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(_ *cobra.Command, args []string) {
	err := internal.NewRemover(configuration).Remove(args)
	cobra.CheckErr(err)
}
