package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/cli/deploy.go")
	s := string(b)
	
	oldStr := `		// Attempt to get the actual remote repository identity if it exists
		ghRepoViewCmd := exec.Command("gh", "repo", "view", "--json", "owner,name", "-q", ".owner.login + \"/\" + .name")
		ghRepoViewCmd.Dir = projectPath
		if repoInfo, err := ghRepoViewCmd.Output(); err == nil {
			parts := strings.Split(strings.TrimSpace(string(repoInfo)), "/")
			if len(parts) == 2 {
				githubOwner = parts[0]
				exactRepoName = parts[1] // Override the local directory name with the true remote repo name
			}
		} else {
			// Fallback for brand new projects that aren't on GitHub yet
			ghApiUserCmd := exec.Command("gh", "api", "user", "-q", ".login")
			if ghUser, err := ghApiUserCmd.Output(); err == nil {
				githubOwner = strings.TrimSpace(string(ghUser))
			}
		}`

	newStr := `		// Attempt to get the actual remote repository identity if it exists
		ghRepoViewCmd := exec.Command("gh", "repo", "view", "--json", "owner,name", "-q", ".owner.login + \"/\" + .name")
		ghRepoViewCmd.Dir = projectPath
		if repoInfo, err := ghRepoViewCmd.Output(); err == nil {
			parts := strings.Split(strings.TrimSpace(string(repoInfo)), "/")
			if len(parts) == 2 {
				githubOwner = parts[0]
				exactRepoName = parts[1] // Override the local directory name with the true remote repo name
			}
		} else {
			// Fallback for brand new projects that aren't on GitHub yet
			ghApiUserCmd := exec.Command("gh", "api", "user", "-q", ".login")
			if ghUser, err := ghApiUserCmd.Output(); err == nil {
				githubOwner = strings.TrimSpace(string(ghUser))
			}
			
			// Try to extract from git remote if gh repo view failed
			gitRemoteCmd := exec.Command("git", "-C", projectPath, "remote", "get-url", "origin")
			if remoteOut, err := gitRemoteCmd.Output(); err == nil {
				remoteStr := strings.TrimSpace(string(remoteOut))
				// Handle both HTTPS and SSH urls
				if strings.HasPrefix(remoteStr, "https://github.com/") {
					parts := strings.Split(strings.TrimPrefix(remoteStr, "https://github.com/"), "/")
					if len(parts) >= 2 {
						githubOwner = parts[0]
						exactRepoName = strings.TrimSuffix(parts[1], ".git")
					}
				} else if strings.HasPrefix(remoteStr, "git@github.com:") {
					parts := strings.Split(strings.TrimPrefix(remoteStr, "git@github.com:"), "/")
					if len(parts) >= 2 {
						githubOwner = parts[0]
						exactRepoName = strings.TrimSuffix(parts[1], ".git")
					}
				}
			}
		}`

	s = strings.Replace(s, oldStr, newStr, 1)
	
	// Also patch the Fast Deploy section
	oldFastStr := `		fastRepoViewCmd := exec.Command("gh", "repo", "view", "--json", "owner,name", "-q", ".owner.login + \"/\" + .name")
		fastRepoViewCmd.Dir = projectPath
		if repoInfo, err := fastRepoViewCmd.Output(); err == nil {
			parts := strings.Split(strings.TrimSpace(string(repoInfo)), "/")
			if len(parts) == 2 {
				githubOwner = parts[0]
				exactRepoName = parts[1]
			}
		} else {
			// Fallback for brand new projects that aren't on GitHub yet
			ghApiUserCmd := exec.Command("gh", "api", "user", "-q", ".login")
			if ghUser, err := ghApiUserCmd.Output(); err == nil {
				githubOwner = strings.TrimSpace(string(ghUser))
			}
		}`

	newFastStr := `		fastRepoViewCmd := exec.Command("gh", "repo", "view", "--json", "owner,name", "-q", ".owner.login + \"/\" + .name")
		fastRepoViewCmd.Dir = projectPath
		if repoInfo, err := fastRepoViewCmd.Output(); err == nil {
			parts := strings.Split(strings.TrimSpace(string(repoInfo)), "/")
			if len(parts) == 2 {
				githubOwner = parts[0]
				exactRepoName = parts[1]
			}
		} else {
			// Fallback for brand new projects that aren't on GitHub yet
			ghApiUserCmd := exec.Command("gh", "api", "user", "-q", ".login")
			if ghUser, err := ghApiUserCmd.Output(); err == nil {
				githubOwner = strings.TrimSpace(string(ghUser))
			}
			
			// Try to extract from git remote if gh repo view failed
			gitRemoteCmd := exec.Command("git", "-C", projectPath, "remote", "get-url", "origin")
			if remoteOut, err := gitRemoteCmd.Output(); err == nil {
				remoteStr := strings.TrimSpace(string(remoteOut))
				// Handle both HTTPS and SSH urls
				if strings.HasPrefix(remoteStr, "https://github.com/") {
					parts := strings.Split(strings.TrimPrefix(remoteStr, "https://github.com/"), "/")
					if len(parts) >= 2 {
						githubOwner = parts[0]
						exactRepoName = strings.TrimSuffix(parts[1], ".git")
					}
				} else if strings.HasPrefix(remoteStr, "git@github.com:") {
					parts := strings.Split(strings.TrimPrefix(remoteStr, "git@github.com:"), "/")
					if len(parts) >= 2 {
						githubOwner = parts[0]
						exactRepoName = strings.TrimSuffix(parts[1], ".git")
					}
				}
			}
		}`

	s = strings.Replace(s, oldFastStr, newFastStr, 1)

	os.WriteFile("internal/cli/deploy.go", []byte(s), 0644)
}
