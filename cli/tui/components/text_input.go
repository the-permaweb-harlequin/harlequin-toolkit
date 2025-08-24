package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TextInputComponent provides a simple text input with label
type TextInputComponent struct {
	input  textinput.Model
	label  string
	width  int
	height int
}

// NewTextInput creates a new text input component
func NewTextInput(label, defaultValue string, width, height int) *TextInputComponent {
	input := textinput.New()
	input.Placeholder = defaultValue
	input.SetValue(defaultValue)
	input.CharLimit = 100
	input.Width = width - 6 // Account for panel border (4) + some padding (2)
	input.Focus()           // Auto-focus for immediate editing

	return &TextInputComponent{
		input:  input,
		label:  label,
		width:  width,
		height: height,
	}
}

// SetSize updates the component dimensions
func (ti *TextInputComponent) SetSize(width, height int) {
	ti.width = width
	ti.height = height
	ti.input.Width = width - 6 // Account for panel border (4) + some padding (2)
}

// GetValue returns the current input value
func (ti *TextInputComponent) GetValue() string {
	return ti.input.Value()
}

// SetValue updates the input value
func (ti *TextInputComponent) SetValue(value string) {
	ti.input.SetValue(value)
}

// Init implements the Bubble Tea model interface
func (ti *TextInputComponent) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles Bubble Tea messages
func (ti *TextInputComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	ti.input, cmd = ti.input.Update(msg)
	return ti, cmd
}

// View renders the text input
func (ti *TextInputComponent) View() string {
	// Create label
	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#874BFD")).
		Margin(0, 0, 1, 0)

	label := labelStyle.Render(ti.label)

	// Create input without border (let panel container handle borders)
	inputField := ti.input.View()

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		MarginTop(1).
		Render("Press Enter to continue")

	// Combine all elements
	content := lipgloss.JoinVertical(lipgloss.Left,
		label,
		inputField,
		instructions,
	)

	return content
}

// CreateOutputDirInput creates a text input for output directory selection
func CreateOutputDirInput(width, height int) *TextInputComponent {
	return NewTextInput("Output Directory", "./dist", width, height)
}
