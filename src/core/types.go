package core

type Package struct {
	// Metadata
	Name        string
	Description string
	Version     string
	Repo        string

	// Package Info
	Maintainer  string
	UpstreamURL string
	LastUpdated string
	Licenses    []string
	OutOfDate   bool

	// Installed
	CurrentBuildLocation string
	Installed            bool
}

type Repo struct {
	Server   string
	SigLevel string
	Include  string
}

type NestConfig struct {
	// Style
	Color bool

	// Other (more categorisation with more options)
	Repos []*Repo
}

type User struct {
	Username   string
	Fullname   string
	HomeDir    string
	ManageHome bool
	Shell      string
	Groups     []string
}
