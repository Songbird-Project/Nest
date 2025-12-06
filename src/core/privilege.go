package core

import (
	"os"
	"os/exec"
	"strings"
)

type PrivilegeManager struct{}

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

func (pm *PrivilegeManager) RunAsAuthUser(command string, args ...string) error {
	cmd := exec.Command("sudo", append([]string{"-E", command}, args...)...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func (pm *PrivilegeManager) ArchiveAndRemove(dir string) error {
	archiveName := strings.TrimSuffix(dir, "/") + ".tar.zst"

	if err := pm.RunAsAuthUser("tar",
		"--zstd",
		"-cvf",
		archiveName,
		dir,
	); err != nil {
		return err
	}

	if err := pm.RunAsAuthUser("tar",
		"--zstd",
		"-tf",
		archiveName,
	); err != nil {
		return err
	}

	return pm.RunAsAuthUser("rm",
		"-rf",
		dir,
	)
}
