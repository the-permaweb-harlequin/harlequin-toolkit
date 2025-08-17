package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/build/builders"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
)

// ViewState represents the current view in the TUI
type ViewState int

const (
	ViewBuildTypeSelection ViewState = iota
	ViewConfiguration
	ViewEntrypointSelection
	ViewOutputDirectory
	ViewConfigReview
	ViewConfigEditing
	ViewBuildRunning
	ViewBuildComplete
)

// StepStatus represents the status of a build step
type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepSuccess
	StepFailed
)

// BuildStep represents a single build step with status
type BuildStep struct {
	Name   string
	Status StepStatus
}

// Model represents the TUI application state
type Model struct {
	state          ViewState
	flow           *BuildFlow
	buildSteps     []BuildStep
	outputLines    []string
	terminalWidth  int
	terminalHeight int
	selectedIndex  int
	availableOptions []string
	configEditFields []string  // Config field values for editing
	configFieldIndex int       // Currently selected config field
	isEditingText    bool       // Whether we're currently editing text in an input
	cursorVisible    bool      // Animation for blinking cursor
}

// Messages for Bubble Tea
type BuildStepStartMsg struct{ StepName string }
type BuildStepCompleteMsg struct{ StepName string; Success bool }
type BuildOutputMsg struct{ Output string }
type tickMsg time.Time

// Command to send tick messages for animations
func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555"))
)

// BuildFlow represents the build configuration flow
type BuildFlow struct {
	BuildType    string
	SubType      string
	Entrypoint   string
	OutputDir    string
	Config       *config.Config
	ConfigEdited bool
}

// RunBuildTUI starts the interactive build TUI
func RunBuildTUI(ctx context.Context) error {
	// Initialize the model
	m := &Model{
		state: ViewBuildTypeSelection,
		flow:  &BuildFlow{},
		buildSteps: []BuildStep{
			{Name: "Copy AOS Files", Status: StepPending},
			{Name: "Bundle Lua", Status: StepPending},
			{Name: "Inject Code", Status: StepPending},
			{Name: "Build WASM", Status: StepPending},
			{Name: "Copy Outputs", Status: StepPending},
			{Name: "Cleanup", Status: StepPending},
		},
		outputLines: []string{},
		availableOptions: []string{"AOS Flavour"},
	}

	// Start the Bubble Tea program
	p := tea.NewProgram(m, tea.WithAltScreen())
	
	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

// Init implements the Bubble Tea model interface
func (m *Model) Init() tea.Cmd {
	// Initialize cursor visibility
	m.cursorVisible = true
	
	return tick()
}

// Update implements the Bubble Tea model interface
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		return m, nil
		
	case tickMsg:
		// Toggle cursor visibility for blinking effect
		m.cursorVisible = !m.cursorVisible
		return m, tick()
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.state == ViewConfigEditing {
				if m.selectedIndex >= 4 {
					// From buttons back to last input field
					m.selectedIndex = 3
					m.isEditingText = true  // Auto-enable editing
				} else if m.selectedIndex > 0 {
					// Navigate up through input fields
					m.selectedIndex--
					m.isEditingText = true  // Auto-enable editing
				}
			} else {
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}
			}
		case "down", "j":
			if m.state == ViewConfigEditing {
				if m.selectedIndex < 3 {
					// Navigate down through input fields
					m.selectedIndex++
					m.isEditingText = true  // Auto-enable editing
				} else if m.selectedIndex == 3 {
					// From last input field to first button
					m.selectedIndex = 4
					m.isEditingText = false  // Disable editing for buttons
				}
			} else {
				if m.selectedIndex < len(m.availableOptions)-1 {
					m.selectedIndex++
				}
			}
		case "left", "h":
			if m.state == ViewConfigEditing {
				if m.selectedIndex == 0 {
					// Handle Target field selector - switch to previous option
					if m.configEditFields[0] == "2" {
						m.configEditFields[0] = "1"  // Switch to WASM 32-bit
					}
				} else if m.selectedIndex >= 4 {
					// In button container, move left (Cancel = 5, Save = 4)
					m.selectedIndex = 5
				}
			}
		case "right", "l":
			if m.state == ViewConfigEditing {
				if m.selectedIndex == 0 {
					// Handle Target field selector - switch to next option
					if m.configEditFields[0] == "1" {
						m.configEditFields[0] = "2"  // Switch to WASM 64-bit
					}
				} else if m.selectedIndex >= 4 {
					// In button container, move right (Cancel = 5, Save = 4)
					m.selectedIndex = 4
				}
			}
		case "tab":
			if m.state == ViewConfigEditing && m.selectedIndex < 4 {
				// From input fields to button container
				m.selectedIndex = 4
			}
		case "enter":
			return m.handleSelection()
		case "esc":
			if m.state == ViewConfigEditing {
				// Exit config editing entirely
				m.state = ViewConfigReview
				m.availableOptions = []string{"Use current config", "Edit config"}
				m.selectedIndex = 0
				m.isEditingText = false
				return m, nil
			}
		case "backspace":
			if m.state == ViewConfigEditing && m.isEditingText && m.selectedIndex > 0 && m.selectedIndex < 4 {
				// Handle backspace in text input (skip Target field which is index 0)
				if m.selectedIndex < len(m.configEditFields) && len(m.configEditFields[m.selectedIndex]) > 0 {
					m.configEditFields[m.selectedIndex] = m.configEditFields[m.selectedIndex][:len(m.configEditFields[m.selectedIndex])-1]
				}
				return m, nil
			}

		default:
			// Handle character input for non-Target fields
			if m.state == ViewConfigEditing && m.isEditingText && m.selectedIndex > 0 && m.selectedIndex < 4 {
				// Check if it's a printable character
				if len(msg.String()) == 1 && msg.String()[0] >= 32 && msg.String()[0] <= 126 {
					// Add character to current field
					if m.selectedIndex < len(m.configEditFields) {
						m.configEditFields[m.selectedIndex] += msg.String()
					}
					return m, nil
				}
			}
		}
		
	case BuildStepStartMsg:
		// Update step status to running
		for i := range m.buildSteps {
			if m.buildSteps[i].Name == msg.StepName {
				m.buildSteps[i].Status = StepRunning
				break
			}
		}
		return m, nil
		
	case BuildStepCompleteMsg:
		// Update step status to success or failed
		for i := range m.buildSteps {
			if m.buildSteps[i].Name == msg.StepName {
				if msg.Success {
					m.buildSteps[i].Status = StepSuccess
				} else {
					m.buildSteps[i].Status = StepFailed
				}
				break
			}
		}
		return m, nil
		
	case BuildOutputMsg:
		// Add output line
		m.outputLines = append(m.outputLines, msg.Output)
		// Keep only last 20 lines
		if len(m.outputLines) > 20 {
			m.outputLines = m.outputLines[len(m.outputLines)-20:]
		}
		return m, nil
	}
	
	return m, nil
}

