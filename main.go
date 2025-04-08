package main

import (
	"fmt"
	"gestic/restic"
	"gestic/ui"
	"github.com/charmbracelet/bubbletea"
	"os"
)

func main() {
	snapshots, err := restic.GetSnapshopts()
	if err != nil {
		panic(err)
	}
	_=snapshots

	// for _, s := range snapshots {
	// 	fmt.Printf("%s", s)
	// }

	p := tea.NewProgram(ui.InitialModel(snapshots))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
