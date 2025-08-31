package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/build/builders"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/debug"
	luautils "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/lua_utils"
	components "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tui/components"
)

// ViewState represents the current view in the TUI
type ViewState int

const (
	ViewCommandSelection ViewState = iota
	ViewBuildTypeSelection
	ViewEntrypointSelection
	ViewOutputDirectory
	ViewConfigReview
	ViewConfigEditing
	ViewBuildRunning
	ViewBuildSuccess
	ViewBuildError
	ViewLuaUtilsSelection
	ViewLuaUtilsEntrypoint
	ViewLuaUtilsOutput
	ViewLuaUtilsRunning
	ViewLuaUtilsSuccess
	ViewLuaUtilsError
)

// Model represents the modernized TUI application state
type Model struct {
	// Core state
	state ViewState
	flow  *BuildFlow
	luaUtilsFlow *LuaUtilsFlow
	ctx   context.Context

	// Bubbles components
	keyMap          components.KeyMap
	help            help.Model
	commandSelector *components.ListSelectorComponent
	buildSelector   *components.ListSelectorComponent
	outputInput     *components.TextInputComponent
	actionSelector  *components.ListSelectorComponent
	filePicker      *components.FilePickerComponent
	fileSelector    *components.ListSelectorComponent // For automatic file discovery
	configForm      *components.ConfigFormComponent
	progress        *components.ProgressComponent
	result          *components.ResultComponent

	// Lua Utils components
	luaUtilsSelector    *components.ListSelectorComponent
	luaUtilsFilePicker  *components.FilePickerComponent
	luaUtilsFileSelector *components.ListSelectorComponent
	luaUtilsOutputInput *components.TextInputComponent

	// Layout
	width  int
	height int

	// File selection mode
	useFilePicker bool // true = manual picker, false = automatic list
	useLuaUtilsFilePicker bool // for lua-utils file selection

	// Build process
	buildResult *BuildResult
	luaUtilsResult *LuaUtilsResult
	program     *tea.Program
}

// BuildFlow represents the build configuration flow (unchanged)
type BuildFlow struct {
	BuildType    string
	SubType      string
	Entrypoint   string
	OutputDir    string
	Config       *config.Config
	ConfigEdited bool
	BuildResult  *BuildResult
}

// BuildResult holds the result of a build operation (unchanged)
type BuildResult struct {
	Success bool
	Error   error
	Flow    *BuildFlow
}

// LuaUtilsFlow represents the lua-utils configuration flow
type LuaUtilsFlow struct {
	Command     string // "bundle" for now
	Entrypoint  string
	OutputPath  string
}

// LuaUtilsResult holds the result of a lua-utils operation
type LuaUtilsResult struct {
	Success bool
	Error   error
	Flow    *LuaUtilsFlow
}

// Messages for Bubble Tea
type BuildStepStartMsg struct{ StepName string }
type BuildStepCompleteMsg struct {
	StepName string
	Success  bool
}
type BuildCompleteMsg struct{ Result *BuildResult }
type LuaUtilsCompleteMsg struct{ Result *LuaUtilsResult }
type TickMsg struct{}

// NewModel creates a new modernized TUI model
func NewModel(ctx context.Context) *Model {
	// Initialize components
	keyMap := components.DefaultKeyMap()
	helpModel := help.New()

	// Create command selector
	commandSelector := components.CreateCommandSelector(40, 10)

	// Create build type selector
	buildSelector := components.CreateBuildTypeSelector(40, 10)

	// Initialize progress component
	progress := components.NewProgressComponent(40, 10)

	return &Model{
		state:           ViewCommandSelection,
		flow:            &BuildFlow{},
		luaUtilsFlow:    &LuaUtilsFlow{},
		ctx:             ctx,
		keyMap:          keyMap,
		help:            helpModel,
		commandSelector: commandSelector,
		buildSelector:   buildSelector,
		progress:        progress,
	}
}

// Init implements the Bubble Tea model interface
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.commandSelector.Init(),
		m.buildSelector.Init(),
		tea.EnterAltScreen,
	)
}