// View implements the Bubble Tea model interface
func (m *Model) View() string {
	// Center the container in terminal
	containerWidth := 90
	leftPadding := (m.terminalWidth - containerWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}
	
	// Create the main layout
	content := m.createMainLayout()
	
	// Center horizontally
	centeredContent := lipgloss.NewStyle().
		MarginLeft(leftPadding).
		Render(content)
	
	// Center vertically if terminal is tall enough
	if m.terminalHeight > 20 {
		topPadding := (m.terminalHeight - 20) / 2
		if topPadding > 0 {
			centeredContent = lipgloss.NewStyle().
				MarginTop(topPadding).
				Render(centeredContent)
		}
	}
	
	return centeredContent
}

// handleSelection processes the current selection
func (m *Model) handleSelection() (tea.Model, tea.Cmd) {
	switch m.state {
	case ViewBuildTypeSelection:
		m.flow.BuildType = "aos"
		m.state = ViewConfiguration
		m.availableOptions = []string{"Standard build"}
		m.selectedIndex = 0
		
	case ViewConfiguration:
		m.flow.SubType = "standard"
		m.state = ViewEntrypointSelection
		// TODO: Load actual Lua files
		m.availableOptions = []string{"main.lua", "src/init.lua"}
		m.selectedIndex = 0
		
	case ViewEntrypointSelection:
		m.flow.Entrypoint = m.availableOptions[m.selectedIndex]
		m.state = ViewOutputDirectory
		m.availableOptions = []string{"./dist", "./build", "./output"}
		m.selectedIndex = 0
		
	case ViewOutputDirectory:
		m.flow.OutputDir = m.availableOptions[m.selectedIndex]
		m.state = ViewConfigReview
		m.availableOptions = []string{"Use current config", "Edit config"}
		m.selectedIndex = 0
		
	case ViewConfigReview:
		if m.selectedIndex == 0 {
			// Use current config, start build
			m.state = ViewBuildRunning
			return m, m.startBuild()
		} else {
			// Edit config - load the actual config and switch to edit mode
			return m.loadAndEditConfig()
		}
		
	case ViewConfigEditing:
		if m.selectedIndex < 4 {
			// Editing a config field - this will be handled by keyboard input
		} else if m.selectedIndex == 4 {
			// Save & Build
			return m.saveConfigAndBuild()
		} else if m.selectedIndex == 5 {
			// Cancel - go back to config review
			m.state = ViewConfigReview
			m.availableOptions = []string{"Use current config", "Edit config"}
			m.selectedIndex = 0
		}
		
	case ViewBuildRunning:
		// Build is running, no action needed
		
	case ViewBuildComplete:
		return m, tea.Quit
	}
	
	return m, nil
}

