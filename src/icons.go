package main

import (
    "fmt"
    "os"
    "strings"

    aw "github.com/deanishe/awgo"
)

var (
    iconFolder = &aw.Icon{Value: "icons/group.png"}
    iconBack   = &aw.Icon{Value: "icons/go_back.png"}
    iconSN     = &aw.Icon{Value: "icons/sn.png"}
    iconPW     = &aw.Icon{Value: "icons/password-alt.png"}
    iconDelete = &aw.Icon{Value: "icons/trash.png"}
    iconEdit   = &aw.Icon{Value: "icons/edit.png"}
)

func getIcon(key string) *aw.Icon {
    iconPath := fmt.Sprintf("icons/%s.png", strings.ToLower(key))
    _, err := os.Stat(iconPath)

    if os.IsNotExist(err) {
        return &aw.Icon{Value: "icons/default.png"}
    } else {
        return &aw.Icon{Value: iconPath}
    }
}
