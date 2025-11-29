package subcommands

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/Songbird-Project/nest/scripting"
)

type BuildCmd struct {
	Name string `arg:"-n,--name" help:"name the new build generation rather than using a number"`
}

func SysBuild(args *BuildCmd) error {
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
	} else if id, err := strconv.Atoi(currGenRootID); err != nil && currGenRootID != "" {
		os.Setenv("NEST_GEN_ROOT", filepath.Join(nestRoot, strconv.Itoa(id+1)+"/"))
		os.Setenv("NEST_AUTOGEN", filepath.Join(os.Getenv("NEST_GEN_ROOT"), "autogen"))
	} else {
		os.Setenv("NEST_GEN_ROOT", filepath.Join(nestRoot, "0/"))
		os.Setenv("NEST_AUTOGEN", filepath.Join(os.Getenv("NEST_GEN_ROOT"), "autogen"))
	}

	os.MkdirAll(os.Getenv("NEST_AUTOGEN"), 0777)

	err := scripting.RunExternal("config", nestRoot)
	if err != nil {
		return err
	}

	err = RunBuildScript()
	if err != nil {
		return err
	}

	return nil
}

func RunBuildScript() error {
	autogen := os.Getenv("NEST_AUTOGEN")

	err := scripting.RunExternal("preBuild", autogen)
	if err != nil {
		return err
	}

	err = scripting.RunExternal("postBuild", autogen)
	if err != nil {
		return err
	}

	return nil
}
