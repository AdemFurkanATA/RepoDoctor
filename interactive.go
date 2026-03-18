package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// InteractiveMode provides an interactive CLI experience
type InteractiveMode struct {
	reader *bufio.Reader
}

// NewInteractiveMode creates a new interactive mode instance
func NewInteractiveMode() *InteractiveMode {
	return &InteractiveMode{
		reader: bufio.NewReader(os.Stdin),
	}
}

// Run starts the interactive mode session
func (i *InteractiveMode) Run() {
	fmt.Println("RepoDoctor Interactive Mode")
	fmt.Println(strings.Repeat("═", 50))
	fmt.Println()

	for {
		i.showMainMenu()

		choice := i.readChoice()

		switch choice {
		case 1:
			i.analyzeMenu()
		case 2:
			i.viewHistory()
		case 3:
			i.configureRules()
		case 4:
			fmt.Println("\nExiting RepoDoctor Interactive Mode...")
			return
		default:
			fmt.Println("\nInvalid choice. Please try again.")
		}
	}
}

// showMainMenu displays the main menu
func (i *InteractiveMode) showMainMenu() {
	fmt.Println("Select action:")
	fmt.Println("  1. Analyze repository")
	fmt.Println("  2. View analysis history")
	fmt.Println("  3. Configure rules")
	fmt.Println("  4. Exit")
	fmt.Print("\n> ")
}

// readChoice reads and validates user choice
func (i *InteractiveMode) readChoice() int {
	input, err := i.reader.ReadString('\n')
	if err != nil {
		return -1
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		return -1
	}

	return choice
}

// analyzeMenu handles the analyze submenu
func (i *InteractiveMode) analyzeMenu() {
	fmt.Println("\nSelect analysis scope:")
	fmt.Println("  1. Current repository")
	fmt.Println("  2. Custom path")
	fmt.Println("  3. Back to main menu")
	fmt.Print("\n> ")

	choice := i.readChoice()

	switch choice {
	case 1:
		fmt.Println("\nAnalyzing current repository...")
		runAnalyze(".", "text", false, true, true)
	case 2:
		fmt.Print("\nEnter path to analyze: ")
		path, err := i.reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading path")
			return
		}
		path = strings.TrimSpace(path)
		if path == "" {
			fmt.Println("Path cannot be empty")
			return
		}
		fmt.Printf("\nAnalyzing repository: %s\n", path)
		runAnalyze(path, "text", false, true, true)
	case 3:
		return
	default:
		fmt.Println("Invalid choice")
	}
}

// viewHistory displays analysis history
func (i *InteractiveMode) viewHistory() {
	fmt.Println("\nViewing analysis history...")
	runHistory(".")
}

// configureRules handles rule configuration
func (i *InteractiveMode) configureRules() {
	absPath, err := filepath.Abs(".")
	if err != nil {
		fmt.Printf("\nError resolving repository path: %v\n\n", err)
		return
	}

	loader := NewConfigLoader(GetConfigPath(absPath))
	config, err := loader.Load()
	if err != nil {
		fmt.Printf("\nError loading configuration: %v\n\n", err)
		return
	}

	for {
		i.showConfigMenu(config)
		choice := i.readChoice()

		switch choice {
		case 1:
			i.toggleSizeRule(config)
		case 2:
			i.toggleGodObjectRule(config)
		case 3:
			i.setMaxFileLines(config)
		case 4:
			i.setMaxFunctionLines(config)
		case 5:
			if err := saveConfig(absPath, config); err != nil {
				fmt.Printf("\nError saving configuration: %v\n\n", err)
			} else {
				fmt.Print("\nConfiguration saved successfully.\n\n")
			}
		case 6:
			return
		default:
			fmt.Print("\nInvalid choice. Please enter a number between 1 and 6.\n\n")
		}
	}
}

func (i *InteractiveMode) showConfigMenu(config *Config) {
	fmt.Println("\nRule Configuration")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("Current settings:\n")
	fmt.Printf("  Size Rule: %s\n", boolLabel(*config.Rules.EnableSizeRule))
	fmt.Printf("  God Object Rule: %s\n", boolLabel(*config.Rules.EnableGodObjectRule))
	fmt.Printf("  Max File Lines: %d\n", config.Size.MaxFileLines)
	fmt.Printf("  Max Function Lines: %d\n", config.Size.MaxFunctionLines)
	fmt.Println()
	fmt.Println("  1. Toggle Size Rule")
	fmt.Println("  2. Toggle God Object Rule")
	fmt.Println("  3. Set Max File Lines")
	fmt.Println("  4. Set Max Function Lines")
	fmt.Println("  5. Save Configuration")
	fmt.Println("  6. Back to main menu")
	fmt.Print("\n> ")
}

func (i *InteractiveMode) toggleSizeRule(config *Config) {
	current := *config.Rules.EnableSizeRule
	next := !current
	config.Rules.EnableSizeRule = &next
	if config.Size != nil {
		config.Size.Enabled = &next
	}
	fmt.Printf("\nSize Rule is now %s.\n\n", boolLabel(next))
}

func (i *InteractiveMode) toggleGodObjectRule(config *Config) {
	current := *config.Rules.EnableGodObjectRule
	next := !current
	config.Rules.EnableGodObjectRule = &next
	if config.GodObject != nil {
		config.GodObject.Enabled = &next
	}
	fmt.Printf("\nGod Object Rule is now %s.\n\n", boolLabel(next))
}

func (i *InteractiveMode) setMaxFileLines(config *Config) {
	value, ok := i.readPositiveInt("Enter max file lines")
	if !ok {
		fmt.Print("\nInvalid value. Please enter a positive number.\n\n")
		return
	}
	config.Size.MaxFileLines = value
	fmt.Printf("\nMax file lines set to %d.\n\n", value)
}

func (i *InteractiveMode) setMaxFunctionLines(config *Config) {
	value, ok := i.readPositiveInt("Enter max function lines")
	if !ok {
		fmt.Print("\nInvalid value. Please enter a positive number.\n\n")
		return
	}
	config.Size.MaxFunctionLines = value
	fmt.Printf("\nMax function lines set to %d.\n\n", value)
}

func (i *InteractiveMode) readPositiveInt(prompt string) (int, bool) {
	input := i.readString(prompt + ": ")
	value, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}

func saveConfig(repoPath string, cfg *Config) error {
	if err := EnsureConfigDir(repoPath); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(GetConfigPath(repoPath), data, 0644)
}

func boolLabel(value bool) string {
	if value {
		return "Enabled"
	}
	return "Disabled"
}

// readString reads a string input from user
func (i *InteractiveMode) readString(prompt string) string {
	fmt.Print(prompt)
	input, err := i.reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(input)
}

// confirm asks for yes/no confirmation
func (i *InteractiveMode) confirm(prompt string) bool {
	response := i.readString(prompt + " (y/n): ")
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

// runInteractive starts the interactive mode
func runInteractive() {
	interactive := NewInteractiveMode()
	interactive.Run()
}
