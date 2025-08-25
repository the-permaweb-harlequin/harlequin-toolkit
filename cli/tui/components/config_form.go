package components

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigFormComponent provides a professional configuration editor using Bubbles textinput
type ConfigFormComponent struct {
	inputs         []textinput.Model
	wasmTarget     int // 0 = 32-bit, 1 = 64-bit
	focused        int
	width          int
	height         int
	keyMap         KeyMap
	isInButtons    bool
	selectedButton int // 0 = Save & Build, 1 = Cancel
}

// Field definitions
var configFields = []struct {
	label       string
	placeholder string
	description string
}{
	{"Stack Size (MB)", "3.0", "Stack size in megabytes (MB) for the WASM runtime"},
	{"Initial Memory (MB)", "5.0", "Initial memory allocation in megabytes (MB)"},
	{"Maximum Memory (MB)", "512.0", "Maximum memory limit in megabytes (MB)"},
}

// NewConfigForm creates a new configuration form using textinput components
func NewConfigForm(width, height int) *ConfigFormComponent {
	inputs := make([]textinput.Model, len(configFields))

	for i, field := range configFields {
		input := textinput.New()
		input.Placeholder = field.placeholder
		input.CharLimit = 10
		input.Width = 20

		// Set validation for numeric fields
		input.Validate = func(s string) error {
			if s == "" {
				return nil // Allow empty for validation
			}
			if _, err := strconv.ParseFloat(s, 64); err != nil {
				return fmt.Errorf("must be a number")
			}
			return nil
		}

		inputs[i] = input
	}

	// Focus the first input
	if len(inputs) > 0 {
		inputs[0].Focus()
	}

	return &ConfigFormComponent{
		inputs:     inputs,
		wasmTarget: 0, // Default to 32-bit
		focused:    0,
		width:      width,
		height:     height,
		keyMap:     ConfigEditKeyMap(),
	}
}

// SetFieldValues sets the form values from configuration
func (cf *ConfigFormComponent) SetFieldValues(target int, stackSizeMB, initialMemoryMB, maxMemoryMB float64) {
	if target == 64 {
		cf.wasmTarget = 1
	} else {
		cf.wasmTarget = 0
	}

	values := []float64{stackSizeMB, initialMemoryMB, maxMemoryMB}
	for i, value := range values {
		if i < len(cf.inputs) {
			cf.inputs[i].SetValue(fmt.Sprintf("%.1f", value))
		}
	}
}

// GetFieldValues returns the parsed configuration values
func (cf *ConfigFormComponent) GetFieldValues() (target int, stackSize, initialMemory, maxMemory int, err error) {
	// WASM target
	if cf.wasmTarget == 1 {
		target = 64
	} else {
		target = 32
	}

	// Parse memory values
	values := make([]int, 3)
	fieldNames := []string{"Stack Size", "Initial Memory", "Maximum Memory"}

	for i, input := range cf.inputs {
		if i >= len(values) {
			break
		}

		valueStr := input.Value()
		if valueStr == "" {
			valueStr = input.Placeholder
		}

		valueMB, parseErr := strconv.ParseFloat(valueStr, 64)
		if parseErr != nil {
			return 0, 0, 0, 0, fmt.Errorf("invalid %s: %s", fieldNames[i], valueStr)
		}

		values[i] = int(valueMB * 1024 * 1024) // Convert MB to bytes
	}

	return target, values[0], values[1], values[2], nil
}

// Init implements the Bubble Tea model interface
func (cf *ConfigFormComponent) Init() tea.Cmd {
	return textinput.Blink
}

// SetSize updates the dimensions of the config form
func (cf *ConfigFormComponent) SetSize(width, height int) {
	cf.width = width
	cf.height = height
}

