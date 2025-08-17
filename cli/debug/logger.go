package debug

import (
	"fmt"
	"os"
	"strings"
)

var (
	// DebugEnabled controls whether debug messages are printed
	DebugEnabled = false
)

func init() {
	// Enable debug mode via environment variable
	if strings.ToLower(os.Getenv("HARLEQUIN_DEBUG")) == "true" {
		DebugEnabled = true
	}
}

// Printf prints a debug message if debug mode is enabled
func Printf(format string, args ...interface{}) {
	if DebugEnabled {
		fmt.Printf("[DEBUG] "+format, args...)
	}
}

// Println prints a debug message with newline if debug mode is enabled
func Println(args ...interface{}) {
	if DebugEnabled {
		fmt.Print("[DEBUG] ")
		fmt.Println(args...)
	}
}

// Info prints an informational message (always shown)
func Info(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Infof is an alias for Info for consistency
func Infof(format string, args ...interface{}) {
	Info(format, args...)
}

// SetEnabled allows programmatic control of debug mode
func SetEnabled(enabled bool) {
	DebugEnabled = enabled
}
