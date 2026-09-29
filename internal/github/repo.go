package github

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// InitAndPush is the main orchestrator function for GitHub synchronization
func InitAndPush(projectPath, projectName string, secrets map[string]string, updateProgress func(string)) error {
	if err := EnsureGitRepo(projectPath, updateProgress); err != nil {
		return err
	}
	
	pushUrl, err := EnsureGitHubRepo(projectPath, projectName, updateProgress)
	if err != nil {
		return err
	}
	
	if err := ConfigureSecrets(projectPath, secrets, updateProgress); err != nil {
		return err
	}
	
	if err := SyncAndPush(projectPath, pushUrl, updateProgress); err != nil {
		return err
	}
	
	return nil
}

// EnsureGitRepo initializes the local git repository and creates the initial commit
func EnsureGitRepo(projectPath string, updateProgress func(string)) error {
	updateProgress("Staging files...")

	// Check if git repo exists, init if not
	if _, err := os.Stat(projectPath + "/.git"); os.IsNotExist(err) {
		cmd := exec.Command("git", "init")
		cmd.Dir = projectPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to init git: %w", err)
		}
	}

	// Ensure .gitignore contains .terraform and other heavy/secret files
	gitignorePath := projectPath + "/.gitignore"
	ignoreContent := "\n# Vertex-on-Edge\n.terraform/\n.terraform.*\nterraform.tfstate\nterraform.tfstate.backup\nnode_modules/\n"

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString(ignoreContent)
		f.Close()
	}

	// Un-track files that might have been accidentally added before we created the gitignore
	rmCmd := exec.Command("git", "rm", "-r", "--cached", ".")
	rmCmd.Dir = projectPath
	rmCmd.Run()

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = projectPath
	addCmd.Run()

	// Ensure git identity is configured to prevent commit failures on fresh machines
	if err := exec.Command("git", "config", "user.email").Run(); err != nil {
		emailCmd := exec.Command("git", "-C", projectPath, "config", "user.email", "deploy@vertexonedge.local")
		emailCmd.Run()
		nameCmd := exec.Command("git", "-C", projectPath, "config", "user.name", "Vertex-on-Edge Automator")
		nameCmd.Run()
	}

	commitCmd := exec.Command("git", "commit", "-m", "ci: initialize vertex-on-edge deployment pipeline")
	commitCmd.Dir = projectPath
	commitCmd.Run() // Ignore errors if nothing to commit

	return nil
}

// EnsureGitHubRepo verifies the repo exists on GitHub, creates it if not, and sets up the remote
func EnsureGitHubRepo(projectPath, projectName string, updateProgress func(string)) (string, error) {
	updateProgress(fmt.Sprintf("Verifying repository '%s' on GitHub...", projectName))
	var originalRemote string
	urlOut, err := exec.Command("gh", "repo", "view", projectName, "--json", "url", "-q", ".url").Output()
	
	if err == nil {
		originalRemote = strings.TrimSpace(string(urlOut))
	} else {
		// Repo does not exist on GitHub, create it
		updateProgress(fmt.Sprintf("Creating GitHub repository '%s'...", projectName))
		createCmd := exec.Command("gh", "repo", "create", projectName, "--private")
		createCmd.Dir = projectPath
		if out, err := createCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to create github repo: %w\nOutput: %s", err, string(out))
		}
		
		// Retrieve the URL of the newly created repo
		urlOut2, err2 := exec.Command("gh", "repo", "view", projectName, "--json", "url", "-q", ".url").Output()
		if err2 != nil {
			return "", fmt.Errorf("failed to retrieve newly created repo url: %w", err2)
		}
		originalRemote = strings.TrimSpace(string(urlOut2))
	}

	// Ensure local git 'origin' points to the definitive GitHub HTTPS URL
	exec.Command("git", "-C", projectPath, "remote", "remove", "origin").Run()
	exec.Command("git", "-C", projectPath, "remote", "add", "origin", originalRemote).Run()

	// Convert SSH URL to HTTPS URL for seamless pushes
	pushUrl := originalRemote
	if strings.HasPrefix(originalRemote, "git@github.com:") {
		repoPart := strings.TrimPrefix(originalRemote, "git@github.com:")
		pushUrl = "https://github.com/" + repoPart
	}

	// Verify GitHub API authentication via HTTPS
	updateProgress("Verifying GitHub API authentication...")
	lsCmd := exec.Command("git", "ls-remote", pushUrl)
	lsCmd.Dir = projectPath
	if lsOut, err := lsCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("GitHub authentication failed during ls-remote.\nOutput: %s", string(lsOut))
	}

	return pushUrl, nil
}

