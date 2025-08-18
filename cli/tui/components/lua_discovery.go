package components

import (
	"os"
	"path/filepath"
	"strings"
)

// LuaFileDiscovery provides utilities for finding Lua files
type LuaFileDiscovery struct {
	skipDirs []string
	maxDepth int
}

// NewLuaFileDiscovery creates a new Lua file discovery utility
func NewLuaFileDiscovery() *LuaFileDiscovery {
	return &LuaFileDiscovery{
		skipDirs: []string{
			"node_modules",
			".git", 
			".svn",
			".hg",
			"dist",
			"build",
			"target",
			"vendor",
			".vscode",
			".idea",
			"__pycache__",
			".DS_Store",
		},
		maxDepth: 5, // Prevent deep recursion
	}
}

// FindLuaFiles recursively finds all .lua files in a directory
func (lfd *LuaFileDiscovery) FindLuaFiles(rootDir string) ([]string, error) {
	var luaFiles []string
	
	err := lfd.walkDirectory(rootDir, rootDir, 0, &luaFiles)
	if err != nil {
		return nil, err
	}
	
	return luaFiles, nil
}

// walkDirectory recursively walks through directories finding Lua files
func (lfd *LuaFileDiscovery) walkDirectory(rootDir, currentDir string, depth int, luaFiles *[]string) error {
	// Prevent excessive recursion
	if depth > lfd.maxDepth {
		return nil
	}
	
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		fullPath := filepath.Join(currentDir, entry.Name())
		
		if entry.IsDir() {
			// Skip directories we don't want to search
			if lfd.shouldSkipDir(entry.Name()) {
				continue
			}
			
			// Recursively search subdirectory
			if err := lfd.walkDirectory(rootDir, fullPath, depth+1, luaFiles); err != nil {
				// Don't fail the entire search if one directory has issues
				continue
			}
		} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".lua") {
			// Found a Lua file - convert to relative path
			relPath, err := filepath.Rel(rootDir, fullPath)
			if err != nil {
				relPath = fullPath // Fallback to absolute path
			}
			*luaFiles = append(*luaFiles, relPath)
		}
	}
	
	return nil
}

// shouldSkipDir checks if a directory should be skipped during search
func (lfd *LuaFileDiscovery) shouldSkipDir(dirName string) bool {
	for _, skipDir := range lfd.skipDirs {
		if dirName == skipDir {
			return true
		}
	}
	return false
}

// AddSkipDir adds a directory to the skip list
func (lfd *LuaFileDiscovery) AddSkipDir(dirName string) {
	lfd.skipDirs = append(lfd.skipDirs, dirName)
}

// SetMaxDepth sets the maximum recursion depth
func (lfd *LuaFileDiscovery) SetMaxDepth(depth int) {
	lfd.maxDepth = depth
}

// FindLuaFilesQuick is a convenience function for quick discovery
func FindLuaFilesQuick(rootDir string) ([]string, error) {
	discovery := NewLuaFileDiscovery()
	return discovery.FindLuaFiles(rootDir)
}
