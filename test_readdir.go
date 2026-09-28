package main
import (
	"fmt"
	"os"
)
func main() {
	entries, err := os.ReadDir("/")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Count:", len(entries))
}
