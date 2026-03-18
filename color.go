package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Color codes for ANSI terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// ColorFormatter provides colored output formatting
type ColorFormatter struct {
	enabled bool
}

// NewColorFormatter creates a new color formatter
func NewColorFormatter(enabled bool) *ColorFormatter {
	return &ColorFormatter{
		enabled: enabled && isTerminal(),
	}
}

// isTerminal checks if the output is a terminal
func isTerminal() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if term := strings.ToLower(os.Getenv("TERM")); term == "dumb" {
		return false
	}

	if runtime.GOOS == "windows" {
		if os.Getenv("WT_SESSION") != "" || os.Getenv("ANSICON") != "" || os.Getenv("ConEmuANSI") == "ON" {
			return true
		}
	}

	if fileInfo, err := os.Stdout.Stat(); err == nil {
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	}

	return false
}

// Color applies a color to text if colors are enabled
func (f *ColorFormatter) Color(text string, colorCode string) string {
	if !f.enabled {
		return text
	}
	return colorCode + text + ColorReset
}

// Info formats an info message in blue
func (f *ColorFormatter) Info(message string) string {
	return f.Color(message, ColorBlue)
}

// Warn formats a warning message in yellow
func (f *ColorFormatter) Warn(message string) string {
	return f.Color(message, ColorYellow)
}

// Error formats an error message in red
func (f *ColorFormatter) Error(message string) string {
	return f.Color(message, ColorRed)
}

// Success formats a success message in green
func (f *ColorFormatter) Success(message string) string {
	return f.Color(message, ColorGreen)
}

// Bold makes text bold
func (f *ColorFormatter) Bold(text string) string {
	if !f.enabled {
		return text
	}
	return "\033[1m" + text + ColorReset
}

// FormatMessage formats a message with severity-based coloring
func (f *ColorFormatter) FormatMessage(severity, message string) string {
	var colored string
	switch strings.ToUpper(severity) {
	case "INFO":
		colored = f.Info(fmt.Sprintf("INFO %s", message))
	case "WARN", "WARNING":
		colored = f.Warn(fmt.Sprintf("WARN %s", message))
	case "ERROR":
		colored = f.Error(fmt.Sprintf("ERROR %s", message))
	case "SUCCESS":
		colored = f.Success(fmt.Sprintf("SUCCESS %s", message))
	default:
		colored = message
	}
	return colored
}

// global formatter instance
var globalColorFormatter *ColorFormatter

// InitColorFormatter initializes the global color formatter
func InitColorFormatter(enabled bool) {
	globalColorFormatter = NewColorFormatter(enabled)
}

// GetColorFormatter returns the global color formatter
func GetColorFormatter() *ColorFormatter {
	if globalColorFormatter == nil {
		globalColorFormatter = NewColorFormatter(true)
	}
	return globalColorFormatter
}

// Colorize applies color to text using the global formatter
func Colorize(text string, colorCode string) string {
	return GetColorFormatter().Color(text, colorCode)
}

// ColorInfo formats info message using global formatter
func ColorInfo(message string) string {
	return GetColorFormatter().Info(message)
}

// ColorWarn formats warning message using global formatter
func ColorWarn(message string) string {
	return GetColorFormatter().Warn(message)
}

// ColorError formats error message using global formatter
func ColorError(message string) string {
	return GetColorFormatter().Error(message)
}

// ColorSuccess formats success message using global formatter
func ColorSuccess(message string) string {
	return GetColorFormatter().Success(message)
}

// Printf prints a colored message based on severity
func ColorPrintf(severity, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Println(GetColorFormatter().FormatMessage(severity, message))
}

// Fprintf writes a colored message to the specified writer
func ColorFprintf(writer *os.File, severity, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	colored := GetColorFormatter().FormatMessage(severity, message)
	if severity == "ERROR" {
		fmt.Fprintln(writer, colored)
	} else {
		fmt.Fprintln(writer, colored)
	}
}
