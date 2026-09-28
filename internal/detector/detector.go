package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/yashchokhat/vertex-on-edge/pkg/models"
)

// DefaultDetector implements the Detector interface using rule-based detection.
type DefaultDetector struct {
	rules []Rule
}

// New creates a new DefaultDetector with the standard detection rules.
func New() *DefaultDetector {
	rules := DefaultRules()
	// Sort rules by priority
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})
	return &DefaultDetector{rules: rules}
}

// NewWithRules creates a detector with custom rules (useful for testing).
func NewWithRules(rules []Rule) *DefaultDetector {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})
	return &DefaultDetector{rules: rules}
}

// Detect analyzes the project at the given path and returns detection results.
func (d *DefaultDetector) Detect(projectPath string) (*models.ProjectInfo, error) {
	// Resolve to absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve path: %w", err)
	}

	// Verify the path exists and is a directory
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory does not exist: %s", absPath)
		}
		return nil, fmt.Errorf("unable to access directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", absPath)
	}

	projectName := filepath.Base(absPath)
	fs := NewFileScanner(absPath)

	// Run detection rules on the root directory
	var stacks []models.DetectedStack
	matched := make(map[string]bool) // track matched rule categories to avoid duplicates

	for _, rule := range d.rules {
		result := rule.Detect(absPath, fs)
		if result != nil {
			// Avoid duplicate detections for the same language category
			key := result.Language
			if result.Framework != "" {
				key = result.Framework
			}
			if !matched[key] {
				matched[key] = true
				stacks = append(stacks, *result)
			}
		}
	}

	// Also check subdirectories for multi-component projects
	subdirs := fs.ListSubdirectories()
	for _, dir := range subdirs {
		subPath := filepath.Join(absPath, dir)
		subFS := NewFileScanner(subPath)
		for _, rule := range d.rules {
			result := rule.Detect(subPath, subFS)
			if result != nil {
				key := dir + ":" + result.Language
				if result.Framework != "" {
					key = dir + ":" + result.Framework
				}
				if !matched[key] {
					matched[key] = true
					// Prepend subdirectory to evidence
					for i, e := range result.Evidence {
						result.Evidence[i] = dir + "/" + e
					}
					stacks = append(stacks, *result)
				}
				break // Only take the highest-priority match per subdirectory
			}
		}
	}

	return &models.ProjectInfo{
		Path:       absPath,
		Name:       projectName,
		Stacks:     stacks,
		PrimaryIdx: 0,
	}, nil
}
