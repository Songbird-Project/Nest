package subcommands

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/Songbird-Project/nest/core"
	"github.com/Songbird-Project/nest/scripting"
	"github.com/Songbird-Project/scsv"
	"github.com/charmbracelet/lipgloss/v2"
)

type BuildCmd struct {
	Name               string `arg:"-n,--name" help:"name the new build generation rather than using a number"`
	NoCompress         bool   `arg:"--no-compress" help:"do not compress the previous generation"`
	IgnoreExistingHome bool   `arg:"-i,--ignore-existing-home" help:"do not restore the previous home directory of a previously existing user that has been recreated"`
	Home               bool   `arg:"-H,--home" help:"only rebuild managed home directories rather than the whole system"`
}

func SysBuild(args *BuildCmd) error {
	authTool := os.Getenv("NEST_AUTH")
	if authTool == "" {
		authTool = "sudo"
	}

	pm := core.PrivilegeManager{}

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

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Blue)

	genRoot := getGenRoot(nestRoot, args.Name, currGenRootID)
	if err := updateEnvironment(genRoot); err != nil {
		return err
	}

	if err := pm.RunAsAuthUser("mkdir", "-p", os.Getenv("NEST_AUTOGEN")); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(infoStyle.Render("Generating system config..."))

	if err := scripting.RunExternalAsAuth("config", nestRoot, pm); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(infoStyle.Render("Running pre-build script..."))

	if err := runBuildScript("preBuild", pm); err != nil {
		return err
	}

	if os.Getenv("NEST_DRY_RUN") == "" {
		if err := manageUsers(pm); err != nil {
			return err
		}

		if err := linkSystemConfigs(currGenRoot, pm); err != nil {
			return err
		}

		if err := installPkgs(pm); err != nil {
			return err
		}
	} else {
		fmt.Println("Destructive Build")
	}

	fmt.Println(infoStyle.Render("Running post-build script..."))

	if err := runBuildScript("postBuild", pm); err != nil {
		return err
	}

	if currGenRoot != "" && !args.NoCompress {
		fmt.Println(infoStyle.Render("Cleaning previous generation..."))
		pm.ArchiveAndRemove(currGenRoot)
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

func manageUsers(pm core.PrivilegeManager) error {
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Blue)
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Yellow)

	fmt.Println(infoStyle.Render("Adding users:"))

	users, err := core.GetUserConfigs()
	if err != nil {
		return err
	}

	userList, homes, err := core.GetUsers()
	if err != nil {
		return err
	}

	var cfgUsers []string

	for _, userCfg := range users {
		cfgUsers = append(cfgUsers, userCfg.Username)

		if !slices.Contains(userList, userCfg.Username) {
			fmt.Printf(infoStyle.Render(" - ", userCfg.Fullname))

			args := []string{
				"-d",
				userCfg.HomeDir,
				"-c",
				userCfg.Fullname,
				"-s",
				userCfg.Shell,
			}

			homeDirArchive := strings.TrimSuffix(userCfg.HomeDir, "/") + ".tar.zst"

			existingHome, err := core.PathExists(
				homeDirArchive,
			)
			if err != nil {
				return err
			}

			if existingHome {
				fmt.Printf(warnStyle.Italic(true).Render(" (restoring)\n"))
				pm.RunAsAuthUser("tar", "--zstd", "-xvf", homeDirArchive)
				pm.RunAsAuthUser("rm", "-rf", homeDirArchive)
			} else {
				fmt.Printf("\n")
				args = append(args, "-m")
			}

			if len(userCfg.Groups) > 0 {
				args = append(args, "-G", strings.Join(userCfg.Groups, ","))
			}

			args = append(args, userCfg.Username)

			if err := pm.RunAsAuthUser("useradd", args...); err != nil {
				return err
			}
		}
	}

	fmt.Println()
	fmt.Println(infoStyle.Render("Removing users:"))

	for idx, existingUser := range userList {
		if !slices.Contains(cfgUsers, existingUser) {
			fmt.Println(infoStyle.Render(" - ", existingUser))

			pm.RunAsAuthUser("userdel", existingUser)
			pm.ArchiveAndRemove(homes[idx])
		}
	}

	return nil
}

func installPkgs(pm core.PrivilegeManager) error {
	args := []string{"-r $NEST_GEN_ROOT/bin", "-b $NEST_GEN_ROOT/db", "-S"}
	pkglistPath := filepath.Join(os.Getenv("NEST_GEN_ROOT"), "pkglist.scsv")
	autogenPkglistPath := filepath.Join(os.Getenv("NEST_GEN_ROOT"), "pkglist.scsv")

	var pkglist scsv.KeyValuePairs
	var autogenPkglist scsv.KeyValuePairs

	pkglistExists, err := core.PathExists(pkglistPath)
	if err != nil {
		return err
	}

	autogenPkglistExists, err := core.PathExists(autogenPkglistPath)
	if err != nil {
		return err
	}

	if pkglistExists {
		pkglist, err = scsv.ParseFile(pkglistPath)
		if err != nil {
			return err
		}
	}

	if autogenPkglistExists {
		autogenPkglist, err = scsv.ParseFile(autogenPkglistPath)
		if err != nil {
			return err
		}
	}

	var pkgs []string

	for repo := range pkglist {
		for _, pkg := range pkglist[repo] {
			pkgs = append(pkgs, pkg[0])
		}
	}

	for repo := range autogenPkglist {
		for _, pkg := range autogenPkglist[repo] {
			if slices.Contains(pkgs, pkg[0]) {
				continue
			}

			pkgs = append(pkgs, pkg[0])
		}
	}

	args = append(args, pkgs...)

	if err := pm.RunAsAuthUser("pacman", args...); err != nil {
		return err
	}

	return nil
}

func linkSystemConfigs(currGenRoot string, pm core.PrivilegeManager) error {
	binDir := filepath.Join(
		strings.TrimSuffix(os.Getenv("NEST_GEN_ROOT"), "/")+"/",
		"bin/*",
	)
	dbDir := filepath.Join(
		strings.TrimSuffix(currGenRoot, "/")+"/",
		"db",
	)
	oldDbDir := filepath.Join(
		strings.TrimSuffix(currGenRoot, "/")+"/",
		"db",
	)

	binFiles, err := filepath.Glob(binDir)
	if err != nil {
		return err
	}

	for _, file := range binFiles {
		if err := pm.RunAsAuthUser("ln", "-sf", file, "/usr/bin/"+file); err != nil {
			return err
		}
	}

	installPkgs(pm)

	if err := pm.RunAsAuthUser("cp", "-rf", oldDbDir, dbDir); err != nil {
		return nil
	}

	return nil
}
