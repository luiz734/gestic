package main

import (
	"fmt"
	"gestic/models/selector"
	"gestic/restic"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func main() {
	snapshots, err := restic.GetSnapshopts()
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
