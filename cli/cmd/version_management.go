package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/debug"
)

// Release represents a release from the API
type Release struct {
	TagName   string    `json:"tag_name"`
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
}

// ReleasesResponse represents the API response structure
type ReleasesResponse struct {
	Releases []Release `json:"releases"`
}

// VersionManagementConfig holds configuration for version management
type VersionManagementConfig struct {
	BaseURL     string
	InstallDir  string
	BinaryName  string
	Timeout     time.Duration
}

// DefaultVersionManagementConfig returns the default configuration
func DefaultVersionManagementConfig() *VersionManagementConfig {
	installDir := "/usr/local/bin"
	if runtime.GOOS == "windows" {
		// For Windows, use a more appropriate default
		installDir = filepath.Join(os.Getenv("PROGRAMFILES"), "harlequin")
	}

	return &VersionManagementConfig{
		BaseURL:    "https://install_cli_harlequin.daemongate.io",
		InstallDir: installDir,
		BinaryName: "harlequin",
		Timeout:    30 * time.Second,
	}
}

// HandleInstallCommand handles the install command
func HandleInstallCommand(ctx context.Context, args []string) {
	config := DefaultVersionManagementConfig()
	var targetVersion string
	var showHelp bool

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--version", "-v":
			if i+1 >= len(args) {
				fmt.Printf("Error: --version requires a value\n\n")
				printInstallUsage()
				os.Exit(1)
			}
			targetVersion = args[i+1]
			i++ // Skip the next argument as it's the value
		case "--help", "-h":
			showHelp = true
		default:
			fmt.Printf("Unknown argument: %s\n\n", args[i])
			printInstallUsage()
			os.Exit(1)
		}
	}

	if showHelp {
		printInstallUsage()
		return
	}

	// Enable debug mode if available
	debug.SetEnabled(true)

	if targetVersion != "" {
		// Install specific version
		if err := installSpecificVersion(ctx, config, targetVersion); err != nil {
			fmt.Printf("Error installing version %s: %v\n", targetVersion, err)
			os.Exit(1)
		}
	} else {
		// Interactive version selection
		if err := interactiveVersionSelection(ctx, config); err != nil {
			fmt.Printf("Error during interactive installation: %v\n", err)
			os.Exit(1)
		}
	}
}

// HandleVersionsCommand handles the versions command
func HandleVersionsCommand(ctx context.Context, args []string) {
	config := DefaultVersionManagementConfig()
	var showHelp bool
	var format string = "table" // default format

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format", "-f":
			if i+1 >= len(args) {
				fmt.Printf("Error: --format requires a value (table, json, list)\n\n")
				printVersionsUsage()
				os.Exit(1)
			}
			format = args[i+1]
			if format != "table" && format != "json" && format != "list" {
				fmt.Printf("Error: invalid format '%s'. Valid formats: table, json, list\n\n", format)
				printVersionsUsage()
				os.Exit(1)
			}
			i++ // Skip the next argument as it's the value
		case "--help", "-h":
			showHelp = true
		default:
			fmt.Printf("Unknown argument: %s\n\n", args[i])
			printVersionsUsage()
			os.Exit(1)
		}
	}

	if showHelp {
		printVersionsUsage()
		return
	}

	if err := listAvailableVersions(ctx, config, format); err != nil {
		fmt.Printf("Error fetching versions: %v\n", err)
		os.Exit(1)
	}
}

// HandleUninstallCommand handles the uninstall command
func HandleUninstallCommand(ctx context.Context, args []string) {
	config := DefaultVersionManagementConfig()
	var showHelp bool

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--help", "-h":
			showHelp = true
		default:
			fmt.Printf("Unknown argument: %s\n\n", args[i])
			printUninstallUsage()
			os.Exit(1)
		}
	}

	if showHelp {
		printUninstallUsage()
		return
	}

	if err := uninstallHarlequin(config); err != nil {
		fmt.Printf("Error during uninstall: %v\n", err)
		os.Exit(1)
	}
}

