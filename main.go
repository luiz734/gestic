package main

import (
	// "gestic/models/selector"
	"fmt"
	"gestic/restic"
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

	// p := tea.NewProgram(
	// 	selector.InitialModel(
	// 		selector.Model{},
	// 		snapshots),
	// )
	// 	fmt.Printf("Alas, there's been an error: %v", err)
	// 	os.Exit(1)
	//
	path1 := "/home/tohru/tmp/restic/snapshots/2025-03-31T22:34:04-03:00/home/tohru"
	entries, err := restic.GetDirEntries(path1)
	if err != nil {
		panic(err)
	}
	for _, e := range entries {
		fmt.Printf("[%s] \t%s\n", e.SizeRadable, e.PathReadable)
	}

}
