package ui

import (
	"os"

	"github.com/charmbracelet/huh"
)

// PromptTestingFramework asks the user which testing frameworks they want to configure
func PromptTestingFramework(language, framework string) []string {
	var selected []string

	var options []huh.Option[string]

	switch language {
	case "TypeScript", "JavaScript":
		options = append(options, huh.NewOption("Jest", "jest"))
		options = append(options, huh.NewOption("Vitest", "vitest"))
		if framework == "Next.js" || framework == "React" {
			options = append(options, huh.NewOption("Cypress (E2E)", "cypress"))
			options = append(options, huh.NewOption("Playwright (E2E)", "playwright"))
		}
	case "Python":
		options = append(options, huh.NewOption("pytest", "pytest"))
		options = append(options, huh.NewOption("unittest", "unittest"))
	case "Go":
		options = append(options, huh.NewOption("go test", "go_test"))
	default:
		options = append(options, huh.NewOption("Default Language Test Runner", "default"))
	}

	err := huh.NewMultiSelect[string]().
		Title("Select testing frameworks (Space to select, Enter to confirm)").
		Options(options...).
		Value(&selected).
		Run()

	if err == huh.ErrUserAborted {
		os.Exit(0)
	}

	return selected
}