// Update handles Bubble Tea messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeComponents()

	case tea.KeyMsg:
		// Global key bindings
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.Help):
			// Toggle help - could be implemented later
			return m, nil
		case key.Matches(msg, m.keyMap.Back):
			return m.handleBack()
		}

		// State-specific handling
		switch m.state {
		case ViewCommandSelection:
			return m.updateCommandSelection(msg)
		case ViewBuildTypeSelection:
			return m.updateBuildTypeSelection(msg)
		case ViewEntrypointSelection:
			return m.updateEntrypointSelection(msg)
		case ViewOutputDirectory:
			return m.updateOutputDirectory(msg)
		case ViewConfigReview:
			return m.updateConfigReview(msg)
		case ViewConfigEditing:
			return m.updateConfigEditing(msg)
		case ViewBuildRunning:
			return m.updateBuildRunning(msg)
		case ViewBuildSuccess, ViewBuildError:
			return m.updateBuildResult(msg)
		case ViewLuaUtilsSelection:
			return m.updateLuaUtilsSelection(msg)
		case ViewLuaUtilsEntrypoint:
			return m.updateLuaUtilsEntrypoint(msg)
		case ViewLuaUtilsOutput:
			return m.updateLuaUtilsOutput(msg)
		case ViewLuaUtilsRunning:
			return m.updateLuaUtilsRunning(msg)
		case ViewLuaUtilsSuccess, ViewLuaUtilsError:
			return m.updateLuaUtilsResult(msg)
		}

	case BuildStepStartMsg:
		if m.progress != nil {
			steps := []components.BuildStep{
				{Name: msg.StepName, Status: components.StepRunning},
			}
			m.progress.UpdateSteps(steps)
		}

	case BuildStepCompleteMsg:
		if m.progress != nil {
			status := components.StepSuccess
			if !msg.Success {
				status = components.StepFailed
			}
			steps := []components.BuildStep{
				{Name: msg.StepName, Status: status},
			}
			m.progress.UpdateSteps(steps)
		}

	case BuildCompleteMsg:
		m.buildResult = msg.Result
		if msg.Result.Success {
			m.state = ViewBuildSuccess
		} else {
			m.state = ViewBuildError
		}

		// Create result component
		// Using content methods that don't apply their own sizing/borders
		m.result = components.NewResultComponent(
			msg.Result.Success,
			msg.Result,
			0, // Width not used by content methods
			0, // Height not used by content methods
		)

		return m, nil

	case LuaUtilsCompleteMsg:
		m.luaUtilsResult = msg.Result
		if msg.Result.Success {
			m.state = ViewLuaUtilsSuccess
		} else {
			m.state = ViewLuaUtilsError
		}

		// Create result component for lua-utils
		m.result = components.NewResultComponent(
			msg.Result.Success,
			msg.Result,
			0, // Width not used by content methods
			0, // Height not used by content methods
		)

		return m, nil

	case TickMsg:
		// Update progress animations during build
		if m.state == ViewBuildRunning && m.progress != nil {
			m.progress.UpdateAnimations()
			// Continue ticking while in build running state
			return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
				return TickMsg{}
			})
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Calculate container width with margin to prevent overflow
	containerWidth := m.width - 10
	header := components.CreateHeader(m.getViewTitle(), containerWidth)

	// Create main content based on state
	var content string
	switch m.state {
	case ViewCommandSelection:
		content = m.viewCommandSelection()
	case ViewBuildTypeSelection:
		content = m.viewBuildTypeSelection()
	case ViewEntrypointSelection:
		content = m.viewEntrypointSelection()
	case ViewOutputDirectory:
		content = m.viewOutputDirectory()
	case ViewConfigReview:
		content = m.viewConfigReview()
	case ViewConfigEditing:
		content = m.viewConfigEditing()
	case ViewBuildRunning:
		content = m.viewBuildRunning()
	case ViewBuildSuccess, ViewBuildError:
		content = m.viewBuildResult()
	case ViewLuaUtilsSelection:
		content = m.viewLuaUtilsSelection()
	case ViewLuaUtilsEntrypoint:
		content = m.viewLuaUtilsEntrypoint()
	case ViewLuaUtilsOutput:
		content = m.viewLuaUtilsOutput()
	case ViewLuaUtilsRunning:
		content = m.viewLuaUtilsRunning()
	case ViewLuaUtilsSuccess, ViewLuaUtilsError:
		content = m.viewLuaUtilsResult()
	}

	// Create controls/help with proper width
	controls := m.createControls()

	// Create main layout with proper container
	mainLayout := lipgloss.JoinVertical(lipgloss.Left,
		header,
		content,
		controls,
	)

	// Create bordered container that fits the terminal with some margin
	container := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#902f17")).
		Padding(0, 1).
		Width(containerWidth).
		Render(mainLayout)

	// Center the container horizontally
	leftMargin := (m.width - containerWidth) / 2
	container = lipgloss.NewStyle().
		MarginLeft(leftMargin).
		Render(container)

	// Only center vertically if we have plenty of extra space
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

// resizeComponents updates component sizes when terminal is resized
func (m *Model) resizeComponents() {
	basePanelWidth := m.getPanelWidth()
	panelHeight := m.getPanelHeight()

	// Use the same width calculation as createTwoPanelLayout
	actualPanelWidth := basePanelWidth - 2 // Match the layout's panel width

	if m.commandSelector != nil {
		m.commandSelector.SetSize(actualPanelWidth, panelHeight)
	}
	if m.buildSelector != nil {
		m.buildSelector.SetSize(actualPanelWidth, panelHeight)
	}
	if m.outputInput != nil {
		m.outputInput.SetSize(actualPanelWidth, panelHeight)
	}
	if m.actionSelector != nil {
		m.actionSelector.SetSize(actualPanelWidth, panelHeight)
	}
	if m.filePicker != nil {
		m.filePicker.SetSize(actualPanelWidth, panelHeight)
	}
	if m.fileSelector != nil {
		m.fileSelector.SetSize(actualPanelWidth, panelHeight)
	}
	if m.configForm != nil {
		m.configForm.SetSize(actualPanelWidth, panelHeight)
	}
	// Note: progress and result components no longer need sizing since they use content methods
}

// getPanelWidth calculates the width for panels based on the container width
func (m *Model) getPanelWidth() int {
	// Container width with margin
	containerWidth := m.width - 10
	// Available width inside the container
	// Container has explicit width set, so only padding (2 chars) reduces available space
	layoutWidth := containerWidth - 2

	// Each panel gets half the layout width minus gap, but use more space
	// Gap is 1 char, so each panel gets (layoutWidth - 1) / 2, plus some extra
	basePanelWidth := (layoutWidth - 1) / 2
	panelWidth := basePanelWidth + 3 // Add 3 chars to each panel to use more space

	// Ensure minimum width
	if panelWidth < 15 {
		panelWidth = 15
	}

	return panelWidth
}

// getPanelHeight calculates the height for panels based on their content needs
func (m *Model) getPanelHeight() int {
	// Return a reasonable fixed height for panels - let them be compact
	// The container will size itself naturally based on the content
	return 12
}

// getContentHeight calculates the available height for content area (excluding header and footer)
func (m *Model) getContentHeight() int {
	// Terminal height minus header (2 lines) minus footer (2 lines) minus container borders/padding
	// Container has: 2 chars for top/bottom borders + 2 chars for top/bottom padding = 4 chars total
	// Header and footer are 2 lines each = 4 lines total
	// Make layout 1 line smaller to prevent overflow
	contentHeight := m.height - 4 - 4 - 1 // 4 for header+footer, 4 for container, 1 for safety

	// Ensure minimum height
	if contentHeight < 8 {
		contentHeight = 8
	}

	return contentHeight
}

