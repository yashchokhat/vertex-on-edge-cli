package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yashchokhat/vertex-on-edge/internal/config"
)

// ASCII art for Vertex-on-Edge — compact, original design.
var brandArt = []string{
	`  _    __          __                             ______    __          `,
	` | |  / /__  _____/ /____  _  __      ____  ____ / ____/___/ /____  ___ `,
	` | | / / _ \/ ___/ __/ _ \| |/_/_____/ __ \/ __ \/ __/ / __  / __ \/ _ \`,
	` | |/ /  __/ /  / /_/  __/>  </_____/ /_/ / / / / /___/ /_/ / /_/ /  __/`,
	` |___/\___/_/   \__/\___/_/|_|      \____/_/ /_/_____/\__,_/\__, /\___/ `,
	`                                                           /____/       `,
}

// PrintBanner displays the premium brand banner.
func PrintBanner() {
	w := TerminalWidth()
	centerStyle := lipgloss.NewStyle().Width(w).Align(lipgloss.Center)

	// Styles
	artStyle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true)

	tagStyle := lipgloss.NewStyle().
		Foreground(dimColor).
		Italic(true)

	borderColor := lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#333333"}
	lineStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Build the banner block
	sepWidth := 60
	if sepWidth > w-4 {
		sepWidth = w - 4
	}

	fmt.Println()
	fmt.Println(centerStyle.Render(lineStyle.Render(strings.Repeat("─", sepWidth))))
	fmt.Println()

	// Render ASCII art centered
	for _, line := range brandArt {
		fmt.Println(centerStyle.Render(artStyle.Render(line)))
	}

	fmt.Println()
	fmt.Println(centerStyle.Render(tagStyle.Render(config.AppTagline)))
	fmt.Println()
	fmt.Println(centerStyle.Render(lineStyle.Render(strings.Repeat("─", sepWidth))))
	fmt.Println()
}

// PrintWelcome displays the welcome section with GitHub link.
func PrintWelcome() {
	w := TerminalWidth()
	centerStyle := lipgloss.NewStyle().Width(w).Align(lipgloss.Center)

	urlLabelStyle := lipgloss.NewStyle().
		Foreground(dimColor)

	urlValueStyle := lipgloss.NewStyle().
		Foreground(accentColor).
		Underline(true)

	fmt.Println(centerStyle.Render(fmt.Sprintf("%s  %s", textStyle.Render("Welcome to"), brightStyle.Render(config.AppName))))
	fmt.Println(centerStyle.Render(textStyle.Render("Ship your application to your own infrastructure.")))
	fmt.Println()
	fmt.Println(centerStyle.Render(fmt.Sprintf("%s  %s", urlLabelStyle.Render("GitHub  →"), urlValueStyle.Render(config.GitHubURL))))
	fmt.Println()
}

// PrintSeparator prints a full-width separator line.
func PrintSeparator() {
	w := TerminalWidth()
	sepWidth := 52
	if sepWidth > w-4 {
		sepWidth = w - 4
	}
	borderColor := lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#333333"}
	lineStyle := lipgloss.NewStyle().Foreground(borderColor)
	fmt.Println("  " + lineStyle.Render(strings.Repeat("─", sepWidth)))
	fmt.Println()
}
