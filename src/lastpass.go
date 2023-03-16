package main

import (
    "fmt"
    "os/exec"
    "regexp"
    "strings"
)

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

func getEntries(query string, folder string) ([]LastpassEntry, error) {
    cmd := exec.Command(cfg.LpassBin, "ls", "--format", "%aN [id: %ai] [url: %al] [username: %au] %ap", "--sync=no", folder)
    out, err := cmd.Output()

    if err != nil {
        return nil, err
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
    return entries, nil
}

func getDetails(itemID string) ([]string, map[string]string, error) {
    cmd := exec.Command(cfg.LpassBin, "show", itemID, "--sync=no")
    out, err := cmd.Output()

    if err != nil {
        return nil, nil, err
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
    return keys, details, nil
}
