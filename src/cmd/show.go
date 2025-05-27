package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/rwilgaard/alfred-lastpass-search/src/pkg/util"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:          "show",
	Short:        "show entry details",
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		itemID := args[0]

		keys, details, err := ls.GetDetails(itemID)
		if err != nil {
			wf.FatalError(err)
		}

		excluded := []string{
			"id", "name", "fullname", "last_modified_gmt",
			"last_touch", "extra_fields", "folder", "notetype",
			"language", "bit strength", "format", "date",
		}

		redacted := []string{
			"password", "passphrase", "private key",
			"license key", "rootkey", "unsealkey",
		}

		fullname := fmt.Sprintf("%s/%s", os.Getenv("item_folder"), os.Getenv("item_name"))

		wf.NewItem("Go back").
			Icon(util.IconBack).
			Arg("go_back").
			Valid(true)

		wf.NewItem("Name").
			Icon(util.GetIcon("name")).
			Subtitle(fullname).
			Arg(fullname).
			Var("sensitive", "false").
			Valid(true)

		for _, key := range keys {
			value := details[key]
			if slices.Contains(excluded, strings.ToLower(key)) {
				continue
			}
			if value == "" {
				continue
			}
			sub := value
			sensitive := "false"
			if slices.Contains(redacted, strings.ToLower(key)) {
				sub = strings.Repeat("•", 32)
				sensitive = "true"
			}
			if key == "Notes" {
				wf.NewItem(key).
					Icon(util.GetIcon(key)).
					Subtitle("Press ⏎ to show notes").
					Arg("notes").
					Var("sensitive", sensitive).
					Valid(true)
				continue
			}
			wf.NewItem(key).
				Icon(util.GetIcon(key)).
				Subtitle(sub).
				Arg(value).
				Var("sensitive", sensitive).
				Var("field", key).
				Valid(true)
		}

		wf.NewItem("Edit entry").
			Icon(util.IconEdit).
			Arg("edit").
			Valid(true)

		deleteMsg := fmt.Sprintf(`Are you sure you want to delete this entry?
Name: %s
ID: %s`, fullname, itemID)

		wf.NewItem("Delete entry").
			Icon(util.IconDelete).
			Arg("delete").
			Var("msg", deleteMsg).
			Valid(true)

		alfredutils.HandleFeedback(wf)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