// getViewTitle returns the title for the current view
func (m *Model) getViewTitle() string {
	switch m.state {
	case ViewCommandSelection:
		return "Select Command"
	case ViewBuildTypeSelection:
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
	case ViewBuildSuccess:
		return "Build Successful!"
	case ViewBuildError:
		return "Build Failed"
	case ViewLuaUtilsSelection:
		return "Select Lua Utils Command"
	case ViewLuaUtilsEntrypoint:
		return "Select Lua File to Bundle"
	case ViewLuaUtilsOutput:
		return "Select Output Path"
	case ViewLuaUtilsRunning:
		return "Bundling Lua Files"
	case ViewLuaUtilsSuccess:
		return "Bundle Successful!"
	case ViewLuaUtilsError:
		return "Bundle Failed"
	}
	return "Harlequin"
}

// viewCommandSelection renders the command selection view
func (m *Model) viewCommandSelection() string {
	if m.commandSelector == nil {
		return "Loading commands..."
	}

	leftPanel := m.commandSelector.View()

	// Right panel with description
	selected := m.commandSelector.GetSelected()
	description := "Welcome to Harlequin! Choose a command to get started."
	if selected != nil {
		switch selected.Value() {
		case "init":
			description = "Create a new AO process project from template.\n\nThis will guide you through:\n• Project name selection\n• Template language choice (Lua, C, Rust, AssemblyScript)\n• Author and GitHub information\n• Project directory setup\n\nAvailable templates:\n• Lua - With C trampoline and LuaRocks\n• C - With Conan and CMake\n• Rust - With Cargo and wasm-pack\n• AssemblyScript - With custom JSON handling\n\nEach template includes comprehensive documentation, testing, and build systems."
		case "build":
			description = "Interactive project building with guided configuration.\n\nThis will take you through:\n• Build type selection\n• Entrypoint file selection\n• Output directory configuration\n• Build configuration review\n• Actual build process\n\nThe TUI will guide you step-by-step through the entire build process with helpful descriptions and validation."
		case "lua-utils":
			description = "Lua utilities for bundling and processing Lua files.\n\nCurrently available:\n• Bundle - Combine multiple Lua files into a single executable\n\nThe bundle command will:\n• Analyze require() statements in your main Lua file\n• Recursively resolve all dependencies\n• Create a single bundled file with all modules\n• Handle circular dependencies gracefully"
		default:
			description = selected.Description()
		}
	}

	rightPanel := components.CreateDescriptionPanel(
		"Getting Started",
		description,
		m.getPanelWidth()-2, // Match the panel container width
		0,                   // Height not used anymore - panel sizes naturally
	)

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewBuildTypeSelection renders the build type selection view
func (m *Model) viewBuildTypeSelection() string {
	if m.buildSelector == nil {
		return "Loading build types..."
	}

	leftPanel := m.buildSelector.View()

	// Right panel with description
	selected := m.buildSelector.GetSelected()
	description := "Select a build configuration type to continue."
	if selected != nil {
		description = selected.Description()
	}

	rightPanel := components.CreateDescriptionPanel(
		"AOS Flavour",
		description,
		m.getPanelWidth()-2, // Match the panel container width
		0,                   // Height not used anymore - panel sizes naturally
	)

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewEntrypointSelection renders the entrypoint selection view
func (m *Model) viewEntrypointSelection() string {
	// Initialize the appropriate selector on first view
	if m.fileSelector == nil && !m.useFilePicker {
		// Try automatic discovery first
		cwd, _ := os.Getwd()
		actualPanelWidth := m.getPanelWidth() - 2 // Match layout panel width
		if selector, err := components.CreateEntrypointSelectorWithDiscovery(cwd, actualPanelWidth, m.getPanelHeight()); err == nil {
			m.fileSelector = selector
		} else {
			// Fall back to file picker if discovery fails
			m.useFilePicker = true
		}
	}

	if m.useFilePicker {
		// Use manual file picker
		if m.filePicker == nil {
			cwd, _ := os.Getwd()
			actualPanelWidth := m.getPanelWidth() - 2 // Match layout panel width
			m.filePicker = components.CreateEntrypointFilePicker(cwd, actualPanelWidth, m.getPanelHeight())
		}

		leftPanel := m.filePicker.View()

		// Right panel with file picker instructions
		rightPanel := components.CreateDescriptionPanel(
			"Manual File Selection",
			fmt.Sprintf("Current directory: %s\n\nNavigate with ↑/↓\nEnter directories with →\nSelect .lua files with Enter\n\nPress 'l' to switch to automatic discovery",
				m.filePicker.GetCurrentDirectory()),
			m.getPanelWidth()-2, // Match the panel container width
			0,                   // Height not used anymore
		)

		return m.createTwoPanelLayout(leftPanel, rightPanel)
	} else {
		// Use automatic discovery list
		leftPanel := m.fileSelector.View()

		// Right panel with discovery info
		selectedFile := ""
		if selected := m.fileSelector.GetSelected(); selected != nil {
			selectedFile = selected.Value()
		}

		description := "Automatically discovered .lua files in your project\n\nFiles are found recursively (excluding build directories)\n\nPress 'f' to switch to manual file picker"
		if selectedFile != "" {
			description = fmt.Sprintf("Selected: %s\n\n%s", selectedFile, description)
		}

		rightPanel := components.CreateDescriptionPanel(
			"Auto-discovered Files",
			description,
			m.getPanelWidth()-2, // Match the panel container width
			0,                   // Height not used anymore
		)

		return m.createTwoPanelLayout(leftPanel, rightPanel)
	}
}

// viewOutputDirectory renders the output directory selection view
func (m *Model) viewOutputDirectory() string {
	// Create output directory input if not exists
	if m.outputInput == nil {
		actualPanelWidth := m.getPanelWidth() - 2 // Match layout panel width
		m.outputInput = components.CreateOutputDirInput(actualPanelWidth, m.getPanelHeight())
	}

	leftPanel := m.outputInput.View()

	rightPanel := components.CreateDescriptionPanel(
		"Output Directory",
		"Enter the path where build outputs should be saved.\n\nThe build will create:\n• WASM binary\n• Lua bundle\n• Configuration files\n\nExamples:\n• ./dist\n• examples/dist\n• ./build",
		m.getPanelWidth()-2, // Match the panel container width
		0,                   // Height not used anymore
	)

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewConfigReview renders the configuration review view
func (m *Model) viewConfigReview() string {
	// Create config action selector if not exists
	if m.actionSelector == nil {
		actualPanelWidth := m.getPanelWidth() - 2 // Match layout panel width
		m.actionSelector = components.CreateConfigActionSelector(actualPanelWidth, m.getPanelHeight())
	}

	leftPanel := m.actionSelector.View()

	// Right panel with current config preview
	configPreview := m.formatConfigPreview()
	rightPanel := components.CreateDescriptionPanel(
		"Current Configuration",
		configPreview,
		m.getPanelWidth()-2, // Match the panel container width
		0,                   // Height not used anymore
	)

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewConfigEditing renders the configuration editing view
func (m *Model) viewConfigEditing() string {
	if m.configForm == nil {
		// Initialize config form
		actualPanelWidth := m.getPanelWidth() - 2 // Match layout panel width
		m.configForm = components.NewConfigForm(actualPanelWidth, m.getPanelHeight())

		// Load current config values
		if m.flow.Config != nil {
			m.configForm.SetFieldValues(
				m.flow.Config.Target,
				float64(m.flow.Config.StackSize)/(1024*1024),     // Convert to MB
				float64(m.flow.Config.InitialMemory)/(1024*1024), // Convert to MB
				float64(m.flow.Config.MaximumMemory)/(1024*1024), // Convert to MB
			)
		}
	}

	leftPanel := m.configForm.View()

	// Right panel with field description
	fieldTitle, fieldDesc := m.configForm.GetCurrentDescription()
	rightPanel := components.CreateDescriptionPanel(
		fieldTitle,
		fieldDesc,
		m.getPanelWidth()-2, // Match the panel container width
		0,                   // Height not used anymore
	)

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewBuildRunning renders the build progress view
func (m *Model) viewBuildRunning() string {
	leftPanel := ""
	if m.progress != nil {
		leftPanel = m.progress.ViewContent()
	}

	rightPanel := components.CreateDescriptionPanel(
		"Build Progress",
		"Building your project...\n\nThis may take a few minutes depending on your project size.",
		m.getPanelWidth()-2, // Match the panel container width
		0,                   // Height not used anymore
	)

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewBuildResult renders the build success/error view
func (m *Model) viewBuildResult() string {
	if m.result == nil {
		return "No result available"
	}

	leftPanel := m.result.ViewPanelContent()
	rightPanel := m.result.ViewDetailsContent()

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// createTwoPanelLayout creates a side-by-side panel layout with equal width distribution
func (m *Model) createTwoPanelLayout(leftPanel, rightPanel string) string {
	panelWidth := m.getPanelWidth() - 2 // Reduce width by 1 to prevent overflow
	contentHeight := m.getContentHeight()
	panelHeight := contentHeight - 2 // Make each panel 1 line smaller

	// Apply left panel style with border and calculated width (left-aligned content)
	leftStyled := components.LeftPanelStyle.
		Width(panelWidth).
		Height(panelHeight).
		Render(leftPanel)

	// Apply right panel style with calculated width/height (left-aligned content)
	rightStyled := components.RightPanelStyle.
		Width(panelWidth).
		Height(panelHeight).
		Render(rightPanel)

	// Fixed 1-character gap
	spacer := "" // Exactly 1 space

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyled,
		spacer,
		rightStyled,
	)
}

// createControls creates the bottom controls section
func (m *Model) createControls() string {
	var controls []string

	switch m.state {
	case ViewCommandSelection:
		controls = []string{"↑/↓ Navigate", "Enter Select", "q Quit"}
	case ViewBuildTypeSelection:
		controls = []string{"↑/↓ Navigate", "Enter Select", "Esc Back", "q Quit"}
	case ViewEntrypointSelection:
		if m.useFilePicker {
			controls = []string{"↑/↓ Navigate", "→ Enter Directory", "Enter Select", "l Auto-discover", "Esc Back", "q Quit"}
		} else {
			controls = []string{"↑/↓ Navigate", "Enter Select", "f File Picker", "Esc Back", "q Quit"}
		}
	case ViewOutputDirectory:
		controls = []string{"↑/↓ Navigate", "Enter Select", "Esc Back", "q Quit"}
	case ViewConfigReview:
		controls = []string{"↑/↓ Navigate", "Enter Select", "Esc Back", "q Quit"}
	case ViewConfigEditing:
		if m.configForm != nil && m.configForm.IsInButtons() {
			controls = []string{"←/→ Select Button", "Enter Confirm", "Esc Back", "q Quit"}
		} else {
			controls = []string{"↑/↓ Navigate", "←/→ Edit Values", "Tab Buttons", "Esc Back", "q Quit"}
		}
	case ViewBuildRunning:
		controls = []string{"Please wait...", "q Quit"}
	case ViewBuildSuccess, ViewBuildError:
		controls = []string{"Enter Exit", "q Quit"}
	case ViewLuaUtilsSelection:
		controls = []string{"↑/↓ Navigate", "Enter Select", "Esc Back", "q Quit"}
	case ViewLuaUtilsEntrypoint:
		if m.useLuaUtilsFilePicker {
			controls = []string{"↑/↓ Navigate", "→ Enter Directory", "Enter Select", "l Auto-discover", "Esc Back", "q Quit"}
		} else {
			controls = []string{"↑/↓ Navigate", "Enter Select", "f File Picker", "Esc Back", "q Quit"}
		}
	case ViewLuaUtilsOutput:
		controls = []string{"Type path", "Enter Confirm", "Esc Back", "q Quit"}
	case ViewLuaUtilsRunning:
		controls = []string{"Please wait...", "q Quit"}
	case ViewLuaUtilsSuccess, ViewLuaUtilsError:
		controls = []string{"Enter Exit", "q Quit"}
	}

	// Use container width for controls (with same margin as main container)
	containerWidth := m.width - 10
	return components.CreateControls(controls, containerWidth)
}

// formatConfigPreview formats the current config for preview
func (m *Model) formatConfigPreview() string {
	if m.flow.Config == nil {
		return "No configuration loaded"
	}

	config := m.flow.Config
	return fmt.Sprintf(`Build Type: %s
Entrypoint: %s
Output: %s

WASM Target: %d-bit
Stack Size: %.1f MB
Initial Memory: %.1f MB
Maximum Memory: %.1f MB`,
		m.flow.BuildType,
		m.flow.Entrypoint,
		m.flow.OutputDir,
		config.Target,
		float64(config.StackSize)/(1024*1024),
		float64(config.InitialMemory)/(1024*1024),
		float64(config.MaximumMemory)/(1024*1024),
	)
}

// Update handlers for each state

func (m *Model) updateCommandSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Pass all messages directly to the list component first
	model, cmd := m.commandSelector.Update(tea.Msg(msg))
	if newSelector, ok := model.(*components.ListSelectorComponent); ok {
		m.commandSelector = newSelector
	}

	// Check if enter was pressed after updating the component
	if key.Matches(msg, m.keyMap.Enter) {
		if selected := m.commandSelector.GetSelected(); selected != nil {
			switch selected.Value() {
			case "init":
				// Launch init command in non-interactive mode for now
				// TODO: Implement full TUI for init
				return m, tea.Quit
			case "build":
				// Go to build type selection
				m.state = ViewBuildTypeSelection
				return m, nil
			case "lua-utils":
				// Go to lua-utils selection
				m.state = ViewLuaUtilsSelection
				return m, nil
			}
		}
	}

	return m, cmd
}

func (m *Model) updateBuildTypeSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Pass all messages directly to the list component first
	model, cmd := m.buildSelector.Update(tea.Msg(msg))
	if newSelector, ok := model.(*components.ListSelectorComponent); ok {
		m.buildSelector = newSelector
	}

	// Check if enter was pressed after updating the component
	if key.Matches(msg, m.keyMap.Enter) {
		if selected := m.buildSelector.GetSelected(); selected != nil {
			m.flow.BuildType = selected.Value()
			m.state = ViewEntrypointSelection
			return m, nil
		}
	}

	return m, cmd
}

func (m *Model) updateEntrypointSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f":
		// Switch to file picker mode
		if !m.useFilePicker {
			m.useFilePicker = true
			m.filePicker = nil // Reset to reinitialize
			return m, nil
		}
	case "l":
		// Switch to list/automatic discovery mode
		if m.useFilePicker {
			m.useFilePicker = false
			m.fileSelector = nil // Reset to reinitialize
			return m, nil
		}
	}

	if key.Matches(msg, m.keyMap.Enter) {
		var selectedFile string

		if m.useFilePicker {
			if m.filePicker != nil && m.filePicker.HasSelection() {
				selectedFile = m.filePicker.GetSelectedFile()
			}
		} else {
			if m.fileSelector != nil {
				if selected := m.fileSelector.GetSelected(); selected != nil {
					selectedFile = selected.Value()
				}
			}
		}

		if selectedFile != "" && selectedFile != "No Lua files found" {
			m.flow.Entrypoint = selectedFile
			m.state = ViewOutputDirectory
			return m, nil
		}
	}

	// Update the active component
	if m.useFilePicker && m.filePicker != nil {
		model, cmd := m.filePicker.Update(msg)
		if newPicker, ok := model.(*components.FilePickerComponent); ok {
			m.filePicker = newPicker
		}
		return m, cmd
	} else if !m.useFilePicker && m.fileSelector != nil {
		model, cmd := m.fileSelector.Update(msg)
		if newSelector, ok := model.(*components.ListSelectorComponent); ok {
			m.fileSelector = newSelector
		}
		return m, cmd
	}

	return m, nil
}

func (m *Model) updateOutputDirectory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Pass all messages directly to the output input first
	if m.outputInput != nil {
		model, cmd := m.outputInput.Update(tea.Msg(msg))
		if newInput, ok := model.(*components.TextInputComponent); ok {
			m.outputInput = newInput
		}

		// Check if enter was pressed after updating the component
		if key.Matches(msg, m.keyMap.Enter) {
			value := m.outputInput.GetValue()
			if value != "" {
				m.flow.OutputDir = value

				// Load or create config
				m.flow.Config = m.loadOrCreateConfig()
				m.state = ViewConfigReview
				return m, nil
			}
		}

		return m, cmd
	}

	return m, nil
}

func (m *Model) updateConfigReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Pass all messages directly to the action selector first
	if m.actionSelector != nil {
		model, cmd := m.actionSelector.Update(tea.Msg(msg))
		if newSelector, ok := model.(*components.ListSelectorComponent); ok {
			m.actionSelector = newSelector
		}

		// Check if enter was pressed after updating the component
		if key.Matches(msg, m.keyMap.Enter) {
			if selected := m.actionSelector.GetSelected(); selected != nil {
				switch selected.Value() {
				case "use":
					// Proceed with current config - go to build
					m.state = ViewBuildRunning
					go m.runBuild() // Run build in background
					// Start progress animations
					return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
						return TickMsg{}
					})
				case "edit":
					// Go to config editing
					m.state = ViewConfigEditing
					return m, nil
				}
			}
		}

		return m, cmd
	}

	return m, nil
}

