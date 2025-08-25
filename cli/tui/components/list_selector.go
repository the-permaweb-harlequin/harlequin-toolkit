package components

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListItem represents an item in our selector list
type ListItem struct {
	title       string
	description string
	value       string
}

// Implement the list.Item interface
func (i ListItem) FilterValue() string { return i.title }
func (i ListItem) Title() string       { return i.title }
func (i ListItem) Description() string { return i.description }
func (i ListItem) Value() string       { return i.value }

// ListSelectorComponent provides a feature-rich selector using Bubbles list
type ListSelectorComponent struct {
	list   list.Model
	choice string
}

// NewListSelector creates a new list-based selector
func NewListSelector(title string, items []ListItem, width, height int) *ListSelectorComponent {
	// Convert our items to list.Item interface
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	// Create delegate for custom styling
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#902f17")).
		Bold(true).
		Underline(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#564f41"))

	// Create the list model
	listModel := list.New(listItems, delegate, width, height)
	listModel.Title = title
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true) // Enable for keyboard navigation
	listModel.Styles.Title = listModel.Styles.Title.
		Foreground(lipgloss.Color("#902f17")).
		Background(lipgloss.Color("")).
		Bold(true).
		Padding(0, 0, 1, 0)

	return &ListSelectorComponent{
		list: listModel,
	}
}

// SetSize updates the list dimensions
func (ls *ListSelectorComponent) SetSize(width, height int) {
	ls.list.SetSize(width, height)
}

// SetItems updates the list items
func (ls *ListSelectorComponent) SetItems(items []ListItem) {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}
	ls.list.SetItems(listItems)
}

// GetSelected returns the currently selected item
func (ls *ListSelectorComponent) GetSelected() *ListItem {
	if item := ls.list.SelectedItem(); item != nil {
		if listItem, ok := item.(ListItem); ok {
			return &listItem
		}
	}
	return nil
}

// Init implements the Bubble Tea model interface
func (ls *ListSelectorComponent) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages
func (ls *ListSelectorComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Debug: print key presses to see if they're reaching the component
		switch msg.String() {
		case "enter":
			// Store the choice when user selects
			if selected := ls.GetSelected(); selected != nil {
				ls.choice = selected.Value()
			}
		case "up", "down", "j", "k":
			// Arrow keys should be passed through to the list
		}
	}

	var cmd tea.Cmd
	ls.list, cmd = ls.list.Update(msg)
	return ls, cmd
}

// View renders the list selector
func (ls *ListSelectorComponent) View() string {
	// Let the list render without forced width constraints
	return ls.list.View()
}

// GetChoice returns the user's choice (set when they press enter)
func (ls *ListSelectorComponent) GetChoice() string {
	return ls.choice
}

// HasChoice returns true if the user has made a selection
func (ls *ListSelectorComponent) HasChoice() bool {
	return ls.choice != ""
}

// CreateBuildTypeSelector creates a selector for build types
func CreateBuildTypeSelector(width, height int) *ListSelectorComponent {
	items := []ListItem{
		{
			title:       "AOS Flavour",
			description: "Builds a WASM binary with your Lua injected into the base AOS process",
			value:       "aos",
		},
	}

	return NewListSelector("Select Build Configuration", items, width, height)
}

// CreateEntrypointSelector creates a selector for entrypoint files
func CreateEntrypointSelector(luaFiles []string, width, height int) *ListSelectorComponent {
	items := make([]ListItem, len(luaFiles))
	for i, file := range luaFiles {
		items[i] = ListItem{
			title:       file,
			description: "Main Lua file for your project",
			value:       file,
		}
	}

	return NewListSelector("Select Entrypoint File", items, width, height)
}

// CreateEntrypointSelectorWithDiscovery creates a selector with automatic Lua file discovery
func CreateEntrypointSelectorWithDiscovery(rootDir string, width, height int) (*ListSelectorComponent, error) {
	// Discover Lua files automatically
	luaFiles, err := FindLuaFilesQuick(rootDir)
	if err != nil {
		return nil, err
	}

	// If no files found, provide helpful message
	if len(luaFiles) == 0 {
		items := []ListItem{
			{
				title:       "No Lua files found",
				description: "No .lua files found in current directory. Use manual file picker instead.",
				value:       "",
			},
		}
		return NewListSelector("Select Entrypoint File", items, width, height), nil
	}

	return CreateEntrypointSelector(luaFiles, width, height), nil
}

// CreateOutputDirSelector creates a selector for output directories
func CreateOutputDirSelector(width, height int) *ListSelectorComponent {
	items := []ListItem{
		{
			title:       "examples/dist",
			description: "Output to examples/dist (recommended)",
			value:       "examples/dist",
		},
		{
			title:       "examples/build",
			description: "Output to examples/build",
			value:       "examples/build",
		},
		{
			title:       "./dist",
			description: "Output to current directory",
			value:       "./dist",
		},
		{
			title:       "./build",
			description: "Output to current directory",
			value:       "./build",
		},
	}

	return NewListSelector("Select Output Directory", items, width, height)
}

// CreateConfigActionSelector creates a selector for config review actions
func CreateConfigActionSelector(width, height int) *ListSelectorComponent {
	items := []ListItem{
		{
			title:       "Use current config",
			description: "Proceed with the existing configuration",
			value:       "use",
		},
		{
			title:       "Edit config",
			description: "Modify configuration before building",
			value:       "edit",
		},
	}

	return NewListSelector("Configuration Review", items, width, height)
}

// CreateCommandSelector creates a selector for available commands
func CreateCommandSelector(width, height int) *ListSelectorComponent {
	items := []ListItem{
		{
			title:       "Build Project",
			description: "Build an AOS project with interactive configuration",
			value:       "build",
		},
	}

	return NewListSelector("Welcome to Harlequin", items, width, height)
}
