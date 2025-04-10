package compare

import (
	"fmt"
	"gestic/restic"
	"math"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
	"github.com/golang-collections/collections/stack"
)

type Model struct {
	width       int
	height      int
	dirNew      restic.DirData
	dirOld      restic.DirData
	rootDir     string
	stackDirNew *stack.Stack
	stackDirOld *stack.Stack
	table       table.Model
}

func InitialModel(width, height int, dirNew, dirOld restic.DirData) Model {
	stackDirNew := stack.New()
	stackDirNew.Push(dirNew)
	stackDirOld := stack.New()
	stackDirOld.Push(dirOld)

	columns := []table.Column{
		{Title: "New", Width: 20},
		{Title: "Old", Width: 20},
		{Title: "Diff", Width: 10},
	}
	m := Model{
		width:       width,
		height:      height,
		dirNew:      dirNew,
		dirOld:      dirOld,
		rootDir:     dirNew.Path,
		stackDirNew: stackDirNew,
		stackDirOld: stackDirOld,
		table: table.New(
			table.WithColumns(columns),
			table.WithFocused(true),
			table.WithHeight(10),
		),
	}
	m = m.updateTable()
	return m
}

func (m Model) updateTable() Model {
	rows, err := generateStringSlice(m.dirNew, m.dirOld)
	if err != nil {
		panic(err)
	}
	m.table.SetRows(rows)
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.ClearScreen,
		func() tea.Msg { return tea.WindowSizeMsg{Width: m.width, Height: m.height} },
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		c1Width := int(math.Floor(float64(m.width) * 0.4))
		c2Width := int(math.Ceil(float64(m.width) * 0.4))
		c3Width := m.width - c1Width - c2Width
		columns := []table.Column{
			{Title: "New", Width: c1Width},
			{Title: "Old", Width: c2Width},
			{Title: "Diff", Width: c3Width},
		}
		m.table = table.New(
			table.WithColumns(columns),
			table.WithFocused(true),
			table.WithHeight(10),
		)
		return m.updateTable(), nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "l":
			childDir := m.dirNew.Children[m.table.Cursor()]
			// Files or empty dirs
			if len(childDir.Children) == 0 {
				return m, nil
			}
			m.stackDirNew.Push(m.dirNew)
			m.dirNew = childDir
			eq := findEquivalent(childDir, m.dirOld.Children)
			if eq == nil {
				eq = &restic.DirData{}
			}
			m.stackDirOld.Push(m.dirOld)
			m.dirOld = *eq
			m.table.GotoTop()
			m = m.updateTable()
			return m, nil
		case "h":
			if m.stackDirNew.Len() > 1 {
				parentDir := m.stackDirNew.Pop().(restic.DirData)
				m.dirNew = parentDir
				smallerDir := m.stackDirOld.Pop().(restic.DirData)
				m.dirOld = smallerDir
			}
			m = m.updateTable()
			m.table.GotoTop()
			return m, nil
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("%s\n\n", m.rootDir))
	output.WriteString(m.table.View())
	output.WriteString(fmt.Sprintf("\n\n%s\n", m.dirNew.Children[m.table.Cursor()].Path))

	var footer string
	footer += fmt.Sprintf("\nCursor: %d", m.table.Cursor())
	footer += fmt.Sprintf("\nStackDir: %#v", m.stackDirNew)
	footer += fmt.Sprintf("\nStackSmaller: %#v", m.stackDirOld)
	//footer += fmt.Sprintf("\nDEBUG: %#v", tableData)
	output.WriteString(footer)

	return output.String()
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
		newerStr := fmt.Sprintf("%s %s", d.newer.SizeRadable, d.newer.PathReadable)
		eqStr := fmt.Sprintf("%s %s", d.older.SizeRadable, d.older.PathReadable)
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
