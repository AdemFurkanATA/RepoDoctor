package main

import "fmt"

type InteractiveSession struct {
	mode *InteractiveMode
}

func NewInteractiveSession(mode *InteractiveMode) *InteractiveSession {
	return &InteractiveSession{mode: mode}
}

func (s *InteractiveSession) Run() {
	for {
		s.mode.showMainMenu()

		choice := s.mode.io.readChoice()

		switch choice {
		case 1:
			s.mode.analyzeMenu()
		case 2:
			s.mode.viewHistory()
		case 3:
			s.mode.configController.configureRules()
		case 4:
			fmt.Println("\nExiting RepoDoctor Interactive Mode...")
			return
		default:
			fmt.Println("\nInvalid choice. Please try again.")
		}
	}
}
