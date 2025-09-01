package components

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UploadWizardState represents the current step in the upload wizard
type UploadWizardState int

const (
	UploadStateWasmFile UploadWizardState = iota
	UploadStateConfig
	UploadStateWalletFile
	UploadStateVersion
	UploadStateGitHash
	UploadStateDryRun
	UploadStateConfirmation
	UploadStateCompleted
)

// UploadWizardComponent represents the upload wizard TUI following existing patterns
type UploadWizardComponent struct {
	state UploadWizardState
	width int
	height int

	// File pickers
	wasmFilePicker   *FilePickerComponent
	configFilePicker *FilePickerComponent
	walletFilePicker *FilePickerComponent

	// Text inputs
	versionInput textinput.Model
	gitHashInput textinput.Model

	// List selector for dry run
	dryRunSelector *ListSelectorComponent

	// Selected values
	WasmFile   string
	ConfigFile string
	WalletFile string
	Version    string
	GitHash    string
	DryRun     bool

	// Callback for completion
	OnComplete func(wasmFile, configFile, walletFile, version, gitHash string, dryRun bool)

	err error
}

// NewUploadWizardComponent creates a new upload wizard following existing patterns
func NewUploadWizardComponent() *UploadWizardComponent {
	// Create file pickers using existing pattern - smaller size to prevent overflow
	wasmPicker := NewFilePicker(40, 10)
	wasmPicker.SetAllowedTypes([]string{".wasm"})

	configPicker := NewFilePicker(40, 10)
	configPicker.SetAllowedTypes([]string{".yml", ".yaml"})

	walletPicker := NewFilePicker(40, 10)
	walletPicker.SetAllowedTypes([]string{".json"})

	// Set up current directory for file pickers
	if cwd, err := os.Getwd(); err == nil {
		wasmPicker.SetCurrentDirectory(cwd)
		configPicker.SetCurrentDirectory(cwd)
		walletPicker.SetCurrentDirectory(cwd)
	}

	// Create text inputs
	versionInput := textinput.New()
	versionInput.Placeholder = "e.g., v1.0.0"
	versionInput.CharLimit = 50
	versionInput.Width = 30
	versionInput.Focus()

	gitHashInput := textinput.New()
	gitHashInput.Placeholder = "e.g., abc123def (leave empty for auto-detect)"
	gitHashInput.CharLimit = 40
	gitHashInput.Width = 50

	// Create dry run selector
	dryRunSelector := CreateDryRunSelector(50, 10)

	return &UploadWizardComponent{
		state:            UploadStateWasmFile,
		wasmFilePicker:   wasmPicker,
		configFilePicker: configPicker,
		walletFilePicker: walletPicker,
		versionInput:     versionInput,
		gitHashInput:     gitHashInput,
		dryRunSelector:   dryRunSelector,
		ConfigFile:       "build_configs/ao-build-config.yml", // Default
		DryRun:           true, // Default to dry run
	}
}

// CreateDryRunSelector creates a selector for dry run mode
func CreateDryRunSelector(width, height int) *ListSelectorComponent {
	items := []ListItem{
		{title: "Dry Run", value: "dry", description: "Preview upload without actually uploading to Arweave"},
		{title: "Actual Upload", value: "upload", description: "Perform the real upload to Arweave (requires wallet with sufficient balance)"},
	}

	return NewListSelector("Upload Mode", items, width, height)
}

// Init initializes the component
func (m *UploadWizardComponent) Init() tea.Cmd {
	var cmds []tea.Cmd
	if m.wasmFilePicker != nil {
		cmds = append(cmds, m.wasmFilePicker.Init())
	}
	if m.configFilePicker != nil {
		cmds = append(cmds, m.configFilePicker.Init())
	}
	if m.walletFilePicker != nil {
		cmds = append(cmds, m.walletFilePicker.Init())
	}
	if m.dryRunSelector != nil {
		cmds = append(cmds, m.dryRunSelector.Init())
	}
	return tea.Batch(cmds...)
}

