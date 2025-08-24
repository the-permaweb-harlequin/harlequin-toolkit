package components

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigField represents a single configuration field
type ConfigField struct {
	Name        string
	Value       string
	Description string
	FieldType   FieldType
}

// FieldType represents the type of input field
type FieldType int

const (
	FieldTypeText FieldType = iota
	FieldTypeSelector
)

// ConfigEditorComponent provides a reusable configuration editor
type ConfigEditorComponent struct {
	fields            []ConfigField
	selectedIndex     int
	isEditingText     bool
	cursorVisible     bool
	width             int
	height            int
	wasmOptions       []string // For WASM target selector
	wasmSelectedIndex int
}

// NewConfigEditor creates a new configuration editor
func NewConfigEditor() *ConfigEditorComponent {
	return &ConfigEditorComponent{
		fields: []ConfigField{
			{Name: "WASM Target", Value: "1", Description: "WASM target architecture\n\nSelect between 32-bit and 64-bit WASM compilation targets. Use ←/→ to toggle between options.", FieldType: FieldTypeSelector},
			{Name: "Stack Size (MB)", Value: "3.0", Description: "Stack size in megabytes (MB) for the WASM runtime", FieldType: FieldTypeText},
			{Name: "Initial Memory (MB)", Value: "5.0", Description: "Initial memory allocation in megabytes (MB)", FieldType: FieldTypeText},
			{Name: "Maximum Memory (MB)", Value: "512.0", Description: "Maximum memory limit in megabytes (MB)", FieldType: FieldTypeText},
		},
		wasmOptions:       []string{"WASM 32", "WASM 64"},
		wasmSelectedIndex: 0,
		width:             41,
		height:            12,
		isEditingText:     true, // Auto-enable editing
	}
}

// SetSize updates the component dimensions
func (c *ConfigEditorComponent) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// SetFieldValues sets the field values from config
func (c *ConfigEditorComponent) SetFieldValues(target int, stackSizeMB, initialMemoryMB, maxMemoryMB float64) {
	targetValue := "1" // Default to WASM 32
	if target == 64 {
		targetValue = "2" // WASM 64
		c.wasmSelectedIndex = 1
	} else {
		c.wasmSelectedIndex = 0
	}

	c.fields[0].Value = targetValue
	c.fields[1].Value = fmt.Sprintf("%.1f", stackSizeMB)
	c.fields[2].Value = fmt.Sprintf("%.1f", initialMemoryMB)
	c.fields[3].Value = fmt.Sprintf("%.1f", maxMemoryMB)
}

// GetFieldValues returns the parsed field values
func (c *ConfigEditorComponent) GetFieldValues() (target int, stackSize, initialMemory, maxMemory int, err error) {
	// Convert selector value back to actual target value
	if c.fields[0].Value == "1" {
		target = 32 // WASM 32-bit
	} else {
		target = 64 // WASM 64-bit
	}

	// Convert MB values back to bytes
	if stackSizeMB, err := strconv.ParseFloat(c.fields[1].Value, 64); err == nil {
		stackSize = int(stackSizeMB * 1024 * 1024)
	} else {
		return 0, 0, 0, 0, fmt.Errorf("invalid stack size: %s", c.fields[1].Value)
	}

	if initialMemoryMB, err := strconv.ParseFloat(c.fields[2].Value, 64); err == nil {
		initialMemory = int(initialMemoryMB * 1024 * 1024)
	} else {
		return 0, 0, 0, 0, fmt.Errorf("invalid initial memory: %s", c.fields[2].Value)
	}

	if maximumMemoryMB, err := strconv.ParseFloat(c.fields[3].Value, 64); err == nil {
		maxMemory = int(maximumMemoryMB * 1024 * 1024)
	} else {
		return 0, 0, 0, 0, fmt.Errorf("invalid maximum memory: %s", c.fields[3].Value)
	}

	return target, stackSize, initialMemory, maxMemory, nil
}

