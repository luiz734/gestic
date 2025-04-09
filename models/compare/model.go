package compare

import (
	"fmt"
	"gestic/restic"
	"github.com/golang-collections/collections/stack"
	"strings"

	"github.com/charmbracelet/bubbletea"
)

type Model struct {
	largerDir  restic.DirData
	smallerDir restic.DirData
	dirStack   *stack.Stack
	cursor     int
}

func InitialModel(l, s restic.DirData) Model {
	dirStack := stack.New()
	dirStack.Push(l)
	m := Model{
		largerDir:  l,
		smallerDir: s,
		dirStack:   dirStack,
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
			m.dirStack.Push(m.largerDir)
			m.largerDir = childDir
			return m, nil
		case "h":
			if m.dirStack.Len() > 1 {
				parentDir := m.dirStack.Pop().(restic.DirData)
				m.largerDir = parentDir
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

	maxLines := 10
	startIndex := max(min(m.cursor-maxLines/2, len(m.largerDir.Children)-maxLines), 0)
	endIndex := min(len(m.largerDir.Children)-1, startIndex+maxLines)
	i := startIndex

	for i <= endIndex {
		e := m.largerDir.Children[i]
		if i == m.cursor {
			output.WriteString(fmt.Sprintf("> [%s] %s\n", e.SizeRadable, e.PathReadable))
		} else {
			output.WriteString(fmt.Sprintf("  [%s] %s\n", e.SizeRadable, e.PathReadable))
		}
		i += 1
	}
	var footer string
	footer += fmt.Sprintf("\nCursor: %d", m.cursor)
	output.WriteString(footer)

	return output.String()
}
