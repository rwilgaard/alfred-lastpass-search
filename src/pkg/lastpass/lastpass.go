package lastpass

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/rwilgaard/alfred-lastpass-search/src/pkg/util"
)

type LastpassService struct {
	BinPath string
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

func NewLastpassService(binPath string) (*LastpassService, error) {
	if len(binPath) == 0 {
		return nil, errors.New("binPath is empty")
	}

	svc := &LastpassService{
		BinPath: binPath,
	}

	return svc, nil
}

func (ls *LastpassService) IsLoggedIn() bool {
	cmd := exec.Command(ls.BinPath, "status", "--quiet")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func (ls *LastpassService) GetFolders() ([]LastpassFolder, error) {
	cmd := ls.BinPath + " ls --format %aN,%al --sync=no | grep ',http://group$' | cut -d, -f1 | sort -u"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return nil, err
	}

	var folders []LastpassFolder
	for _, folder := range strings.Split(string(out), "\n") {
		lf := LastpassFolder{
			Name: folder,
		}
		folders = append(folders, lf)
	}

	return folders, nil
}

func (ls *LastpassService) GetEntries(query string, folders []string, fuzzy bool) ([]LastpassEntry, error) {
	var output string
	if len(folders) == 0 {
		folders = append(folders, "")
	}

	for _, folder := range folders {
		cmd := exec.Command(ls.BinPath, "ls", "--format", "%aN [id: %ai] [url: %al] [username: %au] %ap", "--sync=no", folder)
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("Error running lpass ls: %s", err)
		}

		output += string(out)
	}

	idRegex := regexp.MustCompile(`\[id: ([0-9]+)\]`)
	urlRegex := regexp.MustCompile(`\[url: (.+?)\]`)
	nameRegex := regexp.MustCompile(`^.*?\/(.*?)\s\[`)
	folderRegex := regexp.MustCompile(`^(.*?)\/`)
	usernameRegex := regexp.MustCompile(`\[username: (.+?)\]`)
	passwordRegex := regexp.MustCompile(`.*\] (.*)$`)

	var entries []LastpassEntry
	for _, l := range strings.Split(string(output), "\n") {
		id := util.RegexSearch(idRegex, l)
		name := util.RegexSearch(nameRegex, l)
		folder := util.RegexSearch(folderRegex, l)
		url := util.RegexSearch(urlRegex, l)
		username := util.RegexSearch(usernameRegex, l)
		password := util.RegexSearch(passwordRegex, l)
		e := fmt.Sprintf("%s %s %s %s %s", id, name, folder, url, username)
		if id == "" {
			continue
		}
		if url == "http://group" {
			continue
		}
		if !util.HasAll(strings.ToLower(e), strings.Split(strings.ToLower(query), " ")) && !fuzzy {
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

func (ls *LastpassService) GetDetails(itemID string) ([]string, map[string]string, error) {
	if len(itemID) == 0 {
		return nil, nil, errors.New("itemID is empty")
	}

	cmd := exec.Command(ls.BinPath, "show", itemID, "--sync=no")
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
		key := util.RegexSearch(keyRegex, l)
		val := util.RegexSearch(valRegex, l)
		keys = append(keys, key)
		details[key] = val

		if key == "Notes" {
			break
		}
	}
	return keys, details, nil
}

func (ls *LastpassService) CheckValidity(entry LastpassEntry, action string) bool {
	if action == "Copy Password" && entry.Password == "" {
		return false
	} else if action == "Copy Username" && entry.Username == "" {
		return false
	}
	return true
}
