# TUI Components

This directory contains reusable Bubble Tea components for the Harlequin TUI.

## Components

### 🎯 SelectorComponent (`selector.go`)
A reusable option selector with keyboard navigation and highlighting.

**Features:**
- Up/down navigation
- Visual highlighting of selected option
- Customizable options and dimensions
- Clean, bordered display

**Usage:**
```go
selector := components.NewSelector("Build Types", []string{"AOS Flavour"})
selector.SetSize(41, 12)
view := selector.View()
```

### 📊 ProgressComponent (`progress.go`)
An animated progress tracker for multi-step processes.

**Features:**
- Animated spinners for running steps
- Success/failure status indicators
- Customizable step names
- Real-time status updates

**Usage:**
```go
steps := []string{"Copy AOS Files", "Bundle Lua", "Build WASM"}
progress := components.NewProgress("Build Progress", steps)
progress.SetStepStatus("Copy AOS Files", components.StepRunning)
view := progress.View()
```

### ✅ ResultComponent (`result.go`)
Success/error display screens with exit functionality.

**Features:**
- Success and error variants
- Detailed information display
- Exit button integration
- Dual-panel layout (result + details)

**Usage:**
```go
result := components.NewResult(components.ResultSuccess, "Build completed!", detailsText)
leftPanel := result.ViewPanel()
rightPanel := result.ViewDetails()
```

### ⚙️ ConfigEditorComponent (`config_editor.go`)
A comprehensive configuration editor with form validation.

**Features:**
- Text inputs with blinking cursors
- WASM target selector (32/64-bit)
- Memory value formatting (MB display)
- Save/Cancel buttons
- Field descriptions

**Usage:**
```go
editor := components.NewConfigEditor()
editor.SetFieldValues(32, 3.0, 5.0, 512.0) // target, stack, initial, max memory
if editor.HandleKeyPress("down") {
    // Navigation handled
}
view := editor.View()
```

### 🎨 LayoutUtils (`layout.go`)
Common layout utilities and styling helpers.

**Features:**
- Consistent header/panel/controls styling
- Two-column layout management
- Content centering utilities
- Responsive width calculations

**Usage:**
```go
layout := components.NewLayoutUtils()
header := layout.CreateHeader("Build Configuration", 86)
content := layout.CreateTwoColumnLayout(leftPanel, rightPanel, 90)
controls := layout.CreateControls("↑/↓ Navigate • Enter Select", 84)
mainLayout := layout.CreateMainLayout(header, content, controls)
final := layout.CenterContent(mainLayout, terminalWidth, terminalHeight)
```

### 📚 HelpRenderer (`help.go`)
Markdown-based help system using Glamour and Bubbles viewport.

**Features:**
- Markdown rendering with syntax highlighting
- Scrollable viewport
- Auto-styling for terminal themes
- File path resolution for docs

**Usage:**
```go
err := components.ShowHelp("build") // Shows docs/commands/build.md
```

## Benefits of Componentization

### Before (1494 lines in main.go):
```go
// Massive Model struct with everything
type Model struct {
    state          ViewState
    flow           *BuildFlow
    buildSteps     []BuildStep
    outputLines    []string
    terminalWidth  int
    terminalHeight int
    selectedIndex  int
    availableOptions []string
    configEditFields []string
    configFieldIndex int
    isEditingText    bool
    cursorVisible    bool
    program        *tea.Program
}

// Giant createConfigEditPanel method (200+ lines)
func (m *Model) createConfigEditPanel() string {
    fieldNames := []string{"WASM Target", "Stack Size (MB)", ...}
    // ... 200+ lines of complex layout logic
}
```

### After (with components):
```go
// Clean, focused Model struct
type Model struct {
    state      ViewState
    flow       *BuildFlow
    // Components
    selector   *components.SelectorComponent
    progress   *components.ProgressComponent
    configEditor *components.ConfigEditorComponent
    result     *components.ResultComponent
    layout     *components.LayoutUtils
    // Terminal state
    terminalWidth  int
    terminalHeight int
    program       *tea.Program
}

// Simple, clean panel creation
func (m *Model) createConfigEditPanel() string {
    return m.configEditor.View()
}
```

## Impact

- **Reduced complexity**: Main TUI file goes from 1494 lines to ~500 lines
- **Reusable components**: Each component can be used across different views
- **Better testing**: Components can be unit tested independently
- **Easier maintenance**: Changes to UI elements happen in one place
- **Consistent styling**: All components use the same design language
- **Type safety**: Clear interfaces and encapsulation
- **Future extensibility**: Easy to add new views and features

## Next Steps

To fully refactor the main TUI:

1. Update the main Model struct to use components
2. Replace create*Panel methods with component calls
3. Move keyboard handling to components where appropriate
4. Update the Update() method to delegate to components
5. Simplify the View() method using layout utilities

This will result in a much cleaner, more maintainable TUI codebase!
