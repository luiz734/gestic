package compare

import (
	"fmt"
	"gestic/restic"
	"golang.design/x/clipboard"
	"math"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
)

type Model struct {
	prevModel tea.Model
	width     int
	height    int

	metadata restic.SnapshotsMetadata
	dirNew   restic.DirData
	dirOld   restic.DirData
	table    table.Model

	clipboard []string
}

func InitialModel(prevModel tea.Model, width, height int, dirNew, dirOld restic.DirData, metadata restic.SnapshotsMetadata) *Model {
	columns := []table.Column{
		{Title: "New", Width: 20},
		{Title: "Old", Width: 20},
		{Title: "Diff", Width: 10},
	}
	m := Model{
		prevModel: prevModel,
		width:     width,
		height:    height,
		dirNew:    dirNew,
		dirOld:    dirOld,
		metadata:  metadata,
		table: table.New(
			table.WithColumns(columns),
			table.WithFocused(true),
			table.WithHeight(10),
		),
	}
	m = *m.updateTable(-1)
	return &m
}

func (m *Model) updateTable(cursor int) *Model {
	rows, err := generateStringSlice(m.dirNew, m.dirOld)
	if err != nil {
		panic(err)
	}
	m.table.SetRows(rows)
	m.table.SetCursor(cursor)
	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.ClearScreen,
		func() tea.Msg { return tea.WindowSizeMsg{Width: m.width, Height: m.height} },
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		c1Width := int(math.Floor(float64(m.width) * 0.4))
		c2Width := int(math.Ceil(float64(m.width) * 0.4))
		c3Width := m.width - c1Width - c2Width
		columns := []table.Column{
			{Title: fmt.Sprintf("New (%s)", m.metadata.NewerId), Width: c1Width},
			{Title: fmt.Sprintf("Old (%s)", m.metadata.OlderId), Width: c2Width},
			{Title: "Diff", Width: c3Width},
		}
		// Restore cursor or set to 0 if there is no previous cursor
		oldCursor := m.table.Cursor()
		m.table = table.New(
			table.WithColumns(columns),
			table.WithFocused(true),
			table.WithHeight(10),
		)
		return m.updateTable(oldCursor), nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "l":
			nextNewDir := m.dirNew.Children[m.table.Cursor()]
			// Files or empty dirs
			if len(nextNewDir.Children) == 0 {
				return m, nil
			}
			nextOldDir := findEquivalent(nextNewDir, m.dirOld.Children)
			if nextOldDir == nil {
				nextOldDir = &restic.DirData{}
			}
			nextModel := InitialModel(m, m.width, m.height, nextNewDir, *nextOldDir, m.metadata)
			return nextModel, nextModel.Init()
		case "h":
			// Notifies if the window have changed size
			if m.prevModel != nil {
				return m.prevModel, func() tea.Msg {
					return tea.WindowSizeMsg{Width: m.width, Height: m.height}
				}
			}
		case "1", "2", "3":
			targetClipboard, err := strconv.Atoi(msg.String())
			if err != nil {
				panic(err)
			}
			targetClipboard -= 1
			if targetClipboard < len(m.clipboard) {
				err := clipboard.Init()
				if err != nil {
					panic(err)
				}
				clipboard.Write(clipboard.FmtText, []byte(m.clipboard[targetClipboard]))
			}
		}
	}
	m.table, cmd = m.table.Update(msg)
	// Order matters here
	m.clipboard = m.updateClipboard()
	return m, cmd
}

func (m *Model) View() string {
	var output strings.Builder

	output.WriteString(m.table.View())
	output.WriteString(m.metadataView())

	var footer string
	footer += fmt.Sprintf("\nCursor: %d", m.table.Cursor())
	//footer += fmt.Sprintf("\nDEBUG: %#v", tableData)
	output.WriteString(footer)

	return output.String()
}

func (m *Model) metadataView() string {
	var output strings.Builder
	output.WriteString("\n\n")
	for index, c := range m.clipboard {
		output.WriteString(fmt.Sprintf("[%d] %s\n", index+1, c))
	}
	return output.String()
}

func (m *Model) updateClipboard() []string {
	c := []string{
		m.dirNew.Children[m.table.Cursor()].Path,
		m.dirOld.Children[m.table.Cursor()].Path,
	}
	fileSystemPath, err := filepath.Rel(m.metadata.NewerFullPath, m.dirNew.Children[m.table.Cursor()].Path)
	if err == nil {
		c = append(c, "/"+fileSystemPath)
	}
	return c
}

func generateStringSlice(newer, older restic.DirData) ([]table.Row, error) {
	type TableData struct {
		newer   restic.DirData
		older   restic.DirData
		absDiff uint64
		diff    int
	}
	var data []TableData
	for i := range len(newer.Children) {
		n := newer.Children[i]
		eq := findEquivalent(n, older.Children)
		if eq == nil {
			eq = &restic.DirData{Size: 0, PathReadable: "<???>"}
		}
		diff := int(n.Size) - int(eq.Size)
		absDiff := uint64(math.Abs(float64(diff)))
		data = append(data, TableData{n, *eq, absDiff, diff})
	}

	for i := range len(older.Children) {
		o := older.Children[i]
		// Skip entries already in "newer"
		if slices.ContainsFunc(newer.Children, func(e restic.DirData) bool {
			return o.PathReadable == e.PathReadable
		}) {
			continue
		}
		eq := findEquivalent(o, newer.Children)
		if eq == nil {
			eq = &restic.DirData{Size: 0, PathReadable: "<???>"}
		}
		diff := int(o.Size) - int(eq.Size)
		absDiff := uint64(math.Abs(float64(diff)))
		data = append(data, TableData{o, *eq, absDiff, diff})
	}

	// TODO: sort by diff
	//slices.SortFunc(data, func(a, b TableData) int {
	//	return int(b.diff - a.diff)
	//})

	var t []table.Row
	for _, d := range data {
		signStr := "+"
		if d.diff < 0 {
			signStr = "-"
		}
		diffStr := fmt.Sprintf("%s%s", signStr, humanize.Bytes(d.absDiff))
		newerStr := fmt.Sprintf("%s %s", d.newer.SizeReadable, d.newer.PathReadable)
		eqStr := fmt.Sprintf("%s %s", d.older.SizeReadable, d.older.PathReadable)
		t = append(t, []string{newerStr, eqStr, diffStr})
	}
	return t, nil
}

func findEquivalent(like restic.DirData, options []restic.DirData) *restic.DirData {
	for _, option := range options {
		if option.PathReadable == like.PathReadable {
			return &option
		}
	}
	return nil

}