func (m *Model) updateConfigEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keyMap.Enter) && m.configForm.IsInButtons() {
		action := m.configForm.GetSelectedAction()
		if action == "save" {
			// Parse and save config
			if err := m.saveConfigFromForm(); err != nil {
				debug.Printf("Error saving config: %v", err)
				return m, nil
			}

			// Start build
			return m.startBuild()
		} else if action == "cancel" {
			m.state = ViewConfigReview
			return m, nil
		}
	}

	model, cmd := m.configForm.Update(msg)
	if newForm, ok := model.(*components.ConfigFormComponent); ok {
		m.configForm = newForm
	}
	return m, cmd
}

func (m *Model) updateBuildRunning(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Only allow quit during build
	return m, nil
}

func (m *Model) updateBuildResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle exit
	if key.Matches(msg, m.keyMap.Enter) {
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) handleBack() (tea.Model, tea.Cmd) {
	switch m.state {
	case ViewBuildTypeSelection:
		m.state = ViewCommandSelection
	case ViewEntrypointSelection:
		m.state = ViewBuildTypeSelection
	case ViewOutputDirectory:
		m.state = ViewEntrypointSelection
	case ViewConfigReview:
		m.state = ViewOutputDirectory
	case ViewConfigEditing:
		m.state = ViewConfigReview
	case ViewLuaUtilsSelection:
		m.state = ViewCommandSelection
	case ViewLuaUtilsEntrypoint:
		m.state = ViewLuaUtilsSelection
	case ViewLuaUtilsOutput:
		m.state = ViewLuaUtilsEntrypoint
	}

	return m, nil
}

