package detector

import (
	"strings"

	"github.com/yashchokhat/vertex-on-edge/pkg/models"
)

// DefaultRules returns the ordered set of detection rules.
// Rules are sorted by priority (lower = higher priority).
// Framework-specific rules are evaluated before generic language rules.
func DefaultRules() []Rule {
	return []Rule{
		nextjsRule(),
		angularRule(),
		vueRule(),
		reactRule(),
		nodejsRule(),
		djangoRule(),
		fastapiRule(),
		flaskRule(),
		pythonRule(),
		goRule(),
		springBootRule(),
		javaRule(),
		rustRule(),
		laravelRule(),
		phpRule(),
		railsRule(),
		rubyRule(),
		flutterRule(),
		reactNativeRule(),
		capacitorRule(),
		androidNativeRule(),
		iosNativeRule(),
	}
}

func pythonHasDependency(fs FileScanner, dep string) (bool, []string) {
	files := []string{"requirements.txt", "pyproject.toml", "Pipfile"}
	var evidence []string
	found := false
	depLower := strings.ToLower(dep)
	for _, file := range files {
		if !fs.FileExists(file) {
			continue
		}
		data, err := fs.ReadFile(file)
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(string(data)), depLower) {
			found = true
			evidence = append(evidence, file)
		}
	}
	return found, evidence
}

func javaHasDependency(fs FileScanner, dep string) (bool, []string) {
	files := []string{"pom.xml", "build.gradle", "build.gradle.kts"}
	var evidence []string
	found := false
	depLower := strings.ToLower(dep)
	for _, file := range files {
		if !fs.FileExists(file) {
			continue
		}
		data, err := fs.ReadFile(file)
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(string(data)), depLower) {
			found = true
			evidence = append(evidence, file)
		}
	}
	return found, evidence
}

func nextjsRule() Rule {
	return Rule{
		Name:     "Next.js",
		Priority: 10,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("package.json") {
				return nil
			}
			pkg, err := fs.ReadPackageJSON()
			if err != nil || pkg == nil || !pkg.HasDependency("next") {
				return nil
			}
			lang := "JavaScript"
			if fs.FileExists("tsconfig.json") {
				lang = "TypeScript"
			}
			pm := fs.DetectJSPackageManager()
			evidence := []string{"package.json", "next dependency"}
			lockfile := ""
			switch pm {
			case "npm":
				lockfile = "package-lock.json"
			case "yarn":
				lockfile = "yarn.lock"
			case "pnpm":
				lockfile = "pnpm-lock.yaml"
			case "bun":
				lockfile = "bun.lock"
			}
			if lockfile != "" && fs.FileExists(lockfile) {
				evidence = append(evidence, lockfile)
			}
			return &models.DetectedStack{
				Language:       lang,
				Framework:      "Next.js",
				Runtime:        "Node.js",
				PackageManager: pm,
				Confidence:     models.ConfidenceHigh,
				Evidence:       evidence,
				ProjectType:    models.ProjectTypeWebApp,
			}
		},
	}
}

func angularRule() Rule {
	return Rule{
		Name:     "Angular",
		Priority: 20,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("package.json") {
				return nil
			}
			pkg, err := fs.ReadPackageJSON()
			if err != nil || pkg == nil || !pkg.HasDependency("@angular/core") {
				return nil
			}
			pm := fs.DetectJSPackageManager()
			return &models.DetectedStack{
				Language:       "TypeScript",
				Framework:      "Angular",
				Runtime:        "Node.js",
				PackageManager: pm,
				Confidence:     models.ConfidenceHigh,
				Evidence:       []string{"package.json", "@angular/core dependency"},
			}
		},
	}
}

func vueRule() Rule {
	return Rule{
		Name:     "Vue",
		Priority: 30,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("package.json") {
				return nil
			}
			pkg, err := fs.ReadPackageJSON()
			if err != nil || pkg == nil {
				return nil
			}
			if !pkg.HasDependency("vue") || pkg.HasDependency("@angular/core") || pkg.HasDependency("next") {
				return nil
			}
			lang := "JavaScript"
			if fs.FileExists("tsconfig.json") {
				lang = "TypeScript"
			}
			pm := fs.DetectJSPackageManager()
			return &models.DetectedStack{
				Language:       lang,
				Framework:      "Vue",
				Runtime:        "Node.js",
				PackageManager: pm,
				Confidence:     models.ConfidenceHigh,
				Evidence:       []string{"package.json", "vue dependency"},
			}
		},
	}
}

