package selector

import (
	"fmt"
	"gestic/models/compare"
	"gestic/restic"
	"strings"

	"github.com/charmbracelet/bubbletea"
)

type Model struct {
	prevModel tea.Model

	snapshots []restic.Snapshot
	first     int
	second    int
	cursor    int

	debug string
}

type SnapshotSelectionMsg struct {
	first  restic.Snapshot
	second restic.Snapshot
}

func (m Model) advaceToCompare() (tea.Model, tea.Cmd) {
	entries, err := restic.GetDirEntries(m.snapshots[m.first].Path)
	if err != nil {
		panic(err)
	}
	if len(entries) != 1 {
		panic("Expected 1 entry")
	}
	entries2, err := restic.GetDirEntries(m.snapshots[m.second].Path)
	if err != nil {
		panic(err)
	}
	if len(entries2) != 1 {
		panic("Expected 1 entry")
	}
	compareModel := compare.InitialModel(entries[0], entries2[0])

	return compareModel, tea.Batch(
		compareModel.Init(),
		// m.prevModel.Init(),
		// more commands
		//func() tea.Msg {
		//	return SnapshotSelectionMsg{
		//		first:  m.snapshots[m.first],
		//		second: m.snapshots[m.second],
		//	}
	)
}

func InitialModel(s []restic.Snapshot) Model {
	m := Model{
		//prevModel: prevModel,
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

		case " ":
			if m.first == -1 {
				m.first = m.cursor
			} else if m.first != m.cursor {
				m.second = m.cursor
			}
			return m, nil
		case "backspace":
			m.first = -1
			m.second = -1
			return m, nil
		case "enter":
			if m.first == -1 || m.second == -1 {
				return m, nil
			}
			return m.advaceToCompare()

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
		var lineEnd = ""
		if m.first == index {
			lineEnd = " [1]"
		} else if m.second == index {
			lineEnd = " [2]"
		}

		if index == m.cursor {
			output.WriteString(fmt.Sprintf(">%s%s\n", s, lineEnd))
		} else {
			output.WriteString(fmt.Sprintf(" %s%s\n", s, lineEnd))
		}
	}

	var footer string
	if m.first != -1 {
		footer += fmt.Sprintf("\n%s %s", "[1]", m.snapshots[m.first].Path)
	}
	if m.second != -1 {
		footer += fmt.Sprintf("\n%s %s", "[2]", m.snapshots[m.second].Path)
	}
	output.WriteString(footer)

	output.WriteString(fmt.Sprintf("\n\nDEBUG: %s", m.debug))

	return output.String()
}
