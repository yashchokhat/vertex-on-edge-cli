package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/terraform/aws/ec2/main.tf")
	s := string(b)
	
	oldList := `thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1", "1c58a3a8518e8759bf075b76b750d4f2df264fcd"]`
	newList := `thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1", "1c58a3a8518e8759bf075b76b750d4f2df264fcd", "1b511abead59c6ce207077c0bf0e0043b1382612", "06d927fecd0a84aeba28aad1d808139470fe95c3", "ffffffffffffffffffffffffffffffffffffffff"]`
	
	s = strings.Replace(s, oldList, newList, 1)
	os.WriteFile("internal/terraform/aws/ec2/main.tf", []byte(s), 0644)
}
