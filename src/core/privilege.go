package core

import (
	"fmt"
	"os/exec"
)

type PrivilegeManager struct {
	AuthTool string
}

func (pm *PrivilegeManager) AuthenticateUser() error {
	var cmd *exec.Cmd

	switch pm.AuthTool {
	case "sudo":
		cmd = exec.Command("sudo", "-v")
	case "doas":
		cmd = exec.Command("doas", "true")
	case "run0":
		cmd = exec.Command("run0", "true")
	default:
		return fmt.Errorf("unkown authentication tool: %s", pm.AuthTool)
	}

	return cmd.Run()
}

func (pm *PrivilegeManager) DeauthenticateUser() error {
	var cmd *exec.Cmd

	switch pm.AuthTool {
	case "sudo":
		cmd = exec.Command("sudo", "-K")
	case "doas":
		cmd = exec.Command("doas", "-L")
	case "run0":
		cmd = exec.Command("pkill", "-f", "polkit.*agent")
	default:
		return fmt.Errorf("unkown authentication tool: %s", pm.AuthTool)
	}

	return cmd.Run()
}