// Update handles messages
func (m *UploadWizardComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate appropriate size for components to prevent overflow
		componentWidth := (msg.Width - 20) // Conservative width
		if componentWidth < 30 {
			componentWidth = 30
		}
		componentHeight := 10

		// Resize file pickers
		if m.wasmFilePicker != nil {
			m.wasmFilePicker.SetSize(componentWidth, componentHeight)
		}
		if m.configFilePicker != nil {
			m.configFilePicker.SetSize(componentWidth, componentHeight)
		}
		if m.walletFilePicker != nil {
			m.walletFilePicker.SetSize(componentWidth, componentHeight)
		}

		// Resize selector
		if m.dryRunSelector != nil {
			m.dryRunSelector.SetSize(componentWidth, 6)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		case "esc":
			return m.handleBack()
		}
	}

	// Update current component based on state
	switch m.state {
	case UploadStateWasmFile:
		if m.wasmFilePicker != nil {
			model, cmd := m.wasmFilePicker.Update(msg)
			if newPicker, ok := model.(*FilePickerComponent); ok {
				m.wasmFilePicker = newPicker
			}
			if m.wasmFilePicker.HasSelection() {
				m.WasmFile = m.wasmFilePicker.GetSelectedFile()
				m.state = UploadStateConfig
			}
			cmds = append(cmds, cmd)
		}

	case UploadStateConfig:
		if m.configFilePicker != nil {
			model, cmd := m.configFilePicker.Update(msg)
			if newPicker, ok := model.(*FilePickerComponent); ok {
				m.configFilePicker = newPicker
			}
			if m.configFilePicker.HasSelection() {
				m.ConfigFile = m.configFilePicker.GetSelectedFile()
				m.state = UploadStateWalletFile
			}
			cmds = append(cmds, cmd)
		}

	case UploadStateWalletFile:
		if m.walletFilePicker != nil {
			model, cmd := m.walletFilePicker.Update(msg)
			if newPicker, ok := model.(*FilePickerComponent); ok {
				m.walletFilePicker = newPicker
			}
			if m.walletFilePicker.HasSelection() {
				m.WalletFile = m.walletFilePicker.GetSelectedFile()
				m.state = UploadStateVersion
				m.versionInput.Focus()
			}
			cmds = append(cmds, cmd)
		}

	case UploadStateVersion:
		m.versionInput, cmd = m.versionInput.Update(msg)
		cmds = append(cmds, cmd)

	case UploadStateGitHash:
		m.gitHashInput, cmd = m.gitHashInput.Update(msg)
		cmds = append(cmds, cmd)

	case UploadStateDryRun:
		if m.dryRunSelector != nil {
			model, cmd := m.dryRunSelector.Update(msg)
			if newSelector, ok := model.(*ListSelectorComponent); ok {
				m.dryRunSelector = newSelector
			}
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// handleEnter processes the enter key for the current state
func (m *UploadWizardComponent) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case UploadStateVersion:
		if m.versionInput.Value() == "" {
			m.err = fmt.Errorf("version is required")
			return m, nil
		}
		m.Version = m.versionInput.Value()
		m.state = UploadStateGitHash
		m.gitHashInput.Focus()

	case UploadStateGitHash:
		m.GitHash = m.gitHashInput.Value()
		m.state = UploadStateDryRun

	case UploadStateDryRun:
		if m.dryRunSelector != nil {
			if selected := m.dryRunSelector.GetSelected(); selected != nil {
				m.DryRun = selected.Value() == "dry"
				m.state = UploadStateConfirmation
			}
		}

	case UploadStateConfirmation:
		if m.OnComplete != nil {
			m.OnComplete(m.WasmFile, m.ConfigFile, m.WalletFile, m.Version, m.GitHash, m.DryRun)
		}
		m.state = UploadStateCompleted
		return m, tea.Quit
	}

	m.err = nil
	return m, nil
}

// handleBack processes the escape key to go back
func (m *UploadWizardComponent) handleBack() (tea.Model, tea.Cmd) {
	switch m.state {
	case UploadStateConfig:
		m.state = UploadStateWasmFile
	case UploadStateWalletFile:
		m.state = UploadStateConfig
	case UploadStateVersion:
		m.state = UploadStateWalletFile
		m.versionInput.Blur()
	case UploadStateGitHash:
		m.state = UploadStateVersion
		m.gitHashInput.Blur()
		m.versionInput.Focus()
	case UploadStateDryRun:
		m.state = UploadStateGitHash
		m.gitHashInput.Focus()
	case UploadStateConfirmation:
		m.state = UploadStateDryRun
	default:
		return m, tea.Quit
	}

	m.err = nil
	return m, nil
}

// View renders the component using existing TUI layout patterns
func (m *UploadWizardComponent) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Use exact same sizing calculation as the main TUI
	containerWidth := m.width - 10
	header := CreateHeader(m.getStateTitle(), containerWidth)

	// Create main content with size constraints
	content := m.renderStateContent()

	// Create controls/help with proper width
	controls := m.createControls()

	// Create main layout with proper container - exact match to main TUI
	mainLayout := lipgloss.JoinVertical(lipgloss.Left,
		header,
		content,
		controls,
	)

	// Create bordered container that fits the terminal with some margin - exact match
	container := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#902f17")).
		Padding(0, 1).
		Width(containerWidth).
		Render(mainLayout)

	// Center the container horizontally - exact match
	leftMargin := (m.width - containerWidth) / 2
	container = lipgloss.NewStyle().
		MarginLeft(leftMargin).
		Render(container)

	// Only center vertically if we have plenty of extra space - exact match
	// Be conservative to avoid overflow
	if m.height > 35 {
		// Only add a small top margin, don't try to center completely
		topPadding := 2
		container = lipgloss.NewStyle().
			MarginTop(topPadding).
			Render(container)
	}

	return container
}

