package selector

import (
	"fmt"
	"gestic/models/compare"
	"gestic/restic"
	"github.com/charmbracelet/bubbles/spinner"
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
	spinner     spinner.Model
	waiting     bool
}

type SnapshotSelectionMsg struct {
	Newer restic.DirData
	Older restic.DirData
}

func (m Model) LoadSnapshots() tea.Msg {
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
	return SnapshotSelectionMsg{
		Newer: newEntries[0],
		Older: oldEntries[0],
	}
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
	spin := spinner.New()
	spin.Spinner = spinner.Line
	m := Model{
		snapshots:   s,
		snapshotNew: -1,
		snapshotOld: -1,
		table: table.New(
			table.WithColumns(columns),
			table.WithFocused(true),
			table.WithHeight(10),
		),
		spinner: spin,
		waiting: false,
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
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case SnapshotSelectionMsg:
		metadata := restic.SnapshotsMetadata{
			NewerFullPath: m.snapshots[m.snapshotNew].Path,
			NewerId:       m.snapshots[m.snapshotNew].Id,
			OlderFullPath: m.snapshots[m.snapshotOld].Path,
			OlderId:       m.snapshots[m.snapshotOld].Id,
		}
		compareModel := compare.InitialModel(nil, m.width, m.height, msg.Newer, msg.Older, metadata)
		return compareModel, tea.Batch(
			compareModel.Init(),
		)

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
			m.waiting = true
			return m, tea.Batch(m.spinner.Tick, m.LoadSnapshots)
		}
	}
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
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

	if m.waiting {
		output.WriteString(fmt.Sprintf("\n\n%s Loading repositories", m.spinner.View()))
	}

	return output.String()
}
