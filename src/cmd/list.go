package cmd

import (
    "fmt"

    aw "github.com/deanishe/awgo"
    "github.com/rwilgaard/go-alfredutils/alfredutils"
    "github.com/spf13/cobra"
)

var (
    foldersFlag []string
    listCmd     = &cobra.Command{
        Use:          "list",
        Short:        "list entries",
        SilenceUsage: false,
        Args:         cobra.RangeArgs(0, 1),
        Run: func(cmd *cobra.Command, args []string) {
            var query string
            if len(args) > 0 {
                query = args[0]
            }

            entries, err := ls.GetEntries(query, foldersFlag, cfg.FuzzySearch)
            if err != nil {
                wf.FatalError(err)
            }

            for _, e := range entries {
                it := wf.NewItem(e.Name).
                    Subtitle(fmt.Sprintf("%s  •  ID: %s", e.Folder, e.ID)).
                    Match(fmt.Sprintf("%s %s %s %s", e.ID, e.Folder, e.Name, e.URL)).
                    UID(e.ID).
                    Var("item_id", e.ID).
                    Var("item_name", e.Name).
                    Var("item_url", e.URL).
                    Var("item_folder", e.Folder).
                    Var("query", query).
                    Var("action", cfg.ModifierReturn).
                    Valid(ls.CheckValidity(e, cfg.ModifierReturn))

                if ls.CheckValidity(e, cfg.ModifierCtrl) {
                    it.NewModifier(aw.ModCtrl).
                        Subtitle(cfg.ModifierCtrl).
                        Var("action", cfg.ModifierCtrl).
                        Valid(true)
                }

                if ls.CheckValidity(e, cfg.ModifierOpt) {
                    it.NewModifier(aw.ModOpt).
                        Subtitle(cfg.ModifierOpt).
                        Var("action", cfg.ModifierOpt).
                        Valid(true)
                }

                if ls.CheckValidity(e, cfg.ModifierCmd) {
                    it.NewModifier(aw.ModCmd).
                        Subtitle(cfg.ModifierCmd).
                        Var("action", cfg.ModifierCmd).
                        Valid(true)
                }
            }

            if cfg.FuzzySearch && len(query) > 0 {
                wf.Filter(query)
            }

            alfredutils.HandleFeedback(wf)
        },
    }
)

func init() {
    listCmd.Flags().StringSliceVarP(&foldersFlag, "folders", "f", []string{}, "Filter entries by folders")

    rootCmd.AddCommand(listCmd)
}
