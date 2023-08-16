package main

import (
    "log"
    "os"
    "os/exec"
    "regexp"
    "strings"

    aw "github.com/deanishe/awgo"
    "github.com/deanishe/awgo/update"
    "github.com/sethvargo/go-password/password"
)

type WorkflowConfig struct {
    LpassBin            string
    ModifierReturn      string `env:"modifier_return"`
    ModifierCmd         string `env:"modifier_cmd"`
    ModifierOpt         string `env:"modifier_opt"`
    ModifierCtrl        string `env:"modifier_ctrl"`
    AllowedSymbols      string `env:"allowed_symbols"`
    FuzzySearch         bool   `env:"fuzzy_search"`
    IntelligentOrdering bool   `env:"intelligent_ordering"`
}

const (
    repo          = "rwilgaard/alfred-lastpass-search"
    updateJobName = "checkForUpdates"
)

var (
    wf  *aw.Workflow
    cfg *WorkflowConfig
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

func generatePassword(length int, symbols bool) (string, error) {
    input := password.GeneratorInput{
        Symbols: cfg.AllowedSymbols,
    }

    sc := 0
    if symbols {
        sc = length / 4
    }

    gen, err := password.NewGenerator(&input)
    if err != nil {
        return "", err
    }

    pw, err := gen.Generate(length, length/4, sc, false, true)
    if err != nil {
        return "", err
    }

    return pw, nil
}

func run() {
    if err := cli.Parse(wf.Args()); err != nil {
        wf.FatalError(err)
    }
    opts.Query = cli.Arg(0)
    cfg = &WorkflowConfig{LpassBin: "lpass"}
    if err := wf.Config.To(cfg); err != nil {
        panic(err)
    }

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

    if !isLoggedIn() {
        wf.NewItem("You're not logged in to Lastpass.").
            Subtitle("Press ⏎ to login.").
            Arg("auth").
            Valid(true)
        wf.SendFeedback()
        return
    }

    if opts.Generate {
        runGenerate()
        wf.SendFeedback()
        return
    }

    if opts.ListFolders {
        runListFolders()
        wf.Filter(opts.Query)
        wf.SendFeedback()
        return
    }

    if opts.Details {
        runDetails()
        wf.SendFeedback()
        return
    }

    runSearch()

    if cfg.FuzzySearch && len(opts.Query) > 0 {
        wf.Filter(opts.Query)
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
