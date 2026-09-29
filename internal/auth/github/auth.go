package github

import (
	"os"
	"os/exec"
	"strings"
)

type GitHubAuthenticator struct {
	Username string
}

func New() *GitHubAuthenticator {
	return &GitHubAuthenticator{}
}

func (a *GitHubAuthenticator) IsAuthenticated() bool {
	cmd := exec.Command("gh", "auth", "status")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	// Parse username from "Logged in to github.com as <username>" or "Logged in to github.com account <username>"
	outStr := string(out)
	if strings.Contains(outStr, "Logged in to github.com as") || strings.Contains(outStr, "Logged in to github.com account") {
		var parts []string
		if strings.Contains(outStr, "Logged in to github.com as") {
			parts = strings.Split(outStr, "Logged in to github.com as ")
		} else {
			parts = strings.Split(outStr, "Logged in to github.com account ")
		}
		
		if len(parts) > 1 {
			userPart := strings.Split(parts[1], " ")[0]
			a.Username = strings.TrimSpace(userPart)
		}
		return true
	}
	return false
}

func (a *GitHubAuthenticator) Authenticate() error {
	cmd := exec.Command("gh", "auth", "login")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (a *GitHubAuthenticator) GetIdentifier() string {
	return a.Username
}
