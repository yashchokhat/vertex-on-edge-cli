package detector

import "github.com/yashchokhat/vertex-on-edge/pkg/models"

// Detector defines the interface for project technology detection.
type Detector interface {
	// Detect analyzes the given directory and returns project information.
	Detect(projectPath string) (*models.ProjectInfo, error)
}

// Rule represents a single detection rule that checks for a specific technology.
type Rule struct {
	// Name is the human-readable name of this rule (e.g., "Next.js").
	Name string

	// Priority determines the order of evaluation. Lower values are checked first.
	// Framework-specific rules should have lower priority values than generic ones.
	Priority int

	// Detect is the function that performs the actual detection.
	// It receives the project path and a FileScanner for filesystem access.
	Detect func(path string, fs FileScanner) *models.DetectedStack
}
