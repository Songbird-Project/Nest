package scripting

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Songbird-Project/nest/core"
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
			cmd = exec.Command("ruby", filepath.Join(dir, name+".rb"))
		case name + ".go":
			cmd = exec.Command("go", filepath.Join(dir, name+".go"))
		}
	}

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if cmd != nil {
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func RunExternalAsAuth(name string, dir string, pm core.PrivilegeManager) error {
	items, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var cmd string
	var args []string

	for _, item := range items {
		switch item.Name() {
		case name + ".py":
			cmd, args = "python", []string{filepath.Join(dir, name+".py")}
		case name + ".rb":
			cmd, args = "ruby", []string{filepath.Join(dir, name+".rb")}
		case name + ".go":
			cmd, args = "python", []string{filepath.Join(dir, name+".go")}
		}
	}

	return pm.RunAsAuthUser(cmd, args...)
}