// ConfigureSecrets sets up the necessary GitHub Action secrets
func ConfigureSecrets(projectPath string, secrets map[string]string, updateProgress func(string)) error {
	updateProgress("Configuring GitHub secrets...")
	var wg sync.WaitGroup
	errCh := make(chan error, len(secrets))

	for k, v := range secrets {
		wg.Add(1)
		go func(key, value string) {
			defer wg.Done()
			updateProgress(fmt.Sprintf("Setting GitHub Secret: %s...", key))
			setCmd := exec.Command("gh", "secret", "set", key, "-b", value)
			setCmd.Dir = projectPath
			out, err := setCmd.CombinedOutput()
			if err != nil {
				errCh <- fmt.Errorf("failed to set secret %s: %w\nOutput: %s", key, err, string(out))
			}
		}(k, v)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

// SyncAndPush proactively merges remote changes and pushes the local repository
func SyncAndPush(projectPath, pushUrl string, updateProgress func(string)) error {
	updateProgress("Checking for remote changes (auto-merge)...")
	
	// Proactively pull remote changes before attempting push
	// -X ours favors local modifications (like .gitignore) during unrelated history collisions
	pullCmd := exec.Command("git", "pull", "--no-rebase", "-X", "ours", pushUrl, "main", "--allow-unrelated-histories")
	pullCmd.Dir = projectPath
	if pullOut, pullErr := pullCmd.CombinedOutput(); pullErr != nil {
		exec.Command("git", "-C", projectPath, "merge", "--abort").Run()
		// Only fail if it's a genuine merge conflict. If remote 'main' simply doesn't exist yet, that's fine!
		if strings.Contains(string(pullOut), "couldn't find remote ref main") {
			// Remote is empty or main doesn't exist, safe to proceed
		} else {
			return fmt.Errorf("failed to sync with remote due to complex merge conflicts.\nOutput: %s", string(pullOut))
		}
	}

	updateProgress("Pushing final changes to GitHub via HTTPS...")
	pushCmd := exec.Command("git", "push", pushUrl, "HEAD:main")
	pushCmd.Dir = projectPath
	out, err := pushCmd.CombinedOutput()
	
	if err != nil {
		// Fallback to push current branch if main fails
		branchCmd := exec.Command("git", "branch", "--show-current")
		branchCmd.Dir = projectPath
		if b, bErr := branchCmd.Output(); bErr == nil {
			branch := strings.TrimSpace(string(b))
			pushCmd2 := exec.Command("git", "push", pushUrl, "HEAD:"+branch)
			pushCmd2.Dir = projectPath
			out2, err2 := pushCmd2.CombinedOutput()
			if err2 != nil {
				return fmt.Errorf("failed to push to github via HTTPS: %w\nOutput1: %s\nOutput2: %s", err2, string(out), string(out2))
			}
		} else {
			return fmt.Errorf("failed to push to github via HTTPS: %w\nOutput: %s", err, string(out))
		}
	}

	// Setup tracking branch gracefully
	exec.Command("git", "-C", projectPath, "branch", "--set-upstream-to=origin/main", "main").Run()

	return nil
}
