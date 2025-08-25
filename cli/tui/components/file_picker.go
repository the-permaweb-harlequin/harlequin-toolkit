package components

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FilePickerComponent provides a professional file picker for entrypoint selection
type FilePickerComponent struct {
	filepicker   filepicker.Model
	selectedFile string
	width        int
	height       int
}

// NewFilePicker creates a new file picker component
func NewFilePicker(width, height int) *FilePickerComponent {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".lua"}
	fp.CurrentDirectory = "."
	fp.Styles.Cursor = fp.Styles.Cursor.Foreground(lipgloss.Color("#902f17"))
	fp.Styles.Symlink = fp.Styles.Symlink.Foreground(lipgloss.Color("#564f41"))
	fp.Styles.Directory = fp.Styles.Directory.Foreground(lipgloss.Color("#93513a"))
	fp.Styles.File = fp.Styles.File.Foreground(lipgloss.Color("#efdec2"))
	fp.Styles.Permission = fp.Styles.Permission.Foreground(lipgloss.Color("#564f41"))
	fp.Styles.Selected = fp.Styles.Selected.Foreground(lipgloss.Color("#902f17")).Bold(true)
	fp.Styles.FileSize = fp.Styles.FileSize.Foreground(lipgloss.Color("#564f41"))

	return &FilePickerComponent{
		filepicker: fp,
		width:      width,
		height:     height,
	}
}

// SetCurrentDirectory sets the starting directory for the file picker
func (fpc *FilePickerComponent) SetCurrentDirectory(dir string) {
	fpc.filepicker.CurrentDirectory = dir
}

// SetSize updates the file picker dimensions
func (fpc *FilePickerComponent) SetSize(width, height int) {
	fpc.width = width
	fpc.height = height
}

// GetSelectedFile returns the selected file path
func (fpc *FilePickerComponent) GetSelectedFile() string {
	return fpc.selectedFile
}

// HasSelection returns true if a file has been selected
func (fpc *FilePickerComponent) HasSelection() bool {
	return fpc.selectedFile != ""
}

// Init implements the Bubble Tea model interface
func (fpc *FilePickerComponent) Init() tea.Cmd {
	return fpc.filepicker.Init()
}

// Update handles Bubble Tea messages
func (fpc *FilePickerComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Check if a Lua file is selected
			if ok, selectedFile := fpc.filepicker.DidSelectFile(msg); ok {
				if strings.HasSuffix(strings.ToLower(selectedFile), ".lua") {
					fpc.selectedFile = selectedFile
					return fpc, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	fpc.filepicker, cmd = fpc.filepicker.Update(msg)
	return fpc, cmd
}

// View renders the file picker
func (fpc *FilePickerComponent) View() string {
	content := fpc.filepicker.View()

	// Add title and instructions
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#902f17")).
		Render("Select Entrypoint File")

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#564f41")).
		Render("Navigate with ↑/↓, Enter directories with →, Select .lua files with Enter")

	// Combine title, instructions, and file picker
	fullContent := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		instructions,
		"",
		content,
	)

	// Create bordered panel
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(1, 1).
		Width(fpc.width).
		Height(fpc.height).
		Render(fullContent)
}

// GetCurrentDirectory returns the current directory being browsed
func (fpc *FilePickerComponent) GetCurrentDirectory() string {
	return fpc.filepicker.CurrentDirectory
}

// GetCurrentPath returns the currently highlighted path
func (fpc *FilePickerComponent) GetCurrentPath() string {
	if fpc.filepicker.FileSelected != "" {
		return filepath.Join(fpc.filepicker.CurrentDirectory, fpc.filepicker.FileSelected)
	}
	return fpc.filepicker.CurrentDirectory
}

// CreateEntrypointFilePicker creates a file picker specifically for Lua entrypoint selection
func CreateEntrypointFilePicker(startingDir string, width, height int) *FilePickerComponent {
	picker := NewFilePicker(width, height)
	picker.SetCurrentDirectory(startingDir)
	return picker
}
