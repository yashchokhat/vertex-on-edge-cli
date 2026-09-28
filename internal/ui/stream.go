package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// StreamPrint smoothly prints long text line-by-line so the user can read it as it appears.
func StreamPrint(text string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		fmt.Println(line)
		time.Sleep(15 * time.Millisecond) // Smooth animation
	}
}

// PrintDim streams text in a dim color
func PrintDim(text string) {
	lines := strings.Split(text, "\n")
	style := lipgloss.NewStyle().Foreground(dimColor)
	for _, line := range lines {
		fmt.Println(style.Render(line))
		time.Sleep(10 * time.Millisecond)
	}
}
