package components

import (
	"github.com/charmbracelet/lipgloss"
)

// LayoutUtils provides reusable layout utilities and styles
type LayoutUtils struct{}

// NewLayoutUtils creates a new layout utilities instance
func NewLayoutUtils() *LayoutUtils {
	return &LayoutUtils{}
}

// Common styles
var (
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#874BFD")).
		Align(lipgloss.Center).
		Height(2)

	PanelStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(0, 1)

	LeftPanelStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(0, 1)

	RightPanelStyle = lipgloss.NewStyle().
		Padding(0, 1)

	ControlsStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		Align(lipgloss.Center).
		Height(2)

	ContainerStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(0, 1).
		Width(90)
)

// CreateHeader creates a centered header with the given title and available width
func (l *LayoutUtils) CreateHeader(title string, availableWidth int) string {
	return HeaderStyle.
		Width(availableWidth).
		Render(title)
}

// CreateTwoColumnLayout creates a side-by-side layout with left and right panels
func (l *LayoutUtils) CreateTwoColumnLayout(leftPanel, rightPanel string, totalWidth int) string {
	// Calculate panel widths (50% each, accounting for gap)
	availableContentWidth := totalWidth - 5 // Container borders and gap
	panelWidth := availableContentWidth / 2
	
	// Apply consistent width to both panels
	leftStyled := PanelStyle.Width(panelWidth).Height(12).Render(leftPanel)
	rightStyled := PanelStyle.Width(panelWidth).Height(12).Render(rightPanel)
	
	return lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, " ", rightStyled)
}

// CreateControls creates a controls panel with the given text and available width
func (l *LayoutUtils) CreateControls(controlsText string, availableWidth int) string {
	if controlsText == "" {
		return ""
	}
	
	return ControlsStyle.
		Width(availableWidth).
		Render(controlsText)
}

// CreateMainLayout creates the complete application layout with header, content, and controls
func (l *LayoutUtils) CreateMainLayout(header, content, controls string) string {
	sections := []string{header, content}
	
	if controls != "" {
		sections = append(sections, controls)
	}
	
	fullLayout := lipgloss.JoinVertical(lipgloss.Left, sections...)
	
	return ContainerStyle.Render(fullLayout)
}

// CenterContent centers content within the given terminal dimensions
func (l *LayoutUtils) CenterContent(content string, terminalWidth, terminalHeight int) string {
	containerWidth := 90
	leftPadding := (terminalWidth - containerWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}
	
	// Center horizontally
	centeredContent := lipgloss.NewStyle().
		MarginLeft(leftPadding).
		Render(content)
	
	// Center vertically if terminal is tall enough
	if terminalHeight > 20 {
		topPadding := (terminalHeight - 20) / 2
		if topPadding > 0 {
			centeredContent = lipgloss.NewStyle().
				MarginTop(topPadding).
				Render(centeredContent)
		}
	}
	
	return centeredContent
}

// GetPanelWidth calculates the standard panel width for two-column layouts
func (l *LayoutUtils) GetPanelWidth() int {
	availableContentWidth := 88 - 5 // Container width minus borders and gap
	return availableContentWidth / 2 // About 41 chars each
}

// Description styles
var (
	DescriptionHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#874BFD"))

	DescriptionBodyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA"))
)

// CreateDescription creates a formatted description panel with header and body
func (l *LayoutUtils) CreateDescription(header, body string, width, height int) string {
	styledHeader := DescriptionHeaderStyle.Render(header)
	content := lipgloss.JoinVertical(lipgloss.Left, styledHeader, "", body)
	
	return PanelStyle.
		Width(width).
		Height(height).
		Render(content)
}

// Standalone functions for the modernized TUI
// These match the API used in the modernized main.go

// CreateHeader creates a centered header (standalone function)
func CreateHeader(title string, width int) string {
	return HeaderStyle.
		Width(width).
		Render(title)
}

// CreateControls creates a controls panel (standalone function)
func CreateControls(controls []string, width int) string {
	controlsText := lipgloss.JoinHorizontal(lipgloss.Left, 
		lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render(" • "),
	)
	
	for i, control := range controls {
		if i > 0 {
			controlsText += lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render(" • ")
		}
		controlsText += control
	}
	
	return ControlsStyle.
		Width(width).
		Render(controlsText)
}

// CreateDescriptionPanel creates a bordered panel with title and description (standalone function)
func CreateDescriptionPanel(title, description string, width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#874BFD")).
		Padding(0, 0, 1, 0)
	
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA"))
	
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		descStyle.Render(description),
	)
	
	// Account for borders (2 chars width) and padding (2 chars width)
	contentWidth := width - 4  // 2 for borders + 2 for padding
	
	// Ensure minimum width
	if contentWidth < 1 {
		contentWidth = 1
	}
	
	// Wrap content to fit within the available width (let height be natural)
	wrappedContent := lipgloss.NewStyle().
		Width(contentWidth).
		Render(content)
	
	return PanelStyle.
		Width(width - 7).
		Render(wrappedContent)
}
