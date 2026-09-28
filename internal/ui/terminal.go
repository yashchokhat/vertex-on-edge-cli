package ui

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func readLine() string {
	var b [1]byte
	var result []byte
	for {
		n, err := os.Stdin.Read(b[:])
		if n > 0 {
			if b[0] == '\n' {
				break
			}
			if b[0] != '\r' {
				result = append(result, b[0])
			}
		}
		if err != nil {
			break
		}
	}
	return strings.TrimSpace(string(result))
}

// ClearScreen clears the terminal and moves the cursor to the top-left.
func ClearScreen() {
	fmt.Fprint(os.Stdout, "\033[2J\033[H")
}

// HideCursor hides the terminal cursor.
func HideCursor() {
	fmt.Fprint(os.Stdout, "\033[?25l")
}

// ShowCursor shows the terminal cursor.
func ShowCursor() {
	fmt.Fprint(os.Stdout, "\033[?25h")
}

// MoveCursor moves the cursor to the specified row and column (1-indexed).
func MoveCursor(row, col int) {
	fmt.Fprintf(os.Stdout, "\033[%d;%dH", row, col)
}

// ClearLine clears the current terminal line.
func ClearLine() {
	fmt.Fprint(os.Stdout, "\033[2K\r")
}

// TerminalWidth returns the width of the terminal, or a default of 80.
func TerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// TerminalHeight returns the height of the terminal, or a default of 24.
func TerminalHeight() int {
	_, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || h <= 0 {
		return 24
	}
	return h
}

// PrintEmptyLines prints n empty lines.
func PrintEmptyLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Println()
	}
}
