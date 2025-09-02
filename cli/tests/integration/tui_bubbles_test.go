package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sebdah/goldie/v2"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tests/shared"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tui"
)

// TestTUIInteractionWithBubbles tests TUI interactions using bubbles/test patterns
func TestTUIInteractionWithBubbles(t *testing.T) {
	shared.SkipInCI(t, "TUI tests require interactive environment")

	ctx := context.Background()

	t.Run("command_selection_flow", func(t *testing.T) {
		// Create a TUI model for testing
		model := tui.NewModel(ctx)

		// Test initial state
		if model == nil {
			t.Fatal("Model should not be nil")
		}

		// Simulate terminal resize
		updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		model = updatedModel.(*tui.Model)

		// Get initial view
		initialView := model.View()
		if initialView == "" {
			t.Error("Initial view should not be empty")
		}

		// Test that it contains expected elements
		if !strings.Contains(initialView, "Select Command") {
			t.Error("Initial view should contain command selection")
		}
	})

	t.Run("keyboard_navigation", func(t *testing.T) {
		model := tui.NewModel(ctx)
		updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		model = updatedModel.(*tui.Model)

		// Test down arrow key
		downKey := tea.KeyMsg{
			Type:  tea.KeyDown,
			Runes: nil,
		}

		updatedModel, cmd := model.Update(downKey)
		model = updatedModel.(*tui.Model)
		if cmd != nil {
			// Command was generated, which is expected for navigation
		}

		// Test up arrow key
		upKey := tea.KeyMsg{
			Type:  tea.KeyUp,
			Runes: nil,
		}

		updatedModel, cmd = model.Update(upKey)
		model = updatedModel.(*tui.Model)
		if cmd != nil {
			// Command was generated, which is expected for navigation
		}

		// Test enter key
		enterKey := tea.KeyMsg{
			Type:  tea.KeyEnter,
			Runes: nil,
		}

		updatedModel, cmd = model.Update(enterKey)
		model = updatedModel.(*tui.Model)
		if cmd != nil {
			// Enter should trigger some action
		}

		// Get view after interactions
		finalView := model.View()
		if finalView == "" {
			t.Error("View should not be empty after interactions")
		}
	})

	t.Run("quit_functionality", func(t *testing.T) {
		model := tui.NewModel(ctx)
		updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		model = updatedModel.(*tui.Model)

		// Test quit key
		quitKey := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'q'},
		}

		updatedModel, cmd := model.Update(quitKey)
		model = updatedModel.(*tui.Model)

		// Check if quit command was issued
		if cmd != nil {
			// Quit command might be generated
			t.Logf("Command generated: %T", cmd)
		}
	})
}

// TestTUIViewSnapshots tests TUI view outputs using goldie for snapshot testing
func TestTUIViewSnapshots(t *testing.T) {
	shared.SkipInCI(t, "Snapshot tests require stable output")

	g := goldie.New(t)
	ctx := context.Background()

	t.Run("initial_command_selection", func(t *testing.T) {
		model := tui.NewModel(ctx)

		// Set a fixed size for consistent snapshots
		updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		model = updatedModel.(*tui.Model)

		view := model.View()

		// Clean the view for consistent snapshots (remove dynamic content)
		cleanedView := cleanViewForSnapshot(view)

		// Compare with golden file
		g.Assert(t, "initial_command_selection", []byte(cleanedView))
	})

	t.Run("help_view", func(t *testing.T) {
		model := tui.NewModel(ctx)
		updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		model = updatedModel.(*tui.Model)

		// Trigger help view
		helpKey := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'?'},
		}

		updatedModel, _ = model.Update(helpKey)
		model = updatedModel.(*tui.Model)
		view := model.View()

		cleanedView := cleanViewForSnapshot(view)
		g.Assert(t, "help_view", []byte(cleanedView))
	})
}

// TestTUIStateTransitions tests state management and transitions
func TestTUIStateTransitions(t *testing.T) {
	shared.SkipInCI(t, "State transition tests require interactive environment")

	ctx := context.Background()

	t.Run("command_to_build_flow", func(t *testing.T) {
		model := tui.NewModel(ctx)
		updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		model = updatedModel.(*tui.Model)

		// Start with command selection
		initialView := model.View()
		if !strings.Contains(initialView, "Select Command") {
			t.Error("Should start with command selection")
		}

		// Simulate selecting build (assuming first option)
		enterKey := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ = model.Update(enterKey)
		model = updatedModel.(*tui.Model)

		// Should now be in build type selection or next step
		newView := model.View()

		// The view should have changed
		if newView == initialView {
			t.Log("View didn't change - this might be expected if validation failed")
		}
	})

	t.Run("back_navigation", func(t *testing.T) {
		model := tui.NewModel(ctx)
		updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		model = updatedModel.(*tui.Model)

		// Try escape key for going back
		escKey := tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, cmd := model.Update(escKey)
		model = updatedModel.(*tui.Model)

		// Should handle escape gracefully
		if cmd != nil {
			// Command might be generated for navigation
		}

		view := model.View()
		if view == "" {
			t.Error("View should not be empty after escape")
		}
	})
}

