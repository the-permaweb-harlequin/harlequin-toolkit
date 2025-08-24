package components

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the TUI application
type KeyMap struct {
	// Navigation
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding

	// Actions
	Enter  key.Binding
	Select key.Binding
	Back   key.Binding
	Tab    key.Binding

	// Editing
	Edit   key.Binding
	Save   key.Binding
	Cancel key.Binding

	// Global
	Help  key.Binding
	Quit  key.Binding
	Debug key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),

		// Actions
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select/confirm"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select item"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "go back"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next section"),
		),

		// Editing
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Save: key.NewBinding(
			key.WithKeys("s", "ctrl+s"),
			key.WithHelp("s", "save"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc", "cancel"),
		),

		// Global
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "show help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Debug: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "toggle debug"),
		),
	}
}

// ShortHelp returns a slice of key bindings to be shown in the short help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Back, k.Quit}
}

// FullHelp returns all key bindings to be shown in the full help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},    // Navigation
		{k.Enter, k.Select, k.Back, k.Tab}, // Actions
		{k.Edit, k.Save, k.Cancel},         // Editing
		{k.Help, k.Quit, k.Debug},          // Global
	}
}

// ConfigEditKeyMap returns key bindings specific to config editing
func ConfigEditKeyMap() KeyMap {
	km := DefaultKeyMap()

	// Override some bindings for config editing context
	km.Save = key.NewBinding(
		key.WithKeys("ctrl+s", "enter"),
		key.WithHelp("ctrl+s/enter", "save & build"),
	)

	return km
}

// BuildRunningKeyMap returns key bindings for the build running state
func BuildRunningKeyMap() KeyMap {
	km := DefaultKeyMap()

	// Disable most actions during build
	km.Up = key.NewBinding(key.WithDisabled())
	km.Down = key.NewBinding(key.WithDisabled())
	km.Select = key.NewBinding(key.WithDisabled())
	km.Edit = key.NewBinding(key.WithDisabled())

	return km
}

// ResultKeyMap returns key bindings for success/error screens
func ResultKeyMap() KeyMap {
	return KeyMap{
		Enter: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "exit"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/esc", "exit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "show help"),
		),
	}
}
