package subcommands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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
	nestRoot := os.Getenv("NEST_ROOT")
	currGenRootID := os.Getenv("NEST_GEN_ROOT_ID")

	if nestRoot == "" || nestRoot == "." {
		nestRoot = "./"
	} else if nestRoot[len(nestRoot)-1:] != "/" {
		nestRoot += "/"
	}

	if len(args.Name) > 0 && args.Name[len(args.Name)-1:] != "/" {
		args.Name += "/"
	}

	if args.Name != "" {
		os.Setenv("NEST_GEN_ROOT", filepath.Join(nestRoot, args.Name))
		os.Setenv("NEST_AUTOGEN", filepath.Join(os.Getenv("NEST_GEN_ROOT"), "autogen"))
	} else if id, err := strconv.Atoi(currGenRootID); err == nil && currGenRootID != "" {
		os.Setenv("NEST_GEN_ROOT", filepath.Join(nestRoot, strconv.Itoa(id+1)+"/"))
		os.Setenv("NEST_AUTOGEN", filepath.Join(os.Getenv("NEST_GEN_ROOT"), "autogen"))
	} else {
		os.Setenv("NEST_GEN_ROOT", filepath.Join(nestRoot, "0/"))
		os.Setenv("NEST_AUTOGEN", filepath.Join(os.Getenv("NEST_GEN_ROOT"), "autogen"))
	}

	pm.RunAsAuthUser("mkdir", []string{"-p", os.Getenv("NEST_AUTOGEN")})

	err := scripting.RunExternalAsAuth("config", nestRoot, pm)
	if err != nil {
		return err
	}

	err = runBuildScript("preBuild", pm)
	if err != nil {
		return err
	}

	if os.Getenv("NEST_DRY_RUN") == "" {
	} else {
		fmt.Println("Destructive Build")
	}

	err = runBuildScript("postBuild", pm)
	if err != nil {
		return err
	}

	return nil
}

func runBuildScript(stage string, pm core.PrivilegeManager) error {
	autogen := os.Getenv("NEST_AUTOGEN")

	err := scripting.RunExternalAsAuth(stage, autogen, pm)
	if err != nil {
		return err
	}

	return nil
}
