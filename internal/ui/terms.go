package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yashchokhat/vertex-on-edge/internal/config"
)

// Terms and conditions text.
var termsText = []string{
	"By using Vertex-on-Edge, you agree to our terms.",
	"We generate infrastructure code, but you are responsible",
	"for deployments and cloud provider costs.",
	"",
	"Full terms: github.com/yashchokhat/vertex-on-edge/blob/main/TERMS.md",
}

// PromptTermsAcceptance displays the terms and conditions and asks the user
// to accept before proceeding. Returns true if accepted, false otherwise.
func PromptTermsAcceptance() bool {
	w := TerminalWidth()
	if w > 80 {
		w = 80
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#0077B6", Dark: "#00B4D8"})

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#888888", Dark: "#666666"})

	bodyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#CCCCCC"})

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#CCCCCC", Dark: "#444444"}).
		Padding(1, 2).
		Width(w)

	warnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))

	// Build terms content
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Terms & Conditions"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  " + config.AppName + " v" + config.Version))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  " + strings.Repeat("─", w-8)))
	b.WriteString("\n\n")

	for _, line := range termsText {
		b.WriteString(bodyStyle.Render("  " + line))
		b.WriteString("\n")
	}

	centerStyle := lipgloss.NewStyle().Width(TerminalWidth()).Align(lipgloss.Center)
	fmt.Println(centerStyle.Render(boxStyle.Render(b.String())))
	fmt.Println()

	// Prompt
	promptText := fmt.Sprintf("%s  %s ",
		lipgloss.NewStyle().Bold(true).Foreground(accentColor).Render("Do you accept the terms and conditions?"),
		dimStyle.Render("[Y/n]"),
	)

	padLen := (TerminalWidth() - lipgloss.Width(promptText)) / 2
	if padLen < 0 {
		padLen = 0
	}
	fmt.Print(strings.Repeat(" ", padLen) + promptText)

	ShowCursor()
	input := strings.ToLower(readLine())
	HideCursor()

	if input == "" || input == "y" || input == "yes" {
		fmt.Printf("\r\033[K%s\n\n",
			centerStyle.Render(fmt.Sprintf("%s  %s", successMark, bodyStyle.Render("Terms accepted. Welcome aboard."))),
		)
		return true
	}

	fmt.Printf("\r\033[K%s\n\n",
		centerStyle.Render(fmt.Sprintf("%s  %s", warnStyle.Render("!"), bodyStyle.Render("Terms declined. Exiting."))),
	)
	return false
}
