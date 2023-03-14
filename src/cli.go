package main

import "flag"

var (
    opts = &options{}
    cli  = flag.NewFlagSet("alfred-lastpass-search", flag.ContinueOnError)
)

type options struct {
    // Arguments
    Query   string
    Folders string
    Name    string
    Length  int

    // Commands
    Details     bool
    Update      bool
    ListFolders bool
    Generate    bool
}

func init() {
    cli.StringVar(&opts.Folders, "folders", "", "only search in specified folders")
    cli.StringVar(&opts.Name, "name", "", "name for new entry")
    cli.IntVar(&opts.Length, "length", 32, "length of password to generate")
    cli.BoolVar(&opts.Details, "details", false, "item details")
    cli.BoolVar(&opts.Update, "update", false, "check for updates")
    cli.BoolVar(&opts.ListFolders, "listfolders", false, "list all folders")
    cli.BoolVar(&opts.Generate, "generate", false, "generate new password")
}
