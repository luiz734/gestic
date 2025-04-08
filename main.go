package main

import (
	"fmt"
	"gestic/models/selector"
	"gestic/restic"
	"github.com/charmbracelet/bubbletea"
	"os"
)

func main() {
	snapshots, err := restic.GetSnapshopts()
	if err != nil {
		panic(err)
	}
	_ = snapshots

	// for _, s := range snapshots {
	// 	fmt.Printf("%s", s)
	// }

	p := tea.NewProgram(
		selector.InitialModel(
			selector.Model{},
			snapshots),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