func reactRule() Rule {
	return Rule{
		Name:     "React",
		Priority: 40,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("package.json") {
				return nil
			}
			pkg, err := fs.ReadPackageJSON()
			if err != nil || pkg == nil {
				return nil
			}
			if !pkg.HasDependency("react") || pkg.HasDependency("next") || pkg.HasDependency("@angular/core") || pkg.HasDependency("vue") {
				return nil
			}
			lang := "JavaScript"
			if fs.FileExists("tsconfig.json") {
				lang = "TypeScript"
			}
			pm := fs.DetectJSPackageManager()
			return &models.DetectedStack{
				Language:       lang,
				Framework:      "React",
				Runtime:        "Node.js",
				PackageManager: pm,
				Confidence:     models.ConfidenceHigh,
				Evidence:       []string{"package.json", "react dependency"},
			}
		},
	}
}

func nodejsRule() Rule {
	return Rule{
		Name:     "Node.js",
		Priority: 50,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("package.json") {
				return nil
			}
			pkg, err := fs.ReadPackageJSON()
			if err != nil || pkg == nil {
				return nil
			}
			if pkg.HasDependency("next") || pkg.HasDependency("react") || pkg.HasDependency("vue") || pkg.HasDependency("@angular/core") {
				return nil
			}
			pm := fs.DetectJSPackageManager()
			return &models.DetectedStack{
				Language:       "JavaScript",
				Framework:      "Node.js",
				Runtime:        "Node.js",
				PackageManager: pm,
				Confidence:     models.ConfidenceMedium,
				Evidence:       []string{"package.json"},
			}
		},
	}
}

func djangoRule() Rule {
	return Rule{
		Name:     "Django",
		Priority: 60,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			found, evidence := pythonHasDependency(fs, "django")
			if !found && !fs.FileExists("manage.py") {
				return nil
			}
			if fs.FileExists("manage.py") {
				evidence = append(evidence, "manage.py")
			}
			return &models.DetectedStack{
				Language:   "Python",
				Framework:  "Django",
				Confidence: models.ConfidenceHigh,
				Evidence:   evidence,
			}
		},
	}
}

func fastapiRule() Rule {
	return Rule{
		Name:     "FastAPI",
		Priority: 70,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			found, evidence := pythonHasDependency(fs, "fastapi")
			if !found {
				return nil
			}
			return &models.DetectedStack{
				Language:    "Python",
				Framework:   "FastAPI",
				Confidence:  models.ConfidenceHigh,
				Evidence:    evidence,
				ProjectType: models.ProjectTypeAPI,
			}
		},
	}
}

func flaskRule() Rule {
	return Rule{
		Name:     "Flask",
		Priority: 80,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			found, evidence := pythonHasDependency(fs, "flask")
			if !found {
				return nil
			}
			return &models.DetectedStack{
				Language:    "Python",
				Framework:   "Flask",
				Confidence:  models.ConfidenceHigh,
				Evidence:    evidence,
				ProjectType: models.ProjectTypeAPI,
			}
		},
	}
}

func pythonRule() Rule {
	return Rule{
		Name:     "Python",
		Priority: 90,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			var evidence []string
			for _, file := range []string{"requirements.txt", "pyproject.toml", "Pipfile"} {
				if fs.FileExists(file) {
					evidence = append(evidence, file)
				}
			}
			if len(evidence) == 0 {
				return nil
			}
			foundDjango, _ := pythonHasDependency(fs, "django")
			foundFastAPI, _ := pythonHasDependency(fs, "fastapi")
			foundFlask, _ := pythonHasDependency(fs, "flask")
			if foundDjango || foundFastAPI || foundFlask || fs.FileExists("manage.py") {
				return nil
			}
			return &models.DetectedStack{
				Language:   "Python",
				Confidence: models.ConfidenceMedium,
				Evidence:   evidence,
			}
		},
	}
}

