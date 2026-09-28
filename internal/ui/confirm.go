package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ConfirmAction asks the user a yes/no question.
func ConfirmAction(question string) bool {
	accentStyle := lipgloss.NewStyle().Bold(true).Foreground(accentColor)
	dimStyle := lipgloss.NewStyle().Foreground(dimColor)

	fmt.Printf("\n    %s  %s ", accentStyle.Render(question), dimStyle.Render("[y/N]"))
	ShowCursor()
	defer HideCursor()

	input := strings.ToLower(readLine())
	if input == "y" || input == "yes" {
		return true
	}
	return false
}
