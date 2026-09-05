//go:build !windows

package main

import (
	"fmt"
	"github.com/ECO-evidence-casework-one/eco/internal/eco"
	"os"
)

func main() {
	root := eco.DefaultDevelopmentWorkspaceRoot(os.TempDir())
	if err := eco.CheckWorkspaceRecoveryState(root); err != nil {
		panic(err)
	}
	var v *eco.Vault
	var err error
	if _, checkErr := os.Lstat(root); os.IsNotExist(checkErr) {
		v, err = eco.CreateVault(root)
	} else {
		v, err = eco.OpenVault(root)
	}
	if err != nil {
		panic(err)
	}
	defer v.Close()
	fmt.Println(eco.BuildName)
	fmt.Printf("Development core opened: %d evidence items at %s\n", len(v.Workspace.Evidence), root)
	if len(os.Args) > 1 {
		for _, p := range os.Args[1:] {
			item, dup, err := v.ImportFile(p, nil)
			fmt.Printf("import %s duplicate=%v err=%v item=%s\n", p, dup, err, item.ID)
		}
	}
}
