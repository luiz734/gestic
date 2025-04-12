package compare

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	NextDir   key.Binding
	PrevDir   key.Binding
	Clipboard key.Binding
	Quit      key.Binding
	Help      key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextDir, k.PrevDir, k.Clipboard}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextDir, k.Help}, // first column
		{k.PrevDir, k.Quit}, // second column
		{k.Clipboard},
	}
}

// DefaultKeyMap returns a set of pager-like default keybindings.
func DefaultKeyMap() keymap {
	return keymap{
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c", "quit"),
		),
		NextDir: key.NewBinding(
			key.WithKeys("l", "right", "enter"),
			key.WithHelp("l/right", "Open"),
		),
		PrevDir: key.NewBinding(
			key.WithKeys("h", "left", "backspace"),
			key.WithHelp("h/left", "Back"),
		),
		Clipboard: key.NewBinding(
			key.WithKeys("1", "2", "3"),
			key.WithHelp("1,2,3", "Copy"),
		),
	}
}
