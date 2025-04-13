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

func main() {

	var cli config.CLI
	_ = kong.Parse(&cli)

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