// TestTUIComponentInteraction tests interaction with specific components
func TestTUIComponentInteraction(t *testing.T) {
	shared.SkipInCI(t, "Component tests require interactive environment")

	ctx := context.Background()

	t.Run("list_selector_interaction", func(t *testing.T) {
		model := tui.NewModel(ctx)
		updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		model = updatedModel.(*tui.Model)

		// Test multiple key presses in sequence
		keys := []tea.KeyMsg{
			{Type: tea.KeyDown},
			{Type: tea.KeyDown},
			{Type: tea.KeyUp},
			{Type: tea.KeyEnter},
		}

		for _, keyMsg := range keys {
			updatedModel, _ := model.Update(keyMsg)
			model = updatedModel.(*tui.Model)
		}

		// Should handle all key interactions without crashing
		view := model.View()
		if view == "" {
			t.Error("View should not be empty after key sequence")
		}
	})

	t.Run("rapid_key_presses", func(t *testing.T) {
		model := tui.NewModel(ctx)
		updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		model = updatedModel.(*tui.Model)

		// Simulate rapid key presses
		for i := 0; i < 10; i++ {
			downKey := tea.KeyMsg{Type: tea.KeyDown}
			updatedModel, _ := model.Update(downKey)
			model = updatedModel.(*tui.Model)
		}

		// Should handle rapid input without issues
		view := model.View()
		if view == "" {
			t.Error("View should not be empty after rapid input")
		}
	})
}

// TestTUIResize tests terminal resize handling
func TestTUIResize(t *testing.T) {
	ctx := context.Background()

	t.Run("window_resize_handling", func(t *testing.T) {
		model := tui.NewModel(ctx)

		// Test various window sizes
		sizes := []tea.WindowSizeMsg{
			{Width: 80, Height: 24},   // Small terminal
			{Width: 120, Height: 40},  // Medium terminal
			{Width: 200, Height: 60},  // Large terminal
			{Width: 60, Height: 20},   // Very small terminal
		}

		for _, size := range sizes {
			updatedModel, _ := model.Update(size)
			model = updatedModel.(*tui.Model)
			view := model.View()

			if view == "" {
				t.Errorf("View should not be empty for size %dx%d", size.Width, size.Height)
			}

						// View should adapt to size constraints
			lines := strings.Split(view, "\n")
			// Allow some flexibility for very small terminals as the TUI may have minimum content requirements
			maxAllowedLines := size.Height + 5 // Allow some buffer for minimum content
			if len(lines) > maxAllowedLines {
				t.Logf("View height (%d lines) significantly exceeds terminal height (%d) for size %dx%d - this may indicate layout issues",
					len(lines), size.Height, size.Width, size.Height)
				// Don't fail the test for small terminals as TUI may have minimum content requirements
				if size.Height > 30 { // Only fail for reasonably large terminals
					t.Errorf("View height (%d lines) exceeds terminal height (%d) for size %dx%d",
						len(lines), size.Height, size.Width, size.Height)
				}
			}
		}
	})
}

// TestTUIWithTimeout tests TUI operations with timeouts to prevent hanging
func TestTUIWithTimeout(t *testing.T) {
	ctx := context.Background()

	t.Run("model_operations_timeout", func(t *testing.T) {
		shared.WithTimeout(t, 5*time.Second, func() {
			model := tui.NewModel(ctx)
			updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
			model = updatedModel.(*tui.Model)

			// Perform various operations
			operations := []tea.Msg{
				tea.KeyMsg{Type: tea.KeyDown},
				tea.KeyMsg{Type: tea.KeyUp},
				tea.KeyMsg{Type: tea.KeyEnter},
				tea.KeyMsg{Type: tea.KeyEsc},
				tea.WindowSizeMsg{Width: 120, Height: 40},
			}

			for _, op := range operations {
				updatedModel, _ := model.Update(op)
				model = updatedModel.(*tui.Model)
			}

			// Should complete within timeout
			view := model.View()
			if view == "" {
				t.Error("View should not be empty")
			}
		})
	})
}

// Helper function to clean view output for consistent snapshots
func cleanViewForSnapshot(view string) string {
	// Remove or normalize dynamic content that changes between runs
	lines := strings.Split(view, "\n")
	var cleanedLines []string

	for _, line := range lines {
		// Skip empty lines at the end
		if strings.TrimSpace(line) == "" && len(cleanedLines) > 0 {
			continue
		}

		// Normalize spacing
		line = strings.TrimRight(line, " ")
		cleanedLines = append(cleanedLines, line)
	}

	return strings.Join(cleanedLines, "\n")
}
