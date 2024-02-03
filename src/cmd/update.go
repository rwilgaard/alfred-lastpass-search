package cmd

import (
	"log"

	aw "github.com/deanishe/awgo"
	"github.com/spf13/cobra"
)

var (
    updateCmd = &cobra.Command{
        Use:          "update",
        Short:        "check for updates",
        SilenceUsage: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            wf.Configure(aw.TextErrors(true))
            log.Println("Checking for updates...")
            if err := wf.CheckForUpdate(); err != nil {
                return err
            }
            return nil
        },
    }
)

func init() {
    rootCmd.AddCommand(updateCmd)
}
