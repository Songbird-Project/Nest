package scripting

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func RunExternal(name string, dir string) error {
	items, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd

	for _, item := range items {
		switch item.Name() {
		case name + ".py":
			cmd = exec.Command("python", filepath.Join(dir, name+".py"))
		case name + ".rb":
			cmd = exec.Command("python", filepath.Join(dir, name+".rb"))
		case name + ".go":
			cmd = exec.Command("python", filepath.Join(dir, name+".go"))
		}
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", output)
		return err
	}

	fmt.Printf("%s", output)

	return nil
}
