package cmd

import (
	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	"github.com/rwilgaard/alfred-lastpass-search/src/pkg/lastpass"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

type workflowConfig struct {
	ModifierReturn      string `env:"modifier_return"`
	ModifierCmd         string `env:"modifier_cmd"`
	ModifierOpt         string `env:"modifier_opt"`
	ModifierCtrl        string `env:"modifier_ctrl"`
	AllowedSymbols      string `env:"allowed_symbols"`
	FuzzySearch         bool   `env:"fuzzy_search"`
	IntelligentOrdering bool   `env:"intelligent_ordering"`
}

const (
	repo       = "rwilgaard/alfred-lastpass-search"
	maxResults = 25
)

var (
	wf      *aw.Workflow
	ls      *lastpass.Service
	cfg     = &workflowConfig{}
	rootCmd = &cobra.Command{
		Use:   "lastpass-alfred",
		Short: "lastpass-alfred is a CLI to be used by Alfred for searching in your Lastpass Vault",
	}
)

func Execute() {
	wf.Run(run)
}

func run() {
	if err := alfredutils.InitWorkflow(wf, cfg); err != nil {
		wf.FatalError(err)
	}

	if err := alfredutils.CheckForUpdates(wf); err != nil {
		wf.FatalError(err)
	}

	var err error
	ls, err = lastpass.NewService("lpass")
	if err != nil {
		wf.FatalError(err)
	}

	if !ls.IsLoggedIn() {
		wf.NewItem("You're not logged in to Lastpass.").
			Subtitle("Press ⏎ to login.").
			Arg("auth").
			Valid(true)
		alfredutils.HandleFeedback(wf)
	}

	if cfg.IntelligentOrdering {
		wf.Configure(aw.SuppressUIDs(false))
	}

	if err := rootCmd.Execute(); err != nil {
		wf.FatalError(err)
	}
}

func init() {
	wf = aw.New(
		aw.MaxResults(maxResults),
		update.GitHub(repo),
		aw.SuppressUIDs(true),
	)
}