// Update handles Bubble Tea messages
func (cf *ConfigFormComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, cf.keyMap.Up):
			cf.navigateUp()
			return cf, nil
		case key.Matches(msg, cf.keyMap.Down):
			cf.navigateDown()
			return cf, nil
		case key.Matches(msg, cf.keyMap.Left):
			if cf.focused == -1 { // WASM Target selector
				cf.wasmTarget = 0 // Switch to 32-bit
			} else if cf.isInButtons {
				cf.selectedButton = 1 // Cancel
			}
			return cf, nil
		case key.Matches(msg, cf.keyMap.Right):
			if cf.focused == -1 { // WASM Target selector
				cf.wasmTarget = 1 // Switch to 64-bit
			} else if cf.isInButtons {
				cf.selectedButton = 0 // Save & Build
			}
			return cf, nil
		case key.Matches(msg, cf.keyMap.Tab):
			if !cf.isInButtons {
				cf.moveToButtons()
			}
			return cf, nil
		}
	}

	// Update the focused text input
	if cf.focused >= 0 && cf.focused < len(cf.inputs) && !cf.isInButtons {
		var cmd tea.Cmd
		cf.inputs[cf.focused], cmd = cf.inputs[cf.focused].Update(msg)
		cmds = append(cmds, cmd)
	}

	return cf, tea.Batch(cmds...)
}

// navigateUp moves focus up through the form
func (cf *ConfigFormComponent) navigateUp() {
	if cf.isInButtons {
		// From buttons to last input
		cf.isInButtons = false
		cf.focused = len(cf.inputs) - 1
		cf.focusInput(cf.focused)
	} else if cf.focused > -1 {
		// Move up through inputs, -1 is WASM target
		cf.blurCurrentInput()
		cf.focused--
		if cf.focused >= 0 {
			cf.focusInput(cf.focused)
		}
	}
}

// navigateDown moves focus down through the form
func (cf *ConfigFormComponent) navigateDown() {
	if cf.focused < len(cf.inputs)-1 {
		// Move down through inputs
		cf.blurCurrentInput()
		cf.focused++
		cf.focusInput(cf.focused)
	} else if cf.focused == len(cf.inputs)-1 {
		// From last input to buttons
		cf.moveToButtons()
	}
}

// moveToButtons moves focus to the button area
func (cf *ConfigFormComponent) moveToButtons() {
	cf.blurCurrentInput()
	cf.isInButtons = true
	cf.selectedButton = 0 // Default to Save & Build
}

// focusInput focuses the specified input
func (cf *ConfigFormComponent) focusInput(index int) {
	if index >= 0 && index < len(cf.inputs) {
		cf.inputs[index].Focus()
	}
}

// blurCurrentInput blurs the currently focused input
func (cf *ConfigFormComponent) blurCurrentInput() {
	if cf.focused >= 0 && cf.focused < len(cf.inputs) {
		cf.inputs[cf.focused].Blur()
	}
}

// GetSelectedAction returns the current action ("save" or "cancel")
func (cf *ConfigFormComponent) GetSelectedAction() string {
	if cf.isInButtons {
		if cf.selectedButton == 0 {
			return "save"
		}
		return "cancel"
	}
	return ""
}

// IsInButtons returns true if focus is in the button area
func (cf *ConfigFormComponent) IsInButtons() bool {
	return cf.isInButtons
}

// GetCurrentDescription returns the description for the focused field
func (cf *ConfigFormComponent) GetCurrentDescription() (string, string) {
	if cf.focused == -1 {
		return "WASM Target", "WASM target architecture\n\nSelect between 32-bit and 64-bit WASM compilation targets. Use ←/→ to toggle between options."
	} else if cf.focused >= 0 && cf.focused < len(configFields) {
		field := configFields[cf.focused]
		return field.label, field.description
	} else if cf.isInButtons {
		if cf.selectedButton == 0 {
			return "Save & Build", "Save the configuration changes and start the build process"
		}
		return "Cancel", "Cancel editing and return to configuration review"
	}
	return "Configuration", "Edit the build configuration values"
}

