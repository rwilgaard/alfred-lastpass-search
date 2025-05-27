package cmd

import (
	"log"

	aw "github.com/deanishe/awgo"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:          "update",
	Short:        "check for updates",
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		wf.Configure(aw.TextErrors(true))
		log.Println("Checking for updates...")
		return wf.CheckForUpdate()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