// loadOrCreateConfig loads existing config or creates default
func (m *Model) loadOrCreateConfig() *config.Config {
	// Try to load existing config
	configPath := ".harlequin.yaml"
	if _, err := os.Stat(configPath); err == nil {
		if cfg := config.ReadConfigFile(configPath); cfg != nil {
			return cfg
		}
	}

	// Try build config
	buildConfigPath := filepath.Join("build_configs", "ao-build-config.yml")
	if _, err := os.Stat(buildConfigPath); err == nil {
		if cfg := config.ReadConfigFile(buildConfigPath); cfg != nil {
			return cfg
		}
	}

	// Create default config
	return config.NewConfig(nil)
}

// saveConfigFromForm saves the configuration from the form
func (m *Model) saveConfigFromForm() error {
	target, stackSize, initialMemory, maxMemory, err := m.configForm.GetFieldValues()
	if err != nil {
		return err
	}

	// Update config
	m.flow.Config.Target = target
	m.flow.Config.StackSize = stackSize
	m.flow.Config.InitialMemory = initialMemory
	m.flow.Config.MaximumMemory = maxMemory
	m.flow.ConfigEdited = true

	return nil
}

// startBuild initiates the build process
func (m *Model) startBuild() (tea.Model, tea.Cmd) {
	m.state = ViewBuildRunning

	// Start the build in a goroutine
	return m, func() tea.Msg {
		go m.runBuild()
		return nil
	}
}

