package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everFinance/goar/utils"
	"github.com/project-kardeshev/go-ardrive-turbo/pkg/signers"
	"github.com/project-kardeshev/go-ardrive-turbo/pkg/turbo"
	"github.com/project-kardeshev/go-ardrive-turbo/pkg/types"
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
	ViewInitWizard
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
	ViewUploadWasmSelection
	ViewUploadConfigSelection
	ViewUploadWalletSelection
	ViewUploadVersion
	ViewUploadGitHash
	ViewUploadDryRun
	ViewUploadConfirmation
	ViewUploadRunning
	ViewUploadSuccess
	ViewUploadError
)

// Model represents the modernized TUI application state
type Model struct {
	// Core state
	state ViewState
	flow  *BuildFlow
	luaUtilsFlow *LuaUtilsFlow
	uploadFlow *UploadFlow
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

	// Upload Module components
	uploadWasmSelector    *components.ListSelectorComponent
	uploadWasmFilePicker  *components.FilePickerComponent
	uploadConfigSelector  *components.ListSelectorComponent
	uploadConfigFilePicker *components.FilePickerComponent
	uploadWalletSelector  *components.ListSelectorComponent
	uploadWalletFilePicker *components.FilePickerComponent
	uploadVersionInput    *components.TextInputComponent
	uploadGitHashInput    *components.TextInputComponent
	uploadDryRunSelector  *components.ListSelectorComponent
	uploadConfirmSelector *components.ListSelectorComponent
	uploadProgress        *components.ProgressComponent // Separate progress for uploads

	// Init wizard component
	initWizard *components.InitWizardComponent

	// Layout
	width  int
	height int

	// File selection mode
	useFilePicker bool // true = manual picker, false = automatic list
	useLuaUtilsFilePicker bool // for lua-utils file selection
	useUploadWasmFilePicker bool // for upload wasm file selection
	useUploadConfigFilePicker bool // for upload config file selection
	useUploadWalletFilePicker bool // for upload wallet file selection

	// Build process
	buildResult *BuildResult
	luaUtilsResult *LuaUtilsResult
	uploadResult *UploadResult
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

// UploadFlow represents the upload-module configuration flow
type UploadFlow struct {
	WasmFile   string
	ConfigFile string
	WalletFile string
	Version    string
	GitHash    string
	DryRun     bool
	Balance    string
	EstimatedCost string
	BalanceCheckError string  // Store balance check error for display
}

// UploadResult holds the result of an upload operation
type UploadResult struct {
	Success    bool
	Error      error
	Flow       *UploadFlow
	DataItemID string // Store the transaction/data item ID
	Output     string // Store the complete output for parsing
}

// Messages for Bubble Tea
type BuildStepStartMsg struct{ StepName string }
type BuildStepCompleteMsg struct {
	StepName string
	Success  bool
}
type UploadStepStartMsg struct{ StepName string }
type UploadStepCompleteMsg struct {
	StepName string
	Success  bool
}
type BuildCompleteMsg struct{ Result *BuildResult }
type LuaUtilsCompleteMsg struct{ Result *LuaUtilsResult }
type UploadCompleteMsg struct{ Result *UploadResult }
type TickMsg struct{}

// NewModel creates a new modernized TUI model
func NewModel(ctx context.Context) *Model {
	// Initialize components
	keyMap := components.DefaultKeyMap()
	helpModel := help.New()

	// Create model with initial size
	model := &Model{
		state:        ViewCommandSelection,
		flow:         &BuildFlow{},
		luaUtilsFlow: &LuaUtilsFlow{},
		uploadFlow:   &UploadFlow{},
		ctx:          ctx,
		keyMap:       keyMap,
		help:         helpModel,
		width:        80, // Default width
		height:       24, // Default height
	}

	// Calculate initial panel dimensions
	panelWidth := model.getPanelWidth() - 2
	panelHeight := model.getPanelHeight()

	// Create components with proper dimensions
	model.commandSelector = components.CreateCommandSelector(panelWidth, panelHeight)
	model.buildSelector = components.CreateBuildTypeSelector(panelWidth, panelHeight)
	model.progress = components.NewProgressComponent(panelWidth, panelHeight)

	return model
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
		case ViewInitWizard:
			return m.updateInitWizard(msg)
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
		case ViewUploadWasmSelection:
			return m.updateUploadWasmSelection(msg)
		case ViewUploadConfigSelection:
			return m.updateUploadConfigSelection(msg)
		case ViewUploadWalletSelection:
			return m.updateUploadWalletSelection(msg)
		case ViewUploadVersion:
			return m.updateUploadVersion(msg)
		case ViewUploadGitHash:
			return m.updateUploadGitHash(msg)
		case ViewUploadDryRun:
			return m.updateUploadDryRun(msg)
		case ViewUploadConfirmation:
			return m.updateUploadConfirmation(msg)
		case ViewUploadRunning:
			return m.updateUploadRunning(msg)
		case ViewUploadSuccess, ViewUploadError:
			return m.updateUploadResult(msg)
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

	case UploadStepStartMsg:
		if m.uploadProgress != nil {
			steps := []components.BuildStep{
				{Name: msg.StepName, Status: components.StepRunning},
			}
			m.uploadProgress.UpdateSteps(steps)
		}

	case UploadStepCompleteMsg:
		if m.uploadProgress != nil {
			status := components.StepSuccess
			if !msg.Success {
				status = components.StepFailed
			}
			steps := []components.BuildStep{
				{Name: msg.StepName, Status: status},
			}
			m.uploadProgress.UpdateSteps(steps)
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

	case UploadCompleteMsg:
		m.uploadResult = msg.Result
		if msg.Result.Success {
			m.state = ViewUploadSuccess
		} else {
			m.state = ViewUploadError
		}

		// Create result component for upload
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
	case ViewInitWizard:
		content = m.viewInitWizard()
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
	case ViewUploadWasmSelection:
		content = m.viewUploadWasmSelection()
	case ViewUploadConfigSelection:
		content = m.viewUploadConfigSelection()
	case ViewUploadWalletSelection:
		content = m.viewUploadWalletSelection()
	case ViewUploadVersion:
		content = m.viewUploadVersion()
	case ViewUploadGitHash:
		content = m.viewUploadGitHash()
	case ViewUploadDryRun:
		content = m.viewUploadDryRun()
	case ViewUploadConfirmation:
		content = m.viewUploadConfirmation()
	case ViewUploadRunning:
		content = m.viewUploadRunning()
	case ViewUploadSuccess, ViewUploadError:
		content = m.viewUploadResult()
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

	// Resize all list components with the updated height
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

	// Resize upload flow components
	if m.uploadWasmSelector != nil {
		m.uploadWasmSelector.SetSize(actualPanelWidth, panelHeight)
	}
	if m.uploadConfigSelector != nil {
		m.uploadConfigSelector.SetSize(actualPanelWidth, panelHeight)
	}
	if m.uploadWalletSelector != nil {
		m.uploadWalletSelector.SetSize(actualPanelWidth, panelHeight)
	}
	if m.uploadDryRunSelector != nil {
		m.uploadDryRunSelector.SetSize(actualPanelWidth, panelHeight)
	}

	// Resize lua utils components
	if m.luaUtilsSelector != nil {
		m.luaUtilsSelector.SetSize(actualPanelWidth, panelHeight)
	}
	if m.luaUtilsFileSelector != nil {
		m.luaUtilsFileSelector.SetSize(actualPanelWidth, panelHeight)
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

// getPanelHeight calculates the height for panels based on available space
func (m *Model) getPanelHeight() int {
	// Use the full available content height minus some padding
	contentHeight := m.getContentHeight()
	// Reserve 4 lines for panel borders/padding/title
	panelHeight := contentHeight - 4

	// Ensure minimum height
	if panelHeight < 8 {
		panelHeight = 8
	}

	// Ensure maximum reasonable height
	if panelHeight > 20 {
		panelHeight = 20
	}

	return panelHeight
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
	case ViewInitWizard:
		return "Initialize New Project"
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
	case ViewUploadWasmSelection:
		return "Select WASM File"
	case ViewUploadConfigSelection:
		return "Select Config File"
	case ViewUploadWalletSelection:
		return "Select Wallet File"
	case ViewUploadVersion:
		return "Enter Version"
	case ViewUploadGitHash:
		return "Enter Git Hash"
	case ViewUploadDryRun:
		return "Select Upload Mode"
	case ViewUploadConfirmation:
		return "Confirm Upload"
	case ViewUploadRunning:
		return "Uploading Module"
	case ViewUploadSuccess:
		return "Upload Successful!"
	case ViewUploadError:
		return "Upload Failed"
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
		case "upload-module":
			description = "Upload built WASM modules to Arweave with comprehensive metadata.\n\nThis will guide you through:\n• WASM file selection\n• Configuration file selection\n• Wallet file selection\n• Version and git hash configuration\n• Upload mode selection (dry run vs actual)\n\nFeatures:\n• Automatic WASM metadata extraction\n• JSON export analysis\n• Comprehensive tagging\n• Progress tracking\n• Dry run validation"
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
	panelHeight := contentHeight - 4 // Reserve space for borders and padding

	// Ensure minimum panel height
	if panelHeight < 8 {
		panelHeight = 8
	}

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
	spacer := " " // Exactly 1 space

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
	case ViewUploadWasmSelection:
		if m.useUploadWasmFilePicker {
			controls = []string{"↑/↓ Navigate", "→ Enter Directory", "Enter Select", "l Auto-discover", "Esc Back", "q Quit"}
		} else {
			controls = []string{"↑/↓ Navigate", "Enter Select", "f File Picker", "Esc Back", "q Quit"}
		}
	case ViewUploadConfigSelection:
		if m.useUploadConfigFilePicker {
			controls = []string{"↑/↓ Navigate", "→ Enter Directory", "Enter Select", "l Auto-discover", "Esc Back", "q Quit"}
		} else {
			controls = []string{"↑/↓ Navigate", "Enter Select", "f File Picker", "Esc Back", "q Quit"}
		}
	case ViewUploadWalletSelection:
		if m.useUploadWalletFilePicker {
			controls = []string{"↑/↓ Navigate", "→ Enter Directory", "Enter Select", "l Auto-discover", "Esc Back", "q Quit"}
		} else {
			controls = []string{"↑/↓ Navigate", "Enter Select", "f File Picker", "Esc Back", "q Quit"}
		}
	case ViewUploadVersion:
		controls = []string{"Type version", "Enter Confirm", "Esc Back", "q Quit"}
	case ViewUploadGitHash:
		controls = []string{"Type git hash", "Enter Confirm", "Esc Back", "q Quit"}
	case ViewUploadDryRun:
		controls = []string{"↑/↓ Navigate", "Enter Select", "Esc Back", "q Quit"}
	case ViewUploadConfirmation:
		controls = []string{"↑/↓ Navigate", "Enter Select", "Esc Back", "q Quit"}
	case ViewUploadRunning:
		controls = []string{"Please wait...", "q Quit"}
	case ViewUploadSuccess, ViewUploadError:
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
				// Go to init wizard
				m.state = ViewInitWizard
				return m, nil
			case "build":
				// Go to build type selection
				m.state = ViewBuildTypeSelection
				return m, nil
			case "upload-module":
				// Go to upload WASM selection
				m.state = ViewUploadWasmSelection
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
	case ViewInitWizard:
		m.state = ViewCommandSelection
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
	case ViewUploadWasmSelection:
		m.state = ViewCommandSelection
	case ViewUploadConfigSelection:
		m.state = ViewUploadWasmSelection
	case ViewUploadWalletSelection:
		m.state = ViewUploadConfigSelection
	case ViewUploadVersion:
		m.state = ViewUploadWalletSelection
	case ViewUploadGitHash:
		m.state = ViewUploadVersion
	case ViewUploadDryRun:
		m.state = ViewUploadVersion
	case ViewUploadConfirmation:
		m.state = ViewUploadDryRun
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

// Upload Module update handlers

// updateUploadWasmSelection handles WASM file selection
func (m *Model) updateUploadWasmSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f":
		// Switch to file picker mode
		if !m.useUploadWasmFilePicker {
			m.useUploadWasmFilePicker = true
			m.uploadWasmFilePicker = nil // Reset to reinitialize
			return m, nil
		}
	case "l":
		// Switch to list/automatic discovery mode
		if m.useUploadWasmFilePicker {
			m.useUploadWasmFilePicker = false
			m.uploadWasmSelector = nil // Reset to reinitialize
			return m, nil
		}
	}

	if key.Matches(msg, m.keyMap.Enter) {
		var selectedFile string

		if m.useUploadWasmFilePicker {
			if m.uploadWasmFilePicker != nil && m.uploadWasmFilePicker.HasSelection() {
				selectedFile = m.uploadWasmFilePicker.GetSelectedFile()
			}
		} else {
			if m.uploadWasmSelector != nil {
				if selected := m.uploadWasmSelector.GetSelected(); selected != nil {
					selectedFile = selected.Value()
				}
			}
		}

		if selectedFile != "" && selectedFile != "No WASM files found" {
			m.uploadFlow.WasmFile = selectedFile
			m.state = ViewUploadConfigSelection
			return m, nil
		}
	}

	// Update the active component
	if m.useUploadWasmFilePicker && m.uploadWasmFilePicker != nil {
		model, cmd := m.uploadWasmFilePicker.Update(msg)
		if newPicker, ok := model.(*components.FilePickerComponent); ok {
			m.uploadWasmFilePicker = newPicker
		}
		return m, cmd
	} else if !m.useUploadWasmFilePicker && m.uploadWasmSelector != nil {
		model, cmd := m.uploadWasmSelector.Update(msg)
		if newSelector, ok := model.(*components.ListSelectorComponent); ok {
			m.uploadWasmSelector = newSelector
		}
		return m, cmd
	}

	return m, nil
}

// updateUploadConfigSelection handles config file selection
func (m *Model) updateUploadConfigSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f":
		// Switch to file picker mode
		if !m.useUploadConfigFilePicker {
			m.useUploadConfigFilePicker = true
			m.uploadConfigFilePicker = nil // Reset to reinitialize
			return m, nil
		}
	case "l":
		// Switch to list/automatic discovery mode
		if m.useUploadConfigFilePicker {
			m.useUploadConfigFilePicker = false
			m.uploadConfigSelector = nil // Reset to reinitialize
			return m, nil
		}
	}

	if key.Matches(msg, m.keyMap.Enter) {
		var selectedFile string

		if m.useUploadConfigFilePicker {
			if m.uploadConfigFilePicker != nil && m.uploadConfigFilePicker.HasSelection() {
				selectedFile = m.uploadConfigFilePicker.GetSelectedFile()
			}
		} else {
			if m.uploadConfigSelector != nil {
				if selected := m.uploadConfigSelector.GetSelected(); selected != nil {
					selectedFile = selected.Value()
				}
			}
		}

		if selectedFile != "" && selectedFile != "No config files found" {
			m.uploadFlow.ConfigFile = selectedFile
			m.state = ViewUploadWalletSelection
			return m, nil
		}
	}

	// Update the active component
	if m.useUploadConfigFilePicker && m.uploadConfigFilePicker != nil {
		model, cmd := m.uploadConfigFilePicker.Update(msg)
		if newPicker, ok := model.(*components.FilePickerComponent); ok {
			m.uploadConfigFilePicker = newPicker
		}
		return m, cmd
	} else if !m.useUploadConfigFilePicker && m.uploadConfigSelector != nil {
		model, cmd := m.uploadConfigSelector.Update(msg)
		if newSelector, ok := model.(*components.ListSelectorComponent); ok {
			m.uploadConfigSelector = newSelector
		}
		return m, cmd
	}

	return m, nil
}

// updateUploadWalletSelection handles wallet file selection
func (m *Model) updateUploadWalletSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f":
		// Switch to file picker mode
		if !m.useUploadWalletFilePicker {
			m.useUploadWalletFilePicker = true
			m.uploadWalletFilePicker = nil // Reset to reinitialize
			return m, nil
		}
	case "l":
		// Switch to list/automatic discovery mode
		if m.useUploadWalletFilePicker {
			m.useUploadWalletFilePicker = false
			m.uploadWalletSelector = nil // Reset to reinitialize
			return m, nil
		}
	}

	if key.Matches(msg, m.keyMap.Enter) {
		var selectedFile string

		if m.useUploadWalletFilePicker {
			if m.uploadWalletFilePicker != nil && m.uploadWalletFilePicker.HasSelection() {
				selectedFile = m.uploadWalletFilePicker.GetSelectedFile()
			}
		} else {
			if m.uploadWalletSelector != nil {
				if selected := m.uploadWalletSelector.GetSelected(); selected != nil {
					selectedFile = selected.Value()
				}
			}
		}

		if selectedFile != "" && selectedFile != "No wallet files found" {
			m.uploadFlow.WalletFile = selectedFile
			m.state = ViewUploadVersion
			return m, nil
		}
	}

	// Update the active component
	if m.useUploadWalletFilePicker && m.uploadWalletFilePicker != nil {
		model, cmd := m.uploadWalletFilePicker.Update(msg)
		if newPicker, ok := model.(*components.FilePickerComponent); ok {
			m.uploadWalletFilePicker = newPicker
		}
		return m, cmd
	} else if !m.useUploadWalletFilePicker && m.uploadWalletSelector != nil {
		model, cmd := m.uploadWalletSelector.Update(msg)
		if newSelector, ok := model.(*components.ListSelectorComponent); ok {
			m.uploadWalletSelector = newSelector
		}
		return m, cmd
	}

	return m, nil
}

