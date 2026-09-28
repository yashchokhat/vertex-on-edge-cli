package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	successMark = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render("✓")
	failMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("✕")

	accentColor    = lipgloss.AdaptiveColor{Light: "#0077B6", Dark: "#00B4D8"}
	dimColor       = lipgloss.AdaptiveColor{Light: "#AAAAAA", Dark: "#555555"}
	textColor      = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#CCCCCC"}
	brightColor    = lipgloss.AdaptiveColor{Light: "#111111", Dark: "#FFFFFF"}
	progressColors = []string{"#003049", "#005F73", "#0077B6", "#0096C7", "#00B4D8", "#48CAE4", "#90E0EF"}

	spinnerAccent = lipgloss.NewStyle().Foreground(accentColor)
	dimStyle      = lipgloss.NewStyle().Foreground(dimColor)
	textStyle     = lipgloss.NewStyle().Foreground(textColor)
	brightStyle   = lipgloss.NewStyle().Foreground(brightColor).Bold(true)
)

// RunStartupAnimation provides a minimal, clean startup.
func RunStartupAnimation() {
	ClearScreen()
}

func getVersion() string {
	return "0.1.0"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
