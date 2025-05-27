package cmd

import (
	"fmt"

	aw "github.com/deanishe/awgo"
	"github.com/rwilgaard/alfred-lastpass-search/src/pkg/util"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

const passwordLengthDefault = 32

var (
	lengthFlag  int
	generateCmd = &cobra.Command{
		Use:          "generate",
		Short:        "generate new password",
		SilenceUsage: true,
		Run: func(_ *cobra.Command, _ []string) {
			pws, err := util.GeneratePassword(lengthFlag, true, cfg.AllowedSymbols)
			if err != nil {
				wf.FatalError(err)
			}
			pwn, err := util.GeneratePassword(lengthFlag, false, "")
			if err != nil {
				wf.FatalError(err)
			}

			sub := fmt.Sprintf("⏎ to copy to clipboard  •  ⌘⏎ to add to LastPass  •  Length: %d", lengthFlag)
			wf.NewItem(pws).
				Subtitle(sub).
				Var("password", pws).
				Arg("copy").
				Valid(true).
				NewModifier(aw.ModCmd).
				Arg("add")

			wf.NewItem(pwn).
				Subtitle(sub+"  •  No symbols").
				Var("password", pwn).
				Arg("copy").
				Valid(true).
				NewModifier(aw.ModCmd).
				Arg("add")

			alfredutils.HandleFeedback(wf)
		},
	}
)

func init() {
	generateCmd.Flags().IntVarP(&lengthFlag, "length", "l", passwordLengthDefault, "length of password to generate")
	rootCmd.AddCommand(generateCmd)
}