// fetchAvailableVersions fetches available versions from the releases API
// TODO: Consider adding GitHub releases API integration for changelog information
// when GitHub releases are available at: https://api.github.com/repos/the-permaweb-harlequin/harlequin-toolkit/releases
func fetchAvailableVersions(ctx context.Context, config *VersionManagementConfig) ([]Release, error) {
	client := &http.Client{
		Timeout: config.Timeout,
	}

	url := fmt.Sprintf("%s/releases", config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response ReleasesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Sort by created date (newest first)
	sort.Slice(response.Releases, func(i, j int) bool {
		return response.Releases[i].CreatedAt.After(response.Releases[j].CreatedAt)
	})

	return response.Releases, nil
}

// VersionItem represents a version in the list
type VersionItem struct {
	TagName string
	Version string
}

func (v VersionItem) FilterValue() string { return v.TagName }
func (v VersionItem) Title() string {
	return v.TagName
}
func (v VersionItem) Description() string {
	if v.Version != "" && v.Version != v.TagName {
		return fmt.Sprintf("Version %s • No changelog available", v.Version)
	}
	return "Stable release • No changelog available"
}

// VersionSelectorModel represents the version selection TUI model
type VersionSelectorModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m VersionSelectorModel) Init() tea.Cmd {
	return nil
}

func (m VersionSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(VersionItem)
			if ok {
				m.choice = i.TagName
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m VersionSelectorModel) View() string {
	if m.choice != "" {
		return ""
	}
	if m.quitting {
		return "Cancelled.\n"
	}
	return "\n" + m.list.View()
}

// interactiveVersionSelection provides a TUI for version selection
func interactiveVersionSelection(ctx context.Context, config *VersionManagementConfig) error {
	fmt.Println("🎭 Harlequin Version Manager")
	fmt.Println("Fetching available versions...")

	releases, err := fetchAvailableVersions(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to fetch versions: %w", err)
	}

	if len(releases) == 0 {
		return fmt.Errorf("no releases found")
	}

	// Convert releases to list items
	items := make([]list.Item, len(releases))
	for i, release := range releases {
		items[i] = VersionItem{
			TagName: release.TagName,
			Version: release.Version,
		}
	}

	// Create list model
	const defaultWidth = 80
	const listHeight = 14

	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, listHeight)
	l.Title = "Select a version to install"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().MarginLeft(2)

	m := VersionSelectorModel{list: l}

	// Run the TUI
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Get the final model and check the choice
	if finalModel, ok := finalModel.(VersionSelectorModel); ok {
		if finalModel.quitting && finalModel.choice == "" {
			fmt.Println("No version selected. Exiting.")
			return nil
		}
		if finalModel.choice != "" {
			return installSpecificVersion(ctx, config, finalModel.choice)
		}
	}

	return fmt.Errorf("no version selected")
}

// listAvailableVersions fetches and displays available versions in the specified format
func listAvailableVersions(ctx context.Context, config *VersionManagementConfig, format string) error {
	fmt.Println("🎭 Fetching available Harlequin versions...")

	releases, err := fetchAvailableVersions(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to fetch versions: %w", err)
	}

	if len(releases) == 0 {
		fmt.Println("No versions found.")
		return nil
	}

	switch format {
	case "json":
		return printVersionsJSON(releases)
	case "list":
		return printVersionsList(releases)
	case "table":
		return printVersionsTable(releases)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// printVersionsJSON outputs versions in JSON format
func printVersionsJSON(releases []Release) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(releases)
}

// printVersionsList outputs versions in simple list format
func printVersionsList(releases []Release) error {
	for _, release := range releases {
		fmt.Println(release.TagName)
	}
	return nil
}

// printVersionsTable outputs versions in a formatted table
func printVersionsTable(releases []Release) error {
	fmt.Printf("\n%-15s %-15s %-25s\n", "TAG", "VERSION", "CREATED")
	fmt.Printf("%-15s %-15s %-25s\n", "───", "───────", "───────")

	for _, release := range releases {
		version := release.Version
		if version == "" {
			version = "-"
		}

		// Format the date nicely
		createdAt := release.CreatedAt.Format("2006-01-02 15:04:05")

		fmt.Printf("%-15s %-15s %-25s\n", release.TagName, version, createdAt)
	}

	fmt.Printf("\nTotal: %d versions available\n", len(releases))
	fmt.Println("\nTo install a specific version:")
	fmt.Println("  harlequin install --version <tag>")
	fmt.Println("\nFor interactive selection:")
	fmt.Println("  harlequin install")

	return nil
}

// installSpecificVersion installs a specific version
func installSpecificVersion(ctx context.Context, config *VersionManagementConfig, version string) error {
	fmt.Printf("🎭 Installing Harlequin version %s...\n", version)

	// Clean version string (remove 'v' prefix if present)
	cleanVersion := strings.TrimPrefix(version, "v")

	// Detect platform and architecture
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go arch to expected format
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	case "386":
		arch = "386"
	default:
		return fmt.Errorf("unsupported architecture: %s", arch)
	}

	// Build download URL
	downloadURL := fmt.Sprintf("%s/releases/%s/%s/%s", config.BaseURL, cleanVersion, platform, arch)

	fmt.Printf("📥 Downloading from: %s\n", downloadURL)

	// Use the installation script approach
	installScript := fmt.Sprintf(`curl -sSL '%s' | VERSION='%s' sh`, config.BaseURL, cleanVersion)

	cmd := exec.CommandContext(ctx, "sh", "-c", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Printf("✅ Successfully installed Harlequin version %s\n", version)
	return nil
}

// uninstallHarlequin removes the harlequin binary
func uninstallHarlequin(config *VersionManagementConfig) error {
	fmt.Println("🎭 Uninstalling Harlequin...")

	// Find the current installation
	binaryPath, err := exec.LookPath(config.BinaryName)
	if err != nil {
		// Try common installation paths
		commonPaths := []string{
			filepath.Join(config.InstallDir, config.BinaryName),
			filepath.Join("/usr/local/bin", config.BinaryName),
			filepath.Join("/usr/bin", config.BinaryName),
		}

		if runtime.GOOS == "windows" {
			commonPaths = append(commonPaths,
				filepath.Join(config.InstallDir, config.BinaryName+".exe"),
				filepath.Join("C:\\Program Files\\harlequin", config.BinaryName+".exe"),
			)
		}

		found := false
		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				binaryPath = path
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("harlequin binary not found. It may not be installed or not in PATH")
		}
	}

	fmt.Printf("📍 Found harlequin at: %s\n", binaryPath)

	// Confirm uninstall
	fmt.Printf("Are you sure you want to uninstall harlequin from %s? [y/N]: ", binaryPath)
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Uninstall cancelled.")
		return nil
	}

	// Remove the binary
	if err := os.Remove(binaryPath); err != nil {
		return fmt.Errorf("failed to remove binary: %w", err)
	}

	fmt.Printf("✅ Successfully uninstalled harlequin from %s\n", binaryPath)
	fmt.Println("💡 To reinstall, run: curl -sSL https://install_cli_harlequin.daemongate.io | sh")

	return nil
}