// runBuild executes the actual build process
func (m *Model) runBuild() {
	debug.Printf("Starting build process")
	debug.Printf("Build config: %+v", m.flow)

	var buildErr error
	success := true

	// Execute the build using AOSBuilder directly
	buildErr = m.executeRealBuild()
	if buildErr != nil {
		debug.Printf("Build failed: %v", buildErr)
		success = false
	} else {
		debug.Printf("Build completed successfully")
	}

	// Send final result
	result := &BuildResult{
		Success: success,
		Error:   buildErr,
		Flow:    m.flow,
	}

	if m.program != nil {
		m.program.Send(BuildCompleteMsg{Result: result})
	}
}

// executeRealBuild runs the actual build process with progress updates
func (m *Model) executeRealBuild() error {
	debug.Printf("Executing real build for entrypoint: %s", m.flow.Entrypoint)

	// Debug: Print build configuration
	debug.Printf("Build configuration:")
	debug.Printf("  Entrypoint: %s", m.flow.Entrypoint)
	debug.Printf("  OutputDir: %s", m.flow.OutputDir)
	debug.Printf("  Config: %+v", m.flow.Config)

	// Create AOSBuilder and execute the complete build process
	builder := builders.NewAOSBuilder(builders.AOSBuilderParams{
		Config:     m.flow.Config,
		Entrypoint: m.flow.Entrypoint,
		OutputDir:  m.flow.OutputDir,
		Callbacks:  builders.NoOpCallbacks(), // Silent for now
	})
	debug.Printf("AOSBuilder created successfully")

	// Define build steps to match the progress component expectations
	steps := []struct {
		name string
		fn   func() error
	}{
		{"Copy AOS Files", func() error {
			debug.Printf("Step: Copy AOS Files")
			return nil // This will be handled by the full build
		}},
		{"Bundle Lua", func() error {
			debug.Printf("Step: Bundle Lua")
			return nil // This will be handled by the full build
		}},
		{"Inject Code", func() error {
			debug.Printf("Step: Inject Code")
			return nil // This will be handled by the full build
		}},
		{"Build WASM", func() error {
			debug.Printf("Step: Build WASM - executing full build process")
			err := builder.Build(m.ctx)
			if err != nil {
				debug.Printf("AOSBuilder.Build() failed: %v", err)
			} else {
				debug.Printf("AOSBuilder.Build() completed successfully")
			}
			return err
		}},
		{"Copy Outputs", func() error {
			debug.Printf("Step: Copy Outputs")
			return nil // This was handled by the build
		}},
		{"Cleanup", func() error {
			debug.Printf("Step: Cleanup")
			return nil // This was handled by the build
		}},
	}

	// Execute each step with progress updates
	for _, step := range steps {
		// Send step start message
		if m.program != nil {
			m.program.Send(BuildStepStartMsg{StepName: step.name})
		}

		// Execute the step
		err := step.fn()

		// Send step completion message
		if m.program != nil {
			m.program.Send(BuildStepCompleteMsg{StepName: step.name, Success: err == nil})
		}

		// If step failed, return the error
		if err != nil {
			return fmt.Errorf("step '%s' failed: %w", step.name, err)
		}

		// Small delay for visual feedback
		time.Sleep(200 * time.Millisecond)
	}

	return nil
}

