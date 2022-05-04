/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/carsonreinke/mailutils/internal"
	"github.com/spf13/cobra"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "TODO",
	Long:  `TODO`,
	Run:   runReport,
}

func init() {
	rootCmd.AddCommand(reportCmd)
}

func runReport(_ *cobra.Command, args []string) {
	err := internal.NewReporter(configuration).Report(args[0])
	cobra.CheckErr(err)
}
