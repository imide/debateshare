package cmd

import (
	"github.com/imide/debateshare/internal"
	"github.com/spf13/cobra"
)

var (
	migrate bool
)

func init() {
	runCmd.Flags().BoolVar(&migrate, "migrate", true, "enable database migration")

	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run application",
	Run: func(cmd *cobra.Command, args []string) {
		internal.Run(cmd.Context(), migrate)
	},
}
