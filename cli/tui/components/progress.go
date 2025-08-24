package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
)

// StepStatus represents the status of a progress step
type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepSuccess
	StepFailed
)

// ProgressStep represents a single step in a progress sequence
type ProgressStep struct {
	Name      string
	Status    StepStatus
	Spinner   harmonica.Spring
	SpinPhase float64
}

// ProgressComponent provides a reusable progress display with animated steps
type ProgressComponent struct {
	steps  []ProgressStep
	title  string
	width  int
	height int
}

// NewProgress creates a new progress component
func NewProgress(title string, stepNames []string) *ProgressComponent {
	steps := make([]ProgressStep, len(stepNames))
	for i, name := range stepNames {
		steps[i] = ProgressStep{
			Name:    name,
			Status:  StepPending,
			Spinner: harmonica.NewSpring(1.0, 0.8, 0.0),
		}
	}

	return &ProgressComponent{
		title:  title,
		steps:  steps,
		width:  41, // Default width
		height: 12, // Default height
	}
}

// SetSize updates the component dimensions
func (p *ProgressComponent) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetStepStatus updates the status of a specific step
func (p *ProgressComponent) SetStepStatus(stepName string, status StepStatus) {
	for i := range p.steps {
		if p.steps[i].Name == stepName {
			p.steps[i].Status = status
			if status == StepRunning {
				p.steps[i].SpinPhase = 0 // Reset spin phase
			}
			break
		}
	}
}

// UpdateAnimations updates spinner animations for running steps
func (p *ProgressComponent) UpdateAnimations() {
	for i := range p.steps {
		if p.steps[i].Status == StepRunning {
			p.steps[i].SpinPhase += 0.1
			if p.steps[i].SpinPhase > 2*3.14159 {
				p.steps[i].SpinPhase = 0
			}
		}
	}
}

// Update handles Bubble Tea messages
func (p *ProgressComponent) Update(msg tea.Msg) tea.Cmd {
	return nil
}

// View renders the progress panel
func (p *ProgressComponent) View() string {
	content := ""

	// Add title
	if p.title != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#874BFD")).
			Margin(0, 0, 1, 0)
		content += titleStyle.Render(p.title) + "\n"
	}

	// Add steps with status indicators
	for _, step := range p.steps {
		icon := ""
		switch step.Status {
		case StepPending:
			icon = "○" // Circle for pending
		case StepRunning:
			// Animated spinner using rotation
			spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			spinnerIndex := int(step.SpinPhase*float64(len(spinnerChars))/6.28) % len(spinnerChars)
			icon = spinnerChars[spinnerIndex]
		case StepSuccess:
			icon = "✓" // Check for success
		case StepFailed:
			icon = "✗" // X for failed
		}
		content += fmt.Sprintf("%s %s\n", icon, step.Name)
	}

	// Create bordered panel
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(0, 1).
		Width(p.width).
		Height(p.height).
		Render(content)
}

// NewProgressComponent creates a new progress component (alias for compatibility)
func NewProgressComponent(width, height int) *ProgressComponent {
	stepNames := []string{"Copy AOS Files", "Bundle Lua", "Inject Code", "Build WASM", "Copy Outputs", "Cleanup"}
	pc := NewProgress("Building Project", stepNames)
	pc.SetSize(width, height)
	return pc
}

// BuildStep represents a build step for compatibility
type BuildStep struct {
	Name   string
	Status StepStatus
}

// UpdateSteps updates multiple steps at once (for compatibility)
func (p *ProgressComponent) UpdateSteps(steps []BuildStep) {
	for _, step := range steps {
		p.SetStepStatus(step.Name, step.Status)
	}
}

// ViewContent renders the progress content without borders/styling for use in layouts
func (p *ProgressComponent) ViewContent() string {
	content := ""

	// Add title
	if p.title != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#874BFD")).
			Margin(0, 0, 1, 0)
		content += titleStyle.Render(p.title) + "\n"
	}

	// Add steps with status indicators
	for _, step := range p.steps {
		icon := ""
		switch step.Status {
		case StepPending:
			icon = "○" // Circle for pending
		case StepRunning:
			// Animated spinner using rotation
			spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			spinnerIndex := int(step.SpinPhase*float64(len(spinnerChars))/6.28) % len(spinnerChars)
			icon = spinnerChars[spinnerIndex]
		case StepSuccess:
			icon = "✓" // Check for success
		case StepFailed:
			icon = "✗" // X for failed
		}
		content += fmt.Sprintf("%s %s\n", icon, step.Name)
	}

	// Return content without border/sizing for layout containers to handle
	return content
}
