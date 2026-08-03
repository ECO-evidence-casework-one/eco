//go:build !windows

package main

import (
	"fmt"
	"github.com/ECO-evidence-casework-one/eco/internal/eco"
	"os"
	"path/filepath"
)

func main() {
	root, err := os.UserConfigDir()
	if err != nil || root == "" {
		panic("ECO could not find a private application-state folder")
	}
	candidate, err := eco.StartCandidate(filepath.Join(root, "EvidenceCaseworkOne"), eco.CurrentRuntime())
	if err != nil {
		panic(err)
	}
	session := candidate.Current
	v := session.Vault
	fmt.Println(eco.BuildName)
	fmt.Printf("Workspace: %s\nIdentity: %s\nPath: %s\nStatus: %s\n", session.Identity.Name, session.Identity.ID, session.Path, session.StatusText())
	fmt.Printf("Development core opened: %d evidence items\n", len(v.Workspace.Evidence))
	if len(os.Args) > 1 {
		for _, p := range os.Args[1:] {
			item, dup, err := v.ImportFile(p, nil)
			fmt.Printf("import %s duplicate=%v err=%v item=%s\n", p, dup, err, item.ID)
		}
	}
}