// Lua Utils view and update functions

// viewLuaUtilsSelection renders the lua-utils command selection view
func (m *Model) viewLuaUtilsSelection() string {
	// Create lua-utils selector if not exists
	if m.luaUtilsSelector == nil {
		actualPanelWidth := m.getPanelWidth() - 2
		m.luaUtilsSelector = components.CreateLuaUtilsSelector(actualPanelWidth, m.getPanelHeight())
	}

	leftPanel := m.luaUtilsSelector.View()

	// Right panel with description
	selected := m.luaUtilsSelector.GetSelected()
	description := "Select a Lua utility command to run."
	if selected != nil {
		switch selected.Value() {
		case "bundle":
			description = "Bundle multiple Lua files into a single executable.\n\nThis command will:\n• Analyze your main Lua file for require() statements\n• Recursively resolve all dependencies\n• Handle circular dependencies gracefully\n• Create a self-contained Lua script\n• Preserve module structure and functionality\n\nThe bundled output includes all required modules as local functions with package loading mappings for require() compatibility."
		default:
			description = selected.Description()
		}
	}

	rightPanel := components.CreateDescriptionPanel(
		"Lua Utils",
		description,
		m.getPanelWidth()-2,
		0,
	)

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewLuaUtilsEntrypoint renders the lua-utils entrypoint selection view
func (m *Model) viewLuaUtilsEntrypoint() string {
	// Initialize the appropriate selector on first view
	if m.luaUtilsFileSelector == nil && !m.useLuaUtilsFilePicker {
		// Try automatic discovery first
		cwd, _ := os.Getwd()
		actualPanelWidth := m.getPanelWidth() - 2
		if selector, err := components.CreateEntrypointSelectorWithDiscovery(cwd, actualPanelWidth, m.getPanelHeight()); err == nil {
			m.luaUtilsFileSelector = selector
		} else {
			// Fall back to file picker if discovery fails
			m.useLuaUtilsFilePicker = true
		}
	}

	if m.useLuaUtilsFilePicker {
		// Use manual file picker
		if m.luaUtilsFilePicker == nil {
			cwd, _ := os.Getwd()
			actualPanelWidth := m.getPanelWidth() - 2
			m.luaUtilsFilePicker = components.CreateEntrypointFilePicker(cwd, actualPanelWidth, m.getPanelHeight())
		}

		leftPanel := m.luaUtilsFilePicker.View()

		rightPanel := components.CreateDescriptionPanel(
			"Manual File Selection",
			fmt.Sprintf("Current directory: %s\n\nNavigate with ↑/↓\nEnter directories with →\nSelect .lua files with Enter\n\nPress 'l' to switch to automatic discovery",
				m.luaUtilsFilePicker.GetCurrentDirectory()),
			m.getPanelWidth()-2,
			0,
		)

		return m.createTwoPanelLayout(leftPanel, rightPanel)
	} else {
		// Use automatic discovery list
		leftPanel := m.luaUtilsFileSelector.View()

		// Right panel with discovery info
		selectedFile := ""
		if selected := m.luaUtilsFileSelector.GetSelected(); selected != nil {
			selectedFile = selected.Value()
		}

		description := "Select the main Lua file to bundle.\n\nFiles are found recursively (excluding build directories)\n\nPress 'f' to switch to manual file picker"
		if selectedFile != "" {
			description = fmt.Sprintf("Selected: %s\n\n%s", selectedFile, description)
		}

		rightPanel := components.CreateDescriptionPanel(
			"Auto-discovered Files",
			description,
			m.getPanelWidth()-2,
			0,
		)

		return m.createTwoPanelLayout(leftPanel, rightPanel)
	}
}

// viewLuaUtilsOutput renders the lua-utils output path selection view
func (m *Model) viewLuaUtilsOutput() string {
	// Create output path input if not exists
	if m.luaUtilsOutputInput == nil {
		actualPanelWidth := m.getPanelWidth() - 2
		m.luaUtilsOutputInput = components.CreateOutputDirInput(actualPanelWidth, m.getPanelHeight())

		// Set a default value based on entrypoint
		if m.luaUtilsFlow.Entrypoint != "" {
			dir := filepath.Dir(m.luaUtilsFlow.Entrypoint)
			base := filepath.Base(m.luaUtilsFlow.Entrypoint)
			ext := filepath.Ext(base)
			name := strings.TrimSuffix(base, ext)
			defaultOutput := filepath.Join(dir, name+".bundled.lua")
			m.luaUtilsOutputInput.SetValue(defaultOutput)
		}
	}

	leftPanel := m.luaUtilsOutputInput.View()

	rightPanel := components.CreateDescriptionPanel(
		"Output Path",
		"Enter the path where the bundled Lua file should be saved.\n\nThe bundled file will contain:\n• All required modules as local functions\n• Package loading mappings\n• Your main file content\n\nExample paths:\n• ./bundle.lua\n• dist/app.bundled.lua\n• output/combined.lua\n\nLeave empty to use default path based on entrypoint.",
		m.getPanelWidth()-2,
		0,
	)

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewLuaUtilsRunning renders the lua-utils progress view
func (m *Model) viewLuaUtilsRunning() string {
	leftPanel := ""
	if m.progress != nil {
		leftPanel = m.progress.ViewContent()
	}

	rightPanel := components.CreateDescriptionPanel(
		"Bundling Progress",
		"Bundling your Lua files...\n\nThis process:\n• Analyzes dependency tree\n• Resolves require() statements\n• Handles circular dependencies\n• Creates bundled output",
		m.getPanelWidth()-2,
		0,
	)

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewLuaUtilsResult renders the lua-utils result view
func (m *Model) viewLuaUtilsResult() string {
	if m.result == nil {
		return "No result available"
	}

	leftPanel := m.result.ViewPanelContent()
	rightPanel := m.result.ViewDetailsContent()

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// Update functions for lua-utils states

// updateLuaUtilsSelection handles lua-utils command selection
func (m *Model) updateLuaUtilsSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Pass all messages directly to the list component first
	model, cmd := m.luaUtilsSelector.Update(tea.Msg(msg))
	if newSelector, ok := model.(*components.ListSelectorComponent); ok {
		m.luaUtilsSelector = newSelector
	}

	// Check if enter was pressed after updating the component
	if key.Matches(msg, m.keyMap.Enter) {
		if selected := m.luaUtilsSelector.GetSelected(); selected != nil {
			switch selected.Value() {
			case "bundle":
				m.luaUtilsFlow.Command = "bundle"
				m.state = ViewLuaUtilsEntrypoint
				return m, nil
			}
		}
	}

	return m, cmd
}

// updateLuaUtilsEntrypoint handles lua-utils entrypoint selection
func (m *Model) updateLuaUtilsEntrypoint(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f":
		// Switch to file picker mode
		if !m.useLuaUtilsFilePicker {
			m.useLuaUtilsFilePicker = true
			m.luaUtilsFilePicker = nil // Reset to reinitialize
			return m, nil
		}
	case "l":
		// Switch to list/automatic discovery mode
		if m.useLuaUtilsFilePicker {
			m.useLuaUtilsFilePicker = false
			m.luaUtilsFileSelector = nil // Reset to reinitialize
			return m, nil
		}
	}

	if key.Matches(msg, m.keyMap.Enter) {
		var selectedFile string

		if m.useLuaUtilsFilePicker {
			if m.luaUtilsFilePicker != nil && m.luaUtilsFilePicker.HasSelection() {
				selectedFile = m.luaUtilsFilePicker.GetSelectedFile()
			}
		} else {
			if m.luaUtilsFileSelector != nil {
				if selected := m.luaUtilsFileSelector.GetSelected(); selected != nil {
					selectedFile = selected.Value()
				}
			}
		}

		if selectedFile != "" && selectedFile != "No Lua files found" {
			m.luaUtilsFlow.Entrypoint = selectedFile
			m.state = ViewLuaUtilsOutput
			return m, nil
		}
	}

	// Update the active component
	if m.useLuaUtilsFilePicker && m.luaUtilsFilePicker != nil {
		model, cmd := m.luaUtilsFilePicker.Update(msg)
		if newPicker, ok := model.(*components.FilePickerComponent); ok {
			m.luaUtilsFilePicker = newPicker
		}
		return m, cmd
	} else if !m.useLuaUtilsFilePicker && m.luaUtilsFileSelector != nil {
		model, cmd := m.luaUtilsFileSelector.Update(msg)
		if newSelector, ok := model.(*components.ListSelectorComponent); ok {
			m.luaUtilsFileSelector = newSelector
		}
		return m, cmd
	}

	return m, nil
}

// updateLuaUtilsOutput handles lua-utils output path selection
func (m *Model) updateLuaUtilsOutput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Pass all messages directly to the output input first
	if m.luaUtilsOutputInput != nil {
		model, cmd := m.luaUtilsOutputInput.Update(tea.Msg(msg))
		if newInput, ok := model.(*components.TextInputComponent); ok {
			m.luaUtilsOutputInput = newInput
		}

		// Check if enter was pressed after updating the component
		if key.Matches(msg, m.keyMap.Enter) {
			value := m.luaUtilsOutputInput.GetValue()

			if value != "" {
				m.luaUtilsFlow.OutputPath = value
				m.state = ViewLuaUtilsRunning
				go m.runLuaUtilsBundle() // Run bundle in background
				return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
					return TickMsg{}
				})
			}
		}

		return m, cmd
	}

	return m, nil
}

