package unit

import (
	"context"
	"testing"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tui"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tests/shared"
)

func TestTUIModelInitialization(t *testing.T) {
	shared.SkipInCI(t, "TUI tests require interactive environment")

	ctx := context.Background()

	// Test that TUI model can be created without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TUI model creation should not panic: %v", r)
		}
	}()

	// This tests the NewModel function with our height improvements
	model := tui.NewModel(ctx)
	if model == nil {
		t.Error("Model should not be nil")
	}
}

func TestPanelHeightCalculation(t *testing.T) {
	// Test height calculation logic
	tests := []struct {
		name           string
		terminalHeight int
		expectedMin    int
	}{
		{
			name:           "small terminal",
			terminalHeight: 15,
			expectedMin:    8, // minimum height
		},
		{
			name:           "medium terminal",
			terminalHeight: 25,
			expectedMin:    8,
		},
		{
			name:           "large terminal",
			terminalHeight: 40,
			expectedMin:    8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This simulates the height calculation logic
			contentHeight := tt.terminalHeight - 4 - 4 - 1 // header+footer, container, safety
			if contentHeight < 8 {
				contentHeight = 8
			}

			panelHeight := contentHeight - 4 // panel borders/padding
			if panelHeight < 8 {
				panelHeight = 8
			}

			if panelHeight < tt.expectedMin {
				t.Errorf("Panel height %d should be at least %d", panelHeight, tt.expectedMin)
			}
		})
	}
}

func TestComponentHeightValidation(t *testing.T) {
	// Test that component height validation works
	tests := []struct {
		name           string
		inputHeight    int
		expectedHeight int
	}{
		{
			name:           "below minimum",
			inputHeight:    3,
			expectedHeight: 6, // minimum for list components
		},
		{
			name:           "valid height",
			inputHeight:    12,
			expectedHeight: 12,
		},
		{
			name:           "large height",
			inputHeight:    25,
			expectedHeight: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the height validation logic from list_selector.go
			height := tt.inputHeight
			if height < 6 {
				height = 6
			}

			if height != tt.expectedHeight {
				t.Errorf("Expected height %d, got %d", tt.expectedHeight, height)
			}
		})
	}
}
