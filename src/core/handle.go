package core

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"

	"github.com/Jguer/go-alpm/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

func AlpmInit() (*alpm.Handle, error) {
	handler, err := alpm.Initialize("/", "/var/lib/pacman")
	if err != nil {
		return nil, err
	}

	return handler, nil
}

func OrganisePkgsByRepo(packages []Package) map[string][]Package {
	repoMap := make(map[string][]Package)

	for _, pkg := range packages {
		repoMap[pkg.Repo] = append(repoMap[pkg.Repo], pkg)
	}

	return repoMap
}

func GetRepoColors() map[string]color.Color {
	return map[string]color.Color{
		"core":     lipgloss.Color("10"),
		"extra":    lipgloss.Color("11"),
		"multilib": lipgloss.Color("12"),
		"testing":  lipgloss.Color("9"),
		"aur":      lipgloss.Color("14"),
	}
}

func GetStyleForRepo(repo string, repoColors map[string]color.Color) color.Color {
	repoLower := strings.ToLower(repo)

	if repoLower == "aur" {
		return repoColors["aur"]
	}

	for key, style := range repoColors {
		if key != "aur" && strings.Contains(repoLower, key) {
			return style
		}
	}

	return lipgloss.Color("13")
}

func ColoriseByRepo(repo string, repoColors map[string]color.Color) string {
	return lipgloss.NewStyle().Foreground(GetStyleForRepo(repo, repoColors)).Bold(true).Render(repo)
}

func SortRepos(organisedPkgs map[string][]Package) []string {
	order := []string{"core", "extra", "multilib", "testing", "aur"}
	result := []string{}
	seen := make(map[string]bool)

	for _, repoName := range order {
		for repo := range organisedPkgs {
			if strings.ToLower(repo) == repoName {
				result = append(result, repo)
				seen[repo] = true
			}
		}
	}

	remainingRepos := []string{}
	for repo := range organisedPkgs {
		if !seen[repo] {
			remainingRepos = append(remainingRepos, repo)
		}
	}

	sort.Strings(remainingRepos)

	result = append(result, remainingRepos...)

	return result
}

func HandleSignal(
	exit int,
	action func() error,
	pm PrivilegeManager,
	signalTypes ...os.Signal,
) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, signalTypes...)
	done := make(chan bool, 1)

	go func() {
		<-sig
		signal.Stop(sig)

		if err := pm.DeauthenticateUser(); err != nil {
			fmt.Printf("%s\n", err)
		}

		if action != nil {
			if err := action(); err != nil {
				fmt.Printf("%s\n", err)
			}
		}

		cmd := exec.Command("pkill", "-9", "-f", "sudo -v")
		if err := cmd.Run(); err != nil {
			fmt.Printf("%s\n", err)
		}

		done <- true
	}()

	<-done
	os.Exit(exit)
}
