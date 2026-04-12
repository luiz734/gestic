package clip

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"
	//"golang.design/x/clipboard"
)

type Model struct {
	keymap keymap
	debug  bool

	timer timer.Model

	blinkStyle  lipgloss.Style
	activeIndex int
	rows        []string
}
type CopyMsg int
type BlinkStartMsg int
type BlinkFinishMsg struct{}

type UpdateClipboardMsg struct {
	First  string
	Second string
	Third  string
}

var timeout = time.Millisecond * 100
var interval = time.Millisecond * 5

func InitialModel() Model {
	return Model{
		keymap:      DefaultKeyMap(),
		debug:       false,
		timer:       timer.New(-1),
		blinkStyle:  defaultStyle,
		activeIndex: -1,
		rows:        []string{"not set", "not set", "not set"},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.CopyOne, m.keymap.CopyTwo, m.keymap.CopyThree):
			targetClipboard, err := strconv.Atoi(msg.String())
			if err != nil {
				panic(err)
			}
			return m, func() tea.Msg {
				return CopyMsg(targetClipboard)
			}
		}

	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		return m, func() tea.Msg { return BlinkFinishMsg{} }

	case CopyMsg:
		blinkStartCmd := func() tea.Msg {
			return BlinkStartMsg(msg)
		}

		m.timer = timer.NewWithInterval(timeout, interval)
		return m, tea.Batch(m.timer.Init(), blinkStartCmd)

	case BlinkStartMsg:
		m.activeIndex = int(msg)
		m.blinkStyle = blinkStyle

	case BlinkFinishMsg:
		// Slice starts at 0
		m.activeIndex -= 1
		if m.activeIndex < len(m.rows) {
			err := clipboard.Init()
			if err != nil {
				panic(err)
			}
			clipboard.Write(clipboard.FmtText, []byte(m.rows[m.activeIndex]))
		}
		log.Printf("Clipboard copied: %s", m.rows[m.activeIndex])
		m.activeIndex = -1
		m.blinkStyle = defaultStyle

	case UpdateClipboardMsg:
		m.rows = []string{msg.First, msg.Second, msg.Third}
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	var output strings.Builder
	output.WriteString("\n\n")

	for index, c := range m.rows {
		if index+1 == m.activeIndex {
			output.WriteString(m.blinkStyle.Render(fmt.Sprintf("[%d] %s", index+1, c)))
		} else {
			output.WriteString(defaultStyle.Render(fmt.Sprintf("[%d] %s", index+1, c)))
		}
		output.WriteString("\n")
	}

	if m.debug {
		output.WriteString(fmt.Sprintf("\n\n---\nDEBUG"))
		output.WriteString(fmt.Sprintf("\n\n%d", m.activeIndex))
		output.WriteString(fmt.Sprintf("\n\n%s", m.timer.View()))
		output.WriteString(fmt.Sprintf("\n\n%v", m.timer.Running()))
	}

	return output.String()
}
