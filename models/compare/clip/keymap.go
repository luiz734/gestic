package clip

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	CopyOne   key.Binding
	CopyTwo   key.Binding
	CopyThree key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.CopyOne, k.CopyTwo, k.CopyThree}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.CopyOne, k.CopyTwo}, // first column
		{k.CopyThree},          // second column
	}
}

// DefaultKeyMap returns a set of pager-like default keybindings.
func DefaultKeyMap() keymap {
	return keymap{
		CopyOne: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "Copy [1]"),
		),
		CopyTwo: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "Copy [2]"),
		),
		CopyThree: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "Copy [3]"),
		),
	}
}
