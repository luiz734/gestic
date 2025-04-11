package selector

import (
	"fmt"
	"gestic/models/compare"
	"gestic/restic"
	"github.com/charmbracelet/bubbles/table"
	"strings"

	"github.com/charmbracelet/bubbletea"
)

type Model struct {
	width       int
	height      int
	snapshots   []restic.Snapshot
	snapshotNew int
	snapshotOld int
	table       table.Model
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
	metadata := restic.SnapshotsMetadata{
		NewerFullPath: m.snapshots[m.snapshotNew].Path,
		NewerId:       m.snapshots[m.snapshotNew].Id,
		OlderFullPath: m.snapshots[m.snapshotOld].Path,
		OlderId:       m.snapshots[m.snapshotOld].Id,
	}
	compareModel := compare.InitialModel(nil, m.width, m.height, newEntries[0], oldEntries[0], metadata)
	return compareModel, tea.Batch(
		compareModel.Init(),
	)
}

func (m Model) UpdateRows() []table.Row {
	var t []table.Row

	for index, s := range m.snapshots {
		checked := " "
		switch index {
		case m.snapshotNew:
			checked = "1"
		case m.snapshotOld:
			checked = "2"
		}
		t = append(t, []string{
			checked,
			s.Id,
			s.Date.Format("2006-01-02 15:04:05"),
			s.SizeStr,
		})
	}

	return t
}
func InitialModel(s []restic.Snapshot) Model {
	columns := []table.Column{
		{Title: " ", Width: 1},
		{Title: "Id", Width: 10},
		{Title: "Date", Width: 20},
		{Title: "Size", Width: 10},
	}
	m := Model{
		snapshots:   s,
		snapshotNew: -1,
		snapshotOld: -1,
		table: table.New(
			table.WithColumns(columns),
			table.WithFocused(true),
			table.WithHeight(10),
		),
	}
	m.table.SetRows(m.UpdateRows())
	m.table.GotoBottom()
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
		case " ":
			if m.snapshotNew == -1 {
				m.snapshotNew = m.table.Cursor()
			} else if m.snapshotNew != m.table.Cursor() {
				m.snapshotOld = m.table.Cursor()
			}
			m.table.SetRows(m.UpdateRows())
			return m, nil
		case "backspace":
			m.snapshotNew = -1
			m.snapshotOld = -1
			m.table.SetRows(m.UpdateRows())
			return m, nil
		case "enter":
			if m.snapshotNew == -1 || m.snapshotOld == -1 {
				return m, nil
			}
			return m.advanceToCompare()
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var output strings.Builder

	output.WriteString(m.table.View())
	output.WriteString("\n")

	var footer string
	if m.snapshotNew != -1 {
		footer += fmt.Sprintf("\n%s %s", "[1]", m.snapshots[m.snapshotNew].Path)
	}
	if m.snapshotOld != -1 {
		footer += fmt.Sprintf("\n%s %s", "[2]", m.snapshots[m.snapshotOld].Path)
	}
	output.WriteString(footer)

	return output.String()
}
