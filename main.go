package main

import (
	// "gestic/models/selector"
	"fmt"
	"gestic/models/selector"
	"gestic/restic"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"strings"
	// "github.com/charmbracelet/bubbletea"
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

	//path1 := "/home/tohru/tmp/restic/snapshots/2025-03-31T22:34:04-03:00/home"
	//path2 := "/home/tohru/tmp/restic/snapshots/2025-04-09T15:01:43-03:00/home"

	// test
	//path1 := "/home/tohru/tmp/restic/snapshots/2025-04-09T17:45:55-03:00/home"
	//path2 := "/home/tohru/tmp/restic/snapshots/2025-04-09T17:46:45-03:00/home"

	p := tea.NewProgram(
		selector.InitialModel(snapshots),
		//.InitialModel(entries[0], entries2[0]),
		// Fix debug on Goland
		//tea.WithInput(os.Stdin),
	)

	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func PrintNodes(n restic.DirData, level int) {
	spaces := strings.Repeat(" ", level*2.0)
	fmt.Printf(spaces)
	fmt.Printf("[%s] \t%s\n", n.SizeRadable, n.PathReadable)
	if len(n.Children) > 0 && level <= 999 {
		for _, c := range n.Children {
			PrintNodes(c, level+1)
		}
	}
}