// updateUploadVersion handles version input
func (m *Model) updateUploadVersion(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Pass all messages directly to the input first
	if m.uploadVersionInput != nil {
		model, cmd := m.uploadVersionInput.Update(tea.Msg(msg))
		if newInput, ok := model.(*components.TextInputComponent); ok {
			m.uploadVersionInput = newInput
		}

		// Check if enter was pressed after updating the component
		if key.Matches(msg, m.keyMap.Enter) {
			value := m.uploadVersionInput.GetValue()
			if value != "" {
				m.uploadFlow.Version = value
				// Auto-set git hash from environment if available, skip manual input
				if gitHash := os.Getenv("GITHUB_SHA"); gitHash != "" {
					m.uploadFlow.GitHash = gitHash
				}
				m.state = ViewUploadDryRun
				return m, nil
			}
		}

		return m, cmd
	}

	return m, nil
}

// updateUploadGitHash handles git hash input
func (m *Model) updateUploadGitHash(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Pass all messages directly to the input first
	if m.uploadGitHashInput != nil {
		model, cmd := m.uploadGitHashInput.Update(tea.Msg(msg))
		if newInput, ok := model.(*components.TextInputComponent); ok {
			m.uploadGitHashInput = newInput
		}

		// Check if enter was pressed after updating the component
		if key.Matches(msg, m.keyMap.Enter) {
			value := m.uploadGitHashInput.GetValue()
			m.uploadFlow.GitHash = value // Can be empty
			m.state = ViewUploadDryRun
			return m, nil
		}

		return m, cmd
	}

	return m, nil
}

