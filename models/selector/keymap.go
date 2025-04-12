package selector

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	Quit   key.Binding
	Help   key.Binding
	Select key.Binding
	Clear  key.Binding
	Accept key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Clear, k.Accept}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Select, k.Help}, // first column
		{k.Clear, k.Quit},  // second column
		{k.Accept},
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
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("<space>", "Select"),
		),
		Clear: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("<backspace>", "Clear"),
		),
		Accept: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("<enter>", "Open repositories"),
		),
	}
}
