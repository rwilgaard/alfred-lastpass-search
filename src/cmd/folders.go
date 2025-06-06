package cmd

import (
	"github.com/rwilgaard/alfred-lastpass-search/src/pkg/lastpass"
	"github.com/rwilgaard/alfred-lastpass-search/src/pkg/util"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var foldersCmd = &cobra.Command{
	Use:          "folders",
	Short:        "list folders",
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		ls, err := lastpass.NewService("lpass")
		if err != nil {
			wf.FatalError(err)
		}

		folders, err := ls.GetFolders()
		if err != nil {
			wf.FatalError(err)
		}

		wf.NewItem("Select folder").
			Match("*").
			Subtitle("Type to search").
			Valid(false)

		for _, f := range folders {
			wf.NewItem(f.Name).
				UID(f.Name).
				Icon(util.IconFolder).
				Var("folder", f.Name).
				Valid(true)
		}

		wf.Filter(args[0])
		alfredutils.HandleFeedback(wf)
	},
}

func init() {
	rootCmd.AddCommand(foldersCmd)
}
