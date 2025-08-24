package luautils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Module represents a Lua module with its metadata
type Module struct {
	Name    string
	Path    string
	Content *string // Use pointer to represent undefined (nil)
}

// Bundle creates a bundled Lua executable from an entry Lua file
func Bundle(entryLuaPath string) (string, error) {
	project, err := createProjectStructure(entryLuaPath)
	if err != nil {
		return "", fmt.Errorf("failed to create project structure: %w", err)
	}

	bundledLua, err := createExecutableFromProject(project)
	if err != nil {
		return "", fmt.Errorf("failed to create executable: %w", err)
	}

	return bundledLua, nil
}

// createExecutableFromProject converts a project structure into executable Lua code
func createExecutableFromProject(project []Module) (string, error) {
	if len(project) == 0 {
		return "", fmt.Errorf("empty project")
	}

	var contents []Module

	// Process all modules except the main file (last one)
	for i := 0; i < len(project)-1; i++ {
		mod := project[i]

		// Check if we already have this module path
		var existing *Module
		for j := range contents {
			if contents[j].Path == mod.Path {
				existing = &contents[j]
				break
			}
		}

		var moduleContent string
		if existing == nil && mod.Content != nil {
			// Create the module function
			modFnName := getModFnName(mod.Name)
			moduleContent = fmt.Sprintf("-- module: \"%s\"\nlocal function _loaded_mod_%s()\n%s\nend\n",
				mod.Name, modFnName, *mod.Content)
		}

		// Create the require mapper
		var targetModName string
		if existing != nil {
			targetModName = existing.Name
		} else {
			targetModName = mod.Name
		}

		requireMapper := fmt.Sprintf("\n_G.package.loaded[\"%s\"] = _loaded_mod_%s()",
			mod.Name, getModFnName(targetModName))

		finalContent := moduleContent + requireMapper
		contents = append(contents, Module{
			Name:    mod.Name,
			Path:    mod.Path,
			Content: &finalContent,
		})
	}

	// Add the main file
	contents = append(contents, project[len(project)-1])

	// Combine all content
	var result strings.Builder
	for _, mod := range contents {
		if mod.Content != nil {
			result.WriteString("\n\n")
			result.WriteString(*mod.Content)
		}
	}

	return result.String(), nil
}

// createProjectStructure builds the project dependency tree from the main file
func createProjectStructure(mainFile string) ([]Module, error) {
	var sorted []Module
	cwd := filepath.Dir(mainFile)

	// Track visited nodes to avoid cycles
	visited := make(map[string]bool)

	// isSorted checks if the module is already in sorted list
	isSorted := func(nodePath string) bool {
		for _, sortedNode := range sorted {
			if sortedNode.Path == nodePath {
				return true
			}
		}
		return false
	}

	// DFS traversal
	var dfs func(Module) error
	dfs = func(currentNode Module) error {
		if visited[currentNode.Path] {
			return nil // Avoid cycles
		}
		visited[currentNode.Path] = true

		// Read the content of current node if it exists
		if _, err := os.Stat(currentNode.Path); err == nil {
			content, err := os.ReadFile(currentNode.Path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", currentNode.Path, err)
			}
			contentStr := string(content)
			currentNode.Content = &contentStr
		}

		childNodes, err := exploreNodes(currentNode, cwd)
		if err != nil {
			return fmt.Errorf("failed to explore nodes for %s: %w", currentNode.Path, err)
		}

		// Visit unvisited child nodes
		for _, childNode := range childNodes {
			if !isSorted(childNode.Path) {
				if err := dfs(childNode); err != nil {
					return err
				}
			}
		}

		if !isSorted(currentNode.Path) {
			sorted = append(sorted, currentNode)
		}

		return nil
	}

	// Start DFS from main file
	mainModule := Module{Path: mainFile}
	if err := dfs(mainModule); err != nil {
		return nil, err
	}

	// Filter out modules that don't exist locally (content is nil)
	var result []Module
	for _, mod := range sorted {
		if mod.Content != nil {
			result = append(result, mod)
		}
	}

	return result, nil
}

// exploreNodes finds child dependencies for a given module
func exploreNodes(node Module, cwd string) ([]Module, error) {
	// Check if file exists
	if _, err := os.Stat(node.Path); os.IsNotExist(err) {
		return []Module{}, nil
	}

	// Read file content
	content, err := os.ReadFile(node.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", node.Path, err)
	}

	contentStr := string(content)
	// Note: We don't modify the input node here, the content will be set in the DFS

	// Find require statements using regex
	requirePattern := regexp.MustCompile(`(?:require\s*\(\s*["'])([^"']+)(?:["']\s*\))`)
	matches := requirePattern.FindAllStringSubmatch(contentStr, -1)

	var requiredModules []Module
	for _, match := range matches {
		if len(match) > 1 {
			moduleName := match[1]
			// Convert dot notation to file path
			modulePath := filepath.Join(cwd, strings.ReplaceAll(moduleName, ".", string(filepath.Separator))+".lua")

			requiredModules = append(requiredModules, Module{
				Name:    moduleName,
				Path:    modulePath,
				Content: nil, // Will be set when the node is explored
			})
		}
	}

	return requiredModules, nil
}

// getModFnName converts a module name to a valid function name
func getModFnName(name string) string {
	// Replace dots with underscores and remove leading underscore
	result := strings.ReplaceAll(name, ".", "_")
	if strings.HasPrefix(result, "_") {
		result = result[1:]
	}
	return result
}
