package compare

import (
	"fmt"
	"gestic/restic"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/dustin/go-humanize"
	"github.com/golang-collections/collections/stack"
)

type Model struct {
	width        int
	height       int
	largerDir    restic.DirData
	smallerDir   restic.DirData
	rootDir      string
	dirStack     *stack.Stack
	smallerStack *stack.Stack
	cursor       int
}

func InitialModel(l, s restic.DirData) Model {
	dirStack := stack.New()
	dirStack.Push(l)
	smallerStack := stack.New()
	smallerStack.Push(s)
	m := Model{
		largerDir:    l,
		smallerDir:   s,
		rootDir:      l.Path,
		dirStack:     dirStack,
		smallerStack: smallerStack,
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
			if m.cursor > len(m.largerDir.Children)-1 {
				m.cursor -= 1
			}
			return m, nil
		case "k":
			m.cursor -= 1
			if m.cursor < 0 {
				m.cursor += 1
			}
			return m, nil
		case "l":
			childDir := m.largerDir.Children[m.cursor]
			// Files or empty dirs
			if len(childDir.Children) == 0 {
				return m, nil
			}
			m.dirStack.Push(m.largerDir)
			m.largerDir = childDir
			eq := findEquivalent(childDir, m.smallerDir.Children)
			if eq == nil {
				eq = &restic.DirData{}
			}
			m.smallerStack.Push(m.smallerDir)
			m.smallerDir = *eq
			m.cursor = 0
			return m, nil
		case "h":
			if m.dirStack.Len() > 1 {
				parentDir := m.dirStack.Pop().(restic.DirData)
				m.largerDir = parentDir
				smallerDir := m.smallerStack.Pop().(restic.DirData)
				m.smallerDir = smallerDir
			}
			return m, nil

		default:
			//m.debug = fmt.Sprintf("%#v", msg.String())
			return m, nil
		}

	}

	return m, nil
}

func (m Model) View() string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("%s\n\n", m.rootDir))

	linesVisible := 10
	startIndex := max(min(m.cursor-linesVisible/2, len(m.largerDir.Children)-linesVisible), 0)
	endIndex := min(len(m.largerDir.Children)-1, startIndex+linesVisible)

	tableData, err := generateStringSlice(
		m.largerDir.Children[startIndex:endIndex],
		m.smallerDir.Children,
	)
	if err != nil {
		panic(err)
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		Width(m.width).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == m.cursor+1-startIndex:
				return focusStyle
			default:
				return defaultStyle
			}
		}).
		Headers("NEWER", "OLDER", "DIFF").
		Rows(tableData...)

	output.WriteString(t.Render())

	output.WriteString(fmt.Sprintf("\n\n%s\n", m.largerDir.Children[m.cursor].Path))

	var footer string
	//footer += fmt.Sprintf("\nCursor: %d", m.cursor)
	//footer += fmt.Sprintf("\nStackDir: %#v", m.dirStack)
	//footer += fmt.Sprintf("\nStackSmaller: %#v", m.smallerStack)
	//footer += fmt.Sprintf("\nDEBUG: %#v", tableData)
	output.WriteString(footer)

	return output.String()
}

func generateStringSlice(newer, older []restic.DirData) ([][]string, error) {
	var t [][]string

	for i := range len(newer) {
		n := newer[i]

		eqStr := "???"
		diff := n.Size
		eq := findEquivalent(n, older)
		if eq != nil {
			greater := max(n.Size, eq.Size)
			lesser := min(n.Size, eq.Size)
			diff = greater - lesser
			eqStr = fmt.Sprintf("%s %s", eq.SizeRadable, eq.PathReadable)
		}

		diffStr := humanize.Bytes(diff)
		newerStr := fmt.Sprintf("%s %s", n.SizeRadable, n.PathReadable)
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
