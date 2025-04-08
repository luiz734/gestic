package ui

import (
	"fmt"
	"gestic/restic"
	"strings"

	"github.com/charmbracelet/bubbletea"
)

type Model struct {
	snapshots []restic.Snapshot
	cursor    int
}

func InitialModel(s []restic.Snapshot) Model {
	m := Model{
		snapshots: s,
		cursor:    0,
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.ClearScreen,
		// func() tea.Msg {
		// 	return tea.WindowSizeMsg{Width: m.width, Height: m.height}
		// },
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		case "j":
			m.cursor += 1
			if m.cursor > len(m.snapshots)-1 {
				m.cursor -= 1
			}
			return m, nil

		case "k":
			m.cursor -= 1
			if m.cursor < 0 {
				m.cursor += 1
			}
			return m, nil

		default:
			return m, nil
		}

	}

	return m, nil
}

func (m Model) View() string {
	var output strings.Builder

	for index, s := range m.snapshots {
		if index == m.cursor {
			output.WriteString(fmt.Sprintf(">%s", s))
		} else {
			output.WriteString(fmt.Sprintf(" %s", s))
		}
	}
	return output.String()
}