// View renders the configuration form
func (cf *ConfigFormComponent) View() string {
	var rows []string

	// WASM Target selector
	wasmLabel := "WASM Target"
	if cf.focused == -1 {
		wasmLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#902f17")).
			Bold(true).
			Render("WASM Target")
	}

	// WASM options
	wasm32Style := lipgloss.NewStyle().Padding(0, 1)
	wasm64Style := lipgloss.NewStyle().Padding(0, 1)

	if cf.wasmTarget == 0 {
		wasm32Style = wasm32Style.
			Background(lipgloss.Color("#902f17")).
			Foreground(lipgloss.Color("#FFFFFF"))
		wasm64Style = wasm64Style.
			Background(lipgloss.Color("#444")).
			Foreground(lipgloss.Color("#AAA"))
	} else {
		wasm32Style = wasm32Style.
			Background(lipgloss.Color("#444")).
			Foreground(lipgloss.Color("#AAA"))
		wasm64Style = wasm64Style.
			Background(lipgloss.Color("#902f17")).
			Foreground(lipgloss.Color("#FFFFFF"))
	}

	wasmOptions := lipgloss.JoinHorizontal(lipgloss.Left,
		wasm32Style.Render("WASM 32"),
		" ",
		wasm64Style.Render("WASM 64"),
	)

	wasmRow := lipgloss.JoinHorizontal(lipgloss.Left, wasmLabel, "  ", wasmOptions)

	// Calculate width for bordered elements
	inputWidth := cf.width - 4 // Account for element borders and some margin

	wasmBorderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(inputWidth).
		Height(1)

	if cf.focused == -1 {
		wasmBorderStyle = wasmBorderStyle.BorderForeground(lipgloss.Color("#902f17"))
	} else {
		wasmBorderStyle = wasmBorderStyle.BorderForeground(lipgloss.Color("#564f41"))
	}

	rows = append(rows, wasmBorderStyle.Render(wasmRow))

	// Text input fields
	for i, input := range cf.inputs {
		field := configFields[i]

		label := field.label
		if cf.focused == i {
			label = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#902f17")).
				Bold(true).
				Render(label)
		}

		inputRow := lipgloss.JoinHorizontal(lipgloss.Left, label, "  ", input.View())

		// Create border - use available width
		borderStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0, 1).
			Width(inputWidth).
			Height(1)

		if cf.focused == i {
			borderStyle = borderStyle.BorderForeground(lipgloss.Color("#902f17"))
		} else {
			borderStyle = borderStyle.BorderForeground(lipgloss.Color("#564f41"))
		}

		rows = append(rows, borderStyle.Render(inputRow))
	}

	// Buttons - calculate width more conservatively
	// Each button has border (2) + padding (2) = 4 chars overhead each
	// Plus 1 space between buttons = 9 chars total overhead
	buttonWidth := (inputWidth - 10) / 2 // Account for button borders, padding, and gap

	saveStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(buttonWidth).
		Align(lipgloss.Center)

	cancelStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(buttonWidth).
		Align(lipgloss.Center)

	if cf.isInButtons {
		if cf.selectedButton == 0 {
			saveStyle = saveStyle.
				Foreground(lipgloss.Color("#902f17")).
				Bold(true).
				BorderForeground(lipgloss.Color("#902f17"))
			cancelStyle = cancelStyle.BorderForeground(lipgloss.Color("#564f41"))
		} else {
			saveStyle = saveStyle.BorderForeground(lipgloss.Color("#564f41"))
			cancelStyle = cancelStyle.
				Foreground(lipgloss.Color("#902f17")).
				Bold(true).
				BorderForeground(lipgloss.Color("#902f17"))
		}
	} else {
		saveStyle = saveStyle.BorderForeground(lipgloss.Color("#666"))
		cancelStyle = cancelStyle.BorderForeground(lipgloss.Color("#666"))
	}

	saveBtn := saveStyle.Render("Save & Build")
	cancelBtn := cancelStyle.Render("Cancel")

	buttonContainer := lipgloss.JoinHorizontal(lipgloss.Top, cancelBtn, " ", saveBtn)
	centeredButtons := lipgloss.NewStyle().
		Width(inputWidth). // Match the input width exactly
		Align(lipgloss.Center).
		Render(buttonContainer)

	rows = append(rows, "", centeredButtons) // Add spacing before buttons

	// Join all rows and center them
	content := strings.Join(rows, "\n")

	// Return content directly - let the panel layout handle sizing and positioning
	return content
}
