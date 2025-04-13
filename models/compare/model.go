package compare

import (
	"fmt"
	"gestic/restic"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"golang.design/x/clipboard"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
)

type Row struct {
	dirA    *restic.DirData
	dirB    *restic.DirData
	absDiff uint64
	diff    int
}

type Model struct {
	prevModel tea.Model
	help      help.Model
	keyMap    keymap
	width     int
	height    int

	metadata  restic.SnapshotsMetadata
	rows      []Row
	table     table.Model
	clipboard []string
}

func InitialModel(prevModel tea.Model, width, height int, dirNew, dirOld restic.DirData, metadata restic.SnapshotsMetadata) *Model {
	columns := []table.Column{
		{Title: "New", Width: 20},
		{Title: "Old", Width: 20},
		{Title: "Diff", Width: 10},
	}
	rows := CreateRows(&dirNew, &dirOld, metadata)
	m := Model{
		prevModel: prevModel,
		help:      help.New(),
		keyMap:    DefaultKeyMap(),
		width:     width,
		height:    height,
		rows:      rows,
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
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keyMap.NextDir):
			nextNewDir := m.rows[m.table.Cursor()].dirA
			// Don't try to advance if is an empty directory or a file
			if len(nextNewDir.Children) == 0 {
				return m, nil
			}
			nextOldDir := m.rows[m.table.Cursor()].dirB
			nextModel := InitialModel(m, m.width, m.height, *nextNewDir, *nextOldDir, m.metadata)
			return nextModel, nextModel.Init()
		case key.Matches(msg, m.keyMap.PrevDir):
			// Notifies if the window have changed size
			if m.prevModel != nil {
				return m.prevModel, func() tea.Msg {
					return tea.WindowSizeMsg{Width: m.width, Height: m.height}
				}
			}
		case key.Matches(msg, m.keyMap.Clipboard):
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
	// Order matters here?
	m.clipboard = m.updateClipboard()
	return m, cmd
}

func (m *Model) View() string {
	var output strings.Builder

	output.WriteString(m.table.View())
	output.WriteString(m.metadataView())
	output.WriteString("\n")
	output.WriteString(m.help.View(m.keyMap))

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

func (m *Model) updateTable(cursor int) *Model {
	rows, err := generateStringSlice(m.rows)
	if err != nil {
		panic(err)
	}
	m.table.SetRows(rows)
	m.table.SetCursor(cursor)
	return m
}

func (m *Model) updateClipboard() []string {
	c := []string{
		m.rows[m.table.Cursor()].dirA.Path,
		m.rows[m.table.Cursor()].dirB.Path,
	}
	fileSystemPath, err := filepath.Rel(m.metadata.NewerFullPath, m.rows[m.table.Cursor()].dirA.Path)
	if err == nil {
		c = append(c, "/"+fileSystemPath)
	}
	return c
}

func generateStringSlice(rows []Row) ([]table.Row, error) {
	var t []table.Row
	for _, r := range rows {
		signStr := "+"
		if r.diff < 0 {
			signStr = "-"
		}
		diffStr := fmt.Sprintf("%s%s", signStr, humanize.Bytes(r.absDiff))
		newerStr := fmt.Sprintf("%s %s", r.dirA.SizeReadable, r.dirA.PathReadable)
		eqStr := fmt.Sprintf("%s %s", r.dirB.SizeReadable, r.dirB.PathReadable)
		t = append(t, []string{newerStr, eqStr, diffStr})
	}
	return t, nil
}

func CreateRows(dirA, dirB *restic.DirData, metadata restic.SnapshotsMetadata) []Row {
	dumbDir := restic.DirData{
		Path:         "???",
		PathReadable: "???",
		Size:         0,
	}

	// Generate maps for each directory
	mapA := make(map[string]restic.DirData)
	mapB := make(map[string]restic.DirData)
	for _, a := range dirA.Children {
		p, err := filepath.Rel(metadata.NewerFullPath, a.Path)
		if err != nil {
			panic(err)
		}
		mapA[p] = a
	}
	for _, b := range dirB.Children {
		p, err := filepath.Rel(metadata.OlderFullPath, b.Path)
		if err != nil {
			panic(err)
		}
		mapB[p] = b
	}
	var rows []Row
	seen := make(map[string]bool)

	// Generate unique entries for elements on A and B
	for path, a := range mapA {
		seen[path] = true
		if b, ok := mapB[path]; ok {
			diff := int(a.Size) - int(b.Size)
			absDiff := uint64(math.Abs(float64(diff)))
			rows = append(rows, Row{dirA: &a, dirB: &b, diff: diff, absDiff: absDiff})
		} else {
			diff := int(a.Size)
			absDiff := uint64(math.Abs(float64(a.Size)))
			rows = append(rows, Row{dirA: &a, dirB: &dumbDir, diff: diff, absDiff: absDiff})
		}
	}
	for path, b := range mapB {
		if seen[path] {
			continue
		}
		diff := -int(b.Size)
		absDiff := uint64(math.Abs(float64(b.Size)))
		rows = append(rows, Row{dirA: &dumbDir, dirB: &b, diff: diff, absDiff: absDiff})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].diff > rows[j].diff
	})

	return rows
}
