package subcommands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Songbird-Project/nest/core"
	"github.com/Songbird-Project/nest/scripting"
)

type BuildCmd struct {
	Name string `arg:"-n,--name" help:"name the new build generation rather than using a number"`
	Home bool   `arg:"-H,--home" help:"only rebuild managed home directories rather than the whole system"`
}

func SysBuild(args *BuildCmd) error {
	authTool := os.Getenv("NEST_AUTH")
	if authTool == "" {
		authTool = "sudo"
	}

	pm := core.PrivilegeManager{
		AuthTool: authTool,
	}

	if args.Home {
		return nil
	} else {
		pm.AuthenticateUser()
		defer pm.DeauthenticateUser()

		err := SysBuildAll(args, pm)
		if err != nil {
			return err
		}
	}

	return nil
}

func SysBuildAll(args *BuildCmd, pm core.PrivilegeManager) error {
	nestRoot := fixRootPath(os.Getenv("NEST_ROOT"))
	currGenRootID := os.Getenv("NEST_GEN_ROOT_ID")
	currGenRoot := os.Getenv("NEST_GEN_ROOT")

	genRoot := getGenRoot(nestRoot, args.Name, currGenRootID)
	if err := updateEnvironment(genRoot); err != nil {
		return err
	}

	if err := pm.RunAsAuthUser("mkdir", []string{"-p", os.Getenv("NEST_AUTOGEN")}); err != nil {
		return err
	}

	if err := scripting.RunExternalAsAuth("config", nestRoot, pm); err != nil {
		return err
	}

	if err := runBuildScript("preBuild", pm); err != nil {
		return err
	}

	if os.Getenv("NEST_DRY_RUN") == "" {
	} else {
		fmt.Println("Destructive Build")
	}

	if err := runBuildScript("postBuild", pm); err != nil {
		return err
	}

	if currGenRoot != "" {
		if err := pm.RunAsAuthUser("zstd", []string{"-r", currGenRoot}); err != nil {
			return err
		}
	}

	return nil
}

func runBuildScript(stage string, pm core.PrivilegeManager) error {
	autogen := os.Getenv("NEST_AUTOGEN")

	return scripting.RunExternalAsAuth(stage, autogen, pm)
}

func fixRootPath(nestRoot string) string {
	if nestRoot == "" || nestRoot == "." {
		nestRoot = "./"
	}

	if !strings.HasSuffix(nestRoot, "/") {
		return nestRoot + "/"
	}

	return nestRoot
}

func getGenRoot(nestRoot string, name string, currGenRootID string) string {
	if name != "" {
		return filepath.Join(nestRoot, strings.TrimSuffix(name, "/")+"/")
	}

	if id, err := strconv.Atoi(currGenRootID); err == nil && currGenRootID != "" {
		return filepath.Join(nestRoot, fmt.Sprintf("%d/", id+1))
	}

	return filepath.Join(nestRoot, "0/")
}

func updateEnvironment(nestGenRoot string) error {
	if err := os.Setenv("NEST_GEN_ROOT", nestGenRoot); err != nil {
		return err
	}

	return os.Setenv("NEST_AUTOGEN", filepath.Join(nestGenRoot, "autogen"))
}
