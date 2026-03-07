package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
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
		runAnalyze(".", "text", false)
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
		runAnalyze(path, "text", false)
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
	fmt.Println("\nRule Configuration")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println("Rule configuration is not yet available in interactive mode.")
	fmt.Println("Please use command-line flags to configure rules.")
	fmt.Println()
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
