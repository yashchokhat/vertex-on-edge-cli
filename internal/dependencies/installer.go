package dependencies

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/yashchokhat/vertex-on-edge/internal/ui"
)

// CheckAndInstall verifies if a command exists, and if not, asks to install it.
func CheckAndInstall(cmdName, displayName string) bool {
	_, err := exec.LookPath(cmdName)
	if err == nil {
		return true // Already installed
	}

	ui.PrintError(fmt.Sprintf("%s is not installed.", displayName), "")
	if !ui.ConfirmAction(fmt.Sprintf("Would you like vertex-on-edge to automatically install %s?", displayName)) {
		return false
	}

	spinner := ui.SpinnerStart(fmt.Sprintf("Installing %s... (this may take a minute)", displayName))
	err = installTool(cmdName)
	if err != nil {
		spinner.Stop(false)
		ui.PrintError(fmt.Sprintf("Failed to install %s", displayName), err.Error())
		ui.PrintInfo("Please install it manually and try again.")
		return false
	}
	spinner.Stop(true)
	ui.PrintSuccess(fmt.Sprintf("Successfully installed %s!", displayName))
	return true
}

func installTool(cmdName string) error {
	osType := runtime.GOOS

	switch osType {
	case "darwin": // macOS
		return installMac(cmdName)
	case "windows":
		return installWindows(cmdName)
	case "linux":
		return installLinux(cmdName)
	default:
		return fmt.Errorf("unsupported OS for auto-installation: %s", osType)
	}
}

func installMac(cmdName string) error {
	// Require Homebrew
	_, err := exec.LookPath("brew")
	if err != nil {
		return fmt.Errorf("homebrew is required for auto-installation on macOS")
	}

	var pkg string
	switch cmdName {
	case "terraform":
		// HashiCorp tap is required on modern brew
		exec.Command("brew", "tap", "hashicorp/tap").Run()
		pkg = "hashicorp/tap/terraform"
	case "aws":
		pkg = "awscli"
	case "gh":
		pkg = "gh"
	default:
		return fmt.Errorf("unknown package: %s", cmdName)
	}

	cmd := exec.Command("brew", "install", pkg)
	return cmd.Run()
}

func installWindows(cmdName string) error {
	// Try winget first, then choco
	var installer, installCmd string
	
	if _, err := exec.LookPath("winget"); err == nil {
		installer = "winget"
		installCmd = "install"
	} else if _, err := exec.LookPath("choco"); err == nil {
		installer = "choco"
		installCmd = "install"
	} else {
		return fmt.Errorf("winget or choco is required for auto-installation on Windows")
	}

	var pkg string
	switch cmdName {
	case "terraform":
		pkg = "Hashicorp.Terraform" // Winget ID
		if installer == "choco" {
			pkg = "terraform"
		}
	case "aws":
		pkg = "Amazon.AWSCLI"
		if installer == "choco" {
			pkg = "awscli"
		}
	case "gh":
		pkg = "GitHub.cli"
		if installer == "choco" {
			pkg = "gh"
		}
	default:
		return fmt.Errorf("unknown package: %s", cmdName)
	}

	cmd := exec.Command(installer, installCmd, pkg, "--accept-package-agreements", "--accept-source-agreements")
	if installer == "choco" {
		cmd = exec.Command(installer, installCmd, pkg, "-y")
	}
	return cmd.Run()
}

func installLinux(cmdName string) error {
	// Try apt (Debian/Ubuntu)
	_, err := exec.LookPath("apt-get")
	if err != nil {
		return fmt.Errorf("only apt-get is supported for Linux auto-installation currently")
	}

	var pkg string
	switch cmdName {
	case "terraform":
		// Needs custom repo setup, fallback to snap if available
		if _, snapErr := exec.LookPath("snap"); snapErr == nil {
			cmd := exec.Command("sudo", "snap", "install", "terraform", "--classic")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
		return fmt.Errorf("snap is required to install terraform on linux")
	case "aws":
		pkg = "awscli"
	case "gh":
		pkg = "gh"
	default:
		return fmt.Errorf("unknown package: %s", cmdName)
	}

	cmd := exec.Command("sudo", "apt-get", "install", "-y", pkg)
	// Bind to IO to allow sudo password prompt
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