func goRule() Rule {
	return Rule{
		Name:     "Go",
		Priority: 100,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("go.mod") {
				return nil
			}
			lines, err := fs.ReadFileLines("go.mod")
			var evidence []string
			if err == nil && len(lines) > 0 {
				evidence = append(evidence, "go.mod: "+strings.TrimSpace(lines[0]))
			} else {
				evidence = append(evidence, "go.mod")
			}
			return &models.DetectedStack{
				Language:   "Go",
				Runtime:    "Go",
				Confidence: models.ConfidenceHigh,
				Evidence:   evidence,
			}
		},
	}
}

func springBootRule() Rule {
	return Rule{
		Name:     "Spring Boot",
		Priority: 110,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			found, evidence := javaHasDependency(fs, "spring-boot")
			if !found {
				return nil
			}
			return &models.DetectedStack{
				Language:   "Java",
				Framework:  "Spring Boot",
				Runtime:    "JVM",
				Confidence: models.ConfidenceHigh,
				Evidence:   evidence,
			}
		},
	}
}

func javaRule() Rule {
	return Rule{
		Name:     "Java",
		Priority: 120,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			foundSB, _ := javaHasDependency(fs, "spring-boot")
			if foundSB {
				return nil
			}
			var pm string
			var evidence []string
			if fs.FileExists("pom.xml") {
				pm = "Maven"
				evidence = append(evidence, "pom.xml")
			} else if fs.FileExists("build.gradle") || fs.FileExists("build.gradle.kts") {
				pm = "Gradle"
				if fs.FileExists("build.gradle") {
					evidence = append(evidence, "build.gradle")
				} else {
					evidence = append(evidence, "build.gradle.kts")
				}
			}
			if len(evidence) == 0 {
				return nil
			}
			return &models.DetectedStack{
				Language:       "Java",
				Runtime:        "JVM",
				PackageManager: pm,
				Confidence:     models.ConfidenceMedium,
				Evidence:       evidence,
			}
		},
	}
}

func rustRule() Rule {
	return Rule{
		Name:     "Rust",
		Priority: 130,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("Cargo.toml") {
				return nil
			}
			return &models.DetectedStack{
				Language:       "Rust",
				Runtime:        "Rust",
				PackageManager: "Cargo",
				Confidence:     models.ConfidenceHigh,
				Evidence:       []string{"Cargo.toml"},
			}
		},
	}
}

func laravelRule() Rule {
	return Rule{
		Name:     "Laravel",
		Priority: 140,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("composer.json") {
				return nil
			}
			data, err := fs.ReadFile("composer.json")
			if err != nil || !strings.Contains(strings.ToLower(string(data)), "laravel") {
				return nil
			}
			return &models.DetectedStack{
				Language:   "PHP",
				Framework:  "Laravel",
				Confidence: models.ConfidenceHigh,
				Evidence:   []string{"composer.json (laravel)"},
			}
		},
	}
}

func phpRule() Rule {
	return Rule{
		Name:     "PHP",
		Priority: 150,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("composer.json") {
				return nil
			}
			data, err := fs.ReadFile("composer.json")
			if err == nil && strings.Contains(strings.ToLower(string(data)), "laravel") {
				return nil
			}
			return &models.DetectedStack{
				Language:       "PHP",
				PackageManager: "Composer",
				Confidence:     models.ConfidenceMedium,
				Evidence:       []string{"composer.json"},
			}
		},
	}
}

func railsRule() Rule {
	return Rule{
		Name:     "Rails",
		Priority: 160,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("Gemfile") {
				return nil
			}
			data, err := fs.ReadFile("Gemfile")
			if err != nil || !strings.Contains(strings.ToLower(string(data)), "rails") {
				return nil
			}
			return &models.DetectedStack{
				Language:   "Ruby",
				Framework:  "Rails",
				Confidence: models.ConfidenceHigh,
				Evidence:   []string{"Gemfile (rails)"},
			}
		},
	}
}

func rubyRule() Rule {
	return Rule{
		Name:     "Ruby",
		Priority: 170,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("Gemfile") {
				return nil
			}
			data, err := fs.ReadFile("Gemfile")
			if err == nil && strings.Contains(strings.ToLower(string(data)), "rails") {
				return nil
			}
			return &models.DetectedStack{
				Language:       "Ruby",
				PackageManager: "Bundler",
				Confidence:     models.ConfidenceMedium,
				Evidence:       []string{"Gemfile"},
			}
		},
	}
}

