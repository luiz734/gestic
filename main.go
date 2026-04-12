package main

import (
	"fmt"
	"gestic/config"
	"gestic/models/selector"
	"gestic/restic"
	"os"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	var cli config.CLI
	ctx := kong.Parse(&cli,
		kong.Name("gestic"),
		kong.Description("A diff tool for restic snapshots."),
		kong.Vars{
			"version": fmt.Sprintf("%s (%s)", version, commit),
		},
	)

	err := ctx.Run()
	// We don't use subcommands and are ignoring errors.
	// if err != nil {
	// 	ctx.FatalIfErrorf(err)
	// }

	snapshots, err := restic.GetSnapshots(cli.RepoPath, cli.MountPath)
	if err != nil {
		panic(err)
	}

	p := tea.NewProgram(
		selector.InitialModel(snapshots),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
