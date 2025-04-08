package ui

import (
	"fmt"
	"gestic/restic"
	"strings"

	"github.com/charmbracelet/bubbletea"
)

type Model struct {
	snapshots []restic.Snapshot
	first     int
	second    int
	cursor    int

	debug string
}

func InitialModel(s []restic.Snapshot) Model {
	m := Model{
		snapshots: s,
		first:     -1,
		second:    -1,
		cursor:    len(s) - 1,
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

		case "backspace":
			m.first = -1
			m.second = -1
			return m, nil
		case " ":
			if m.first == -1 {
				m.first = m.cursor
			} else {
				m.second = m.cursor
			}
			return m, nil

		default:
			m.debug = fmt.Sprintf("%#v", msg.String())
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

	var footer string
	if m.first != -1 {
		footer += fmt.Sprintf("\n%s", m.snapshots[m.first].Path)
	}
	if m.second != -1 {
		footer += fmt.Sprintf("\n%s", m.snapshots[m.second].Path)
	}
	output.WriteString(footer)


	output.WriteString(fmt.Sprintf("\n\nDEBUG: %s", m.debug))

	return output.String()
}
