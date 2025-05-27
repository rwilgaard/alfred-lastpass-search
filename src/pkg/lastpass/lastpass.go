package lastpass

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/rwilgaard/alfred-lastpass-search/src/pkg/util"
)

// Service handles interactions with the LastPass CLI.
type Service struct {
	BinPath     string
	ExecCommand func(name string, arg ...string) *exec.Cmd
}

type Folder struct {
	Name string
}

type Entry struct {
	ID       string
	Name     string
	Folder   string
	URL      string
	Username string
	Password string
}

// NewService creates a new Service.
func NewService(binPath string) (*Service, error) {
	if len(binPath) == 0 {
		return nil, errors.New("binPath is empty")
	}

	svc := &Service{
		BinPath:     binPath,
		ExecCommand: exec.Command, // Default to the real exec.Command
	}

	return svc, nil
}

// IsLoggedIn checks if the user is logged into LastPass.
func (ls *Service) IsLoggedIn() bool {
	cmd := ls.ExecCommand(ls.BinPath, "status", "--quiet")
	err := cmd.Run()

	return err == nil
}

// GetFolders retrieves all LastPass folder names.
func (ls *Service) GetFolders() ([]Folder, error) {
	cmdString := ls.BinPath + ` ls --format="%/as%/ag" --sync=no | sort -u`
	cmd := ls.ExecCommand("bash", "-c", cmdString)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running command to get folders: %w", err)
	}

	var folders []Folder
	folderOutput := strings.TrimSuffix(string(out), "\n")
	if folderOutput == "" {
		return []Folder{}, nil
	}
	for _, folderName := range strings.Split(folderOutput, "\n") {
		if folderName != "" { // Ensure we don't add empty folder names
			lf := Folder{
				Name: folderName,
			}
			folders = append(folders, lf)
		}
	}

	return folders, nil
}

// GetEntries retrieves LastPass entries, optionally filtered by query and folders.
func (ls *Service) GetEntries(query string, folders []string, fuzzy bool) ([]Entry, error) { //nolint:revive // Allow control flag for fuzzy search
	var outputBuilder strings.Builder

	if len(folders) == 0 {
		folders = []string{""}
	}

	for _, folder := range folders {
		cmd := ls.ExecCommand(ls.BinPath, "ls", "--format", "%aN [id: %ai] [url: %al] [username: %au] %ap", "--sync=no", folder)
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("error running lpass ls for folder '%s': %w", folder, err)
		}
		outputBuilder.Write(out)
		if len(out) > 0 && out[len(out)-1] != '\n' {
			outputBuilder.WriteByte('\n')
		}
	}

	fullOutput := outputBuilder.String()

	idRegex := regexp.MustCompile(`\[id: ([0-9]+)\]`)
	urlRegex := regexp.MustCompile(`\[url: (.+?)\]`)
	nameRegex := regexp.MustCompile(`^.*?\/(.*?)\s\[`)
	folderRegex := regexp.MustCompile(`^(.*?)\/`)
	usernameRegex := regexp.MustCompile(`\[username: (.+?)\]`)
	passwordRegex := regexp.MustCompile(`.*\] (.*)$`)

	entries := make([]Entry, 0)
	trimmedOutput := strings.TrimSuffix(fullOutput, "\n")
	if trimmedOutput == "" {
		return entries, nil
	}

	for _, l := range strings.Split(trimmedOutput, "\n") {
		if l == "" {
			continue
		}

		id := util.RegexSearch(idRegex, l)
		name := util.RegexSearch(nameRegex, l)
		folder := util.RegexSearch(folderRegex, l)
		url := util.RegexSearch(urlRegex, l)
		username := util.RegexSearch(usernameRegex, l)
		password := util.RegexSearch(passwordRegex, l)

		if id == "" {
			// Skip entries without an ID
			continue
		}

		if url == "http://group" {
			// Skip entries with a URL of "http://group"
			continue
		}

		if !fuzzy && query != "" {
			searchableString := fmt.Sprintf("%s %s %s %s %s", id, name, folder, url, username)
			if !util.HasAll(strings.ToLower(searchableString), strings.Split(strings.ToLower(query), " ")) {
				continue
			}
		}

		entries = append(entries, Entry{
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

// GetDetails retrieves detailed information for a specific LastPass item.
func (ls *Service) GetDetails(itemID string) ([]string, map[string]string, error) {
	if len(itemID) == 0 {
		return nil, nil, errors.New("itemID is empty")
	}

	cmd := ls.ExecCommand(ls.BinPath, "show", itemID, "--sync=no")
	out, err := cmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("error running lpass show for itemID '%s': %w", itemID, err)
	}

	keyRegex := regexp.MustCompile(`^(\S.+?):`)
	valRegex := regexp.MustCompile(`^\S.+?: (.*)`)
	keys := []string{}
	details := make(map[string]string)

	outputLines := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n") //nolint:revive // Allow newline without const

	for i, l := range outputLines {
		if l == "" {
			continue
		}
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

// CheckValidity checks if an action can be performed on an entry.
func (ls *Service) CheckValidity(entry Entry, action string) bool {
	if action == "Copy Password" && entry.Password == "" {
		return false
	} else if action == "Copy Username" && entry.Username == "" {
		return false
	}
	return true
}