// updateUploadDryRun handles dry run selection
func (m *Model) updateUploadDryRun(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Pass all messages directly to the selector first
	if m.uploadDryRunSelector != nil {
		model, cmd := m.uploadDryRunSelector.Update(tea.Msg(msg))
		if newSelector, ok := model.(*components.ListSelectorComponent); ok {
			m.uploadDryRunSelector = newSelector
		}

		// Check if enter was pressed after updating the component
		if key.Matches(msg, m.keyMap.Enter) {
			if selected := m.uploadDryRunSelector.GetSelected(); selected != nil {
				m.uploadFlow.DryRun = selected.Value() == "true"

								// If user selected actual upload, check balance before proceeding
				if !m.uploadFlow.DryRun {
					err := m.checkBalanceAndCost()
					if err != nil {
						debug.Printf("Balance check failed: %v", err)
						// Store the error for display in confirmation screen
						m.uploadFlow.BalanceCheckError = err.Error()
					} else {
						// Clear any previous error
						m.uploadFlow.BalanceCheckError = ""
					}
				}

				// Reset confirmation selector to rebuild with new balance info
				m.uploadConfirmSelector = nil
				m.state = ViewUploadConfirmation
				return m, nil
			}
		}

		return m, cmd
	}

	return m, nil
}

// updateUploadConfirmation handles upload confirmation
func (m *Model) updateUploadConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Pass all messages directly to the selector first
	if m.uploadConfirmSelector != nil {
		model, cmd := m.uploadConfirmSelector.Update(tea.Msg(msg))
		if newSelector, ok := model.(*components.ListSelectorComponent); ok {
			m.uploadConfirmSelector = newSelector
		}

		// Check if enter was pressed after updating the component
		if key.Matches(msg, m.keyMap.Enter) {
			if selected := m.uploadConfirmSelector.GetSelected(); selected != nil {
								switch selected.Value() {
						case "confirm":
			m.state = ViewUploadRunning

			// Initialize upload progress component
			panelWidth := m.getPanelWidth() - 2
			panelHeight := m.getPanelHeight()
			m.uploadProgress = components.NewProgressComponent(panelWidth, panelHeight)

			go m.runUpload() // Run upload in background
			return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
				return TickMsg{}
			})
				case "insufficient":
					// Do nothing - insufficient balance prevents upload
					// User can only cancel to go back and fix the issue
					return m, nil
				case "cancel":
					m.state = ViewUploadDryRun
					return m, nil
				}
			}
		}

		return m, cmd
	}

	return m, nil
}

