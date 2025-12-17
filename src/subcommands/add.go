package subcommands

type AddCmd struct {
	PkgNames []string `arg:"positional"`
}

func AddPkg() error {
	return nil
}