// createMainLayout creates the full application layout
func (m *Model) createMainLayout() string {
	// Header - spans full width
	header := m.createHeader()
	
	// Left panel - either selector or build steps
	leftPanel := m.createLeftPanel()
	
	// Right panel - either description or output
	rightPanel := m.createRightPanel()
	
	// Align panels with header and controls (no centering)
	panelsContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel)  // Reduced gap
	
	// Controls section
	controls := m.createControls()
	
	// Combine all sections vertically with reduced spacing
	fullLayout := lipgloss.JoinVertical(lipgloss.Left,
		header,
		panelsContent,  // Aligned panels, no centering
		controls,
	)
	
	// Wrap in main container with reduced padding
	container := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(0, 1).  // Reduced padding: half of original (1,2) -> (0,1)
		Width(90).
		Render(fullLayout)
	
	return container
}

// createHeader creates the header section
func (m *Model) createHeader() string {
	title := m.getViewTitle()
	// Calculate available width: container content width (88) minus border
	availableWidth := 88 - 2  // Subtract border width
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#874BFD")).  // Purple text, no background
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(0, 1).  // Reduced padding
		Width(availableWidth).
		Align(lipgloss.Center).
		Render(title)
}

// createLeftPanel creates the left panel content
func (m *Model) createLeftPanel() string {
	if m.state == ViewBuildRunning || m.state == ViewBuildComplete {
		return m.createBuildStepsPanel()
	} else if m.state == ViewConfigEditing {
		return m.createConfigEditPanel()
	}
	return m.createSelectorPanel()
}

// createRightPanel creates the right panel content
func (m *Model) createRightPanel() string {
	if m.state == ViewBuildRunning || m.state == ViewBuildComplete {
		return m.createOutputPanel()
	} else if m.state == ViewConfigEditing {
		return m.createConfigDescriptionPanel()
	}
	return m.createDescriptionPanel()
}

