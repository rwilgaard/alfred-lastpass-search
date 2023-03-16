# LastPass Search

A workflow for searching in LastPass. This workflow uses the [LastPass CLI](https://github.com/lastpass/lastpass-cli).

The easiest way to install the LastPass CLI is using [Homebrew](https://brew.sh/):
```
brew install lastpass-cli
```

## Features
* Search for entries
* Edit existing entries
* Delete existing entries
* Add new entries & password generation
* Workflow auto update

## Keywords

* `lp` search for entries in the entire LastPass vault. A hotkey can be configured for this keyword.
* `lpp` search for entries only in specified private folders. The private folders can be configured in the **User Configuration**. A hotkey can be configured for this keyword.
* `lpadd` add new entry to LastPass.
* `lpgen` generate a new random password and copy it to the clipboard or add it directly to LastPass. The default length is 32 characters, but you can also specify the length after `lpgen`.
* `lpsync` run a manual sync of the Lastpass Vault.
* `lpout` logout of LastPass.

## Actions
All the mappings below can be changed in the **User Configuration**.

#### Default mappings
The following actions can be used on entries returned from the `lp` & `lpp` keywords:
* `↩` will copy the password to the clipboard.
* `⌘` + `↩` will show details for the entry.
* `⌥` + `↩` will copy the username to the clipboard.
* `⌃` + `↩` will copy the ID to the clipboard.
