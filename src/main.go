package main

import (
    "flag"
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

type LastpassFolder struct {
    Name string
}

type LastpassEntry struct {
    ID       string
    Name     string
    Folder   string
    URL      string
    Username string
    Password string
}

const (
    repo          = "rwilgaard/alfred-lastpass-search"
    updateJobName = "checkForUpdates"
)

var (
    wf              *aw.Workflow
    searchFlag      string
    detailsFlag     string
    foldersFlag     string
    updateFlag      bool
    listFoldersFlag bool
    cfg             *WorkflowConfig
)

func init() {
    wf = aw.New(aw.MaxResults(25), update.GitHub(repo), aw.SuppressUIDs(true))
    flag.StringVar(&searchFlag, "search", "", "search entries")
    flag.StringVar(&detailsFlag, "details", "", "item details")
    flag.StringVar(&foldersFlag, "folders", "", "only search in specified folders")
    flag.BoolVar(&updateFlag, "update", false, "check for updates")
    flag.BoolVar(&listFoldersFlag, "listfolders", false, "list all folders")
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

func isLoggedIn() bool {
    cmd := exec.Command(cfg.LpassBin, "status", "--quiet")
    if err := cmd.Run(); err != nil {
        return false
    }
    return true
}

func getFolders() ([]LastpassFolder, error) {
    cmd := cfg.LpassBin + " ls --format %aN,%al --sync=no | grep ',http://group$' | cut -d, -f1 | sort -u"
    out, err := exec.Command("bash", "-c", cmd).Output()
    if err != nil {
        return nil, err
    }

    var folders []LastpassFolder
    for _, f := range strings.Split(string(out), "\n") {
        lf := LastpassFolder{
            Name: f,
        }
        folders = append(folders, lf)
    }

    return folders, nil
}

func getEntries(query string, folder string) []LastpassEntry {
    cmd := exec.Command(cfg.LpassBin, "ls", "--format", "%aN [id: %ai] [url: %al] [username: %au] %ap", "--sync=no", folder)
    out, err := cmd.Output()

    if err != nil {
        panic(err)
    }

    idRegex := regexp.MustCompile(`\[id: ([0-9]+)\]`)
    urlRegex := regexp.MustCompile(`\[url: (.+?)\]`)
    nameRegex := regexp.MustCompile(`^.*?\/(.*?)\s\[`)
    folderRegex := regexp.MustCompile(`^(.*?)\/`)
    usernameRegex := regexp.MustCompile(`\[username: (.+?)\]`)
    passwordRegex := regexp.MustCompile(`.*\] (.*)$`)

    var entries []LastpassEntry
    for _, l := range strings.Split(string(out), "\n") {
        id := reSearch(idRegex, l)
        name := reSearch(nameRegex, l)
        folder := reSearch(folderRegex, l)
        url := reSearch(urlRegex, l)
        username := reSearch(usernameRegex, l)
        password := reSearch(passwordRegex, l)
        e := fmt.Sprintf("%s %s %s %s %s", id, name, folder, url, username)
        if url == "http://group" {
            continue
        }
        if !hasAll(strings.ToLower(e), strings.Split(strings.ToLower(query), " ")) {
            continue
        }
        entries = append(entries, LastpassEntry{
            ID:       id,
            Name:     name,
            Folder:   folder,
            URL:      url,
            Username: username,
            Password: password,
        })
    }
    return entries
}

func getDetails(itemID string) ([]string, map[string]string) {
    cmd := exec.Command(cfg.LpassBin, "show", itemID, "--sync=no")
    out, err := cmd.Output()

    if err != nil {
        panic(err)
    }

    keyRegex := regexp.MustCompile(`^(\S.+?):`)
    valRegex := regexp.MustCompile(`^\S.+?: (.*)`)
    keys := []string{}
    details := make(map[string]string)
    for i, l := range strings.Split(string(out), "\n") {
        if i == 0 {
            continue
        }
        key := reSearch(keyRegex, l)
        val := reSearch(valRegex, l)
        keys = append(keys, key)
        details[key] = val
        if key == "Notes" {
            break
        }
    }
    return keys, details
}

func run() {
    wf.Args()
    flag.Parse()

    wf.Configure(aw.SuppressUIDs(true))

    if updateFlag {
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

    backIcon := aw.Icon{Value: fmt.Sprintf("%s/icons/go_back.png", wf.Dir())}

    if listFoldersFlag {
        wf.Configure(aw.SuppressUIDs(true))
        folderIcon := aw.Icon{Value: fmt.Sprintf("%s/icons/group.png", wf.Dir())}
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
                Icon(&folderIcon).
                Var("folder", f.Name).
                Valid(true)
        }

        wf.Filter(searchFlag)
        wf.SendFeedback()
        return
    }

    if detailsFlag != "" {
        keys, details := getDetails(detailsFlag)
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
            Icon(&backIcon).
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
            if slices.Contains(redacted, strings.ToLower(key)) {
                sub = strings.Repeat("•", 32)
            }
            if key == "Notes" {
                wf.NewItem(key).
                    Subtitle("Press ⏎ to show notes").
                    Arg("notes").
                    Var("notes", value).
                    Valid(true)
                continue
            }
            wf.NewItem(key).
                Subtitle(sub).
                Arg(value).
                Valid(true)
        }

        wf.NewItem("Edit entry").
            Arg("edit").
            Valid(true)

        fullname := fmt.Sprintf("%s/%s", os.Getenv("item_folder"), os.Getenv("item_name"))
        deleteMsg := fmt.Sprintf(`Are you sure you want to delete this entry?
Name: %s
ID: %s`, fullname, os.Getenv("item_id"))

        wf.NewItem("Delete entry").
            Arg("delete").
            Var("msg", deleteMsg).
            Valid(true)

        wf.SendFeedback()
        return
    }

    var entries []LastpassEntry
    for _, folder := range strings.Split(foldersFlag, ",") {
        entries = append(entries, getEntries(searchFlag, strings.TrimSpace(folder))...)
    }
    for _, e := range entries {
        icon := aw.Icon{Value: fmt.Sprintf("%s/icons/password.png", wf.Dir())}
        if e.URL == "http://sn" {
            icon = aw.Icon{Value: fmt.Sprintf("%s/icons/sn.png", wf.Dir())}
        }
        it := wf.NewItem(e.Name).
            Subtitle(fmt.Sprintf("%s  •  ID: %s", e.Folder, e.ID)).
            Icon(&icon).
            Var("item_id", e.ID).
            Var("item_name", e.Name).
            Var("item_url", e.URL).
            Var("item_folder", e.Folder).
            Var("query", searchFlag).
            Var("action", cfg.ModifierReturn).
            Valid(true)

        it.NewModifier(aw.ModCtrl).
            Subtitle(cfg.ModifierCtrl).
            Var("action", cfg.ModifierCtrl).
            Valid(true)

        it.NewModifier(aw.ModOpt).
            Subtitle(cfg.ModifierOpt).
            Var("action", cfg.ModifierOpt).
            Valid(true)

        it.NewModifier(aw.ModCmd).
            Subtitle(cfg.ModifierCmd).
            Var("action", cfg.ModifierCmd).
            Valid(true)
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