// HandleKeyPress processes navigation and editing keys
func (c *ConfigEditorComponent) HandleKeyPress(key string) bool {
	switch key {
	case "up", "k":
		if c.selectedIndex >= 4 {
			// From buttons back to last input field
			c.selectedIndex = 3
			c.isEditingText = true
		} else if c.selectedIndex > 0 {
			// Navigate up through input fields
			c.selectedIndex--
			c.isEditingText = true
		}
		return true

	case "down", "j":
		if c.selectedIndex < 3 {
			// Navigate down through input fields
			c.selectedIndex++
			c.isEditingText = true
		} else if c.selectedIndex == 3 {
			// From last input field to first button
			c.selectedIndex = 4
			c.isEditingText = false
		}
		return true

	case "left", "h":
		if c.selectedIndex == 0 {
			// Handle Target field selector - switch to previous option
			if c.fields[0].Value == "2" {
				c.fields[0].Value = "1" // Switch to WASM 32-bit
				c.wasmSelectedIndex = 0
			}
		} else if c.selectedIndex >= 4 {
			// In button container, move left (Cancel = 5, Save = 4)
			c.selectedIndex = 5
		}
		return true

	case "right", "l":
		if c.selectedIndex == 0 {
			// Handle Target field selector - switch to next option
			if c.fields[0].Value == "1" {
				c.fields[0].Value = "2" // Switch to WASM 64-bit
				c.wasmSelectedIndex = 1
			}
		} else if c.selectedIndex >= 4 {
			// In button container, move right (Cancel = 5, Save = 4)
			c.selectedIndex = 4
		}
		return true

	case "tab":
		if c.selectedIndex < 4 {
			// From input fields to button container
			c.selectedIndex = 4
		}
		return true

	case "backspace":
		if c.isEditingText && c.selectedIndex > 0 && c.selectedIndex < 4 {
			// Handle backspace in text input (skip Target field which is index 0)
			if len(c.fields[c.selectedIndex].Value) > 0 {
				c.fields[c.selectedIndex].Value = c.fields[c.selectedIndex].Value[:len(c.fields[c.selectedIndex].Value)-1]
			}
		}
		return true
	}

	// Handle character input for non-Target fields
	if c.isEditingText && c.selectedIndex > 0 && c.selectedIndex < 4 && len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
		c.fields[c.selectedIndex].Value += key
		return true
	}

	return false
}

// SetCursorVisible updates cursor visibility for blinking animation
func (c *ConfigEditorComponent) SetCursorVisible(visible bool) {
	c.cursorVisible = visible
}

// GetSelectedIndex returns the currently selected field/button index
func (c *ConfigEditorComponent) GetSelectedIndex() int {
	return c.selectedIndex
}

// GetCurrentDescription returns the description for the currently selected field
func (c *ConfigEditorComponent) GetCurrentDescription() (string, string) {
	if c.selectedIndex < len(c.fields) {
		return c.fields[c.selectedIndex].Name, c.fields[c.selectedIndex].Description
	} else if c.selectedIndex == 4 {
		return "Save & Build", "Save the configuration changes and start the build process"
	} else if c.selectedIndex == 5 {
		return "Cancel", "Cancel editing and return to configuration review"
	}
	return "Configuration", "Edit the build configuration values"
}

// Update handles Bubble Tea messages
func (c *ConfigEditorComponent) Update(msg tea.Msg) tea.Cmd {
	return nil
}

