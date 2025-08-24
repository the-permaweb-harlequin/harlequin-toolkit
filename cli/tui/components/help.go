package components

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// HelpRenderer provides a reusable help documentation viewer using Glamour and Bubbles
type HelpRenderer struct {
	viewport viewport.Model
	ready    bool
	width    int
	height   int
	content  string
	title    string
}

// NewHelpRenderer creates a new help renderer for the given command
func NewHelpRenderer(command string) (*HelpRenderer, error) {
	// Load markdown content
	content, err := loadHelpContent(command)
	if err != nil {
		return nil, fmt.Errorf("failed to load help content for command %s: %w", command, err)
	}

	// Render markdown with Glamour
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create glamour renderer: %w", err)
	}

	renderedContent, err := renderer.Render(content)
	if err != nil {
		return nil, fmt.Errorf("failed to render markdown: %w", err)
	}

	hr := &HelpRenderer{
		content: renderedContent,
		title:   fmt.Sprintf("Help: %s", command),
	}

	return hr, nil
}

// SetSize updates the help renderer dimensions
func (hr *HelpRenderer) SetSize(width, height int) {
	hr.width = width
	hr.height = height

	if !hr.ready {
		// Initialize viewport with content
		hr.viewport = viewport.New(width-4, height-6) // Account for borders and title
		hr.viewport.SetContent(hr.content)
		hr.ready = true
	} else {
		// Update existing viewport
		hr.viewport.Width = width - 4
		hr.viewport.Height = height - 6
	}
}

// Init implements the Bubble Tea model interface
func (hr *HelpRenderer) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages
func (hr *HelpRenderer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Return quit message to parent
			return hr, tea.Quit
		}
	case tea.WindowSizeMsg:
		hr.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	hr.viewport, cmd = hr.viewport.Update(msg)
	return hr, cmd
}

// View renders the help content
func (hr *HelpRenderer) View() string {
	if !hr.ready {
		return "Loading help..."
	}

	// Create title bar
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#874BFD")).
		Padding(0, 1).
		Width(hr.width - 4).
		Align(lipgloss.Center)

	title := titleStyle.Render(hr.title)

	// Create content area with viewport
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(0, 1)

	content := contentStyle.Render(hr.viewport.View())

	// Create footer with navigation help
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		Align(lipgloss.Center).
		Width(hr.width - 4)

	footer := footerStyle.Render("↑/↓ Scroll • q/Esc Exit")

	// Combine all parts
	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		content,
		footer,
	)
}

// loadHelpContent loads markdown content for the specified command
func loadHelpContent(command string) (string, error) {
	// Try multiple paths to find the docs
	possiblePaths := []string{
		// Relative to current working directory
		filepath.Join("docs", "commands", command+".md"),
		// Relative to parent directory (in case we're in cli/)
		filepath.Join("..", "docs", "commands", command+".md"),
		// Absolute path relative to executable
		"",
	}

	// Add executable-relative path
	if execPath, err := os.Executable(); err == nil {
		possiblePaths[2] = filepath.Join(filepath.Dir(execPath), "docs", "commands", command+".md")
	}

	// Try to find the docs by looking for go.mod and building path from there
	if wd, err := os.Getwd(); err == nil {
		for dir := wd; dir != "/" && dir != "."; dir = filepath.Dir(dir) {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				possiblePaths = append(possiblePaths, filepath.Join(dir, "cli", "docs", "commands", command+".md"))
				break
			}
		}
	}

	var lastErr error
	for _, docsPath := range possiblePaths {
		if docsPath == "" {
			continue
		}

		if content, err := os.ReadFile(docsPath); err == nil {
			return string(content), nil
		} else {
			lastErr = err
		}
	}

	return "", fmt.Errorf("failed to find help file for command %s: %w", command, lastErr)
}

// ShowHelp is a convenience function to run a help renderer as a standalone program
func ShowHelp(command string) error {
	renderer, err := NewHelpRenderer(command)
	if err != nil {
		return err
	}

	program := tea.NewProgram(renderer, tea.WithAltScreen())
	_, err = program.Run()
	return err
}