// updateLuaUtilsRunning handles lua-utils running state
func (m *Model) updateLuaUtilsRunning(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Only allow quit during bundling
	return m, nil
}

// updateLuaUtilsResult handles lua-utils result state
func (m *Model) updateLuaUtilsResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle exit
	if key.Matches(msg, m.keyMap.Enter) {
		return m, tea.Quit
	}

	return m, nil
}

// runLuaUtilsBundle executes the lua-utils bundle process
func (m *Model) runLuaUtilsBundle() {
	debug.Printf("Starting lua-utils bundle process")
	debug.Printf("Entrypoint: %s", m.luaUtilsFlow.Entrypoint)
	debug.Printf("Output: %s", m.luaUtilsFlow.OutputPath)

	var bundleErr error
	success := true

	// Import lua utils (already imported at top)

	// Perform the bundling
	bundledContent, err := luautils.Bundle(m.luaUtilsFlow.Entrypoint)
	if err != nil {
		debug.Printf("Bundle failed: %v", err)
		bundleErr = err
		success = false
	} else {
		// Write the bundled content to the output file
		// Ensure output directory exists
		outputDir := filepath.Dir(m.luaUtilsFlow.OutputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			debug.Printf("Failed to create output directory: %v", err)
			bundleErr = fmt.Errorf("failed to create output directory: %w", err)
			success = false
		} else {
			// Write the bundled content
			if err := os.WriteFile(m.luaUtilsFlow.OutputPath, []byte(bundledContent), 0644); err != nil {
				debug.Printf("Failed to write bundled file: %v", err)
				bundleErr = fmt.Errorf("failed to write bundled file: %w", err)
				success = false
			} else {
				debug.Printf("Bundle completed successfully: %s", m.luaUtilsFlow.OutputPath)
			}
		}
	}

	// Send final result
	result := &LuaUtilsResult{
		Success: success,
		Error:   bundleErr,
		Flow:    m.luaUtilsFlow,
	}

	if m.program != nil {
		m.program.Send(LuaUtilsCompleteMsg{Result: result})
	}
}

// RunBuildTUI starts the modernized interactive build TUI
func RunBuildTUI(ctx context.Context) error {
	m := NewModel(ctx)

	// Start the Bubble Tea program
	p := tea.NewProgram(m, tea.WithAltScreen())
	m.program = p // Store reference for sending messages

	_, err := p.Run()
	return err
}