func flutterRule() Rule {
	return Rule{
		Name:     "Flutter",
		Priority: 5,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("pubspec.yaml") {
				return nil
			}
			data, err := fs.ReadFile("pubspec.yaml")
			if err != nil || !strings.Contains(strings.ToLower(string(data)), "flutter") {
				return nil
			}
			return &models.DetectedStack{
				Language:   "Dart",
				Framework:  "Flutter",
				Confidence: models.ConfidenceHigh,
				Evidence:   []string{"pubspec.yaml (flutter)"},
			}
		},
	}
}

func reactNativeRule() Rule {
	return Rule{
		Name:     "React Native",
		Priority: 15,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("package.json") {
				return nil
			}
			pkg, err := fs.ReadPackageJSON()
			if err != nil || pkg == nil {
				return nil
			}
			if !pkg.HasDependency("react-native") {
				return nil
			}
			lang := "JavaScript"
			if fs.FileExists("tsconfig.json") {
				lang = "TypeScript"
			}
			return &models.DetectedStack{
				Language:       lang,
				Framework:      "React Native",
				PackageManager: fs.DetectJSPackageManager(),
				Confidence:     models.ConfidenceHigh,
				Evidence:       []string{"package.json (react-native)"},
			}
		},
	}
}

func androidNativeRule() Rule {
	return Rule{
		Name:     "Android (Native)",
		Priority: 115,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if fs.FileExists("pubspec.yaml") || fs.FileExists("package.json") {
				return nil // Likely Flutter, React Native, or Capacitor
			}
			
			hasGradle := fs.FileExists("build.gradle") || fs.FileExists("build.gradle.kts") || fs.FileExists("settings.gradle")
			hasManifest := fs.FileExists("app/src/main/AndroidManifest.xml")
			
			if !hasGradle || !hasManifest {
				return nil
			}
			
			lang := "Java"
			if fs.FileExists("build.gradle.kts") {
				lang = "Kotlin"
			}
			
			return &models.DetectedStack{
				Language:       lang,
				Framework:      "Android (Native)",
				PackageManager: "Gradle",
				Confidence:     models.ConfidenceHigh,
				Evidence:       []string{"AndroidManifest.xml", "build.gradle"},
			}
		},
	}
}

func iosNativeRule() Rule {
	return Rule{
		Name:     "iOS (Native)",
		Priority: 116,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if fs.FileExists("pubspec.yaml") || fs.FileExists("package.json") {
				return nil // Likely Flutter, React Native, or Capacitor
			}
			
			// Check for .xcodeproj or .xcworkspace directories
			subdirs := fs.ListSubdirectories()
			hasXcode := false
			for _, dir := range subdirs {
				if strings.HasSuffix(dir, ".xcodeproj") || strings.HasSuffix(dir, ".xcworkspace") {
					hasXcode = true
					break
				}
			}
			
			if !hasXcode {
				return nil
			}
			
			pm := ""
			if fs.FileExists("Podfile") {
				pm = "CocoaPods"
			} else if fs.FileExists("Package.swift") {
				pm = "SwiftPM"
			}
			
			return &models.DetectedStack{
				Language:       "Swift/Objective-C",
				Framework:      "iOS (Native)",
				PackageManager: pm,
				Confidence:     models.ConfidenceHigh,
				Evidence:       []string{"Xcode project files"},
			}
		},
	}
}

func capacitorRule() Rule {
	return Rule{
		Name:     "Capacitor",
		Priority: 8,
		Detect: func(path string, fs FileScanner) *models.DetectedStack {
			if !fs.FileExists("capacitor.config.ts") && !fs.FileExists("capacitor.config.json") && !fs.FileExists("capacitor.config.js") {
				return nil
			}
			
			lang := "TypeScript"
			if fs.FileExists("capacitor.config.js") || fs.FileExists("capacitor.config.json") {
				lang = "JavaScript"
			}
			if fs.FileExists("tsconfig.json") {
				lang = "TypeScript"
			}
			
			return &models.DetectedStack{
				Language:       lang,
				Framework:      "Capacitor",
				PackageManager: fs.DetectJSPackageManager(),
				Confidence:     models.ConfidenceHigh,
				Evidence:       []string{"capacitor.config.ts/js/json"},
			}
		},
	}
}
