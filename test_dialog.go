package main
import (
	"fmt"
	"github.com/sqweek/dialog"
)
func main() {
	fmt.Println("Attempting to open dialog...")
	dir, err := dialog.Directory().Title("Select Project Directory").Browse()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Selected:", dir)
}
