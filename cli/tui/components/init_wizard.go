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
	StateTemplateSelection InitWizardState = iota
	StateProjectName
	StateAuthorName
	StateGitHubUser
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
	templateItem{
		name:        "go",
		description: "Go AO Process",
		language:    "Go",
		buildSystem: "Go + TinyGo",
		features: []string{
			"Type-safe Go programming",
			"Goroutines and channels",
			"Standard library support",
			"TinyGo WebAssembly compilation",
			"Efficient memory usage",
		},
	},
}

func NewInitWizardComponent() *InitWizardComponent {
	// Project name input
	projectInput := textinput.New()
	projectInput.Placeholder = "my-ao-process"
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

	// Template list with proper styling
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#902f17")).
		Bold(true).
		Underline(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#564f41"))

	templateList := list.New(availableTemplateItems, delegate, 0, 0)
	templateList.Title = "Select Template"
	templateList.SetShowStatusBar(false)
	templateList.SetFilteringEnabled(false)
	templateList.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#902f17")).
		Bold(true).
		Padding(0, 0, 1, 0)

	return &InitWizardComponent{
		state:        StateTemplateSelection,
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

		// Calculate proper panel width for template list (match main TUI exactly)
		containerWidth := msg.Width - 10
		layoutWidth := containerWidth - 2
		basePanelWidth := (layoutWidth - 1) / 2
		panelWidth := basePanelWidth + 3  // Same as main TUI, no double reduction
		if panelWidth < 15 {
			panelWidth = 15
		}

		iwc.templateList.SetWidth(panelWidth - 2)  // Account for panel padding
		iwc.templateList.SetHeight(12)

	case tea.KeyMsg:
		switch iwc.state {
		case StateTemplateSelection:
			switch msg.String() {
			case "enter":
				if selected := iwc.templateList.SelectedItem(); selected != nil {
					if item, ok := selected.(templateItem); ok {
						iwc.TemplateLang = item.name
						iwc.state = StateProjectName
						iwc.projectInput.Focus()
						return iwc, textinput.Blink
					}
				}
			case "ctrl+c", "esc":
				return iwc, tea.Quit
			}
			iwc.templateList, cmd = iwc.templateList.Update(msg)
			return iwc, cmd

		case StateProjectName:
			switch msg.String() {
			case "enter":
				if strings.TrimSpace(iwc.projectInput.Value()) != "" {
					iwc.ProjectName = strings.TrimSpace(iwc.projectInput.Value())
					iwc.state = StateAuthorName
					iwc.authorInput.Focus()
					return iwc, textinput.Blink
				}
			case "ctrl+c", "esc":
				return iwc, tea.Quit
			}
			iwc.projectInput, cmd = iwc.projectInput.Update(msg)
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
				iwc.state = StateComplete
				iwc.githubInput.Blur()

				// Immediately create the project
				if iwc.OnComplete != nil {
					iwc.OnComplete(iwc.ProjectName, iwc.TemplateLang, iwc.AuthorName, iwc.GitHubUser, iwc.TargetDir)
				}
				return iwc, tea.Quit
			case "ctrl+c", "esc":
				return iwc, tea.Quit
			}
			iwc.githubInput, cmd = iwc.githubInput.Update(msg)
			return iwc, cmd

		case StateComplete:
			return iwc, tea.Quit
		}
	}

	return iwc, tea.Batch(cmds...)
}

func (iwc *InitWizardComponent) View() string {
	switch iwc.state {
	case StateTemplateSelection:
		return iwc.renderTemplateSelection()
	case StateProjectName:
		return iwc.renderProjectNameInput()
	case StateAuthorName:
		return iwc.renderAuthorInput()
	case StateGitHubUser:
		return iwc.renderGitHubInput()
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
	// Left panel: Template list
	listTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#902f17")).
		Bold(true).
		Render("🎭 Select Template")

	listSubtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#564f41")).
		Render("Choose a template for your project")

	leftPanel := lipgloss.JoinVertical(lipgloss.Left,
		listTitle,
		"",
		listSubtitle,
		"",
		iwc.templateList.View(),
	)

	// Right panel: Template details
	var rightPanel string
	if selected := iwc.templateList.SelectedItem(); selected != nil {
		if item, ok := selected.(templateItem); ok {
			detailTitle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#902f17")).
				Bold(true).
				Render(fmt.Sprintf("%s Template", item.language))

			description := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#564f41")).
				Render(item.description)

			buildSystem := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666")).
				Render(fmt.Sprintf("Build System: %s", item.buildSystem))

			featuresTitle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#902f17")).
				Bold(true).
				Render("Features:")

			featureList := make([]string, len(item.features))
			for i, feature := range item.features {
				featureList[i] = fmt.Sprintf("• %s", feature)
			}
			features := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#564f41")).
				Render(strings.Join(featureList, "\n"))

			rightPanel = lipgloss.JoinVertical(lipgloss.Left,
				detailTitle,
				"",
				description,
				"",
				buildSystem,
				"",
				featuresTitle,
				"",
				features,
			)
		}
	} else {
		rightPanel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666")).
			Render("Select a template to see details")
	}

	// Calculate panel dimensions (matching main TUI's getPanelWidth logic)
	totalWidth := iwc.width
	if totalWidth == 0 {
		totalWidth = 90 // Default width
	}

	// Replicate the main TUI's panel width calculation exactly
	containerWidth := totalWidth - 10  // Container width with margin
	layoutWidth := containerWidth - 2  // Available width inside container
	basePanelWidth := (layoutWidth - 1) / 2  // Each panel gets half minus gap
	panelWidth := basePanelWidth + 3  // Add extra space (same as main TUI)

	// Ensure minimum width
	if panelWidth < 15 {
		panelWidth = 15
	}

	// Update template list dimensions
	iwc.templateList.SetWidth(panelWidth - 2)  // Account for panel padding
	iwc.templateList.SetHeight(12)

	// Style the panels (make both have borders for consistent appearance)
	leftStyled := LeftPanelStyle.
		Width(panelWidth).
		Height(15).
		Render(leftPanel)

	// Use LeftPanelStyle for right panel too to make them visually equal
	rightStyled := LeftPanelStyle.
		Width(panelWidth).
		Height(15).
		Render(rightPanel)

	// Instructions at the bottom
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		Align(lipgloss.Center).
		Render("↑/↓ to navigate • Enter to select • Ctrl+C to cancel")

	// Combine panels and instructions
	twoPanelLayout := lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, " ", rightStyled)

	return lipgloss.JoinVertical(lipgloss.Left,
		twoPanelLayout,
		"",
		instructions,
	)
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