// updateUploadRunning handles upload running state
func (m *Model) updateUploadRunning(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Only allow quit during upload
	return m, nil
}

// updateUploadResult handles upload result state
func (m *Model) updateUploadResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle exit
	if key.Matches(msg, m.keyMap.Enter) {
		return m, tea.Quit
	}

	return m, nil
}

// runUpload executes the actual upload process
func (m *Model) runUpload() {
	debug.Printf("Starting upload process")
	debug.Printf("Upload config: %+v", m.uploadFlow)

	// Send upload step start messages
	if m.program != nil {
		m.program.Send(UploadStepStartMsg{StepName: "Analyzing WASM metadata"})
	}

	var uploadErr error
	var output string
	var dataItemID string
	success := true

	// Execute the actual upload using the upload command
	output, uploadErr = m.executeRealUpload()
	if uploadErr != nil {
		debug.Printf("Upload failed: %v", uploadErr)
		success = false
		if m.program != nil {
			m.program.Send(UploadStepCompleteMsg{StepName: "Upload failed", Success: false})
		}
	} else {
		debug.Printf("Upload completed successfully")
		dataItemID = m.extractDataItemID(output)
		if m.program != nil {
			m.program.Send(UploadStepCompleteMsg{StepName: "Analyzing WASM metadata", Success: true})
			m.program.Send(UploadStepStartMsg{StepName: "Creating upload tags"})
			m.program.Send(UploadStepCompleteMsg{StepName: "Creating upload tags", Success: true})
			m.program.Send(UploadStepStartMsg{StepName: "Signing data item"})
			m.program.Send(UploadStepCompleteMsg{StepName: "Signing data item", Success: true})
			m.program.Send(UploadStepStartMsg{StepName: "Uploading to Arweave"})
			m.program.Send(UploadStepCompleteMsg{StepName: "Uploading to Arweave", Success: true})
		}
	}

	// Send final result
	result := &UploadResult{
		Success:    success,
		Error:      uploadErr,
		Flow:       m.uploadFlow,
		DataItemID: dataItemID,
		Output:     output,
	}

	if m.program != nil {
		m.program.Send(UploadCompleteMsg{Result: result})
	}
}