// createSelectorPanel creates the left selector panel
func (m *Model) createSelectorPanel() string {
	content := ""
	
	// Add options with highlighting for selected
	for i, option := range m.availableOptions {
		if i == m.selectedIndex {
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
	
	// Calculate 50% of available width accounting for borders and gap
	// Each panel has 2-char border, plus 1-char gap = 5 chars overhead total
	availableContentWidth := 88 - 5  // Container width minus borders and gap
	panelWidth := availableContentWidth / 2  // About 41 chars each
	
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(0, 1).
		Width(panelWidth).
		Height(12).
		Render(content)
}

// createBuildStepsPanel creates the build steps panel with status indicators
func (m *Model) createBuildStepsPanel() string {
	content := "Build Progress:\n\n"
	
	for _, step := range m.buildSteps {
		icon := ""
		switch step.Status {
		case StepPending:
			icon = "○"  // Circle for pending
		case StepRunning:
			icon = "◐"  // Half circle for running (spinner effect)
		case StepSuccess:
			icon = "✓"  // Check for success
		case StepFailed:
			icon = "✗"  // X for failed
		}
		content += fmt.Sprintf("%s %s\n", icon, step.Name)
	}
	
	// Calculate 50% of available width accounting for borders and gap
	// Each panel has 2-char border, plus 1-char gap = 5 chars overhead total
	availableContentWidth := 88 - 5  // Container width minus borders and gap
	panelWidth := availableContentWidth / 2  // About 41 chars each
	
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(0, 1).
		Width(panelWidth).
		Height(12).
		Render(content)
}

// createConfigEditPanel creates the config editing panel
func (m *Model) createConfigEditPanel() string {
	fieldNames := []string{"WASM Target", "Stack Size (MB)", "Initial Memory (MB)", "Maximum Memory (MB)"}
	
	// Create horizontal label/input layout
	var content string
	
	for i, fieldName := range fieldNames {
		value := ""
		if i < len(m.configEditFields) {
			value = m.configEditFields[i]
		}
		
		// Label styling - simple, no borders or fixed width
		var label string
		if i == m.selectedIndex {
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#874BFD")).
				Bold(true).
				Align(lipgloss.Left).
				Padding(0, 2, 0, 0)
			label = labelStyle.Render(fieldName)
		} else {
			labelStyle := lipgloss.NewStyle().
				Align(lipgloss.Left).
				Padding(0, 2, 0, 0)
			label = labelStyle.Render(fieldName)
		}
		
		// Input styling - different behavior for Target field (selector) vs others (text input)
		var input string
		if i == 0 { // Target field - selector
			// Create both WASM 32 and WASM 64 options
			var wasm32, wasm64 string
			
			if value == "1" {
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
			
			inputStyle := lipgloss.NewStyle().
				Align(lipgloss.Left)
			input = inputStyle.Render(optionsGroup)
		} else { // Other fields - text input
			if i == m.selectedIndex {
				// Selected input: show blinking cursor (auto-editing)
				var cursor string
				if m.cursorVisible {
					cursor = "┃"  // Visible cursor
				} else {
					cursor = " "  // Invisible cursor (blink effect)
				}
				inputText := value + cursor
				
				inputStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#874BFD")).
					Align(lipgloss.Left)
				input = inputStyle.Render(inputText)
			} else {
				// Unselected input: simple text
				inputStyle := lipgloss.NewStyle().
					Align(lipgloss.Left)
				input = inputStyle.Render(value)
			}
		}
		
		// Create label/input group with proper spacing and alignment
		labelInputGroup := lipgloss.JoinHorizontal(lipgloss.Left, label, input)
		
		// Container for the label/input group with border
		// Horizontally justified (content distributed), vertically centered, aligned left
		var containerStyle lipgloss.Style
		if i == m.selectedIndex {
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
	// Calculate 50% width for each button (accounting for gap)
	buttonAreaWidth := 37  // Full panel content width
	buttonWidth := (buttonAreaWidth - 1) / 2  // 18 chars each, minus 1 for gap
	
	var cancelBtn, saveBtn string
	
	if m.selectedIndex == 5 {
		// Cancel selected
		cancelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#874BFD")).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 1).
			Width(buttonWidth).
			MaxWidth(buttonWidth).
			Align(lipgloss.Center).
			Inline(true)
		cancelBtn = cancelStyle.Render("Cancel")
	} else {
		cancelStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666")).
			Padding(0, 1).
			Width(buttonWidth).
			MaxWidth(buttonWidth).
			Align(lipgloss.Center).
			Inline(true)
		cancelBtn = cancelStyle.Render("Cancel")
	}
	
	if m.selectedIndex == 4 {
		// Save & Build selected
		saveStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#874BFD")).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 1).
			Width(buttonWidth).
			MaxWidth(buttonWidth).
			Align(lipgloss.Center).
			Inline(true)
		saveBtn = saveStyle.Render("Save & Build")
	} else {
		saveStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666")).
			Padding(0, 1).
			Width(buttonWidth).
			MaxWidth(buttonWidth).
			Align(lipgloss.Center).
			Inline(true)
		saveBtn = saveStyle.Render("Save & Build")
	}
	
	// Join buttons horizontally and center them
	buttonContainer := lipgloss.JoinHorizontal(lipgloss.Top, cancelBtn, " ", saveBtn)
	centeredButtons := lipgloss.NewStyle().
		Width(37).  // Full panel content width
		Align(lipgloss.Center).
		Render(buttonContainer)
	
	// Center the fields content and use JoinVertical to stick buttons to bottom
	centeredFields := lipgloss.NewStyle().
		Width(37).
		Align(lipgloss.Center).
		Render(content)
	
	// Use JoinVertical with spacing to separate fields from buttons
	finalContent := lipgloss.JoinVertical(lipgloss.Center, centeredFields, centeredButtons)
	
	// Calculate 50% of available width accounting for borders and gap
	// Each panel has 2-char border, plus 1-char gap = 5 chars overhead total
	availableContentWidth := 88 - 5  // Container width minus borders and gap
	panelWidth := availableContentWidth / 2  // About 41 chars each
	
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(0, 1).
		Width(panelWidth).
		Height(12).
		Render(finalContent)
}

// createConfigDescriptionPanel creates the config description panel
func (m *Model) createConfigDescriptionPanel() string {
	var header, body string
	
	if m.selectedIndex < 4 {
		fieldNames := []string{"Target", "Stack Size", "Initial Memory", "Maximum Memory"}
		descriptions := []string{
			"WASM target architecture\n\nSelect between 32-bit and 64-bit WASM compilation targets. Use ←/→ to toggle between options.",
			"Stack size in megabytes (MB) for the WASM runtime",
			"Initial memory allocation in megabytes (MB)",
			"Maximum memory limit in megabytes (MB)",
		}
		
		if m.selectedIndex < len(fieldNames) {
			header = fieldNames[m.selectedIndex]
			body = descriptions[m.selectedIndex]
		}
	} else if m.selectedIndex == 4 {
		header = "Save & Build"
		body = "Save the configuration changes and start the build process"
	} else if m.selectedIndex == 5 {
		header = "Cancel"
		body = "Cancel editing and return to configuration review"
	} else {
		header = "Configuration"
		body = "Edit the build configuration values"
	}
	
	// Create styled content
	styledHeader := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#874BFD")).
		Render(header)
	
	content := fmt.Sprintf("%s\n\n%s", styledHeader, body)
	
	// Calculate 50% of available width accounting for borders and gap
	// Each panel has 2-char border, plus 1-char gap = 5 chars overhead total
	availableContentWidth := 88 - 5  // Container width minus borders and gap
	panelWidth := availableContentWidth / 2  // About 41 chars each
	
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(0, 1).
		Width(panelWidth).
		Height(12).
		Render(content)
}

// createControls creates the controls section
func (m *Model) createControls() string {
	// Don't show controls during build
	if m.state == ViewBuildRunning || m.state == ViewBuildComplete {
		return ""
	}
	
	var controls string
	
	if m.state == ViewConfigEditing {
		if m.selectedIndex < 4 {
			// In input fields (always editing)
			controls = "↑/↓ Navigate   •   Esc Exit   •   q Quit"
		} else {
			// In button container
			controls = "←/→ Select Button  •   Esc Exit   •   q Quit"
		}
	} else {
		controls = "↑/↓ Navigate   •   Enter Select   •   q Quit"
	}
	
	// Calculate available width: container content width (88) minus border + padding
	availableWidth := 88 - 4  // 2 for border + 2 for horizontal padding
	
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(0, 1).  // Reduced padding
		Width(availableWidth).
		Align(lipgloss.Center).
		Render(controls)
}

// createDescriptionPanel creates the description panel
func (m *Model) createDescriptionPanel() string {
	var header, body string
	
	switch m.state {
	case ViewBuildTypeSelection:
		if m.selectedIndex < len(m.availableOptions) && m.availableOptions[m.selectedIndex] == "AOS Flavour" {
			header = "AOS Flavour"
			body = "Builds a wasm binary with your Lua injected into the base AOS process"
		}
	case ViewConfiguration:
		header = "Standard Build"
		body = "Uses the default AOS build configuration with standard optimizations"
	case ViewEntrypointSelection:
		header = "Entrypoint File"
		body = "The main Lua file that will be bundled and injected into the AOS process"
	case ViewOutputDirectory:
		header = "Output Directory"
		body = "Directory where the compiled WASM file and bundled Lua code will be saved"
	case ViewConfigReview:
		header = "Configuration"
		body = "Review and optionally edit the .harlequin.yaml configuration before building"
	default:
		header = "Select an option"
		body = "Choose from the available options to see detailed information"
	}
	
	// Create styled content
	styledHeader := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#874BFD")).
		Render(header)
	
	content := fmt.Sprintf("%s\n\n%s", styledHeader, body)
	
	// Calculate 50% of available width accounting for borders and gap
	// Each panel has 2-char border, plus 1-char gap = 5 chars overhead total
	availableContentWidth := 88 - 5  // Container width minus borders and gap
	panelWidth := availableContentWidth / 2  // About 41 chars each
	
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(0, 1).
		Width(panelWidth).
		Height(12).
		Render(content)
}

// createOutputPanel creates the output panel for build logs
func (m *Model) createOutputPanel() string {
	content := "Build Output:\n\n"
	
	for _, line := range m.outputLines {
		content += line + "\n"
	}
	
	if len(m.outputLines) == 0 {
		content += "Waiting for output..."
	}
	
	// Calculate 50% of available width accounting for borders and gap
	// Each panel has 2-char border, plus 1-char gap = 5 chars overhead total
	availableContentWidth := 88 - 5  // Container width minus borders and gap
	panelWidth := availableContentWidth / 2  // About 41 chars each
	
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(0, 1).
		Width(panelWidth).
		Height(12).
		Render(content)
}

// getViewTitle returns the title for the current view
func (m *Model) getViewTitle() string {
	switch m.state {
	case ViewBuildTypeSelection:
		return "Select Build Type"
	case ViewConfiguration:
		return "Select Build Configuration"
	case ViewEntrypointSelection:
		return "Select Entrypoint File"
	case ViewOutputDirectory:
		return "Select Output Directory"
	case ViewConfigReview:
		return "Review Configuration"
	case ViewConfigEditing:
		return "Edit Configuration"
	case ViewBuildRunning:
		return "Building Project"
	case ViewBuildComplete:
		return "Build Complete"
	default:
		return "Harlequin Build Tool"
	}
}

// startBuild initiates the build process
func (m *Model) startBuild() tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement actual build process with callbacks
		// For now, just simulate
		return BuildStepStartMsg{StepName: "Copy AOS Files"}
	}
}

// loadAndEditConfig loads the config and opens the edit form
func (m *Model) loadAndEditConfig() (tea.Model, tea.Cmd) {
	// Load config from file or use default
	cfg, err := m.loadConfigForEdit()
	if err != nil {
		// Use default config if loading fails - use realistic default values in bytes
		cfg = &config.Config{
			Target:        32,           // WASM 32-bit
			StackSize:     3145728,      // 3 MB in bytes
			InitialMemory: 4194304,      // 4 MB in bytes
			MaximumMemory: 1073741824,   // 1 GB in bytes
		}
	}
	
	// Set the loaded config in flow
	m.flow.Config = cfg
	
	// Initialize config editing fields with user-friendly values
	targetValue := "1" // Default to WASM 32
	if cfg.Target == 64 {
		targetValue = "2" // WASM 64
	}
	m.configEditFields = []string{
		targetValue,                                              // Target: convert 32/64 to selector values
		fmt.Sprintf("%.1f", float64(cfg.StackSize)/1024/1024),     // Stack Size: convert bytes to MB
		fmt.Sprintf("%.1f", float64(cfg.InitialMemory)/1024/1024), // Initial Memory: convert bytes to MB  
		fmt.Sprintf("%.1f", float64(cfg.MaximumMemory)/1024/1024), // Maximum Memory: convert bytes to MB
	}
	m.configFieldIndex = 0
	
	// Switch to config editing view
	m.state = ViewConfigEditing
	m.availableOptions = []string{"WASM Target", "Stack Size (MB)", "Initial Memory (MB)", "Maximum Memory (MB)", "Save & Build", "Cancel"}
	m.selectedIndex = 0
	m.isEditingText = true  // Auto-enable editing for first field
	
	return m, nil
}

// saveConfigAndBuild saves the edited config and starts the build
func (m *Model) saveConfigAndBuild() (tea.Model, tea.Cmd) {
	cfg := m.flow.Config
	if cfg == nil {
		cfg = &config.Config{}
		m.flow.Config = cfg
	}
	
	// Parse the edited field values back to config
	if len(m.configEditFields) >= 4 {
		// Convert selector value back to actual target value
		if m.configEditFields[0] == "1" {
			cfg.Target = 32  // WASM 32-bit
		} else {
			cfg.Target = 64  // WASM 64-bit
		}
		// Convert MB values back to bytes
		if stackSizeMB, err := strconv.ParseFloat(m.configEditFields[1], 64); err == nil {
			cfg.StackSize = int(stackSizeMB * 1024 * 1024)
		}
		if initialMemoryMB, err := strconv.ParseFloat(m.configEditFields[2], 64); err == nil {
			cfg.InitialMemory = int(initialMemoryMB * 1024 * 1024)
		}
		if maximumMemoryMB, err := strconv.ParseFloat(m.configEditFields[3], 64); err == nil {
			cfg.MaximumMemory = int(maximumMemoryMB * 1024 * 1024)
		}
	}
	
	// Mark config as edited and start build
	m.flow.ConfigEdited = true
	m.state = ViewBuildRunning
	return m, m.startBuild()
}

// loadConfigForEdit loads the config from .harlequin.yaml or build config
func (m *Model) loadConfigForEdit() (*config.Config, error) {
	// Try .harlequin.yaml first
	configPath := ".harlequin.yaml"
	if _, err := os.Stat(configPath); err == nil {
		defer func() {
			if r := recover(); r != nil {
				// Config file exists but is invalid
			}
		}()
		
		cfg := config.ReadConfigFile(configPath)
		if cfg != nil {
			return cfg, nil
		}
	}
	
	// Fallback to build config file
	buildConfigPath := "build_configs/ao-build-config.yml"
	if _, err := os.Stat(buildConfigPath); err == nil {
		defer func() {
			if r := recover(); r != nil {
				// Build config file exists but is invalid
			}
		}()
		
		cfg := config.ReadConfigFile(buildConfigPath)
		if cfg != nil {
			return cfg, nil
		}
	}
	
	// If both fail, return error to use defaults
	return nil, fmt.Errorf("no config file found")
}

// showConfigEditForm displays the config editing form
func (m *Model) showConfigEditForm() (tea.Model, tea.Cmd) {
	cfg := m.flow.Config
	if cfg == nil {
		cfg = &config.Config{
			Target:        1,
			StackSize:     512,
			InitialMemory: 32,
			MaximumMemory: 256,
		}
		m.flow.Config = cfg
	}
	
	// Convert int values to strings for the form
	targetStr := fmt.Sprintf("%d", cfg.Target)
	stackSizeStr := fmt.Sprintf("%d", cfg.StackSize)
	initialMemoryStr := fmt.Sprintf("%d", cfg.InitialMemory)
	maximumMemoryStr := fmt.Sprintf("%d", cfg.MaximumMemory)
	
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("WASM Target").
				Value(&targetStr).
				Description("Target configuration"),
			
			huh.NewInput().
				Title("Stack Size (MB)").
				Value(&stackSizeStr).
				Description("Stack size in KB"),
			
			huh.NewInput().
				Title("Initial Memory (MB)").
				Value(&initialMemoryStr).
				Description("Initial memory in MB"),
			
			huh.NewInput().
				Title("Maximum Memory (MB)").
				Value(&maximumMemoryStr).
				Description("Maximum memory in MB"),
		),
	)
	
	err := form.Run()
	if err != nil {
		// If form fails, just proceed with current config
		m.state = ViewBuildRunning
		return m, m.startBuild()
	}
	
	// Parse the string values back to integers
	if target, err := strconv.Atoi(targetStr); err == nil {
		cfg.Target = target
	}
	if stackSize, err := strconv.Atoi(stackSizeStr); err == nil {
		cfg.StackSize = stackSize
	}
	if initialMemory, err := strconv.Atoi(initialMemoryStr); err == nil {
		cfg.InitialMemory = initialMemory
	}
	if maximumMemory, err := strconv.Atoi(maximumMemoryStr); err == nil {
		cfg.MaximumMemory = maximumMemory
	}
	
	// Mark config as edited and start build
	m.flow.ConfigEdited = true
	m.state = ViewBuildRunning
	return m, m.startBuild()
}

// selectBuildType prompts user to select the build type
func selectBuildType(flow *BuildFlow) error {
	var buildType string

	// Show the full layout container
	fmt.Println() // Add some spacing
	fmt.Println(createBuildTypeLayoutContainer())
	fmt.Println()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("").  // Remove title as it's now in the header
				Description("").  // Remove description as it's now in the layout
				Options(
					huh.NewOption("AOS Flavour", "aos"),
				).
				Value(&buildType),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	flow.BuildType = buildType
	return nil
}

// createBuildTypeLayoutContainer creates the full layout with header, panels, and controls
func createBuildTypeLayoutContainer() string {
	// Header
	header := createHeader("Select Build Configuration")
	
	// Left panel - selector
	leftPanel := createSelectorPanel()
	
	// Right panel - description
	rightPanel := createBuildTypeDescriptionPanel("aos")
	
	// Main content area (left + right panels)
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
	
	// Bottom controls
	bottomControls := createBottomControls()
	
	// Combine all sections vertically
	fullLayout := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",  // Spacing
		mainContent,
		"",  // Spacing
		bottomControls,
	)
	
	// Wrap in main container
	container := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2).
		Width(90).
		Render(fullLayout)
	
	return container
}

