package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	// DebugEnabled controls whether debug messages are printed
	DebugEnabled = false
	// logFile holds the debug log file handle
	logFile *os.File
	// LogFilePath holds the absolute path to the debug log file
	LogFilePath string
)

func init() {
	// Enable debug mode via environment variable
	if strings.ToLower(os.Getenv("HARLEQUIN_DEBUG")) == "true" {
		DebugEnabled = true
	}

	// Initialize log file
	initLogFile()
}

// initLogFile creates or opens the debug log file
func initLogFile() {
	// Determine log file location - try user home directory first
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home dir not available
		LogFilePath = filepath.Join(".", "harlequin-debug.log")
	} else {
		// Create .harlequin directory in user home
		harlequinDir := filepath.Join(homeDir, ".harlequin")
		os.MkdirAll(harlequinDir, 0755)
		LogFilePath = filepath.Join(harlequinDir, "harlequin-debug.log")
	}

	// Create/open log file
	var openErr error
	logFile, openErr = os.OpenFile(LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if openErr != nil {
		// If we can't create the log file, disable file logging
		logFile = nil
	}
}

// logToFile writes a message to the log file with timestamp
func logToFile(level, message string) {
	if logFile != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(logFile, "[%s] [%s] %s\n", timestamp, level, message)
		logFile.Sync() // Ensure it's written immediately
	}
}

// Printf prints a debug message if debug mode is enabled
func Printf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	// Always log to file
	logToFile("DEBUG", message)

	// Only print to console if debug enabled
	if DebugEnabled {
		fmt.Printf("[DEBUG] " + message)
	}
}

// Println prints a debug message with newline if debug mode is enabled
func Println(args ...interface{}) {
	message := fmt.Sprint(args...)

	// Always log to file
	logToFile("DEBUG", message)

	// Only print to console if debug enabled
	if DebugEnabled {
		fmt.Print("[DEBUG] ")
		fmt.Println(args...)
	}
}

// Info prints an informational message (always shown)
func Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	// Always log to file
	logToFile("INFO", message)

	// Always print to console
	fmt.Printf(message)
}

// Infof is an alias for Info for consistency
func Infof(format string, args ...interface{}) {
	Info(format, args...)
}

// Error logs an error message to both console and file
func Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	// Always log to file
	logToFile("ERROR", message)

	// Always print to console
	fmt.Printf("❌ Error: " + message)
}

// Fatal logs a fatal error and provides log file location
func Fatal(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	// Always log to file
	logToFile("FATAL", message)

	// Always print to console with additional context
	fmt.Printf("❌ Fatal Error: %s\n", message)
	if LogFilePath != "" {
		absPath, _ := filepath.Abs(LogFilePath)
		fmt.Printf("📄 Detailed error logs available at: %s\n", absPath)
	}
}

// SetEnabled allows programmatic control of debug mode
func SetEnabled(enabled bool) {
	DebugEnabled = enabled
}

// Close closes the log file (should be called on program exit)
func Close() {
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
}
