package compare

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gestic/models/compare/clip"
	"gestic/restic"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
)

const MaxColSize = 6
const ViewportHeight = 12

type Row struct {
	dirA    *restic.DirData
	dirB    *restic.DirData
	absDiff uint64
	diff    int
}

type Model struct {
	prevModel tea.Model
	clipModel tea.Model
	help      help.Model
	keyMap    keymap
	width     int
	height    int

	metadata restic.SnapshotsMetadata
	rows     []Row
	table    table.Model
}

func InitialModel(prevModel tea.Model, width, height int, dirNew, dirOld *restic.DirData, metadata restic.SnapshotsMetadata) *Model {
	columns := []table.Column{
		{Title: "New", Width: 20},
		{Title: "Old", Width: 20},
		{Title: "Diff", Width: 10},
	}
	rows := CreateRows(dirNew, dirOld, metadata)
	m := Model{
		prevModel: prevModel,
		clipModel: clip.InitialModel(),
		help:      help.New(),
		keyMap:    DefaultKeyMap(),
		width:     width,
		height:    height,
		rows:      rows,
		metadata:  metadata,
		table: table.New(
			table.WithColumns(columns),
			table.WithFocused(true),
			table.WithHeight(ViewportHeight),
			table.WithStyles(tableStyles),
		),
	}
	m = *m.updateTable(-1)
	return &m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.ClearScreen,
		func() tea.Msg { return tea.WindowSizeMsg{Width: m.width, Height: m.height} },
		m.updateClipboardCmd,
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		c1Width := int(math.Floor(float64(m.width) * 0.4))
		c2Width := int(math.Ceil(float64(m.width) * 0.4))
		c3Width := m.width - c1Width - c2Width

		columns := []table.Column{
			{Title: fmt.Sprintf("--- New (%s) ---", m.metadata.NewerId), Width: c1Width},
			{Title: fmt.Sprintf("--- Old (%s) ---", m.metadata.OlderId), Width: c2Width},
			{Title: "---  Diff ---", Width: c3Width},
		}

		m.table.SetColumns(columns)
		m.table.SetHeight(ViewportHeight)

		return m.updateTable(m.table.Cursor()), nil

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
			nextModel := InitialModel(m, m.width, m.height, nextNewDir, nextOldDir, m.metadata)
			return nextModel, nextModel.Init()

		case key.Matches(msg, m.keyMap.PrevDir):
			// Notifies if the window have changed size

			// I am not sure if this is necessary at all
			// The idea is to preserve the window dimensions
			// when navigating back, since the parent model
			// does not know about possible changes
			// Lets keep it for now
			if m.prevModel != nil {
				return m.prevModel, func() tea.Msg {
					return tea.WindowSizeMsg{Width: m.width, Height: m.height}
				}
			}
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	oldCursor := m.table.Cursor()
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	// We trigger an update only if the cursor changes
	if m.table.Cursor() != oldCursor {
		cmds = append(cmds, m.updateClipboardCmd)
	}

	m.clipModel, cmd = m.clipModel.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
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
	output.WriteString(m.clipModel.View())
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

func (m *Model) updateClipboardCmd() tea.Msg {

	// This is relative to user files
	// E.g. /home/myuser/foo/bar
	userPath := m.metadata.NewerFullPath

	// This is relative to the snapshots
	// E.g.: /mnt/mountpoint/snapshots/DATE-TIME/home/myuser/foo/bar
	newerSnapshotPath := m.rows[m.table.Cursor()].dirA.Path
	olderSnapshotPath := m.rows[m.table.Cursor()].dirB.Path

	// If a path exists only in one directory the other one will be "???"
	// Since we can't know which one is valid, check the first one
	// If it is not valid, then the second MUST BE VALID
	validSnapshotPath := newerSnapshotPath
	if validSnapshotPath == "???" {
		validSnapshotPath = olderSnapshotPath
		userPath = m.metadata.OlderFullPath
	}
	fileSystemPath, err := filepath.Rel(userPath, validSnapshotPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Could not determine relative path of %s to %s: %s\n", userPath, validSnapshotPath, err.Error())
		os.Exit(1)
	}

	return clip.UpdateClipboardMsg{
		First:  newerSnapshotPath,
		Second: olderSnapshotPath,
		Third:  "/" + fileSystemPath,
	}

}

func renderSizePath(size, path string, col1Length int, isDir bool) (string, error) {
	s := ""
	if len(size) > col1Length {
		return "", fmt.Errorf("Column is to short to fit string %s", size)
	}
	s += strings.Repeat(" ", col1Length-len(size))
	s += size + " " + path
	return s, nil
}

func generateStringSlice(rows []Row) ([]table.Row, error) {
	var t []table.Row
	for _, r := range rows {
		signStr := "+"
		if r.diff < 0 {
			signStr = "-"
		}
		diffStr := fmt.Sprintf("%s%s", signStr, humanize.Bytes(r.absDiff))
		newerStr, err := renderSizePath(r.dirA.SizeReadable, r.dirA.PathReadable, MaxColSize, r.dirA.IsDir)
		if err != nil {
			return t, fmt.Errorf("can't generate table row for newStr: %w", err)
		}
		eqStr, err := renderSizePath(r.dirB.SizeReadable, r.dirB.PathReadable, MaxColSize, r.dirB.IsDir)
		if err != nil {
			return t, fmt.Errorf("can't generate table row for eqStr: %w", err)
		}
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
	mapA := make(map[string]*restic.DirData)
	mapB := make(map[string]*restic.DirData)
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
			rows = append(rows, Row{dirA: a, dirB: b, diff: diff, absDiff: absDiff})
		} else {
			diff := int(a.Size)
			absDiff := uint64(math.Abs(float64(a.Size)))
			rows = append(rows, Row{dirA: a, dirB: &dumbDir, diff: diff, absDiff: absDiff})
		}
	}
	for path, b := range mapB {
		if seen[path] {
			continue
		}
		diff := -int(b.Size)
		absDiff := uint64(math.Abs(float64(b.Size)))
		rows = append(rows, Row{dirA: &dumbDir, dirB: b, diff: diff, absDiff: absDiff})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].diff > rows[j].diff
	})

	return rows
}
