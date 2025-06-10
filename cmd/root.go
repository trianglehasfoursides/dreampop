package cmd

import (
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "dreampop",
	Short: "Your notelist but in terminal",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func Execute() error {
	if err := root.Execute(); err != nil {
		return err
	}
	return nil
}
