package main

import (
	"fmt"
	"os"
	"strings"

	aw "github.com/deanishe/awgo"
	"golang.org/x/exp/slices"
)

func runGenerate() {
    pws, err := generatePassword(opts.Length, true)
    if err != nil {
        wf.FatalError(err)
    }
    pwn, err := generatePassword(opts.Length, false)
    if err != nil {
        wf.FatalError(err)
    }

    sub := fmt.Sprintf("⏎ to copy to clipboard  •  ⌘⏎ to add to LastPass  •  Length: %d", opts.Length)
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
}

func runListFolders() {
    folders, err := getFolders()
    if err != nil {
        wf.FatalError(err)
    }

    wf.NewItem("Select folder").
        Match("*").
        Subtitle("Type to search").
        Valid(false)

    for _, f := range folders {
        wf.NewItem(f.Name).
            Icon(iconFolder).
            Var("folder", f.Name).
            Valid(true)
    }
}

func runSearch() {
    var entries []LastpassEntry
    for _, folder := range strings.Split(opts.Folders, ",") {
        l, err := getEntries(opts.Query, strings.TrimSpace(folder))
        if err != nil {
            wf.FatalError(err)
        }
        entries = append(entries, l...)
    }
    for _, e := range entries {
        icon := iconPW
        if e.URL == "http://sn" {
            icon = iconSN
        }

        it := wf.NewItem(e.Name).
            Subtitle(fmt.Sprintf("%s  •  ID: %s", e.Folder, e.ID)).
            Match(fmt.Sprintf("%s %s %s %s", e.ID, e.Folder, e.Name, e.URL)).
            Icon(icon).
            Var("item_id", e.ID).
            Var("item_name", e.Name).
            Var("item_url", e.URL).
            Var("item_folder", e.Folder).
            Var("query", opts.Query).
            Var("action", cfg.ModifierReturn).
            Valid(checkValidity(e, cfg.ModifierReturn))

        if checkValidity(e, cfg.ModifierCtrl) {
            it.NewModifier(aw.ModCtrl).
                Subtitle(cfg.ModifierCtrl).
                Var("action", cfg.ModifierCtrl).
                Valid(true)
        }

        if checkValidity(e, cfg.ModifierOpt) {
            it.NewModifier(aw.ModOpt).
                Subtitle(cfg.ModifierOpt).
                Var("action", cfg.ModifierOpt).
                Valid(true)
        }

        if checkValidity(e, cfg.ModifierCmd) {
            it.NewModifier(aw.ModCmd).
                Subtitle(cfg.ModifierCmd).
                Var("action", cfg.ModifierCmd).
                Valid(true)
        }
    }
}

func runDetails() {
    keys, details, err := getDetails(opts.Query)
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
        Icon(iconBack).
        Arg("go_back").
        Valid(true)

    wf.NewItem("Name").
        Icon(getIcon("name")).
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
                Icon(getIcon(key)).
                Subtitle("Press ⏎ to show notes").
                Arg("notes").
                Var("sensitive", sensitive).
                Valid(true)
            continue
        }
        wf.NewItem(key).
            Icon(getIcon(key)).
            Subtitle(sub).
            Arg(value).
            Var("sensitive", sensitive).
            Var("field", key).
            Valid(true)
    }

    wf.NewItem("Edit entry").
        Icon(iconEdit).
        Arg("edit").
        Valid(true)

    deleteMsg := fmt.Sprintf(`Are you sure you want to delete this entry?
Name: %s
ID: %s`, fullname, os.Getenv("item_id"))

    wf.NewItem("Delete entry").
        Icon(iconDelete).
        Arg("delete").
        Var("msg", deleteMsg).
        Valid(true)
}