// getStateTitle returns the title for the current state
func (m *UploadWizardComponent) getStateTitle() string {
	switch m.state {
	case UploadStateWasmFile:
		return "Select WASM File"
	case UploadStateConfig:
		return "Select Configuration File"
	case UploadStateWalletFile:
		return "Select Wallet File"
	case UploadStateVersion:
		return "Enter Module Version"
	case UploadStateGitHash:
		return "Enter Git Hash"
	case UploadStateDryRun:
		return "Choose Upload Mode"
	case UploadStateConfirmation:
		return "Confirm Upload Settings"
	case UploadStateCompleted:
		return "Upload Completed"
	}
	return "Module Upload Wizard"
}

// renderStateContent renders the content for the current state
func (m *UploadWizardComponent) renderStateContent() string {
	// Calculate content width (match main TUI patterns)
	contentWidth := m.width - 14 // Account for container borders and padding
	if contentWidth < 30 {
		contentWidth = 30
	}

	var content strings.Builder

	// Error display
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
		content.WriteString(errorStyle.Render("❌ " + m.err.Error()))
		content.WriteString("\n\n")
	}

	// State-specific content with constrained width
	switch m.state {
	case UploadStateWasmFile:
		content.WriteString("Step 1: Select the WASM file to upload\n\n")
		if m.wasmFilePicker != nil {
			// Constrain the file picker content
			pickerContent := lipgloss.NewStyle().Width(contentWidth).Render(m.wasmFilePicker.View())
			content.WriteString(pickerContent)
		}

	case UploadStateConfig:
		content.WriteString("Step 2: Select the build configuration file\n\n")
		if m.configFilePicker != nil {
			pickerContent := lipgloss.NewStyle().Width(contentWidth).Render(m.configFilePicker.View())
			content.WriteString(pickerContent)
		}

	case UploadStateWalletFile:
		content.WriteString("Step 3: Select the Arweave wallet file\n\n")
		if m.walletFilePicker != nil {
			pickerContent := lipgloss.NewStyle().Width(contentWidth).Render(m.walletFilePicker.View())
			content.WriteString(pickerContent)
		}

	case UploadStateVersion:
		content.WriteString("Step 4: Enter the module version\n\n")
		content.WriteString("Version: " + m.versionInput.View())
		content.WriteString("\n\nExample: v1.0.0, v2.1.3, etc.")

	case UploadStateGitHash:
		content.WriteString("Step 5: Enter git hash (optional)\n\n")
		content.WriteString("Git Hash: " + m.gitHashInput.View())
		content.WriteString("\n\nLeave empty for auto-detection")

	case UploadStateDryRun:
		content.WriteString("Step 6: Choose upload mode\n\n")
		if m.dryRunSelector != nil {
			selectorContent := lipgloss.NewStyle().Width(contentWidth).Render(m.dryRunSelector.View())
			content.WriteString(selectorContent)
		}

	case UploadStateConfirmation:
		content.WriteString("Step 7: Confirm your settings\n\n")
		content.WriteString(fmt.Sprintf("WASM File: %s\n", m.WasmFile))
		content.WriteString(fmt.Sprintf("Config File: %s\n", m.ConfigFile))
		content.WriteString(fmt.Sprintf("Wallet File: %s\n", m.WalletFile))
		content.WriteString(fmt.Sprintf("Version: %s\n", m.Version))
		if m.GitHash != "" {
			content.WriteString(fmt.Sprintf("Git Hash: %s\n", m.GitHash))
		}
		mode := "Actual Upload"
		if m.DryRun {
			mode = "Dry Run"
		}
		content.WriteString(fmt.Sprintf("Mode: %s\n", mode))

	case UploadStateCompleted:
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
		content.WriteString(successStyle.Render("✅ Upload configuration completed!"))
		content.WriteString("\n\nStarting upload process...")
	}

	return content.String()
}

// createControls creates the bottom controls section
func (m *UploadWizardComponent) createControls() string {
	var controls []string

	switch m.state {
	case UploadStateWasmFile:
		controls = []string{"↑/↓ Navigate", "→ Enter Dir", "Enter Select", "q Quit"}
	case UploadStateConfig:
		controls = []string{"↑/↓ Navigate", "→ Enter Dir", "Enter Select", "Esc Back", "q Quit"}
	case UploadStateWalletFile:
		controls = []string{"↑/↓ Navigate", "→ Enter Dir", "Enter Select", "Esc Back", "q Quit"}
	case UploadStateVersion:
		controls = []string{"Type version", "Enter Continue", "Esc Back", "q Quit"}
	case UploadStateGitHash:
		controls = []string{"Type hash", "Enter Continue", "Esc Back", "q Quit"}
	case UploadStateDryRun:
		controls = []string{"↑/↓ Navigate", "Enter Select", "Esc Back", "q Quit"}
	case UploadStateConfirmation:
		controls = []string{"Enter Proceed", "Esc Back", "q Quit"}
	case UploadStateCompleted:
		controls = []string{"Starting upload..."}
	}

	// Use the same width calculation as the main TUI to prevent overflow
	containerWidth := m.width - 10
	if containerWidth < 40 {
		containerWidth = 40
	}

	return CreateControls(controls, containerWidth)
}
