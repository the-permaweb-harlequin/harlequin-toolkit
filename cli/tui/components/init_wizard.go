package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InitWizardState int

const (
	StateProjectName InitWizardState = iota
	StateTemplateSelection
	StateAuthorName
	StateGitHubUser
	StateConfirmation
	StateComplete
)

type InitWizardComponent struct {
	state         InitWizardState
	projectInput  textinput.Model
	authorInput   textinput.Model
	githubInput   textinput.Model
	templateList  list.Model
	width         int
	height        int

	// Results
	ProjectName  string
	TemplateLang string
	AuthorName   string
	GitHubUser   string
	TargetDir    string

	// Completion callback
	OnComplete func(projectName, templateLang, authorName, githubUser, targetDir string)
}

type templateItem struct {
	name        string
	description string
	language    string
	buildSystem string
	features    []string
}

func (i templateItem) FilterValue() string { return i.name }
func (i templateItem) Title() string       { return fmt.Sprintf("%s (%s)", i.description, i.language) }
func (i templateItem) Description() string { return fmt.Sprintf("Build System: %s", i.buildSystem) }

var availableTemplateItems = []list.Item{
	templateItem{
		name:        "lua",
		description: "Lua AO Process",
		language:    "Lua",
		buildSystem: "CMake + LuaRocks",
		features: []string{
			"C trampoline with embedded Lua interpreter",
			"LuaRocks package management",
			"WebAssembly compilation",
			"Modular architecture",
			"Comprehensive testing with Busted",
		},
	},
	templateItem{
		name:        "c",
		description: "C AO Process",
		language:    "C",
		buildSystem: "CMake + Conan",
		features: []string{
			"Conan package management",
			"Google Test integration",
			"Emscripten WebAssembly compilation",
			"Memory-efficient implementation",
			"Docker build support",
		},
	},
	templateItem{
		name:        "rust",
		description: "Rust AO Process",
		language:    "Rust",
		buildSystem: "Cargo + wasm-pack",
		features: []string{
			"Thread-safe state management",
			"Serde JSON serialization",
			"wasm-bindgen WebAssembly bindings",
			"Comprehensive error handling",
			"Size-optimized builds",
		},
	},
	templateItem{
		name:        "assemblyscript",
		description: "AssemblyScript AO Process",
		language:    "AssemblyScript",
		buildSystem: "AssemblyScript Compiler",
		features: []string{
			"TypeScript-like syntax",
			"Custom JSON handling",
			"Memory-safe operations",
			"Node.js testing framework",
			"Size optimization",
		},
	},
}

func NewInitWizardComponent() *InitWizardComponent {
	// Project name input
	projectInput := textinput.New()
	projectInput.Placeholder = "my-ao-process"
	projectInput.Focus()
	projectInput.CharLimit = 50
	projectInput.Width = 40

	// Author name input
	authorInput := textinput.New()
	authorInput.Placeholder = "Your Name"
	authorInput.CharLimit = 50
	authorInput.Width = 40

	// GitHub username input
	githubInput := textinput.New()
	githubInput.Placeholder = "your-username"
	githubInput.CharLimit = 50
	githubInput.Width = 40

	// Template list
	templateList := list.New(availableTemplateItems, list.NewDefaultDelegate(), 0, 0)
	templateList.Title = "Select Template"
	templateList.SetShowStatusBar(false)
	templateList.SetFilteringEnabled(false)
	templateList.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#902f17")).
		Bold(true).
		Padding(0, 0, 1, 0)

	return &InitWizardComponent{
		state:        StateProjectName,
		projectInput: projectInput,
		authorInput:  authorInput,
		githubInput:  githubInput,
		templateList: templateList,
	}
}

func (iwc *InitWizardComponent) Init() tea.Cmd {
	return textinput.Blink
}

func (iwc *InitWizardComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		iwc.width = msg.Width
		iwc.height = msg.Height
		iwc.templateList.SetWidth(msg.Width - 4)
		iwc.templateList.SetHeight(msg.Height - 8)

	case tea.KeyMsg:
		switch iwc.state {
		case StateProjectName:
			switch msg.String() {
			case "enter":
				if strings.TrimSpace(iwc.projectInput.Value()) != "" {
					iwc.ProjectName = strings.TrimSpace(iwc.projectInput.Value())
					iwc.state = StateTemplateSelection
					return iwc, nil
				}
			case "ctrl+c", "esc":
				return iwc, tea.Quit
			}
			iwc.projectInput, cmd = iwc.projectInput.Update(msg)
			return iwc, cmd

		case StateTemplateSelection:
			switch msg.String() {
			case "enter":
				if selected := iwc.templateList.SelectedItem(); selected != nil {
					if item, ok := selected.(templateItem); ok {
						iwc.TemplateLang = item.name
						iwc.state = StateAuthorName
						iwc.authorInput.Focus()
						return iwc, textinput.Blink
					}
				}
			case "ctrl+c", "esc":
				return iwc, tea.Quit
			}
			iwc.templateList, cmd = iwc.templateList.Update(msg)
			return iwc, cmd

		case StateAuthorName:
			switch msg.String() {
			case "enter":
				iwc.AuthorName = strings.TrimSpace(iwc.authorInput.Value())
				iwc.state = StateGitHubUser
				iwc.authorInput.Blur()
				iwc.githubInput.Focus()
				return iwc, textinput.Blink
			case "ctrl+c", "esc":
				return iwc, tea.Quit
			}
			iwc.authorInput, cmd = iwc.authorInput.Update(msg)
			return iwc, cmd

		case StateGitHubUser:
			switch msg.String() {
			case "enter":
				iwc.GitHubUser = strings.TrimSpace(iwc.githubInput.Value())
				iwc.state = StateConfirmation
				iwc.githubInput.Blur()
				return iwc, nil
			case "ctrl+c", "esc":
				return iwc, tea.Quit
			}
			iwc.githubInput, cmd = iwc.githubInput.Update(msg)
			return iwc, cmd

		case StateConfirmation:
			switch msg.String() {
			case "enter", "y", "Y":
				iwc.state = StateComplete
				if iwc.OnComplete != nil {
					iwc.OnComplete(iwc.ProjectName, iwc.TemplateLang, iwc.AuthorName, iwc.GitHubUser, iwc.TargetDir)
				}
				return iwc, tea.Quit
			case "n", "N", "ctrl+c", "esc":
				return iwc, tea.Quit
			}

		case StateComplete:
			return iwc, tea.Quit
		}
	}

	return iwc, tea.Batch(cmds...)
}

