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

    // Commands
    Details     bool
    Update      bool
    ListFolders bool
}

func init() {
    cli.StringVar(&opts.Folders, "folders", "", "only search in specified folders")
    cli.BoolVar(&opts.Details, "details", false, "item details")
    cli.BoolVar(&opts.Update, "update", false, "check for updates")
    cli.BoolVar(&opts.ListFolders, "listfolders", false, "list all folders")
}
