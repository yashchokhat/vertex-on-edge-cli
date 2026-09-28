package models

// Confidence represents the confidence level of a detection result.
type Confidence string

const (
	ConfidenceHigh   Confidence = "High"
	ConfidenceMedium Confidence = "Medium"
	ConfidenceLow    Confidence = "Low"
	ConfidenceNone   Confidence = "None"
)

// ProjectType categorizes the kind of project detected.
type ProjectType string

const (
	ProjectTypeWebApp    ProjectType = "Web Application"
	ProjectTypeAPI       ProjectType = "API"
	ProjectTypeCLI       ProjectType = "CLI"
	ProjectTypeLibrary   ProjectType = "Library"
	ProjectTypeFullStack ProjectType = "Full Stack"
	ProjectTypeUnknown   ProjectType = "Unknown"
)

// DetectedStack represents the complete detection result for a project.
type DetectedStack struct {
	Language       string      `json:"language"`
	Framework      string      `json:"framework,omitempty"`
	Runtime        string      `json:"runtime,omitempty"`
	PackageManager string      `json:"package_manager,omitempty"`
	Version        string      `json:"version,omitempty"`
	Confidence     Confidence  `json:"confidence"`
	Evidence       []string    `json:"evidence"`
	ProjectType    ProjectType `json:"project_type"`
}

// ProjectInfo holds metadata about the scanned project.
type ProjectInfo struct {
	Path       string          `json:"path"`
	Name       string          `json:"name"`
	Stacks     []DetectedStack `json:"stacks"`
	PrimaryIdx int             `json:"primary_index"`
}

// Primary returns the primary detected stack, or nil if none.
func (p *ProjectInfo) Primary() *DetectedStack {
	if len(p.Stacks) == 0 {
		return nil
	}
	if p.PrimaryIdx >= 0 && p.PrimaryIdx < len(p.Stacks) {
		return &p.Stacks[p.PrimaryIdx]
	}
	return &p.Stacks[0]
}

// HasDetection returns true if at least one stack was detected.
func (p *ProjectInfo) HasDetection() bool {
	return len(p.Stacks) > 0
}
