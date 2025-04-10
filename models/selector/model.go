package selector

import (
	"fmt"
	"gestic/models/compare"
	"gestic/restic"
	"strings"

	"github.com/charmbracelet/bubbletea"
)

type Model struct {
	width       int
	height      int
	snapshots   []restic.Snapshot
	snapshotNew int
	snapshotOld int
	cursor      int
	debug       string
}

type SnapshotSelectionMsg struct {
	Newer restic.Snapshot
	Older restic.Snapshot
}

func (m Model) advanceToCompare() (tea.Model, tea.Cmd) {
	newEntries, err := restic.GetDirEntries(m.snapshots[m.snapshotNew].Path)
	if err != nil {
		panic(fmt.Errorf("error getting dir newEntries: %w", err))
	}
	oldEntries, err := restic.GetDirEntries(m.snapshots[m.snapshotOld].Path)
	if err != nil {
		panic(fmt.Errorf("error getting dir newEntries: %w", err))
	}
	if len(newEntries) != 1 || len(oldEntries) != 1 {
		panic(fmt.Errorf("root directory should contain 1 child: %w", err))
	}
	compareModel := compare.InitialModel(nil, m.width, m.height, newEntries[0], oldEntries[0])
	return compareModel, tea.Batch(
		compareModel.Init(),
	)
}

func InitialModel(s []restic.Snapshot) Model {
	m := Model{
		snapshots:   s,
		snapshotNew: -1,
		snapshotOld: -1,
		cursor:      len(s) - 1,
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.ClearScreen,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
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
			if m.snapshotNew == -1 {
				m.snapshotNew = m.cursor
			} else if m.snapshotNew != m.cursor {
				m.snapshotOld = m.cursor
			}
			return m, nil
		case "backspace":
			m.snapshotNew = -1
			m.snapshotOld = -1
			return m, nil
		case "enter":
			if m.snapshotNew == -1 || m.snapshotOld == -1 {
				return m, nil
			}
			return m.advanceToCompare()
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
		if m.snapshotNew == index {
			lineEnd = " [1]"
		} else if m.snapshotOld == index {
			lineEnd = " [2]"
		}

		if index == m.cursor {
			output.WriteString(fmt.Sprintf(">%s%s\n", s, lineEnd))
		} else {
			output.WriteString(fmt.Sprintf(" %s%s\n", s, lineEnd))
		}
	}

	var footer string
	if m.snapshotNew != -1 {
		footer += fmt.Sprintf("\n%s %s", "[1]", m.snapshots[m.snapshotNew].Path)
	}
	if m.snapshotOld != -1 {
		footer += fmt.Sprintf("\n%s %s", "[2]", m.snapshots[m.snapshotOld].Path)
	}
	output.WriteString(footer)

	output.WriteString(fmt.Sprintf("\n\nDEBUG: %s", m.debug))

	return output.String()
}