// View renders the configuration editor panel
func (c *ConfigEditorComponent) View() string {
	var content string

	// Render each field
	for i, field := range c.fields {
		// Label styling
		var label string
		if i == c.selectedIndex {
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#874BFD")).
				Bold(true).
				Align(lipgloss.Left).
				Padding(0, 2, 0, 0)
			label = labelStyle.Render(field.Name)
		} else {
			labelStyle := lipgloss.NewStyle().
				Align(lipgloss.Left).
				Padding(0, 2, 0, 0)
			label = labelStyle.Render(field.Name)
		}

		// Input styling - different behavior for Target field (selector) vs others (text input)
		var input string
		if i == 0 { // Target field - selector
			// Create both WASM 32 and WASM 64 options
			var wasm32, wasm64 string

			if c.wasmSelectedIndex == 0 {
				// WASM 32 is selected
				wasm32Style := lipgloss.NewStyle().
					Background(lipgloss.Color("#874BFD")).
					Foreground(lipgloss.Color("#FFFFFF")).
					Padding(0, 1)
				wasm32 = wasm32Style.Render("WASM 32")

				wasm64Style := lipgloss.NewStyle().
					Background(lipgloss.Color("#444")).
					Foreground(lipgloss.Color("#AAA")).
					Padding(0, 1)
				wasm64 = wasm64Style.Render("WASM 64")
			} else {
				// WASM 64 is selected
				wasm32Style := lipgloss.NewStyle().
					Background(lipgloss.Color("#444")).
					Foreground(lipgloss.Color("#AAA")).
					Padding(0, 1)
				wasm32 = wasm32Style.Render("WASM 32")

				wasm64Style := lipgloss.NewStyle().
					Background(lipgloss.Color("#874BFD")).
					Foreground(lipgloss.Color("#FFFFFF")).
					Padding(0, 1)
				wasm64 = wasm64Style.Render("WASM 64")
			}

			// Join the options horizontally
			optionsGroup := lipgloss.JoinHorizontal(lipgloss.Left, wasm32, " ", wasm64)
			input = optionsGroup
		} else { // Other fields - text input
			if i == c.selectedIndex {
				// Selected input: show blinking cursor (auto-editing)
				var cursor string
				if c.cursorVisible {
					cursor = "┃" // Visible cursor
				} else {
					cursor = " " // Invisible cursor (blink effect)
				}
				inputText := field.Value + cursor

				inputStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#874BFD")).
					Align(lipgloss.Left)
				input = inputStyle.Render(inputText)
			} else {
				// Unselected input: simple text
				input = field.Value
			}
		}

		// Create label/input group with proper spacing and alignment
		labelInputGroup := lipgloss.JoinHorizontal(lipgloss.Left, label, input)

		// Container for the label/input group with border
		var containerStyle lipgloss.Style
		if i == c.selectedIndex {
			containerStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#874BFD")).
				Padding(0, 1).
				Width(35).
				Height(1).
				AlignVertical(lipgloss.Center).
				AlignHorizontal(lipgloss.Left)
		} else {
			containerStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#666")).
				Padding(0, 1).
				Width(35).
				Height(1).
				AlignVertical(lipgloss.Center).
				AlignHorizontal(lipgloss.Left)
		}

		containerRow := containerStyle.Render(labelInputGroup)
		content += containerRow + "\n"
	}

	// Add centered horizontal button container at bottom
	buttonAreaWidth := 37
	buttonWidth := (buttonAreaWidth - 1) / 2

	var cancelBtn, saveBtn string

	if c.selectedIndex == 5 {
		// Cancel selected
		cancelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#874BFD")).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 1).
			Width(buttonWidth).
			Align(lipgloss.Center)
		cancelBtn = cancelStyle.Render("Cancel")
	} else {
		cancelStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666")).
			Padding(0, 1).
			Width(buttonWidth).
			Align(lipgloss.Center)
		cancelBtn = cancelStyle.Render("Cancel")
	}

	if c.selectedIndex == 4 {
		// Save & Build selected
		saveStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#874BFD")).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 1).
			Width(buttonWidth).
			Align(lipgloss.Center)
		saveBtn = saveStyle.Render("Save & Build")
	} else {
		saveStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666")).
			Padding(0, 1).
			Width(buttonWidth).
			Align(lipgloss.Center)
		saveBtn = saveStyle.Render("Save & Build")
	}

	// Join buttons horizontally and center them
	buttonContainer := lipgloss.JoinHorizontal(lipgloss.Top, cancelBtn, " ", saveBtn)
	centeredButtons := lipgloss.NewStyle().
		Width(37).
		Align(lipgloss.Center).
		Render(buttonContainer)

	// Center the fields content and use JoinVertical to stick buttons to bottom
	centeredFields := lipgloss.NewStyle().
		Width(37).
		Align(lipgloss.Center).
		Render(content)

	// Use JoinVertical with spacing to separate fields from buttons
	finalContent := lipgloss.JoinVertical(lipgloss.Center, centeredFields, centeredButtons)

	// Create bordered panel
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(0, 1).
		Width(c.width).
		Height(c.height).
		Render(finalContent)
}
