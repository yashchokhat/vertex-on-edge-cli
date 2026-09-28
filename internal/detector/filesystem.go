package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// excludedDirs contains directories that should never be scanned.
var excludedDirs = map[string]bool{
	".git":          true,
	"node_modules":  true,
	"vendor":        true,
	".next":         true,
	"dist":          true,
	"build":         true,
	"target":        true,
	"bin":           true,
	"__pycache__":   true,
	".venv":         true,
	"venv":          true,
	".tox":          true,
	".mypy_cache":   true,
	".pytest_cache": true,
}

// FileScanner provides controlled filesystem access for detection.
type FileScanner struct {
	root string
}

// NewFileScanner creates a new FileScanner rooted at the given path.
func NewFileScanner(root string) FileScanner {
	return FileScanner{root: root}
}

// Root returns the scanner's root directory.
func (fs FileScanner) Root() string {
	return fs.root
}

// FileExists checks if a file exists at the given relative path.
func (fs FileScanner) FileExists(relPath string) bool {
	full := filepath.Join(fs.root, relPath)
	info, err := os.Stat(full)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a directory exists at the given relative path.
func (fs FileScanner) DirExists(relPath string) bool {
	full := filepath.Join(fs.root, relPath)
	info, err := os.Stat(full)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ReadFile reads the contents of a file at the given relative path.
func (fs FileScanner) ReadFile(relPath string) ([]byte, error) {
	full := filepath.Join(fs.root, relPath)
	return os.ReadFile(full)
}

// PackageJSON represents a parsed package.json file.
type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
}

// ReadPackageJSON reads and parses a package.json file.
func (fs FileScanner) ReadPackageJSON() (*PackageJSON, error) {
	data, err := fs.ReadFile("package.json")
	if err != nil {
		return nil, err
	}
	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

// HasDependency checks if a package.json contains a given dependency
// (in either dependencies or devDependencies).
func (pkg *PackageJSON) HasDependency(name string) bool {
	if pkg == nil {
		return false
	}
	if _, ok := pkg.Dependencies[name]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies[name]; ok {
		return true
	}
	return false
}

// DetectJSPackageManager determines the JavaScript package manager used.
func (fs FileScanner) DetectJSPackageManager() string {
	switch {
	case fs.FileExists("pnpm-lock.yaml"):
		return "pnpm"
	case fs.FileExists("yarn.lock"):
		return "yarn"
	case fs.FileExists("bun.lock") || fs.FileExists("bun.lockb"):
		return "bun"
	case fs.FileExists("package-lock.json"):
		return "npm"
	default:
		return "npm"
	}
}

// ListTopLevelFiles returns the names of files and directories in the root.
// Excluded directories are filtered out.
func (fs FileScanner) ListTopLevelFiles() []string {
	entries, err := os.ReadDir(fs.root)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if e.IsDir() && excludedDirs[name] {
			continue
		}
		names = append(names, name)
	}
	return names
}

// ListSubdirectories returns immediate subdirectories (excluding hidden and excluded dirs).
func (fs FileScanner) ListSubdirectories() []string {
	entries, err := os.ReadDir(fs.root)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if excludedDirs[name] {
			continue
		}
		dirs = append(dirs, name)
	}
	return dirs
}

// ReadFileLines reads a file and returns its lines as a string slice.
func (fs FileScanner) ReadFileLines(relPath string) ([]string, error) {
	data, err := fs.ReadFile(relPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	return lines, nil
}
