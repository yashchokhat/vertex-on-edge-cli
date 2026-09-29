package main

import (
	"os"

	"github.com/yashchokhat/vertex-on-edge/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