// executeRealUpload runs the actual upload process
func (m *Model) executeRealUpload() (string, error) {
	debug.Printf("Executing real upload for WASM: %s", m.uploadFlow.WasmFile)

	// Build the arguments for calling the harlequin binary directly
	args := []string{"upload-module"}
	if m.uploadFlow.WasmFile != "" {
		args = append(args, "--wasm-file", m.uploadFlow.WasmFile)
	}
	if m.uploadFlow.ConfigFile != "" {
		args = append(args, "--config", m.uploadFlow.ConfigFile)
	}
	if m.uploadFlow.WalletFile != "" {
		args = append(args, "--wallet-file", m.uploadFlow.WalletFile)
	}
	if m.uploadFlow.Version != "" {
		args = append(args, "--version", m.uploadFlow.Version)
	}
	if m.uploadFlow.GitHash != "" {
		args = append(args, "--git-hash", m.uploadFlow.GitHash)
	}
	if m.uploadFlow.DryRun {
		args = append(args, "--dry-run")
	}

	debug.Printf("Calling harlequin binary with args: %v", args)

	// Find the harlequin binary path
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Call the harlequin binary with upload-module command
	cmd := exec.CommandContext(m.ctx, execPath, args...)
	output, err := cmd.CombinedOutput()

	debug.Printf("Upload command output: %s", string(output))
	if err != nil {
		debug.Printf("Upload command failed: %v", err)
		return string(output), fmt.Errorf("upload failed: %w", err)
	}

	return string(output), nil
}