// createHeader creates the header section
func createHeader(title string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#874BFD")).
		Padding(0, 2).
		Width(82).  // Adjust for container padding
		Align(lipgloss.Center).
		Render(title)
}

// createSelectorPanel creates the left selector panel
func createSelectorPanel() string {
	content := "Build Types:\n\n❯ AOS Flavour"
	
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(1, 2).
		Width(35).
		Height(8).
		Render(content)
}

// createBottomControls creates the bottom control panel
func createBottomControls() string {
	controls := "Controls: ↑/↓ Navigate • Enter Select • q Quit"
	
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Background(lipgloss.Color("#222")).
		Padding(0, 2).
		Width(82).
		Align(lipgloss.Center).
		Render(controls)
}

// createBuildTypeDescriptionPanel creates a styled description panel for the given build type
func createBuildTypeDescriptionPanel(buildType string) string {
	var header, body string
	
	switch buildType {
	case "aos":
		header = "AOS Flavour"
		body = "Builds a wasm binary with your Lua injected into the base AOS process"
	default:
		header = "Select a build type"
		body = "Choose from the available options to see detailed information"
	}
	
	// Create styled content
	styledHeader := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#874BFD")).
		Render(header)
	
	// Create the content
	content := fmt.Sprintf("%s\n\n%s", styledHeader, body)
	
	// Create a bordered panel that matches the selector panel
	panel := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).  // Match selector border
		Padding(1, 2).
		Width(40).  // Adjust width to fit in container
		Height(8).  // Match selector height
		Render(content)
	
	return panel
}

