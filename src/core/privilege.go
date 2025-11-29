package core

import (
	"os"
	"os/exec"
)

type PrivilegeManager struct {
	AuthTool string
}

func (pm *PrivilegeManager) AuthenticateUser() error {
	cmd := exec.Command("sudo", "-v")

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func (pm *PrivilegeManager) DeauthenticateUser() error {
	cmd := exec.Command("sudo", "-K")

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func (pm *PrivilegeManager) RunAsAuthUser(command string, args []string) error {
	cmd := exec.Command("sudo", append([]string{"-E", command}, args...)...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
