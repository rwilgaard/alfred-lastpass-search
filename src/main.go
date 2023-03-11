package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "regexp"
    "strings"

    aw "github.com/deanishe/awgo"
    "github.com/deanishe/awgo/update"
    "golang.org/x/exp/slices"
)

type WorkflowConfig struct {
    LpassBin       string
    ModifierReturn string `env:"modifier_return"`
    ModifierCmd    string `env:"modifier_cmd"`
    ModifierOpt    string `env:"modifier_opt"`
    ModifierCtrl   string `env:"modifier_ctrl"`
}

const (
    repo          = "rwilgaard/alfred-lastpass-search"
    updateJobName = "checkForUpdates"
)

var (
    wf         *aw.Workflow
    cfg        *WorkflowConfig
)

func init() {
    wf = aw.New(
        aw.MaxResults(25),
        update.GitHub(repo),
        aw.SuppressUIDs(true),
    )
}

func reSearch(regex *regexp.Regexp, query string) string {
    if regex.MatchString(query) {
        return regex.FindStringSubmatch(query)[1]
    } else {
        return ""
    }
}

func hasAll(input string, words []string) bool {
    for _, w := range words {
        if strings.Contains(input, w) {
            continue
        }
        return false
    }
    return true
}

func checkValidity(entry LastpassEntry, action string) bool {
    if action == "Copy Password" && entry.Password == "" {
        return false
    } else if action == "Copy Username" && entry.Username == "" {
        return false
    }
    return true
}

func run() {
    if err := cli.Parse(wf.Args()); err != nil {
        wf.FatalError(err)
    }
    opts.Query = cli.Arg(0)

    if opts.Update {
        wf.Configure(aw.TextErrors(true))
        log.Println("Checking for updates...")
        if err := wf.CheckForUpdate(); err != nil {
            wf.FatalError(err)
        }
        return
    }

    if wf.UpdateCheckDue() && !wf.IsRunning(updateJobName) {
        log.Println("Running update check in background...")
        cmd := exec.Command(os.Args[0], "-update")
        if err := wf.RunInBackground(updateJobName, cmd); err != nil {
            log.Printf("Error starting update check: %s", err)
        }
    }

    if wf.UpdateAvailable() {
        wf.NewItem("Update Available!").
            Subtitle("Press ⏎ to install").
            Autocomplete("workflow:update").
            Valid(false).
            Icon(aw.IconInfo)
    }

    cfg = &WorkflowConfig{
        LpassBin: "lpass",
    }
    if err := wf.Config.To(cfg); err != nil {
        panic(err)
    }

    if !isLoggedIn() {
        wf.NewItem("You're not logged in to Lastpass.").
            Subtitle("Press ⏎ to login.").
            Arg("auth").
            Valid(true)
        wf.SendFeedback()
        return
    }

    if opts.ListFolders {
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

        wf.Filter(opts.Query)
        wf.SendFeedback()
        return
    }

    if opts.Details {
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

        wf.NewItem("Go back").
            Icon(iconBack).
            Arg("go_back").
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
                    Var("notes", value).
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

        fullname := fmt.Sprintf("%s/%s", os.Getenv("item_folder"), os.Getenv("item_name"))
        deleteMsg := fmt.Sprintf(`Are you sure you want to delete this entry?
Name: %s
ID: %s`, fullname, os.Getenv("item_id"))

        wf.NewItem("Delete entry").
            Icon(iconDelete).
            Arg("delete").
            Var("msg", deleteMsg).
            Valid(true)

        wf.SendFeedback()
        return
    }

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

    if wf.IsEmpty() {
        wf.NewItem("No results found...").
            Subtitle("Try a different query?").
            Icon(aw.IconInfo)
    }
    wf.SendFeedback()
}

func main() {
    wf.Run(run)
}
