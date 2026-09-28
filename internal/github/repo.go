package github

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// InitAndPush initializes a git repository, creates a GitHub repo if needed, sets secrets, and pushes.
func InitAndPush(projectPath, projectName string, secrets map[string]string, updateProgress func(string)) error {
	// 1. Add files first so they are included if gh repo create pushes
	updateProgress("Staging files...")

	// Check if git repo exists, init if not
	if _, err := os.Stat(projectPath + "/.git"); os.IsNotExist(err) {
		cmd := exec.Command("git", "init")
		cmd.Dir = projectPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to init git: %w", err)
		}
	}

	// 1.5 Ensure .gitignore contains .terraform and other heavy/secret files
	gitignorePath := projectPath + "/.gitignore"
	ignoreContent := "\n# Vertex-on-Edge\n.terraform/\n.terraform.*\nterraform.tfstate\nterraform.tfstate.backup\nnode_modules/\n"

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString(ignoreContent)
		f.Close()
	}

	updateProgress("Staging files...")

	// Un-track files that might have been accidentally added before we created the gitignore
	rmCmd := exec.Command("git", "rm", "-r", "--cached", ".")
	rmCmd.Dir = projectPath
	rmCmd.Run()

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = projectPath
	addCmd.Run()

	commitCmd := exec.Command("git", "commit", "-m", "ci: initialize vertex-on-edge deployment pipeline")
	commitCmd.Dir = projectPath
	commitCmd.Run() // Ignore errors if nothing to commit

	// 2. Check if github remote exists
	updateProgress("Checking GitHub remote...")
	urlCmd := exec.Command("git", "remote", "get-url", "origin")
	urlCmd.Dir = projectPath
	err = urlCmd.Run()
	if err != nil {
		updateProgress(fmt.Sprintf("Creating GitHub repository '%s'...", projectName))
		// Create and push
		createCmd := exec.Command("gh", "repo", "create", projectName, "--private", "--source=.", "--remote=origin", "--push")
		createCmd.Dir = projectPath
		// Do not bind stdin to avoid blocking on prompts silently.
		// If it needs auth, it should have been done in pre-flight.
		out, err := createCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to create github repo: %w\nOutput: %s", err, string(out))
		}
	}

	// 3. Set Secrets securely in parallel
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

	// 4. Final Push (if there are new commits after repo creation)
	updateProgress("Pushing final changes to GitHub...")
	pushCmd := exec.Command("git", "push", "-u", "origin", "HEAD")
	pushCmd.Dir = projectPath
	out, err := pushCmd.CombinedOutput()
	if err != nil {
		branchCmd := exec.Command("git", "branch", "--show-current")
		branchCmd.Dir = projectPath
		b, bErr := branchCmd.Output()
		if bErr == nil {
			branch := strings.TrimSpace(string(b))
			pushCmd2 := exec.Command("git", "push", "--set-upstream", "origin", branch)
			pushCmd2.Dir = projectPath
			out2, err2 := pushCmd2.CombinedOutput()
			if err2 != nil {
				return fmt.Errorf("failed to push to github: %w\nOutput1: %s\nOutput2: %s", err2, string(out), string(out2))
			}
		} else {
			return fmt.Errorf("failed to push to github: %w\nOutput: %s", err, string(out))
		}
	}

	return nil
}
