package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

func main() {
	var selected string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewFilePicker().
				Title("Select Directory").
				DirAllowed(true).
				FileAllowed(false).
				Height(10).
				Value(&selected),
		),
	).Run()

	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Selected:", selected)
	}
}
