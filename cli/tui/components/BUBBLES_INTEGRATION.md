# Bubbles Integration Analysis

## 🎯 Current Components vs. Available Bubbles Components

### **Our Components That Should Use Bubbles:**

| Our Component | Should Use | Benefits |
|---------------|------------|----------|
| **SelectorComponent** | `bubbles/list` | Built-in filtering, pagination, better keyboard handling, status messages |
| **ConfigEditorComponent** | `bubbles/textinput` + `bubbles/key` | Professional text inputs with validation, better cursor handling, key bindings |
| **ProgressComponent** | `bubbles/progress` + `bubbles/spinner` | Smooth animations, better visual progress bars, proven spinner logic |
| **HelpRenderer** | `bubbles/help` + `bubbles/key` | Automatic key binding help generation, consistent help formatting |
| **Layout/Navigation** | `bubbles/key` | Centralized, configurable key bindings with help text |

### **Specific Improvements:**

## 1. 📁 **Entrypoint Selection → `bubbles/filepicker`**

**Current**: Manual file discovery + custom selector
**Upgrade**: Professional file picker with directory navigation

```go
// Instead of our basic file list
entrypoints := findLuaFiles(cwd) // Manual scanning

// Use Bubbles filepicker
picker := filepicker.New()
picker.AllowedTypes = []string{".lua"}
picker.CurrentDirectory = cwd
// Automatic directory navigation, file filtering, better UX
```

## 2. 📋 **Option Selection → `bubbles/list`**

**Current**: Basic string list with manual highlighting
**Upgrade**: Feature-rich list with filtering, help, status

```go
// Instead of manual option rendering
type SelectorComponent struct {
    options []string
    selectedIndex int
}

// Use Bubbles list with rich features
items := []list.Item{
    listItem{title: "AOS Flavour", desc: "Build WASM with Lua injection"},
    listItem{title: "Standard", desc: "Basic build configuration"},
}
listModel := list.New(items, list.NewDefaultDelegate(), 40, 12)
listModel.Title = "Build Types"
// Automatic filtering, pagination, keyboard shortcuts
```

## 3. ⌨️ **Key Bindings → `bubbles/key`**

**Current**: String matching in switch statements
**Upgrade**: Structured key bindings with automatic help

```go
// Instead of manual key handling
switch msg.String() {
case "up", "k":
    // handle up
case "down", "j":
    // handle down
}

// Use structured key bindings
type KeyMap struct {
    Up    key.Binding
    Down  key.Binding
    Enter key.Binding
    Quit  key.Binding
}

var DefaultKeyMap = KeyMap{
    Up: key.NewBinding(
        key.WithKeys("up", "k"),
        key.WithHelp("↑/k", "move up"),
    ),
    Down: key.NewBinding(
        key.WithKeys("down", "j"), 
        key.WithHelp("↓/j", "move down"),
    ),
    // Automatic help generation!
}
```

## 4. 📝 **Text Input → `bubbles/textinput`**

**Current**: Manual cursor blinking, character handling
**Upgrade**: Professional text inputs with validation

```go
// Instead of manual text input handling
if msg.String() == "backspace" && len(field.Value) > 0 {
    field.Value = field.Value[:len(field.Value)-1]
}

// Use Bubbles textinput
input := textinput.New()
input.Placeholder = "3.0"
input.Focus()
input.CharLimit = 10
input.Validate = func(s string) error {
    if _, err := strconv.ParseFloat(s, 64); err != nil {
        return errors.New("must be a number")
    }
    return nil
}
// Automatic cursor, validation, copy/paste support
```

## 5. 📊 **Progress Display → `bubbles/progress`**

**Current**: Manual spinner animation with harmonica
**Upgrade**: Smooth progress bars with built-in animations

```go
// Instead of manual spinner tracking
type ProgressStep struct {
    SpinPhase float64
    Spinner   harmonica.Spring
}

// Use Bubbles progress
progressBar := progress.New(
    progress.WithDefaultGradient(),
    progress.WithWidth(40),
)
// Built-in smooth animations, gradients, percentage display
```

## 6. ❓ **Help System → `bubbles/help`**

**Current**: Manual help text formatting
**Upgrade**: Automatic help generation from key bindings

```go
// Instead of hardcoded help text
controls := "↑/↓ Navigate • Enter Select • q Quit"

// Use Bubbles help with automatic generation
helpModel := help.New()
helpView := helpModel.View(keyMap) // Automatically generates from KeyMap!
// Consistent formatting, responsive layout, automatic updates
```

## 🚀 **Implementation Priority:**

### **High Impact, Low Effort:**
1. **`bubbles/key`** - Replace all manual key handling
2. **`bubbles/help`** - Automatic help generation
3. **`bubbles/textinput`** - Replace config input fields

### **High Impact, Medium Effort:**
4. **`bubbles/list`** - Replace selector component
5. **`bubbles/filepicker`** - Replace entrypoint selection

### **Medium Impact, Medium Effort:**
6. **`bubbles/progress`** - Replace progress animations

## 💡 **Example Integration:**

```go
// Before: Manual everything
type Model struct {
    selectedIndex    int
    availableOptions []string
    configEditFields []string
    cursorVisible    bool
    // ... lots of manual state
}

// After: Bubbles-powered
type Model struct {
    // Bubbles components handle their own state
    list         list.Model
    filepicker   filepicker.Model
    configInputs []textinput.Model
    progress     progress.Model
    help         help.Model
    keyMap       KeyMap
    
    // Minimal app state
    state ViewState
    flow  *BuildFlow
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keyMap.Quit):
            return m, tea.Quit
        case key.Matches(msg, m.keyMap.Select):
            return m.handleSelection()
        }
    }
    
    // Delegate to components
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    cmds = append(cmds, cmd)
    
    m.filepicker, cmd = m.filepicker.Update(msg)
    cmds = append(cmds, cmd)
    
    return m, tea.Batch(cmds...)
}

func (m Model) View() string {
    // Components render themselves
    switch m.state {
    case ViewEntrypointSelection:
        return m.filepicker.View()
    case ViewBuildTypeSelection:
        return m.list.View()
    case ViewConfigEditing:
        return m.renderConfigForm() // Uses textinput models
    }
    
    // Help is automatically generated
    helpView := m.help.View(m.keyMap)
    return layout.CreateMainLayout(header, content, helpView)
}
```

## 🎯 **Benefits:**

- **Less Code**: Remove ~500 lines of manual input/navigation handling
- **Better UX**: Professional text inputs, file navigation, help system
- **Consistent**: All components follow Bubbles design patterns  
- **Maintainable**: Battle-tested components, automatic updates
- **Accessible**: Built-in keyboard navigation patterns
- **Extensible**: Easy to add new features using existing patterns

## 📋 **Next Steps:**

1. **Create new components** using Bubbles primitives
2. **Replace manual key handling** with `bubbles/key` 
3. **Integrate filepicker** for entrypoint selection
4. **Use textinput** for configuration fields
5. **Add automatic help** generation
6. **Test and iterate** on the improved UX