func (iwc *InitWizardComponent) View() string {
	switch iwc.state {
	case StateProjectName:
		return iwc.renderProjectNameInput()
	case StateTemplateSelection:
		return iwc.renderTemplateSelection()
	case StateAuthorName:
		return iwc.renderAuthorInput()
	case StateGitHubUser:
		return iwc.renderGitHubInput()
	case StateConfirmation:
		return iwc.renderConfirmation()
	case StateComplete:
		return iwc.renderComplete()
	}
	return ""
}

func (iwc *InitWizardComponent) renderProjectNameInput() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#902f17")).
		Bold(true).
		Render("🎭 Create New AO Process Project")

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#564f41")).
		Render("Enter a name for your project")

	input := iwc.projectInput.View()

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		Render("Press Enter to continue • Ctrl+C to cancel")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		subtitle,
		"",
		input,
		"",
		instructions,
	)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#902f17")).
		Padding(2).
		Render(content)
}

func (iwc *InitWizardComponent) renderTemplateSelection() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#902f17")).
		Bold(true).
		Render(fmt.Sprintf("🎭 Project: %s", iwc.ProjectName))

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#564f41")).
		Render("Choose a template for your project")

	// Show features of selected template
	var features string
	if selected := iwc.templateList.SelectedItem(); selected != nil {
		if item, ok := selected.(templateItem); ok {
			featureList := make([]string, len(item.features))
			for i, feature := range item.features {
				featureList[i] = fmt.Sprintf("  • %s", feature)
			}
			features = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#564f41")).
				Render(strings.Join(featureList, "\n"))
		}
	}

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		Render("↑/↓ to navigate • Enter to select • Ctrl+C to cancel")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		subtitle,
		"",
		iwc.templateList.View(),
		"",
		features,
		"",
		instructions,
	)

	return content
}

func (iwc *InitWizardComponent) renderAuthorInput() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#902f17")).
		Bold(true).
		Render(fmt.Sprintf("🎭 Project: %s (%s)", iwc.ProjectName, iwc.TemplateLang))

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#564f41")).
		Render("Enter author name (optional)")

	input := iwc.authorInput.View()

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		Render("Press Enter to continue • Ctrl+C to cancel")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		subtitle,
		"",
		input,
		"",
		instructions,
	)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#902f17")).
		Padding(2).
		Render(content)
}

func (iwc *InitWizardComponent) renderGitHubInput() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#902f17")).
		Bold(true).
		Render(fmt.Sprintf("🎭 Project: %s (%s)", iwc.ProjectName, iwc.TemplateLang))

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#564f41")).
		Render("Enter GitHub username (optional)")

	input := iwc.githubInput.View()

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		Render("Press Enter to continue • Ctrl+C to cancel")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		subtitle,
		"",
		input,
		"",
		instructions,
	)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#902f17")).
		Padding(2).
		Render(content)
}

func (iwc *InitWizardComponent) renderConfirmation() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#902f17")).
		Bold(true).
		Render("🎭 Confirm Project Creation")

	details := []string{
		fmt.Sprintf("Project Name: %s", iwc.ProjectName),
		fmt.Sprintf("Template: %s", iwc.TemplateLang),
	}

	if iwc.AuthorName != "" {
		details = append(details, fmt.Sprintf("Author: %s", iwc.AuthorName))
	}

	if iwc.GitHubUser != "" {
		details = append(details, fmt.Sprintf("GitHub: %s", iwc.GitHubUser))
	}

	targetDir := iwc.TargetDir
	if targetDir == "" {
		targetDir = iwc.ProjectName
	}
	details = append(details, fmt.Sprintf("Directory: %s", targetDir))

	detailsText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#564f41")).
		Render(strings.Join(details, "\n"))

	question := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#902f17")).
		Bold(true).
		Render("Create this project? (y/N)")

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		Render("Y to create • N to cancel")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		detailsText,
		"",
		question,
		"",
		instructions,
	)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#902f17")).
		Padding(2).
		Render(content)
}

func (iwc *InitWizardComponent) renderComplete() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#93513a")).
		Bold(true).
		Render("🎉 Project Created Successfully!")

	message := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#564f41")).
		Render(fmt.Sprintf("Your %s project '%s' has been created.", iwc.TemplateLang, iwc.ProjectName))

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		message,
	)
}
