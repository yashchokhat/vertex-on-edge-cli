package ui

import (
	"os"

	"github.com/charmbracelet/huh"
)

// PromptExistingConfig asks the user what to do if a config already exists
func PromptExistingConfig() string {
	var selected string

	err := huh.NewSelect[string]().
		Title("Existing configuration detected. What would you like to do?").
		Options(
			huh.NewOption("Fast Deploy (Use existing configuration)", "deploy"),
			huh.NewOption("Edit Configuration (Re-run setup wizard)", "edit"),
		).
		Value(&selected).
		Run()

	if err == huh.ErrUserAborted {
		os.Exit(0)
	}

	return selected
}
