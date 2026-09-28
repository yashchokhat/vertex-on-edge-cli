package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yashchokhat/vertex-on-edge/pkg/models"
)

// testdataDir returns the absolute path to the testdata directory.
func testdataDir(t *testing.T) string {
	t.Helper()
	// Walk up from internal/detector to the project root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(wd, "..", "..", "testdata")
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func TestDetectNextJS(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "nextjs"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Framework != "Next.js" {
		t.Errorf("expected framework Next.js, got %q", prim.Framework)
	}
	if prim.Language != "TypeScript" {
		t.Errorf("expected language TypeScript, got %q", prim.Language)
	}
	if prim.Runtime != "Node.js" {
		t.Errorf("expected runtime Node.js, got %q", prim.Runtime)
	}
	if prim.PackageManager != "pnpm" {
		t.Errorf("expected package manager pnpm, got %q", prim.PackageManager)
	}
	if prim.Confidence != models.ConfidenceHigh {
		t.Errorf("expected confidence High, got %q", prim.Confidence)
	}
	if len(prim.Evidence) == 0 {
		t.Error("expected evidence, got none")
	}
}

func TestDetectReact(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "react"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Framework != "React" {
		t.Errorf("expected framework React, got %q", prim.Framework)
	}
	if prim.Language != "JavaScript" {
		t.Errorf("expected language JavaScript, got %q", prim.Language)
	}
	if prim.Runtime != "Node.js" {
		t.Errorf("expected runtime Node.js, got %q", prim.Runtime)
	}
}

func TestDetectVue(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "vue"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Framework != "Vue" {
		t.Errorf("expected framework Vue, got %q", prim.Framework)
	}
	if prim.Language != "JavaScript" {
		t.Errorf("expected language JavaScript, got %q", prim.Language)
	}
}

func TestDetectAngular(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "angular"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Framework != "Angular" {
		t.Errorf("expected framework Angular, got %q", prim.Framework)
	}
	if prim.Language != "TypeScript" {
		t.Errorf("expected language TypeScript, got %q", prim.Language)
	}
}

func TestDetectNodeJS(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "node"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Framework != "Node.js" {
		t.Errorf("expected framework Node.js, got %q", prim.Framework)
	}
	if prim.Confidence != models.ConfidenceMedium {
		t.Errorf("expected confidence Medium, got %q", prim.Confidence)
	}
}

func TestDetectPythonGeneric(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "python"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Language != "Python" {
		t.Errorf("expected language Python, got %q", prim.Language)
	}
	if prim.Framework != "" {
		t.Errorf("expected no framework for generic Python, got %q", prim.Framework)
	}
	if prim.Confidence != models.ConfidenceMedium {
		t.Errorf("expected confidence Medium, got %q", prim.Confidence)
	}
}

func TestDetectFastAPI(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "python-fastapi"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Framework != "FastAPI" {
		t.Errorf("expected framework FastAPI, got %q", prim.Framework)
	}
	if prim.Language != "Python" {
		t.Errorf("expected language Python, got %q", prim.Language)
	}
	if prim.Confidence != models.ConfidenceHigh {
		t.Errorf("expected confidence High, got %q", prim.Confidence)
	}
}

func TestDetectFlask(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "python-flask"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Framework != "Flask" {
		t.Errorf("expected framework Flask, got %q", prim.Framework)
	}
	if prim.Language != "Python" {
		t.Errorf("expected language Python, got %q", prim.Language)
	}
}

func TestDetectDjango(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "python-django"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Framework != "Django" {
		t.Errorf("expected framework Django, got %q", prim.Framework)
	}
	if prim.Language != "Python" {
		t.Errorf("expected language Python, got %q", prim.Language)
	}
}

func TestDetectGo(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "go"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Language != "Go" {
		t.Errorf("expected language Go, got %q", prim.Language)
	}
	if prim.Runtime != "Go" {
		t.Errorf("expected runtime Go, got %q", prim.Runtime)
	}
	if prim.Confidence != models.ConfidenceHigh {
		t.Errorf("expected confidence High, got %q", prim.Confidence)
	}
}

func TestDetectJava(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "java"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Language != "Java" {
		t.Errorf("expected language Java, got %q", prim.Language)
	}
	if prim.PackageManager != "Maven" {
		t.Errorf("expected package manager Maven, got %q", prim.PackageManager)
	}
	if prim.Confidence != models.ConfidenceMedium {
		t.Errorf("expected confidence Medium, got %q", prim.Confidence)
	}
}

func TestDetectSpringBoot(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "java-spring"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Framework != "Spring Boot" {
		t.Errorf("expected framework Spring Boot, got %q", prim.Framework)
	}
	if prim.Language != "Java" {
		t.Errorf("expected language Java, got %q", prim.Language)
	}
}

func TestDetectRust(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "rust"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Language != "Rust" {
		t.Errorf("expected language Rust, got %q", prim.Language)
	}
	if prim.PackageManager != "Cargo" {
		t.Errorf("expected package manager Cargo, got %q", prim.PackageManager)
	}
}

func TestDetectPHP(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "php"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Language != "PHP" {
		t.Errorf("expected language PHP, got %q", prim.Language)
	}
	if prim.PackageManager != "Composer" {
		t.Errorf("expected package manager Composer, got %q", prim.PackageManager)
	}
}

func TestDetectLaravel(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "php-laravel"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Framework != "Laravel" {
		t.Errorf("expected framework Laravel, got %q", prim.Framework)
	}
	if prim.Language != "PHP" {
		t.Errorf("expected language PHP, got %q", prim.Language)
	}
}

func TestDetectRuby(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "ruby"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Language != "Ruby" {
		t.Errorf("expected language Ruby, got %q", prim.Language)
	}
	if prim.PackageManager != "Bundler" {
		t.Errorf("expected package manager Bundler, got %q", prim.PackageManager)
	}
}

func TestDetectRails(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "ruby-rails"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasDetection() {
		t.Fatal("expected detection, got none")
	}
	prim := info.Primary()
	if prim.Framework != "Rails" {
		t.Errorf("expected framework Rails, got %q", prim.Framework)
	}
	if prim.Language != "Ruby" {
		t.Errorf("expected language Ruby, got %q", prim.Language)
	}
}

func TestDetectUnknown(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "unknown"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.HasDetection() {
		t.Errorf("expected no detection for unknown project, got %d stacks", len(info.Stacks))
	}
}

func TestDetectMalformedPackageJSON(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "malformed"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not crash, and should not detect a framework
	if info.HasDetection() {
		prim := info.Primary()
		if prim.Framework == "Next.js" || prim.Framework == "React" {
			t.Errorf("should not detect a JS framework from malformed package.json, got %q", prim.Framework)
		}
	}
}

func TestDetectEmptyDirectory(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "empty"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.HasDetection() {
		t.Errorf("expected no detection for empty directory, got %d stacks", len(info.Stacks))
	}
}

func TestDetectNonExistentDirectory(t *testing.T) {
	d := New()
	_, err := d.Detect(filepath.Join(testdataDir(t), "does-not-exist"))
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

// TestFrameworkPriorityOverLanguage verifies that Next.js is detected
// instead of generic Node.js when both 'next' and other dependencies exist.
func TestFrameworkPriorityOverLanguage(t *testing.T) {
	d := New()
	info, err := d.Detect(filepath.Join(testdataDir(t), "nextjs"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	prim := info.Primary()
	if prim.Framework != "Next.js" {
		t.Errorf("expected Next.js (framework priority), got %q", prim.Framework)
	}
	// Should NOT report as generic Node.js
	for _, s := range info.Stacks {
		if s.Framework == "Node.js" {
			t.Error("generic Node.js should not appear alongside Next.js")
		}
	}
}

// TestFileScannerExclusions verifies that excluded directories are skipped.
func TestFileScannerExclusions(t *testing.T) {
	tmp := t.TempDir()
	// Create excluded directories
	for _, d := range []string{"node_modules", ".git", "vendor", "dist"} {
		os.MkdirAll(filepath.Join(tmp, d), 0o755)
	}
	// Create a visible directory
	os.MkdirAll(filepath.Join(tmp, "src"), 0o755)

	fs := NewFileScanner(tmp)
	dirs := fs.ListSubdirectories()

	for _, d := range dirs {
		if d == "node_modules" || d == ".git" || d == "vendor" || d == "dist" {
			t.Errorf("excluded directory %q should not appear in ListSubdirectories", d)
		}
	}

	found := false
	for _, d := range dirs {
		if d == "src" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'src' in ListSubdirectories")
	}
}

// TestPackageJSONHasDependency tests the HasDependency method.
func TestPackageJSONHasDependency(t *testing.T) {
	pkg := &PackageJSON{
		Dependencies:    map[string]string{"next": "14.0.0", "react": "18.0.0"},
		DevDependencies: map[string]string{"typescript": "5.0.0"},
	}

	if !pkg.HasDependency("next") {
		t.Error("expected next to be found in dependencies")
	}
	if !pkg.HasDependency("typescript") {
		t.Error("expected typescript to be found in devDependencies")
	}
	if pkg.HasDependency("express") {
		t.Error("expected express to not be found")
	}

	// Test nil package
	var nilPkg *PackageJSON
	if nilPkg.HasDependency("next") {
		t.Error("nil PackageJSON should return false")
	}
}
