package components

import (
	"fmt"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ResultType represents the type of result to display
type ResultType int

const (
	ResultSuccess ResultType = iota
	ResultError
)

// ResultComponent provides a reusable success/error display with exit button
type ResultComponent struct {
	resultType ResultType
	message    string
	details    string
	width      int
	height     int
}

// NewResult creates a new result component
func NewResult(resultType ResultType, message, details string) *ResultComponent {
	return &ResultComponent{
		resultType: resultType,
		message:    message,
		details:    details,
		width:      41, // Default width
		height:     12, // Default height
	}
}

// SetSize updates the component dimensions
func (r *ResultComponent) SetSize(width, height int) {
	r.width = width
	r.height = height
}

// SetMessage updates the main message
func (r *ResultComponent) SetMessage(message string) {
	r.message = message
}

// SetDetails updates the details text
func (r *ResultComponent) SetDetails(details string) {
	r.details = details
}

// Update handles Bubble Tea messages
func (r *ResultComponent) Update(msg tea.Msg) tea.Cmd {
	return nil
}

// ViewPanel renders the result as a left panel with exit button
func (r *ResultComponent) ViewPanel() string {
	var icon string
	var iconColor lipgloss.Color
	
	if r.resultType == ResultSuccess {
		icon = "✅"
		iconColor = lipgloss.Color("#00FF00")
	} else {
		icon = "❌"
		iconColor = lipgloss.Color("#FF0000")
	}
	
	// Create content with icon and message
	iconStyle := lipgloss.NewStyle().
		Foreground(iconColor).
		Bold(true)
	
	exitButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#874BFD")).
		Bold(true).
		Padding(0, 2).
		Margin(2, 0, 0, 0).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Align(lipgloss.Center)
	
	content := fmt.Sprintf("%s\n\n%s\n\n%s",
		iconStyle.Render(icon+" "+r.message),
		"",
		exitButtonStyle.Render("[ Exit ]"))
	
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(1, 1).
		Width(r.width).
		Height(r.height).
		Align(lipgloss.Center).
		Render(content)
}

// ViewDetails renders the result details as a right panel
func (r *ResultComponent) ViewDetails() string {
	content := r.details
	if content == "" {
		content = "No details available"
	}
	
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(1, 1).
		Width(r.width).
		Height(r.height).
		Render(content)
}

// NewResultComponent creates a new result component (for compatibility)
func NewResultComponent(success bool, result interface{}, width, height int) *ResultComponent {
	var resultType ResultType
	var message, details string
	
	if success {
		resultType = ResultSuccess
		message = "Build completed successfully!"
		details = formatBuildDetails(result, true)
	} else {
		resultType = ResultError
		message = "Build failed"
		details = formatBuildDetails(result, false)
	}
	
	rc := NewResult(resultType, message, details)
	
	rc.SetSize(width, height)
	return rc
}



// formatBuildDetails extracts and formats build configuration and result details
func formatBuildDetails(result interface{}, success bool) string {
	details := "📋 Build Configuration:\n\n"
	
	// Use reflection to extract build flow information
	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	// Try to get Flow field
	if v.Kind() == reflect.Struct {
		flowField := v.FieldByName("Flow")
		if flowField.IsValid() && !flowField.IsNil() {
			flow := flowField.Elem()
			
			// Extract build configuration details
			if buildType := flow.FieldByName("BuildType"); buildType.IsValid() {
				details += fmt.Sprintf("• Build Type: %s\n", buildType.String())
			}
			if entrypoint := flow.FieldByName("Entrypoint"); entrypoint.IsValid() {
				details += fmt.Sprintf("• Entrypoint: %s\n", entrypoint.String())
			}
			if outputDir := flow.FieldByName("OutputDir"); outputDir.IsValid() {
				details += fmt.Sprintf("• Output Directory: %s\n", outputDir.String())
			}
			
			// Try to extract config details
			if configField := flow.FieldByName("Config"); configField.IsValid() && !configField.IsNil() {
				config := configField.Elem()
				
				if target := config.FieldByName("Target"); target.IsValid() {
					details += fmt.Sprintf("• Target: %s\n", target.String())
				}
				if stackSize := config.FieldByName("StackSize"); stackSize.IsValid() {
					var bytes int64
					if stackSize.Kind() == reflect.Int || stackSize.Kind() == reflect.Int64 {
						bytes = stackSize.Int()
					} else if stackSize.Kind() == reflect.Uint || stackSize.Kind() == reflect.Uint64 {
						bytes = int64(stackSize.Uint())
					}
					if bytes > 0 {
						mb := float64(bytes) / (1024 * 1024)
						details += fmt.Sprintf("• Stack Size: %.1f MB\n", mb)
					}
				}
				if initialMem := config.FieldByName("InitialMemory"); initialMem.IsValid() {
					var bytes int64
					if initialMem.Kind() == reflect.Int || initialMem.Kind() == reflect.Int64 {
						bytes = initialMem.Int()
					} else if initialMem.Kind() == reflect.Uint || initialMem.Kind() == reflect.Uint64 {
						bytes = int64(initialMem.Uint())
					}
					if bytes > 0 {
						mb := float64(bytes) / (1024 * 1024)
						details += fmt.Sprintf("• Initial Memory: %.1f MB\n", mb)
					}
				}
				if maxMem := config.FieldByName("MaximumMemory"); maxMem.IsValid() {
					var bytes int64
					if maxMem.Kind() == reflect.Int || maxMem.Kind() == reflect.Int64 {
						bytes = maxMem.Int()
					} else if maxMem.Kind() == reflect.Uint || maxMem.Kind() == reflect.Uint64 {
						bytes = int64(maxMem.Uint())
					}
					if bytes > 0 {
						mb := float64(bytes) / (1024 * 1024)
						details += fmt.Sprintf("• Maximum Memory: %.1f MB\n", mb)
					}
				}
				if gitHash := config.FieldByName("AOSGitHash"); gitHash.IsValid() {
					hash := gitHash.String()
					if len(hash) > 8 {
						hash = hash[:8] + "..."
					}
					details += fmt.Sprintf("• AOS Git Hash: %s\n", hash)
				}
			}
		}
	}
	
	details += "\n"
	
	if success {
		details += "✅ Build completed successfully!\n\n"
		details += "📁 Output files created:\n"
		details += "• process.wasm - Compiled WASM binary\n"
		details += "• bundled.lua - Lua code bundle\n"
		details += "• config.yml - Build configuration"
	} else {
		details += "❌ Build failed\n\n"
		
		// Try to extract error information
		if errorField := v.FieldByName("Error"); errorField.IsValid() && !errorField.IsNil() {
			details += "Error: " + errorField.Elem().String()
		} else {
			details += fmt.Sprintf("Error details: %v", result)
		}
	}
	
	return details
}



// ViewDetailsPanel returns the details panel view (alias for compatibility)
func (r *ResultComponent) ViewDetailsPanel() string {
	return r.ViewDetails()
}

// ViewPanelContent renders the result content without borders/styling for use in layouts
func (r *ResultComponent) ViewPanelContent() string {
	var icon string
	var iconColor lipgloss.Color
	
	if r.resultType == ResultSuccess {
		icon = "✅"
		iconColor = lipgloss.Color("#00FF00")
	} else {
		icon = "❌"
		iconColor = lipgloss.Color("#FF0000")
	}
	
	// Create content with icon and message
	iconStyle := lipgloss.NewStyle().
		Foreground(iconColor).
		Bold(true)
	
	exitButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#874BFD")).
		Bold(true).
		Padding(0, 2).
		Margin(2, 0, 0, 0).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Align(lipgloss.Center)
	
	content := fmt.Sprintf("%s\n\n%s\n\n%s",
		iconStyle.Render(icon+" "+r.message),
		"",
		exitButtonStyle.Render("[ Exit ]"))
	
	// Return content without outer border/sizing for layout containers to handle
	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(content)
}

// ViewDetailsContent renders the result details without borders/styling for use in layouts
func (r *ResultComponent) ViewDetailsContent() string {
	content := r.details
	if content == "" {
		content = "No details available"
	}
	
	// Return content without outer border/sizing for layout containers to handle
	return content
}
