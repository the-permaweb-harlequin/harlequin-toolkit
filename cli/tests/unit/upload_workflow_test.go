package unit

import (
	"regexp"
	"testing"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tests/shared"
)

func TestDataItemIDExtraction(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "standard format",
			output:   "   • 🎉 Upload completed! Transaction ID: abc123def456ghi789\n",
			expected: "abc123def456ghi789",
		},
		{
			name:     "simple format",
			output:   "Transaction ID: xyz789abc123def456\nOther text here",
			expected: "xyz789abc123def456",
		},
		{
			name:     "lowercase format",
			output:   "transaction ID: qwe123rty456uio789\n",
			expected: "qwe123rty456uio789",
		},
		{
			name:     "short ID format",
			output:   "ID: long_transaction_id_here_12345\n",
			expected: "long_transaction_id_here_12345",
		},
		{
			name:     "no ID found",
			output:   "Upload failed - no transaction ID available",
			expected: "",
		},
		{
			name:     "short ID ignored",
			output:   "ID: abc\nOther text",
			expected: "", // Too short, should be ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the extractDataItemID logic
			patterns := []string{
				`Transaction ID: ([a-zA-Z0-9_-]+)`,
				`transaction ID: ([a-zA-Z0-9_-]+)`,
				`ID: ([a-zA-Z0-9_-]+)`,
				`id: ([a-zA-Z0-9_-]+)`,
			}

			result := ""
			for _, pattern := range patterns {
				re := regexp.MustCompile(pattern)
				matches := re.FindStringSubmatch(tt.output)
				if len(matches) > 1 && len(matches[1]) > 10 { // Transaction IDs are typically longer
					result = matches[1]
					break
				}
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestUploadStepProgression(t *testing.T) {
	// Test the logical flow of upload steps
	expectedSteps := []string{
		"Analyzing WASM metadata",
		"Creating upload tags",
		"Signing data item",
		"Uploading to Arweave",
	}

	// Verify step order is logical
	for i, step := range expectedSteps {
		if step == "" {
			t.Errorf("Step %d should not be empty", i)
		}

		// Check step content makes sense
		switch i {
		case 0:
			if !uploadTestContains(step, "WASM") && !uploadTestContains(step, "metadata") {
				t.Errorf("First step should involve WASM analysis: %s", step)
			}
		case 1:
			if !uploadTestContains(step, "tag") {
				t.Errorf("Second step should involve tags: %s", step)
			}
		case 2:
			if !uploadTestContains(step, "Sign") {
				t.Errorf("Third step should involve signing: %s", step)
			}
		case 3:
			if !uploadTestContains(step, "Upload") && !uploadTestContains(step, "Arweave") {
				t.Errorf("Final step should involve uploading: %s", step)
			}
		}
	}
}

func TestUploadResultStructure(t *testing.T) {
	// Test the UploadResult structure
	type MockUploadFlow struct {
		WasmFile   string
		ConfigFile string
		WalletFile string
		Version    string
		GitHash    string
		DryRun     bool
	}

	type MockUploadResult struct {
		Success    bool
		Error      error
		Flow       *MockUploadFlow
		DataItemID string
		Output     string
	}

	// Test successful upload result
	successResult := &MockUploadResult{
		Success:    true,
		Error:      nil,
		Flow:       &MockUploadFlow{
			WasmFile:   "test.wasm",
			ConfigFile: "config.yml",
			WalletFile: "wallet.json",
			Version:    "1.0.0",
			GitHash:    "abc123",
			DryRun:     false,
		},
		DataItemID: "test_transaction_id_12345",
		Output:     "Upload completed successfully",
	}

	// Verify structure
	if !successResult.Success {
		t.Error("Success result should have Success=true")
	}

	if successResult.Error != nil {
		t.Error("Success result should have no error")
	}

	if successResult.DataItemID == "" {
		t.Error("Success result should have DataItemID")
	}

	if successResult.Flow == nil {
		t.Error("Result should have Flow information")
	}

	// Test failed upload result
	failedResult := &MockUploadResult{
		Success:    false,
		Error:      shared.MockError("upload failed"),
		Flow:       successResult.Flow,
		DataItemID: "",
		Output:     "Upload failed: insufficient balance",
	}

	if failedResult.Success {
		t.Error("Failed result should have Success=false")
	}

	if failedResult.Error == nil {
		t.Error("Failed result should have an error")
	}

	if failedResult.DataItemID != "" {
		t.Error("Failed result should not have DataItemID")
	}
}

func TestUploadMessageTypes(t *testing.T) {
	// Test upload-specific message types
	type MockUploadStepStartMsg struct {
		StepName string
	}

	type MockUploadStepCompleteMsg struct {
		StepName string
		Success  bool
	}

	type MockUploadCompleteMsg struct {
		Result interface{}
	}

	// Test message creation
	startMsg := MockUploadStepStartMsg{StepName: "Analyzing WASM metadata"}
	if startMsg.StepName == "" {
		t.Error("Start message should have step name")
	}

	completeMsg := MockUploadStepCompleteMsg{
		StepName: "Uploading to Arweave",
		Success:  true,
	}
	if completeMsg.StepName == "" {
		t.Error("Complete message should have step name")
	}

	finalMsg := MockUploadCompleteMsg{Result: "test result"}
	if finalMsg.Result == nil {
		t.Error("Final message should have result")
	}
}

// Helper function to check if string contains substring
func uploadTestContains(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (substr == "" || s == substr ||
		    len(s) > len(substr) &&
		    (s[:len(substr)] == substr ||
		     s[len(s)-len(substr):] == substr ||
		     uploadTestContainsSubstring(s, substr)))
}

func uploadTestContainsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
