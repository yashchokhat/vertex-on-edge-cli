package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/yashchokhat/vertex-on-edge/internal/config"
	"github.com/yashchokhat/vertex-on-edge/pkg/models"
)

// PromptProjectPath asks the user to choose the project location interactively.
func PromptProjectPath() string {
	var method string
	err := huh.NewSelect[string]().
		Title("Select Project Location").
		Options(
			huh.NewOption("Current Directory (.)", "."),
			huh.NewOption("Enter path manually", "manual"),
			huh.NewOption("Navigate file system interactively (Native GUI)", "picker"),
			huh.NewOption("Navigate file system interactively (Terminal CLI)", "picker_cli"),
		).
		Value(&method).
		Run()

	if err == huh.ErrUserAborted {
		os.Exit(0)
	}

	if method == "." {
		cwd, _ := os.Getwd()
		return cwd
	}

	if method == "picker" {
		PrintInfo("Opening native OS folder selection dialog...")

		startDir, err := os.UserHomeDir()
		if err != nil || startDir == "" {
			startDir = "/"
			if runtime.GOOS == "windows" {
				startDir = "C:\\"
			}
		}

		var selectedPath string

		if runtime.GOOS == "darwin" {
			cmd := exec.Command("osascript",
				"-e", `tell application (path to frontmost application as text) to set theDir to choose folder with prompt "Select Project Directory" default location alias (POSIX file "`+startDir+`")`,
				"-e", `POSIX path of theDir`,
			)
			out, cmdErr := cmd.Output()
			if cmdErr != nil {
				os.Exit(0) // Assume User Cancelled
			} else {
				selectedPath = strings.TrimSpace(string(out))
			}
		} else if runtime.GOOS == "windows" {
			psScript := `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.FolderBrowserDialog; $f.SelectedPath = '` + startDir + `'; if($f.ShowDialog() -eq 'OK'){ $f.SelectedPath }`
			cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
			out, cmdErr := cmd.Output()
			if cmdErr != nil || strings.TrimSpace(string(out)) == "" {
				os.Exit(0)
			} else {
				selectedPath = strings.TrimSpace(string(out))
			}
		} else {
			// Linux: try zenity then kdialog
			cmd := exec.Command("zenity", "--file-selection", "--directory", "--title=Select Project Directory")
			out, cmdErr := cmd.Output()
			if cmdErr != nil {
				cmd = exec.Command("kdialog", "--getexistingdirectory", startDir)
				out, cmdErr = cmd.Output()
				if cmdErr != nil {
					os.Exit(0)
				}
			}
			selectedPath = strings.TrimSpace(string(out))
		}

		if selectedPath != "" {
			return selectedPath
		}

		// Fallback to current if canceled or errored
		cwd, _ := os.Getwd()
		return cwd
	}

	if method == "picker_cli" {
		var selectedPath string

		km := huh.NewDefaultKeyMap()
		km.FilePicker.Open = key.NewBinding(key.WithKeys("l", "right", "enter"), key.WithHelp("enter", "open"))
		km.FilePicker.Select = key.NewBinding(key.WithKeys("ctrl+enter", "ctrl+s", "tab"), key.WithHelp("tab/ctrl+s", "select dir"))

		startDir, err := os.UserHomeDir()
		if err != nil || startDir == "" {
			startDir = "/"
			if runtime.GOOS == "windows" {
				startDir = "C:\\"
			}
		}

		osName := "Linux"
		if runtime.GOOS == "windows" {
			osName = "Windows"
		} else if runtime.GOOS == "darwin" {
			osName = "macOS"
		}

		title := fmt.Sprintf("Select project directory [%s] (Enter to go inside, Tab/Ctrl+S to select)", osName)

		err = huh.NewFilePicker().
			Title(title).
			DirAllowed(true).
			FileAllowed(false).
			ShowHidden(true).
			CurrentDirectory(startDir).
			Height(10).
			Value(&selectedPath).
			WithKeyMap(km).
			Run()

		if err == huh.ErrUserAborted {
			os.Exit(0)
		}
		if err == nil && selectedPath != "" {
			return selectedPath
		}

		// Fallback to current if canceled or errored
		cwd, _ := os.Getwd()
		return cwd
	}

	if method == "manual" {
		var selectedPath string
		err := huh.NewInput().
			Title("Enter project directory path:").
			Placeholder("/Users/name/Projects/my-app").
			Value(&selectedPath).
			Run()
		if err == huh.ErrUserAborted {
			os.Exit(0)
		}

		// Resolve path
		selectedPath = strings.TrimSpace(selectedPath)
		if strings.HasPrefix(selectedPath, "~") {
			home, _ := os.UserHomeDir()
			selectedPath = filepath.Join(home, strings.TrimPrefix(selectedPath, "~"))
		}
		if selectedPath == "" {
			cwd, _ := os.Getwd()
			return cwd
		}
		return selectedPath
	}

	cwd, _ := os.Getwd()
	return cwd
}