// selectSubType prompts user to select the sub-type
func selectSubType(flow *BuildFlow) error {
	var subType string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select build configuration").
				Description("Choose the build configuration for " + strings.ToUpper(flow.BuildType)).
				Options(
					huh.NewOption("Standard build", "standard"),
				).
				Value(&subType),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	flow.SubType = subType
	return nil
}

// selectEntrypoint prompts user to select the entrypoint file
func selectEntrypoint(flow *BuildFlow) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Find Lua files in current directory
	luaFiles, err := findLuaFiles(cwd)
	if err != nil {
		return fmt.Errorf("failed to find Lua files: %w", err)
	}

	if len(luaFiles) == 0 {
		return fmt.Errorf("no Lua files found in current directory")
	}

	var entrypoint string

	// Create options from found Lua files
	options := make([]huh.Option[string], 0, len(luaFiles))
	for _, file := range luaFiles {
		// Show relative path from current directory
		relPath, _ := filepath.Rel(cwd, file)
		options = append(options, huh.NewOption(relPath, file))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select entrypoint file").
				Description("Choose the main Lua file for your project").
				Options(options...).
				Value(&entrypoint),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	flow.Entrypoint = entrypoint
	return nil
}

// selectOutputDirectory prompts user to select the output directory
func selectOutputDirectory(flow *BuildFlow) error {
	var outputDir string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Output directory").
				Description("Enter the directory where build outputs will be saved").
				Value(&outputDir).
				Placeholder("./dist"),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if outputDir == "" {
		outputDir = "./dist"
	}

	flow.OutputDir = outputDir
	return nil
}

