package terraform

import (
	"strings"
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type Runner interface {
	Init() error
	Validate() error
	Plan() (string, error)
	Apply() error
	Destroy() error
	Output(name string) (string, error)
	StateList() ([]string, error)
	Import(address string, id string) error
}

type LocalRunner struct {
	WorkspaceDir string
}

func NewLocalRunner(workspaceDir string) *LocalRunner {
	return &LocalRunner{
		WorkspaceDir: workspaceDir,
	}
}

func (r *LocalRunner) execute(args ...string) (string, error) {
	cmd := exec.Command("terraform", args...)
	cmd.Dir = r.WorkspaceDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("terraform %s failed: %w\nStderr: %s", args[0], err, stderr.String())
	}

	return stdout.String(), nil
}

func (r *LocalRunner) Init() error {
	_, err := r.execute("init", "-input=false")
	return err
}

func (r *LocalRunner) Validate() error {
	_, err := r.execute("validate")
	return err
}

func (r *LocalRunner) Plan() (string, error) {
	return r.execute("plan", "-input=false", "-no-color")
}

func (r *LocalRunner) Apply() error {
	cmd := exec.Command("terraform", "apply", "-auto-approve", "-input=false")
	cmd.Dir = r.WorkspaceDir

	// Stream output to terminal
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}
	return nil
}

func (r *LocalRunner) Destroy() error {
	cmd := exec.Command("terraform", "destroy", "-auto-approve", "-input=false")
	cmd.Dir = r.WorkspaceDir

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("terraform destroy failed: %w", err)
	}
	return nil
}

func (r *LocalRunner) Output(name string) (string, error) {
	return r.execute("output", "-raw", name)
}

func (r *LocalRunner) StateList() ([]string, error) {
	out, err := r.execute("state", "list")
	if err != nil {
		// If state is empty, terraform returns no error but empty string
		return nil, nil
	}
	var resources []string
	for _, line := range strings.Split(out, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			resources = append(resources, trimmed)
		}
	}
	return resources, nil
}

func (r *LocalRunner) Import(address string, id string) error {
	_, err := r.execute("import", "-input=false", address, id)
	return err
}
