package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yashchokhat/vertex-on-edge/internal/config"
	"github.com/yashchokhat/vertex-on-edge/pkg/models"
)

func shortenPath(p string) string {
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(p, home) {
		return strings.Replace(p, home, "~", 1)
	}
	return p
}

// PrintDetectionResult prints the detection summary inside a styled card.
func PrintDetectionResult(info *models.ProjectInfo) {
	if !info.HasDetection() {
		PrintNoDetection(info)
		return
	}

	prim := info.Primary()

	w := TerminalWidth()
	cardWidth := 52
	if cardWidth > w-4 {
		cardWidth = w - 4
	}

	borderColor := lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#333333"}
	lineStyle := lipgloss.NewStyle().Foreground(borderColor)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(accentColor)
	labelStyle := lipgloss.NewStyle().Foreground(dimColor)
	valueStyle := lipgloss.NewStyle().Foreground(brightColor).Bold(true)
	treeStyle := lipgloss.NewStyle().Foreground(borderColor)
	evidenceStyle := lipgloss.NewStyle().Foreground(textColor)

	// Section: Project
	fmt.Println("  " + lineStyle.Render(strings.Repeat("─", cardWidth)))
	fmt.Println()
	fmt.Println("    " + headerStyle.Render("Project"))
	fmt.Println()
	fmt.Printf("    %-14s %s\n", labelStyle.Render("Path"), valueStyle.Render(shortenPath(info.Path)))
	fmt.Printf("    %-14s %s\n", labelStyle.Render("Name"), valueStyle.Render(info.Name))
	fmt.Println()

	// Section: Detected Stack
	fmt.Println("    " + headerStyle.Render("Detected Stack"))
	fmt.Println()

	type field struct{ label, value string }
	var fields []field

	if prim.Language != "" {
		fields = append(fields, field{"Language", prim.Language})
	}
	if prim.Framework != "" {
		fields = append(fields, field{"Framework", prim.Framework})
	}
	if prim.Runtime != "" {
		fields = append(fields, field{"Runtime", prim.Runtime})
	}
	if prim.PackageManager != "" {
		fields = append(fields, field{"Package Mgr", prim.PackageManager})
	}
	fields = append(fields, field{"Confidence", string(prim.Confidence)})

	for i, f := range fields {
		prefix := "├─"
		if i == len(fields)-1 {
			prefix = "└─"
		}
		// Highlight confidence with color
		val := valueStyle.Render(f.value)
		if f.label == "Confidence" {
			switch models.Confidence(f.value) {
			case models.ConfidenceHigh:
				val = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render(f.value)
			case models.ConfidenceMedium:
				val = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render(f.value)
			case models.ConfidenceLow:
				val = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render(f.value)
			}
		}
		fmt.Printf("    %s %-14s %s\n",
			treeStyle.Render(prefix),
			labelStyle.Render(f.label),
			val,
		)
	}

	// Section: Evidence
	if len(prim.Evidence) > 0 {
		fmt.Println()
		fmt.Println("    " + headerStyle.Render("Evidence"))
		fmt.Println()
		for i, ev := range prim.Evidence {
			prefix := "├─"
			if i == len(prim.Evidence)-1 {
				prefix = "└─"
			}
			fmt.Printf("    %s %s\n", treeStyle.Render(prefix), evidenceStyle.Render(ev))
		}
	}

	// Additional stacks (multi-component)
	if len(info.Stacks) > 1 {
		fmt.Println()
		fmt.Println("    " + headerStyle.Render("Additional Components"))
		fmt.Println()
		for i := 0; i < len(info.Stacks); i++ {
			if i == info.PrimaryIdx {
				continue
			}
			s := info.Stacks[i]
			name := s.Language
			if s.Framework != "" {
				name = s.Framework
			}
			prefix := "├─"
			remaining := 0
			for j := i + 1; j < len(info.Stacks); j++ {
				if j != info.PrimaryIdx {
					remaining++
				}
			}
			if remaining == 0 {
				prefix = "└─"
			}
			fmt.Printf("    %s %s\n", treeStyle.Render(prefix), evidenceStyle.Render(name))
		}
	}

	fmt.Println()
	fmt.Println("  " + lineStyle.Render(strings.Repeat("─", cardWidth)))
	fmt.Println()
	fmt.Printf("    %s  %s\n\n", successMark, textStyle.Render("Detection complete."))
}

// PrintNoDetection prints when no stack could be identified.
func PrintNoDetection(info *models.ProjectInfo) {
	w := TerminalWidth()
	cardWidth := 52
	if cardWidth > w-4 {
		cardWidth = w - 4
	}

	borderColor := lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#333333"}
	lineStyle := lipgloss.NewStyle().Foreground(borderColor)
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(dimColor)
	valueStyle := lipgloss.NewStyle().Foreground(textColor)
	treeStyle := lipgloss.NewStyle().Foreground(borderColor)

	fmt.Println()
	fmt.Println("  " + lineStyle.Render(strings.Repeat("─", cardWidth)))
	fmt.Println()
	fmt.Printf("    %s  %s\n\n",
		warnStyle.Render("!"),
		textStyle.Render("Unable to confidently identify the project stack."),
	)

	// Show top-level files as hints
	fmt.Println("    " + labelStyle.Render("Top-level files found:"))
	fmt.Println()

	// Try to list some files from the project path
	topFiles := listTopFiles(info.Path)
	if len(topFiles) > 0 {
		for i, f := range topFiles {
			prefix := "├─"
			if i == len(topFiles)-1 {
				prefix = "└─"
			}
			fmt.Printf("    %s %s\n", treeStyle.Render(prefix), valueStyle.Render(f))
		}
	} else {
		fmt.Printf("    %s %s\n", treeStyle.Render("└─"), valueStyle.Render("(empty directory)"))
	}

	fmt.Println()
	fmt.Println("    " + labelStyle.Render("Vertex-on-Edge currently supports:"))
	fmt.Println()

	stacks := config.SupportedStacks()
	for i, s := range stacks {
		prefix := "├─"
		if i == len(stacks)-1 {
			prefix = "└─"
		}
		fmt.Printf("    %s %s\n", treeStyle.Render(prefix), valueStyle.Render(s))
	}

	fmt.Println()
	fmt.Println("  " + lineStyle.Render(strings.Repeat("─", cardWidth)))
	fmt.Println()
}

// listTopFiles returns up to 8 top-level file names from a directory.
func listTopFiles(path string) []string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if e.IsDir() {
			name += "/"
		}
		names = append(names, name)
		if len(names) >= 8 {
			break
		}
	}
	return names
}

// PrintError prints a formatted error message.
func PrintError(message string, reason string) {
	fmt.Printf("    %s  %s\n", failMark, textStyle.Render(message))
	if reason != "" {
		errDetailStyle := lipgloss.NewStyle().Foreground(dimColor)
		fmt.Printf("       %s\n", errDetailStyle.Render(reason))
	}
	fmt.Println()
}

// PrintSuccess prints a success message.
func PrintSuccess(message string) {
	fmt.Printf("    %s  %s\n", successMark, textStyle.Render(message))
}

// PrintInfo prints an informational message.
func PrintInfo(message string) {
	dot := lipgloss.NewStyle().Foreground(accentColor).Render("●")
	fmt.Printf("    %s  %s\n", dot, textStyle.Render(message))
}