// reviewAndEditConfig loads and allows editing of the .harlequin config
func reviewAndEditConfig(flow *BuildFlow) error {
	// Try to load existing config
	configPath := ".harlequin.yaml"
	var cfg *config.Config

	if _, err := os.Stat(configPath); err == nil {
		cfg = config.ReadConfigFile(configPath)
		fmt.Println(infoStyle.Render("📄 Loaded existing .harlequin.yaml"))
	} else {
		cfg = config.NewConfig(nil)
		fmt.Println(infoStyle.Render("🆕 Using default configuration (no .harlequin.yaml found)"))
	}

	flow.Config = cfg

	// Show current config and ask if user wants to edit
	var action string

	currentConfig := fmt.Sprintf(`Current configuration:
  AOS Git Hash: %s
  Compute Limit: %s
  Module Format: %s
  Target: %d
  Stack Size: %d
  Initial Memory: %d
  Maximum Memory: %d`,
		cfg.AOSGitHash,
		cfg.ComputeLimit,
		cfg.ModuleFormat,
		cfg.Target,
		cfg.StackSize,
		cfg.InitialMemory,
		cfg.MaximumMemory)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Configuration Review").
				Description(currentConfig),

			huh.NewSelect[string]().
				Title("What would you like to do?").
				Options(
					huh.NewOption("Use current configuration", "use"),
					huh.NewOption("Edit configuration", "edit"),
				).
				Value(&action),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if action == "edit" {
		return editConfig(flow)
	}

	return nil
}

