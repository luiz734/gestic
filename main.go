package main

import (
	// "gestic/models/selector"
	"fmt"
	"gestic/models/compare"
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

	path1 := "/home/tohru/tmp/restic/snapshots/2025-03-31T22:34:04-03:00/home/"
	entries, err := restic.GetDirEntries(path1)
	if err != nil {
		panic(err)
	}
	if len(entries) != 1 {
		panic("Expected 1 entry")
	}

	p := tea.NewProgram(
		compare.InitialModel(entries[0], entries[0]),
	)

	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	//for _, entry := range entries {
	//	PrintNodes(entry, 0)
	//}
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