// ConfirmDetection asks the user if the detection is correct using Y/n.
func ConfirmDetection() bool {
	accentStyle := lipgloss.NewStyle().Bold(true).Foreground(accentColor)
	dimStyle := lipgloss.NewStyle().Foreground(dimColor)

	// Left align with 4 spaces to match the spinner checkmark alignment
	fmt.Printf("\n    %s  %s ", accentStyle.Render("Is the detected stack correct?"), dimStyle.Render("[Y/n]"))
	ShowCursor()
	defer HideCursor()

	input := strings.ToLower(readLine())
	if input == "" || input == "y" || input == "yes" {
		return true
	}
	return false
}

// SelectStackManual uses an interactive list to let the user select a framework.
func SelectStackManual() string {
	stacks := config.SupportedStacks()
	var selected string

	var options []huh.Option[string]
	for _, s := range stacks {
		options = append(options, huh.NewOption(s, s))
	}

	err := huh.NewSelect[string]().
		Title("Select your primary framework/tech stack:").
		Options(options...).
		Height(12).
		Value(&selected).
		Run()

	if err == huh.ErrUserAborted {
		os.Exit(0)
	}

	return selected
}

// PromptMissingDetails asks the user for details that might be missing after manual selection using interactive lists.
func PromptMissingDetails(stack *models.DetectedStack) {
	fmt.Println()
	PrintInfo("We need a few more details to configure your project.")
	fmt.Println()

	if stack.Language == "" {
		var langOptions []huh.Option[string]

		switch stack.Framework {
		case "Next.js", "React", "Vue", "Angular", "Svelte / SvelteKit", "Nuxt", "Remix", "NestJS", "Express.js", "Node.js (Generic)":
			langOptions = []huh.Option[string]{
				huh.NewOption("TypeScript", "TypeScript"),
				huh.NewOption("JavaScript", "JavaScript"),
			}
		case "Django", "FastAPI", "Flask", "Python (Generic)":
			langOptions = []huh.Option[string]{huh.NewOption("Python", "Python")}
		case "Go (Standard / Gin / Echo)":
			langOptions = []huh.Option[string]{huh.NewOption("Go", "Go")}
		case "Spring Boot", "Java (Generic)":
			langOptions = []huh.Option[string]{
				huh.NewOption("Java", "Java"),
				huh.NewOption("Kotlin", "Kotlin"),
			}
		case "Laravel", "PHP (Generic)":
			langOptions = []huh.Option[string]{huh.NewOption("PHP", "PHP")}
		case "Ruby on Rails", "Ruby (Generic)":
			langOptions = []huh.Option[string]{huh.NewOption("Ruby", "Ruby")}
		case "Rust":
			langOptions = []huh.Option[string]{huh.NewOption("Rust", "Rust")}
		default:
			langOptions = []huh.Option[string]{
				huh.NewOption("TypeScript", "TypeScript"),
				huh.NewOption("JavaScript", "JavaScript"),
				huh.NewOption("Python", "Python"),
				huh.NewOption("Go", "Go"),
				huh.NewOption("Java", "Java"),
				huh.NewOption("Other", "Other"),
			}
		}

		err := huh.NewSelect[string]().
			Title("Programming Language:").
			Options(langOptions...).
			Value(&stack.Language).
			Run()
		if err == huh.ErrUserAborted {
			os.Exit(0)
		}
	}

	if stack.PackageManager == "" {
		var pmOptions []huh.Option[string]

		switch stack.Language {
		case "JavaScript", "TypeScript":
			pmOptions = []huh.Option[string]{
				huh.NewOption("npm", "npm"),
				huh.NewOption("yarn", "yarn"),
				huh.NewOption("pnpm", "pnpm"),
				huh.NewOption("bun", "bun"),
			}
		case "Python":
			pmOptions = []huh.Option[string]{
				huh.NewOption("pip", "pip"),
				huh.NewOption("poetry", "poetry"),
				huh.NewOption("pipenv", "pipenv"),
				huh.NewOption("uv", "uv"),
			}
		case "Go":
			pmOptions = []huh.Option[string]{huh.NewOption("go mod", "go mod")}
		case "Java", "Kotlin":
			pmOptions = []huh.Option[string]{
				huh.NewOption("Maven", "Maven"),
				huh.NewOption("Gradle", "Gradle"),
			}
		case "PHP":
			pmOptions = []huh.Option[string]{huh.NewOption("Composer", "Composer")}
		case "Ruby":
			pmOptions = []huh.Option[string]{huh.NewOption("Bundler", "Bundler")}
		case "Rust":
			pmOptions = []huh.Option[string]{huh.NewOption("Cargo", "Cargo")}
		default:
			pmOptions = []huh.Option[string]{
				huh.NewOption("npm", "npm"),
				huh.NewOption("yarn", "yarn"),
				huh.NewOption("pnpm", "pnpm"),
				huh.NewOption("pip", "pip"),
				huh.NewOption("go mod", "go mod"),
				huh.NewOption("Maven", "Maven"),
				huh.NewOption("Composer", "Composer"),
				huh.NewOption("None/Other", "None"),
			}
		}

		err := huh.NewSelect[string]().
			Title("Package Manager:").
			Options(pmOptions...).
			Value(&stack.PackageManager).
			Run()
		if err == huh.ErrUserAborted {
			os.Exit(0)
		}
	}

	fmt.Println()
	PrintSuccess(fmt.Sprintf("Configuration complete: %s (%s) with %s", stack.Framework, stack.Language, stack.PackageManager))
}