// editConfig allows user to edit configuration fields
func editConfig(flow *BuildFlow) error {
	cfg := flow.Config

	// Convert int fields to strings for editing
	targetStr := fmt.Sprintf("%d", cfg.Target)
	stackSizeStr := fmt.Sprintf("%d", cfg.StackSize)
	initialMemoryStr := fmt.Sprintf("%d", cfg.InitialMemory)
	maxMemoryStr := fmt.Sprintf("%d", cfg.MaximumMemory)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("AOS Git Hash").
				Description("Git commit hash or branch name for AOS").
				Value(&cfg.AOSGitHash),

			huh.NewInput().
				Title("Compute Limit").
				Description("Maximum compute units for the module").
				Value(&cfg.ComputeLimit),

			huh.NewInput().
				Title("Module Format").
				Description("Target format for the compiled module").
				Value(&cfg.ModuleFormat),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("WASM Target Architecture").
				Description("Target architecture (32 or 64)").
				Value(&targetStr),

			huh.NewInput().
				Title("Stack Size (MB)").
				Description("Stack size in bytes").
				Value(&stackSizeStr),

			huh.NewInput().
				Title("Initial Memory (MB)").
				Description("Initial memory size in bytes").
				Value(&initialMemoryStr),

			huh.NewInput().
				Title("Maximum Memory (MB)").
				Description("Maximum memory size in bytes").
				Value(&maxMemoryStr),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	// Convert string values back to integers
	if target, err := parseInt(targetStr, "WASM Target"); err != nil {
		return err
	} else {
		cfg.Target = target
	}

	if stackSize, err := parseInt(stackSizeStr, "Stack Size (MB)"); err != nil {
		return err
	} else {
		cfg.StackSize = stackSize
	}

	if initialMemory, err := parseInt(initialMemoryStr, "Initial Memory (MB)"); err != nil {
		return err
	} else {
		cfg.InitialMemory = initialMemory
	}

	if maxMemory, err := parseInt(maxMemoryStr, "Maximum Memory (MB)"); err != nil {
		return err
	} else {
		cfg.MaximumMemory = maxMemory
	}

	flow.ConfigEdited = true
	return nil
}

// executeBuild runs the actual build process
func executeBuild(ctx context.Context, flow *BuildFlow) error {
	fmt.Println()
	fmt.Println(titleStyle.Render("🚀 Starting Build"))
	fmt.Println()

	// Create AOSBuilder with the selected parameters
	builder := builders.NewAOSBuilder(builders.AOSBuilderParams{
		Config:         flow.Config,
		ConfigFilePath: nil, // Use default .harlequin.yaml
		Entrypoint:     flow.Entrypoint,
		OutputDir:      flow.OutputDir,
		Callbacks:      builders.CallbacksProgress, // Use progress callbacks for nice output
	})

	// Run the build
	if err := builder.Build(ctx); err != nil {
		fmt.Println()
		fmt.Println(errorStyle.Render("❌ Build failed: " + err.Error()))
		return err
	}

	fmt.Println()
	fmt.Println(titleStyle.Render("✅ Build completed successfully!"))
	fmt.Printf("📁 Output directory: %s\n", flow.OutputDir)

	return nil
}

// findLuaFiles recursively finds all .lua files in the given directory
func findLuaFiles(dir string) ([]string, error) {
	var luaFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip common directories that shouldn't contain entrypoint files
		if info.IsDir() {
			switch info.Name() {
			case "node_modules", ".git", "dist", "build", "target":
				return filepath.SkipDir
			}
		}

		// Add Lua files
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".lua") {
			luaFiles = append(luaFiles, path)
		}

		return nil
	})

	return luaFiles, err
}

// parseInt converts a string to an integer with error handling
func parseInt(s, fieldName string) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid value for %s: %s", fieldName, s)
	}
	return val, nil
}