// extractDataItemID extracts the data item ID from upload output
func (m *Model) extractDataItemID(output string) string {
	// Look for patterns like "Transaction ID: abc123" or "• Transaction ID: abc123"
	// or "🎉 Upload completed! Transaction ID: abc123"

	// Try to find transaction ID patterns
	patterns := []string{
		`Transaction ID: ([a-zA-Z0-9_-]+)`,
		`transaction ID: ([a-zA-Z0-9_-]+)`,
		`ID: ([a-zA-Z0-9_-]+)`,
		`id: ([a-zA-Z0-9_-]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 && len(matches[1]) > 10 { // Transaction IDs are typically longer
			debug.Printf("Extracted data item ID: %s", matches[1])
			return matches[1]
		}
	}

	debug.Printf("Could not extract data item ID from output")
	return ""
}

// checkBalanceAndCost checks wallet balance and estimates upload cost
func (m *Model) checkBalanceAndCost() error {
	debug.Printf("Checking wallet balance and upload cost")

	// We need to load the wallet and WASM file to check balance and cost
	// Read WASM file to get size
	wasmData, err := os.ReadFile(m.uploadFlow.WasmFile)
	if err != nil {
		return fmt.Errorf("failed to read WASM file: %w", err)
	}

	// Load wallet
	var jwk map[string]interface{}
	if os.Getenv("WALLET") != "" {
		err = json.Unmarshal([]byte(os.Getenv("WALLET")), &jwk)
		if err != nil {
			return fmt.Errorf("failed to parse WALLET environment variable: %w", err)
		}
	} else {
		walletContent, err := os.ReadFile(m.uploadFlow.WalletFile)
		if err != nil {
			return fmt.Errorf("failed to read wallet file %s: %w", m.uploadFlow.WalletFile, err)
		}
		err = json.Unmarshal(walletContent, &jwk)
		if err != nil {
			return fmt.Errorf("failed to parse wallet file: %w", err)
		}
	}

	// Create signer
	signer, err := signers.NewArweaveSigner(jwk)
	if err != nil {
		return fmt.Errorf("failed to create Arweave signer: %w", err)
	}

	// Create authenticated client and check balance
	turboClient := turbo.Authenticated(nil, signer)
	balance, err := turboClient.GetBalanceForSigner(m.ctx)
	if err != nil {
		// Check if it's a 404 User Not Found error - treat as 0 balance
		if strings.Contains(err.Error(), "HTTP 404") || strings.Contains(err.Error(), "User Not Found") {
			debug.Printf("User not found (404) - treating as 0 balance")
			balance = &types.Balance{
				WinC:     "0",
				Credits:  "0",
				Currency: "winston",
			}
		} else {
			return fmt.Errorf("failed to check wallet balance: %w", err)
		}
	}

	// Estimate upload cost
	unauthenticatedClient := turbo.Unauthenticated(nil)
	fileSize := int64(len(wasmData))
	debug.Printf("Requesting upload costs for file size: %d bytes", fileSize)

	uploadCosts, err := unauthenticatedClient.GetUploadCosts(m.ctx, []int64{fileSize})
	if err != nil {
		debug.Printf("GetUploadCosts API error: %v", err)

		// Check if it's a JSON parsing error - this is a known issue with the API
		if strings.Contains(err.Error(), "json: cannot unmarshal object into Go value of type []types.UploadCost") {
			debug.Printf("Known issue: API returned object but expected array - continuing without cost estimate")

			// Set a default estimated cost warning and continue
			m.uploadFlow.Balance = balance.WinC
			m.uploadFlow.EstimatedCost = "unknown"
			m.uploadFlow.BalanceCheckError = "Unable to estimate upload cost due to API format issue. Upload may still proceed."

			debug.Printf("Balance: %s, Estimated cost: unknown (API issue)", balance.WinC)
			return nil
		}

		return fmt.Errorf("failed to estimate upload cost: %w", err)
	}

	if len(uploadCosts) == 0 {
		return fmt.Errorf("no upload cost estimate received")
	}

	// Store balance and cost in flow
	m.uploadFlow.Balance = balance.WinC
	m.uploadFlow.EstimatedCost = uploadCosts[0].Winc

	debug.Printf("Balance: %s, Estimated cost: %s", balance.WinC, uploadCosts[0].Winc)

	return nil
}

// checkInsufficientBalance checks if wallet has sufficient balance for upload
func (m *Model) checkInsufficientBalance() error {
	if m.uploadFlow.Balance == "" || m.uploadFlow.EstimatedCost == "" {
		return nil // No balance info available
	}

	// If cost estimate is unknown, we can't determine sufficiency - assume sufficient
	if m.uploadFlow.EstimatedCost == "unknown" {
		return nil
	}

	// Parse balance and cost as integers for comparison
	balanceInt, err := strconv.ParseInt(m.uploadFlow.Balance, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse balance: %w", err)
	}

	costInt, err := strconv.ParseInt(m.uploadFlow.EstimatedCost, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse cost estimate: %w", err)
	}

	// Check if balance is sufficient
	if balanceInt < costInt {
		return fmt.Errorf("insufficient wallet balance: need %s, have %s Winston Credits",
			m.uploadFlow.EstimatedCost, m.uploadFlow.Balance)
	}

	return nil
}

// winstonToCredits converts Winston Credits to Credits (AR denomination)
func winstonToCredits(winston string) string {
	if winston == "" || winston == "0" {
		return "0"
	}

	// Handle special case for unknown/unavailable estimates
	if winston == "unknown" {
		return "unknown"
	}

	// Convert string to big.Int for precision
	winstonBig := new(big.Int)
	winstonBig, ok := winstonBig.SetString(winston, 10)
	if !ok {
		debug.Printf("Error parsing winston string: %s", winston)
		return winston + " Winston" // Fallback to raw winston
	}

	// Convert Winston to AR
	ar := utils.WinstonToAR(winstonBig)

	// Format with reasonable precision (6 decimal places)
	return ar.Text('f', 6)
}

// formatCreditsDisplay formats credits with appropriate units
func formatCreditsDisplay(winston string) string {
	credits := winstonToCredits(winston)
	if credits == "0" {
		return "0 Credits"
	}
	if credits == "unknown" {
		return "Unable to estimate"
	}
	return credits + " Credits"
}

// Upload Module view functions

// viewUploadWasmSelection renders the WASM file selection view
func (m *Model) viewUploadWasmSelection() string {
	// Initialize the appropriate selector on first view
	if m.uploadWasmSelector == nil && !m.useUploadWasmFilePicker {
		// Try automatic discovery first
		cwd, _ := os.Getwd()
		actualPanelWidth := m.getPanelWidth() - 2
		if selector, err := components.CreateWasmSelectorWithDiscovery(cwd, actualPanelWidth, m.getPanelHeight()); err == nil {
			m.uploadWasmSelector = selector
		} else {
			// Fall back to file picker if discovery fails
			m.useUploadWasmFilePicker = true
		}
	}

	if m.useUploadWasmFilePicker {
		// Use manual file picker
		if m.uploadWasmFilePicker == nil {
			cwd, _ := os.Getwd()
			actualPanelWidth := m.getPanelWidth() - 2
			m.uploadWasmFilePicker = components.NewFilePicker(actualPanelWidth, m.getPanelHeight())
			m.uploadWasmFilePicker.SetCurrentDirectory(cwd)
			m.uploadWasmFilePicker.SetAllowedTypes([]string{".wasm"})
		}

		leftPanel := m.uploadWasmFilePicker.View()

		rightPanel := components.CreateDescriptionPanel(
			"Manual File Selection",
			fmt.Sprintf("Current directory: %s\n\nNavigate with ↑/↓\nEnter directories with →\nSelect .wasm files with Enter\n\nPress 'l' to switch to automatic discovery",
				m.uploadWasmFilePicker.GetCurrentDirectory()),
			m.getPanelWidth()-2,
			0,
		)

		return m.createTwoPanelLayout(leftPanel, rightPanel)
	} else {
		// Use automatic discovery list
		leftPanel := m.uploadWasmSelector.View()

		// Right panel with discovery info
		selectedFile := ""
		if selected := m.uploadWasmSelector.GetSelected(); selected != nil {
			selectedFile = selected.Value()
		}

		description := "Select the WASM file to upload.\n\nFiles are found recursively (excluding build directories)\n\nPress 'f' to switch to manual file picker"
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

// viewUploadConfigSelection renders the config file selection view
func (m *Model) viewUploadConfigSelection() string {
	// Initialize the appropriate selector on first view
	if m.uploadConfigSelector == nil && !m.useUploadConfigFilePicker {
		// Try automatic discovery first
		cwd, _ := os.Getwd()
		actualPanelWidth := m.getPanelWidth() - 2
		if selector, err := components.CreateConfigSelectorWithDiscovery(cwd, actualPanelWidth, m.getPanelHeight()); err == nil {
			m.uploadConfigSelector = selector
		} else {
			// Fall back to file picker if discovery fails
			m.useUploadConfigFilePicker = true
		}
	}

	if m.useUploadConfigFilePicker {
		// Use manual file picker
		if m.uploadConfigFilePicker == nil {
			cwd, _ := os.Getwd()
			actualPanelWidth := m.getPanelWidth() - 2
			m.uploadConfigFilePicker = components.NewFilePicker(actualPanelWidth, m.getPanelHeight())
			m.uploadConfigFilePicker.SetCurrentDirectory(cwd)
			m.uploadConfigFilePicker.SetAllowedTypes([]string{".yml", ".yaml"})
		}

		leftPanel := m.uploadConfigFilePicker.View()

		rightPanel := components.CreateDescriptionPanel(
			"Manual File Selection",
			fmt.Sprintf("Current directory: %s\n\nNavigate with ↑/↓\nEnter directories with →\nSelect .yml/.yaml files with Enter\n\nPress 'l' to switch to automatic discovery",
				m.uploadConfigFilePicker.GetCurrentDirectory()),
			m.getPanelWidth()-2,
			0,
		)

		return m.createTwoPanelLayout(leftPanel, rightPanel)
	} else {
		// Use automatic discovery list
		leftPanel := m.uploadConfigSelector.View()

		// Right panel with discovery info
		selectedFile := ""
		if selected := m.uploadConfigSelector.GetSelected(); selected != nil {
			selectedFile = selected.Value()
		}

		description := "Select the build configuration file.\n\nLooking for ao-build-config.yml and .harlequin.yaml files\n\nPress 'f' to switch to manual file picker"
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

// viewUploadWalletSelection renders the wallet file selection view
func (m *Model) viewUploadWalletSelection() string {
	// Initialize the appropriate selector on first view
	if m.uploadWalletSelector == nil && !m.useUploadWalletFilePicker {
		// Try automatic discovery first
		cwd, _ := os.Getwd()
		actualPanelWidth := m.getPanelWidth() - 2
		if selector, err := components.CreateWalletSelectorWithDiscovery(cwd, actualPanelWidth, m.getPanelHeight()); err == nil {
			m.uploadWalletSelector = selector
		} else {
			// Fall back to file picker if discovery fails
			m.useUploadWalletFilePicker = true
		}
	}

	if m.useUploadWalletFilePicker {
		// Use manual file picker
		if m.uploadWalletFilePicker == nil {
			cwd, _ := os.Getwd()
			actualPanelWidth := m.getPanelWidth() - 2
			m.uploadWalletFilePicker = components.NewFilePicker(actualPanelWidth, m.getPanelHeight())
			m.uploadWalletFilePicker.SetCurrentDirectory(cwd)
			m.uploadWalletFilePicker.SetAllowedTypes([]string{".json"})
		}

		leftPanel := m.uploadWalletFilePicker.View()

		rightPanel := components.CreateDescriptionPanel(
			"Manual File Selection",
			fmt.Sprintf("Current directory: %s\n\nNavigate with ↑/↓\nEnter directories with →\nSelect .json wallet files with Enter\n\nPress 'l' to switch to automatic discovery",
				m.uploadWalletFilePicker.GetCurrentDirectory()),
			m.getPanelWidth()-2,
			0,
		)

		return m.createTwoPanelLayout(leftPanel, rightPanel)
	} else {
		// Use automatic discovery list
		leftPanel := m.uploadWalletSelector.View()

		// Right panel with discovery info
		selectedFile := ""
		if selected := m.uploadWalletSelector.GetSelected(); selected != nil {
			selectedFile = selected.Value()
		}

		description := "Select the Arweave wallet file.\n\nLooking for key.json, wallet.json, and similar files\n\nPress 'f' to switch to manual file picker"
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

// viewUploadVersion renders the version input view
func (m *Model) viewUploadVersion() string {
	// Create version input if not exists
	if m.uploadVersionInput == nil {
		actualPanelWidth := m.getPanelWidth() - 2
		m.uploadVersionInput = components.CreateVersionInput(actualPanelWidth, m.getPanelHeight())
	}

	leftPanel := m.uploadVersionInput.View()

	rightPanel := components.CreateDescriptionPanel(
		"Module Version",
		"Enter the version for your module.\n\nThis will be included in the upload tags for tracking.\n\nExamples:\n• v1.0.0\n• v2.1.3\n• v0.1.0-beta\n• dev\n\nSemantic versioning is recommended.",
		m.getPanelWidth()-2,
		0,
	)

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewUploadGitHash renders the git hash input view
func (m *Model) viewUploadGitHash() string {
	// Create git hash input if not exists
	if m.uploadGitHashInput == nil {
		actualPanelWidth := m.getPanelWidth() - 2
		m.uploadGitHashInput = components.CreateOutputDirInput(actualPanelWidth, m.getPanelHeight())
		// Try to set git hash from environment
		if gitHash := os.Getenv("GITHUB_SHA"); gitHash != "" {
			m.uploadGitHashInput.SetValue(gitHash)
		}
	}

	leftPanel := m.uploadGitHashInput.View()

	rightPanel := components.CreateDescriptionPanel(
		"Git Hash (Optional)",
		"Enter the git commit hash for this build.\n\nThis helps track which version of your code was used.\n\nLeave empty for auto-detection or if not using git.\n\nThe hash will be included in upload tags.",
		m.getPanelWidth()-2,
		0,
	)

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewUploadDryRun renders the dry run selection view
func (m *Model) viewUploadDryRun() string {
	// Create dry run selector if not exists
	if m.uploadDryRunSelector == nil {
		actualPanelWidth := m.getPanelWidth() - 2
		m.uploadDryRunSelector = components.CreateUploadDryRunSelector(actualPanelWidth, m.getPanelHeight())
	}

	leftPanel := m.uploadDryRunSelector.View()

	// Right panel with description
	selected := m.uploadDryRunSelector.GetSelected()
	description := "Choose whether to perform a dry run or actual upload."
	if selected != nil {
		switch selected.Value() {
		case "true":
			description = "Dry Run Mode\n\nThis will:\n• Parse and validate the WASM file\n• Read configuration files\n• Generate all upload tags\n• Show exactly what would be uploaded\n• NOT actually upload anything\n\nGreat for testing and validation."
		case "false":
			description = "Actual Upload Mode\n\nThis will:\n• Perform all validation steps\n• Load your Arweave wallet\n• Create and sign the data item\n• Upload to Arweave via Turbo\n• Provide transaction ID and URL\n\nRequires wallet with sufficient funds."
		default:
			description = selected.Description()
		}
	}

	rightPanel := components.CreateDescriptionPanel(
		"Upload Mode",
		description,
		m.getPanelWidth()-2,
		0,
	)

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewUploadConfirmation renders the upload confirmation view
func (m *Model) viewUploadConfirmation() string {
	// Create confirmation selector if not exists or if balance info changed
	if m.uploadConfirmSelector == nil {
		actualPanelWidth := m.getPanelWidth() - 2

				// Check if we have balance information and if it's sufficient for actual uploads
		if !m.uploadFlow.DryRun {
						if m.uploadFlow.BalanceCheckError != "" {
				// Balance check failed - show error state
				// Use regular confirmation selector but with error context
				m.uploadConfirmSelector = components.CreateUploadConfirmationSelector(actualPanelWidth, m.getPanelHeight())
			} else if m.uploadFlow.Balance != "" && m.uploadFlow.EstimatedCost != "" {
				// Parse balance and cost to check sufficiency
				hasSufficientBalance := true
				if err := m.checkInsufficientBalance(); err != nil {
					hasSufficientBalance = false
				}

								m.uploadConfirmSelector = components.CreateUploadConfirmationSelectorWithBalance(
					actualPanelWidth,
					m.getPanelHeight(),
					hasSufficientBalance,
					formatCreditsDisplay(m.uploadFlow.Balance),
					formatCreditsDisplay(m.uploadFlow.EstimatedCost),
				)
			} else {
				// No balance info available
				m.uploadConfirmSelector = components.CreateUploadConfirmationSelector(actualPanelWidth, m.getPanelHeight())
			}
		} else {
			// Use regular confirmation selector for dry runs
			m.uploadConfirmSelector = components.CreateUploadConfirmationSelector(actualPanelWidth, m.getPanelHeight())
		}
	}

	leftPanel := m.formatUploadPreview()
	rightPanel := m.uploadConfirmSelector.View()

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewUploadRunning renders the upload progress view
func (m *Model) viewUploadRunning() string {
	leftPanel := ""
	if m.uploadProgress != nil {
		leftPanel = m.uploadProgress.ViewContent()
	}

	rightPanel := components.CreateDescriptionPanel(
		"Upload Progress",
		"Uploading your module...\n\nThis process:\n• Analyzes WASM metadata\n• Creates upload tags\n• Signs the data item\n• Uploads to Arweave\n\nPlease wait...",
		m.getPanelWidth()-2,
		0,
	)

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// viewUploadResult renders the upload result view
func (m *Model) viewUploadResult() string {
	if m.result == nil {
		return "No result available"
	}

	leftPanel := m.result.ViewPanelContent()
	rightPanel := m.result.ViewDetailsContent()

	return m.createTwoPanelLayout(leftPanel, rightPanel)
}

// formatUploadPreview formats the current upload config for preview
func (m *Model) formatUploadPreview() string {
	mode := "Actual Upload"
	if m.uploadFlow.DryRun {
		mode = "Dry Run"
	}

	preview := fmt.Sprintf(`Upload Configuration

WASM File: %s
Config File: %s
Wallet File: %s
Version: %s
Git Hash: %s
Mode: %s`,
		m.uploadFlow.WasmFile,
		m.uploadFlow.ConfigFile,
		m.uploadFlow.WalletFile,
		m.uploadFlow.Version,
		m.uploadFlow.GitHash,
		mode,
	)

		// Add balance information for actual uploads
	if !m.uploadFlow.DryRun {
		if m.uploadFlow.BalanceCheckError != "" {
			// Show balance check error
			preview += fmt.Sprintf(`

Balance Check:
Status: ❌ Error checking balance
Error: %s`, m.uploadFlow.BalanceCheckError)
		} else if m.uploadFlow.Balance != "" && m.uploadFlow.EstimatedCost != "" {
		// Check if balance is sufficient
		isBalanceSufficient := true
		if err := m.checkInsufficientBalance(); err != nil {
			isBalanceSufficient = false
		}

		balanceStatus := "✅ Sufficient"
		if !isBalanceSufficient {
			balanceStatus = "⚠️  Insufficient"
		}

				preview += fmt.Sprintf(`

Balance Check:
Current Balance: %s
Estimated Cost: %s
Status: %s`,
			formatCreditsDisplay(m.uploadFlow.Balance),
			formatCreditsDisplay(m.uploadFlow.EstimatedCost),
			balanceStatus)

		if !isBalanceSufficient && m.uploadFlow.EstimatedCost != "unknown" {
			// Parse balance and cost to show shortfall (only if cost is not unknown)
			if balanceInt, err1 := strconv.ParseInt(m.uploadFlow.Balance, 10, 64); err1 == nil {
				if costInt, err2 := strconv.ParseInt(m.uploadFlow.EstimatedCost, 10, 64); err2 == nil {
					shortfall := costInt - balanceInt
					shortfallStr := strconv.FormatInt(shortfall, 10)
					preview += fmt.Sprintf(`
Shortfall: %s`, formatCreditsDisplay(shortfallStr))
				}
			}
		}
		}
	}

	preview += fmt.Sprintf(`

Ready to proceed with %s.`, mode)

	return preview
}

// updateInitWizard handles init wizard updates
func (m *Model) updateInitWizard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Initialize init wizard if needed
	if m.initWizard == nil {
		m.initWizard = components.NewInitWizardComponent()

		// Set up completion callback
		m.initWizard.OnComplete = func(projectName, templateLang, authorName, githubUser, targetDir string) {
			// Create the project using the CLI command to avoid import cycle
			args := []string{"init", templateLang, "--name", projectName}
			if authorName != "" {
				args = append(args, "--author", authorName)
			}
			if githubUser != "" {
				args = append(args, "--github", githubUser)
			}
			if targetDir != "" && targetDir != projectName {
				args = append(args, "--dir", targetDir)
			}

			cmd := exec.Command(os.Args[0], args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				fmt.Printf("Error creating project: %v\n", err)
			}

			// Go back to command selection after completion
			m.state = ViewCommandSelection
		}
	}

	// Update the init wizard
	model, cmd := m.initWizard.Update(tea.Msg(msg))
	if newWizard, ok := model.(*components.InitWizardComponent); ok {
		m.initWizard = newWizard
	}

	return m, cmd
}

// viewInitWizard renders the init wizard view
func (m *Model) viewInitWizard() string {
	// Initialize init wizard if needed
	if m.initWizard == nil {
		m.initWizard = components.NewInitWizardComponent()

		// Set up completion callback
		m.initWizard.OnComplete = func(projectName, templateLang, authorName, githubUser, targetDir string) {
			// Create the project using the CLI command to avoid import cycle
			args := []string{"init", templateLang, "--name", projectName}
			if authorName != "" {
				args = append(args, "--author", authorName)
			}
			if githubUser != "" {
				args = append(args, "--github", githubUser)
			}
			if targetDir != "" && targetDir != projectName {
				args = append(args, "--dir", targetDir)
			}

			cmd := exec.Command(os.Args[0], args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				fmt.Printf("Error creating project: %v\n", err)
			}

			// Go back to command selection after completion
			m.state = ViewCommandSelection
		}
	}

	return m.initWizard.View()
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