// printInstallUsage prints usage information for the install command
func printInstallUsage() {
	fmt.Println("🎭 Harlequin Install Command")
	fmt.Println()
	fmt.Println("Install or upgrade harlequin to a specific version")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  harlequin install [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -v, --version <version>  Install specific version (e.g., 1.2.3 or v1.2.3)")
	fmt.Println("  -h, --help              Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin install                    # Interactive version selection")
	fmt.Println("  harlequin install --version 1.2.3   # Install version 1.2.3")
	fmt.Println("  harlequin install --version v1.2.3  # Install version 1.2.3 (with v prefix)")
	fmt.Println()
	fmt.Println("Interactive Mode:")
	fmt.Println("  When run without --version, this command will:")
	fmt.Println("  • Fetch available versions from the release API")
	fmt.Println("  • Present a TUI for version selection")
	fmt.Println("  • Install the selected version")
	fmt.Println()
}

// printVersionsUsage prints usage information for the versions command
func printVersionsUsage() {
	fmt.Println("🎭 Harlequin Versions Command")
	fmt.Println()
	fmt.Println("List all available harlequin versions")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  harlequin versions [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -f, --format <format>  Output format: table, json, list (default: table)")
	fmt.Println("  -h, --help            Show this help message")
	fmt.Println()
	fmt.Println("Output Formats:")
	fmt.Println("  table  Human-readable table with version details (default)")
	fmt.Println("  json   JSON format for programmatic use")
	fmt.Println("  list   Simple list of version tags")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin versions                    # Show versions in table format")
	fmt.Println("  harlequin versions --format json     # Output as JSON")
	fmt.Println("  harlequin versions --format list     # Simple list of tags")
	fmt.Println()
	fmt.Println("This command will:")
	fmt.Println("  • Fetch available versions from the release API")
	fmt.Println("  • Display version information in the specified format")
	fmt.Println("  • Show installation instructions")
	fmt.Println()
}

// printUninstallUsage prints usage information for the uninstall command
func printUninstallUsage() {
	fmt.Println("🎭 Harlequin Uninstall Command")
	fmt.Println()
	fmt.Println("Remove harlequin from your system")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  harlequin uninstall [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -h, --help  Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin uninstall  # Remove harlequin with confirmation")
	fmt.Println()
	fmt.Println("This command will:")
	fmt.Println("  • Locate the harlequin binary")
	fmt.Println("  • Ask for confirmation")
	fmt.Println("  • Remove the binary from your system")
	fmt.Println()
}
