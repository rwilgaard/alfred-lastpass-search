package util

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	aw "github.com/deanishe/awgo"
	"github.com/sethvargo/go-password/password"
)

var (
	IconFolder = &aw.Icon{Value: "icons/group.png"}
	IconBack   = &aw.Icon{Value: "icons/go_back.png"}
	IconSN     = &aw.Icon{Value: "icons/sn.png"}
	IconPW     = &aw.Icon{Value: "icons/password-alt.png"}
	IconDelete = &aw.Icon{Value: "icons/trash.png"}
	IconEdit   = &aw.Icon{Value: "icons/edit.png"}
)

func RegexSearch(regex *regexp.Regexp, query string) string {
	if regex.MatchString(query) {
		return regex.FindStringSubmatch(query)[1]
	}
	return ""
}

func HasAll(input string, words []string) bool {
	for _, w := range words {
		if strings.Contains(input, w) {
			continue
		}
		return false
	}
	return true
}

func GeneratePassword(length int, symbols bool, allowedSymbols string) (string, error) { //nolint:revive // Allow control flag for symbols
	input := password.GeneratorInput{
		Symbols: allowedSymbols,
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

func GetIcon(key string) *aw.Icon {
	iconPath := fmt.Sprintf("icons/%s.png", strings.ToLower(key))
	_, err := os.Stat(iconPath)

	if os.IsNotExist(err) {
		return &aw.Icon{Value: "icons/default.png"}
	}

	return &aw.Icon{Value: iconPath}
}
