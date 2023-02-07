package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	aw "github.com/deanishe/awgo"
	"golang.org/x/exp/slices"
)

type workflowConfig struct {
    LpassBin       string
    PrivateFolders string `env:"PRIVATE_FOLDERS"`
}

type lastpassEntry struct {
    ID       string
    Name     string
    Folder   string
    URL      string
    Username string
    Password string
}

var (
    wf          *aw.Workflow
    searchFlag  string
    detailsFlag string
    privateFlag bool
    cfg         *workflowConfig
)

func init() {
    wf = aw.New(aw.MaxResults(25), aw.SuppressUIDs(true))
    flag.StringVar(&searchFlag, "search", "", "search entries")
    flag.StringVar(&detailsFlag, "details", "", "item details")
    flag.BoolVar(&privateFlag, "private", false, "only search in private folders")
}

func reSearch(regex *regexp.Regexp, query string) string {
    if regex.MatchString(query) {
        return regex.FindStringSubmatch(query)[1]
    } else {
        return ""
    }
}

func isLoggedIn() bool {
    cmd := exec.Command(cfg.LpassBin, "status", "--quiet")
    if err := cmd.Run(); err != nil {
        return false
    }
    return true
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

func getEntries(query string, folder string) []lastpassEntry {
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

    var entries []lastpassEntry
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
        entries = append(entries, lastpassEntry{
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

func getDetails(itemID string) map[string]string {
    cmd := exec.Command(cfg.LpassBin, "show", itemID, "--sync=no")
    out, err := cmd.Output()

    if err != nil {
        panic(err)
    }

    keyRegex := regexp.MustCompile(`^(.+?):`)
    valRegex := regexp.MustCompile(`^.+?: (.*)`)
    details := make(map[string]string)
    for i, l := range strings.Split(string(out), "\n") {
        if i == 0 {
            continue
        }
        key := reSearch(keyRegex, l)
        val := reSearch(valRegex, l)
        details[key] = val
    }
    return details
}

func run() {
    wf.Args()
    flag.Parse()

    cfg = &workflowConfig{
        LpassBin: "lpass",
    }
    if err := wf.Config.To(cfg); err != nil {
        panic(err)
    }

    if !isLoggedIn() {
        wf.NewItem("You're not logged in to Lastpass.").
            Subtitle("Press ENTER to login.").
            Arg("auth").
            Valid(true)
        wf.SendFeedback()
        return
    }

    if detailsFlag != "" {
        backIcon := aw.Icon{Value: fmt.Sprintf("%s/icons/go_back.png", wf.Dir())}
        details := getDetails(detailsFlag)
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

        for key, value := range details {
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

        wf.NewItem("Delete entry").
            Arg("delete").
            Valid(true)

        wf.SendFeedback()
        return
    }

    var entries []lastpassEntry
    if privateFlag {
        entries = getEntries(searchFlag, os.Getenv("PRIVATE_FOLDERS"))
    } else {
        entries = getEntries(searchFlag, "")
    }
    for _, e := range entries {
        it := wf.NewItem(e.Name).
            Subtitle(fmt.Sprintf("Folder: %s  |  ID: %s", e.Folder, e.ID)).
            Arg("details").
            Var("item_id", e.ID).
            Var("item_name", e.Name).
            Var("item_url", e.URL).
            Var("query", searchFlag).
            Copytext("").
            Valid(true)

        it.NewModifier(aw.ModCtrl).
            Subtitle("Copy ID").
            Arg("copy").
            Var("copy_field", "id").
            Valid(true)

        if e.Username != "" {
            it.NewModifier(aw.ModAlt).
                Subtitle("Copy username").
                Arg("copy").
                Var("copy_field", "username").
                Valid(true)
        }
        if e.Password != "" {
            it.NewModifier(aw.ModCmd).
                Subtitle("Copy password").
                Arg("copy").
                Var("copy_field", "password").
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
