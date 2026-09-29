package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/cli/deploy.go")
	s := string(b)
	
	oldBlock := `	if roleArn != "" {
		secrets["AWS_ROLE_ARN"] = strings.TrimSpace(roleArn)
	}`
	
	s = strings.Replace(s, oldBlock, "", 1)
	os.WriteFile("internal/cli/deploy.go", []byte(s), 0644)
}
