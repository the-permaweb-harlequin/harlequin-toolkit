package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SelectorComponent provides a reusable option selector panel
type SelectorComponent struct {
	options       []string
	selectedIndex int
	width         int
	height        int
	title         string
}

// NewSelector creates a new selector component
func NewSelector(title string, options []string) *SelectorComponent {
	return &SelectorComponent{
		title:         title,
		options:       options,
		selectedIndex: 0,
		width:         41, // Default width
		height:        12, // Default height
	}
}

// SetOptions updates the available options
func (s *SelectorComponent) SetOptions(options []string) {
	s.options = options
	if s.selectedIndex >= len(options) {
		s.selectedIndex = 0
	}
}

// SetSize updates the component dimensions
func (s *SelectorComponent) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetSelectedIndex sets the currently selected option
func (s *SelectorComponent) SetSelectedIndex(index int) {
	if index >= 0 && index < len(s.options) {
		s.selectedIndex = index
	}
}

// GetSelectedIndex returns the currently selected index
func (s *SelectorComponent) GetSelectedIndex() int {
	return s.selectedIndex
}

// GetSelectedOption returns the currently selected option
func (s *SelectorComponent) GetSelectedOption() string {
	if s.selectedIndex < len(s.options) {
		return s.options[s.selectedIndex]
	}
	return ""
}

// HandleKeyPress processes navigation keys and returns true if handled
func (s *SelectorComponent) HandleKeyPress(key string) bool {
	switch key {
	case "up", "k":
		if s.selectedIndex > 0 {
			s.selectedIndex--
		}
		return true
	case "down", "j":
		if s.selectedIndex < len(s.options)-1 {
			s.selectedIndex++
		}
		return true
	}
	return false
}

// Update handles Bubble Tea messages (for future extensibility)
func (s *SelectorComponent) Update(msg tea.Msg) tea.Cmd {
	return nil
}

// View renders the selector panel
func (s *SelectorComponent) View() string {
	content := ""
	
	// Add title if provided
	if s.title != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#874BFD")).
			Margin(0, 0, 1, 0)
		content += titleStyle.Render(s.title) + "\n"
	}
	
	// Add options with highlighting for selected
	for i, option := range s.options {
		if i == s.selectedIndex {
			// Selected option: highlighted and underlined
			selectedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#874BFD")).
				Bold(true).
				Underline(true)
			content += "❯ " + selectedStyle.Render(option) + "\n"
		} else {
			content += "  " + option + "\n"
		}
	}
	
	// Create bordered panel
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(0, 1).
		Width(s.width).
		Height(s.height).
		Render(content)
}
