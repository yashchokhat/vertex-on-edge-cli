package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/yashchokhat/vertex-on-edge/internal/auth/aws"
	ghauth "github.com/yashchokhat/vertex-on-edge/internal/auth/github"
	"github.com/yashchokhat/vertex-on-edge/internal/dependencies"
	"github.com/yashchokhat/vertex-on-edge/internal/detector"
	"github.com/yashchokhat/vertex-on-edge/internal/docker"
	ghactions "github.com/yashchokhat/vertex-on-edge/internal/github"
	"github.com/yashchokhat/vertex-on-edge/internal/platform"
	"github.com/yashchokhat/vertex-on-edge/internal/terraform"
	"github.com/yashchokhat/vertex-on-edge/internal/ui"
	"github.com/yashchokhat/vertex-on-edge/pkg/models"

	"github.com/charmbracelet/lipgloss"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the application to the cloud",
	Long: `Deploy your application to the cloud with a single command.

This walks you through project detection, cloud provider selection,
infrastructure provisioning (via Terraform), and CI/CD pipeline
configuration (via GitHub Actions).`,
	RunE: runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

// runDeploy is the main deployment orchestrator. It follows a strict sequence:
//
//  1. Environment check (Git, Docker, Terraform)
//  2. GitHub and AWS authentication
//  3. Project selection and framework detection
//  4. Cloud provider and deployment target selection
//  5. AWS infrastructure configuration (VPC, Subnet, Key Pair)
//  6. CI/CD file generation (Dockerfile, GitHub Actions workflows)
//  7. Terraform init, validate, plan, and apply
//  8. Post-apply validation (OIDC, role ARN from Terraform outputs)
//  9. GitHub secrets injection and repository push
func runDeploy(cmd *cobra.Command, args []string) error {
	ui.RunStartupAnimation()
	ui.PrintBanner()
	ui.PrintWelcome()

	// -----------------------------------------------------------------------
	// Step 1: Environment Check
	// -----------------------------------------------------------------------
	ui.PrintInfo("Checking environment...")

	if _, err := exec.LookPath("git"); err != nil {
		ui.PrintError("Git is not installed", "Git is required for deployments.")
		ui.PrintInfo("Install guide: https://git-scm.com/downloads")
		os.Exit(1)
	}
	fmt.Println("  ✓ Git")

	if _, err := exec.LookPath("docker"); err != nil {
		ui.PrintError("Docker is not installed", "Docker is required to build containers.")
		ui.PrintInfo("Install guide: https://docs.docker.com/get-docker/")
		os.Exit(1)
	}
	fmt.Println("  ✓ Docker")

	if _, err := exec.LookPath("terraform"); err != nil {
		fmt.Println("  ✕ Terraform is missing")
		if !dependencies.CheckAndInstall("terraform", "Terraform") {
			ui.PrintError("Terraform is required", "Please install it to continue.")
			ui.PrintInfo("Install guide (macOS): brew tap hashicorp/tap && brew install hashicorp/tap/terraform")
			ui.PrintInfo("Install guide (Windows): winget install Hashicorp.Terraform")
			ui.PrintInfo("Official docs: https://developer.hashicorp.com/terraform/install")
			os.Exit(1)
		}
	} else {
		fmt.Println("  ✓ Terraform")
	}
	fmt.Println()

	// -----------------------------------------------------------------------
	// Step 2: Authentication
	// -----------------------------------------------------------------------
	ui.PrintInfo("Authentication")

	if !dependencies.CheckAndInstall("gh", "GitHub CLI (gh)") {
		ui.PrintInfo("Continuing without GitHub auth. You can still generate local files.")
	} else {
		ghAuth := ghauth.New()
		if !ghAuth.IsAuthenticated() {
			ui.PrintInfo("GitHub authentication required")
			err := ghAuth.Authenticate()
			if err != nil {
				ui.PrintError("GitHub authentication failed", err.Error())
				ui.PrintInfo("Continuing without GitHub auth. You can still generate local files.")
			} else {
				ui.PrintSuccess("GitHub Authenticated")
			}
		} else {
			ui.PrintSuccess("GitHub Authenticated")
		}
	}

	// -----------------------------------------------------------------------
	// Step 3: Project Selection
	// -----------------------------------------------------------------------
	projectPath := ui.PromptProjectPath()
	ui.PrintSuccess(fmt.Sprintf("Project selected: %s\n", projectPath))

	// Check for an existing Vertex-on-Edge configuration in this project.
	hasConfig, configuredProvider, configuredTarget, tfDir := platform.DetectExistingConfig(projectPath)

	action := "edit"
	if hasConfig {
		action = ui.PromptExistingConfig()
	}

	var safeProjectName string
	var exactRepoName string
	var providerID, targetID string
	awsRegion := "ap-south-1"
	var err error
	var spinner *ui.Spinner

	if action == "edit" {
		// -------------------------------------------------------------------
		// Configuration Wizard (edit / first-time setup)
		// -------------------------------------------------------------------

		// Step 4: Framework Detection
		d := detector.New()
		spinner = ui.SpinnerStart("Detecting project stack...")
		info, dErr := d.Detect(projectPath)
		if dErr != nil {
			spinner.Stop(false)
			ui.PrintError("Detection failed", dErr.Error())
			return nil
		}

		if info.HasDetection() {
			spinner.Stop(true)
			ui.PrintDetectionResult(info)
		} else {
			spinner.Stop(false)
			ui.PrintNoDetection(info)
			manualStack := ui.SelectStackManual()
			info.Stacks = append(info.Stacks, models.DetectedStack{Framework: manualStack})
		}

		confirmed := ui.ConfirmDetection()
		if !confirmed {
			manualStack := ui.SelectStackManual()
			info.Stacks = []models.DetectedStack{{Framework: manualStack}}
		}

		primary := info.Primary()
		if primary != nil {
			ui.PromptMissingDetails(primary)
		}

		// Step 5: Cloud Provider Selection
		selectedProviderID := ui.PromptCloudProvider()
		providerID = string(selectedProviderID)

		// AWS-specific authentication and region selection.
		var infraChoices ui.AWSInfraChoices
		if selectedProviderID == platform.ProviderAWS {
			if !dependencies.CheckAndInstall("aws", "AWS CLI") {
				ui.PrintInfo("Continuing without AWS auth. Cloud deployment step will likely fail later.")
			} else {
				awsAuth := aws.New()
				if !awsAuth.IsAuthenticated() {
					ui.PrintInfo("AWS authentication required")
					err := awsAuth.Authenticate()
					if err != nil {
						ui.PrintError("AWS authentication failed", err.Error())
						ui.PrintInfo("Continuing without AWS auth. Cloud deployment step will likely fail later.")
					} else {
						ui.PrintSuccess("AWS Authenticated")
					}
				} else {
					ui.PrintSuccess("AWS Authenticated")
				}
			}

			// Show the connected AWS account ID.
			if callerOut, err := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text").Output(); err == nil {
				ui.PrintInfo(fmt.Sprintf("AWS Account ID: %s", strings.TrimSpace(string(callerOut))))
			}

			// Prompt for AWS region.
			awsRegion = ui.PromptAWSRegion()

			// Prompt for VPC, Subnet, Key Pair.
			infraChoices = ui.PromptAWSInfrastructure()
			infraChoices.AWSRegion = awsRegion
		}

		// Step 6: Target Selection
		selectedTargetID := ui.PromptDeploymentTarget(selectedProviderID)
		targetID = string(selectedTargetID)

		// Step 7: Testing Strategy
		testingFrameworks := ui.PromptTestingFramework(primary.Language, primary.Framework)

		// Step 8: Generate CI/CD Pipeline Files
		ui.PrintInfo("Generating CI/CD pipeline...")

		var genWg sync.WaitGroup
		genWg.Add(3)

		go func() {
			defer genWg.Done()
			if err := docker.GenerateDockerfile(projectPath, primary.Framework); err != nil {
				ui.PrintError("Failed to generate Dockerfile", err.Error())
			} else {
				ui.PrintSuccess("Dockerfile attached to project")
			}
		}()

		go func() {
			defer genWg.Done()
			if err := ghactions.GenerateTestWorkflow(projectPath, primary.Language, testingFrameworks); err != nil {
				ui.PrintError("Failed to generate testing workflow", err.Error())
			} else if len(testingFrameworks) > 0 {
				ui.PrintSuccess(fmt.Sprintf("Testing CI attached (.github/workflows/test.yml) - using %s", strings.Join(testingFrameworks, ", ")))
			}
		}()

		go func() {
			defer genWg.Done()
			if err := ghactions.GenerateDeployWorkflow(projectPath, providerID, targetID); err != nil {
				ui.PrintError("Failed to generate deployment workflow", err.Error())
			} else {
				ui.PrintSuccess("Deployment CI attached (.github/workflows/deploy.yml)")
			}
		}()

		genWg.Wait()

		// Step 9: Resolve GitHub Repository Identity
		ui.PrintInfo(fmt.Sprintf("Configuring deployment for %s on %s...", primary.Framework, targetID))

		exactRepoName = info.Name
		safeProjectName = strings.ToLower(strings.ReplaceAll(info.Name, " ", "-"))

		githubOwner := resolveGitHubOwner(projectPath)
		exactRepoName = resolveGitHubRepoName(projectPath, exactRepoName)

		// Build Terraform variables map.
		vars := map[string]string{
			"aws_region":        awsRegion,
			"project_name":      safeProjectName,
			"instance_type":     "t3.micro",
			"app_port":          "3000",
			"github_owner":      githubOwner,
			"github_repository": exactRepoName,
		}

		// Inject the infrastructure choices into Terraform variables.
		if providerID == "aws" {
			if infraChoices.CreateVPC {
				vars["create_vpc"] = "true"
			} else {
				vars["create_vpc"] = "false"
			}
			if infraChoices.CreateSubnet {
				vars["create_subnet"] = "true"
			} else {
				vars["create_subnet"] = "false"
			}
			if infraChoices.ExistingKeyPairName != "" {
				vars["existing_key_pair_name"] = infraChoices.ExistingKeyPairName
			}
		}

		tfDir, err = terraform.RenderTemplates(projectPath, providerID, targetID, vars)
		if err != nil {
			ui.PrintError("Failed to generate Terraform templates", err.Error())
			return nil
		}
		ui.PrintSuccess(fmt.Sprintf("Terraform templates generated at %s", tfDir))
	} else {
		// -------------------------------------------------------------------
		// Fast Deploy (reuse existing config)
		// -------------------------------------------------------------------
		ui.PrintInfo("Using existing Vertex-on-Edge configuration...")
		providerID = configuredProvider
		targetID = configuredTarget

		exactRepoName = filepath.Base(projectPath)
		githubOwner := resolveGitHubOwner(projectPath)
		exactRepoName = resolveGitHubRepoName(projectPath, exactRepoName)
		safeProjectName = strings.ToLower(strings.ReplaceAll(exactRepoName, " ", "-"))

		// Keep Terraform variables in sync with the current GitHub repo identity,
		// so the IAM trust policy matches on the next apply.
		syncTfVarsGitHub(tfDir, githubOwner, exactRepoName)
	}

	// -----------------------------------------------------------------------
	// Step 10: Terraform Init, Validate, Plan, Apply
	// -----------------------------------------------------------------------
	runner := terraform.NewLocalRunner(tfDir)

	spinner = ui.SpinnerStart("Initializing Terraform...")
	err = runner.Init()
	if err != nil {
		spinner.Stop(false)
		ui.PrintError("Terraform Init Failed (Is Terraform installed?)", err.Error())
		return nil
	}
	spinner.Stop(true)

	spinner = ui.SpinnerStart("Validating Terraform configuration...")
	err = runner.Validate()
	if err != nil {
		spinner.Stop(false)
		ui.PrintError("Terraform Validate Failed", err.Error())
		return nil
	}
	spinner.Stop(true)

	// Reconcile existing AWS resources into the Terraform state.
	if providerID == string(platform.ProviderAWS) {
		if err := ReconcileAWSInfrastructure(runner, safeProjectName, awsRegion); err != nil {
			ui.PrintError("Reconciliation Failed", err.Error())
		}
	}

	ui.PrintInfo("Generating Terraform plan...")
	planOut, err := runner.Plan()
	if err != nil {
		ui.PrintError("Terraform Plan Failed", err.Error())
		return nil
	}

	fmt.Println()
	ui.StreamPrint(planOut)
	fmt.Println()

	if !ui.ConfirmAction("Continue with Terraform apply?") {
		ui.PrintInfo("Deployment aborted by user.")
		os.Exit(0)
	}

	ui.PrintInfo("Applying infrastructure...")
	err = runner.Apply()
	if err != nil {
		ui.PrintError("Terraform Apply Failed", err.Error())
		return nil
	}
	ui.PrintSuccess("Infrastructure deployed successfully!")

	// -----------------------------------------------------------------------
	// Step 11: Post-Apply -- Read outputs and configure GitHub secrets
	// -----------------------------------------------------------------------
	ui.PrintInfo("Extracting infrastructure outputs...")

	appUrl, err := runner.Output("application_url")
	if err != nil {
		ui.PrintError("Failed to get Application URL from Terraform", err.Error())
	}

	instanceIp, ipErr := runner.Output("instance_public_ip")
	sshKey, sshErr := runner.Output("ssh_private_key")

	if providerID == string(platform.ProviderAWS) {
		if ipErr != nil || sshErr != nil || instanceIp == "" || sshKey == "" {
			ui.PrintError("SSH_CREDENTIALS_MISSING", "Failed to read the EC2 IP or SSH key from Terraform outputs.")
			ui.PrintInfo("Check that the Terraform apply completed successfully and that outputs.tf includes ssh_private_key and instance_public_ip.")
			return nil
		}

		fmt.Println("    ✓  SSH Key generated")
		fmt.Printf("    ✓  Host IP: %s\n", strings.TrimSpace(instanceIp))
	}

	// Build the set of GitHub Actions secrets.
	ui.PrintInfo("Securing GitHub Actions environment...")
	secrets := map[string]string{}
	
	if providerID == string(platform.ProviderAWS) {
		secrets["EC2_HOST"] = strings.TrimSpace(instanceIp)
		secrets["EC2_SSH_KEY"] = strings.TrimSpace(sshKey)
	}

	// -----------------------------------------------------------------------
	// Step 12: Push to GitHub and inject secrets
	// -----------------------------------------------------------------------
	spinner = ui.SpinnerStart("Initializing GitHub configuration...")
	err = ghactions.InitAndPush(projectPath, exactRepoName, secrets, func(status string) {
		spinner.Update(status)
	})
	if err != nil {
		spinner.Stop(false)
		ui.PrintError("GitHub repository synchronization failed", err.Error())
		fmt.Println("\n  ✓ AWS infrastructure deployed")
		fmt.Println("  ✓ EC2 instance ready")
		fmt.Println("  ✓ SSH key generated")
		fmt.Println("\n  ✕ GitHub repository synchronization failed")
		fmt.Println("\n  Reason:\n    GitHub authentication failed, missing 'workflow' token scope, or unresolved merge conflicts.")
		fmt.Println("\n  Nothing else needs to be provisioned on AWS.")
		return err
	}

	spinner.Stop(true)
	ui.PrintSuccess("Repository pushed to GitHub with secure CI/CD secrets!")
	ui.PrintSuccess("Vertex-on-Edge Deployment Handoff Complete!")
	if appUrl != "" {
		fmt.Printf("\n  Your application will be live at: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#00B4D8")).Render(strings.TrimSpace(appUrl)))
		fmt.Printf("  (Please allow 3-5 minutes for the first GitHub Actions pipeline to finish building and deploying your container.)\n\n")
	}

	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// resolveGitHubOwner determines the GitHub owner (username or org) for the
// current project. It tries: gh repo view -> gh api user -> git remote -> $USER.
func resolveGitHubOwner(projectPath string) string {
	// Try gh repo view first (works when the repo already exists on GitHub).
	ghRepoViewCmd := exec.Command("gh", "repo", "view", "--json", "owner,name", "-q", ".owner.login + \"/\" + .name")
	ghRepoViewCmd.Dir = projectPath
	if repoInfo, err := ghRepoViewCmd.Output(); err == nil {
		parts := strings.Split(strings.TrimSpace(string(repoInfo)), "/")
		if len(parts) == 2 {
			return parts[0]
		}
	}

	// Fallback: ask the GitHub API for the authenticated user.
	ghApiUserCmd := exec.Command("gh", "api", "user", "-q", ".login")
	if ghUser, err := ghApiUserCmd.Output(); err == nil {
		if owner := strings.TrimSpace(string(ghUser)); owner != "" {
			return owner
		}
	}

	// Fallback: parse from git remote.
	gitRemoteCmd := exec.Command("git", "-C", projectPath, "remote", "get-url", "origin")
	if remoteOut, err := gitRemoteCmd.Output(); err == nil {
		if owner := parseGitHubOwnerFromRemote(strings.TrimSpace(string(remoteOut))); owner != "" {
			return owner
		}
	}

	// Last resort.
	return os.Getenv("USER")
}

// resolveGitHubRepoName determines the GitHub repository name. It prefers
// the remote identity over the local directory name.
func resolveGitHubRepoName(projectPath, fallback string) string {
	ghRepoViewCmd := exec.Command("gh", "repo", "view", "--json", "owner,name", "-q", ".owner.login + \"/\" + .name")
	ghRepoViewCmd.Dir = projectPath
	if repoInfo, err := ghRepoViewCmd.Output(); err == nil {
		parts := strings.Split(strings.TrimSpace(string(repoInfo)), "/")
		if len(parts) == 2 {
			return parts[1]
		}
	}

	// Try git remote origin.
	gitRemoteCmd := exec.Command("git", "-C", projectPath, "remote", "get-url", "origin")
	if remoteOut, err := gitRemoteCmd.Output(); err == nil {
		if repo := parseGitHubRepoFromRemote(strings.TrimSpace(string(remoteOut))); repo != "" {
			return repo
		}
	}

	return fallback
}

// parseGitHubOwnerFromRemote extracts the owner from an HTTPS or SSH GitHub URL.
func parseGitHubOwnerFromRemote(remote string) string {
	if strings.HasPrefix(remote, "https://github.com/") {
		parts := strings.Split(strings.TrimPrefix(remote, "https://github.com/"), "/")
		if len(parts) >= 2 {
			return parts[0]
		}
	} else if strings.HasPrefix(remote, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(remote, "git@github.com:"), "/")
		if len(parts) >= 2 {
			return parts[0]
		}
	}
	return ""
}

// parseGitHubRepoFromRemote extracts the repo name from an HTTPS or SSH GitHub URL.
func parseGitHubRepoFromRemote(remote string) string {
	if strings.HasPrefix(remote, "https://github.com/") {
		parts := strings.Split(strings.TrimPrefix(remote, "https://github.com/"), "/")
		if len(parts) >= 2 {
			return strings.TrimSuffix(parts[1], ".git")
		}
	} else if strings.HasPrefix(remote, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(remote, "git@github.com:"), "/")
		if len(parts) >= 2 {
			return strings.TrimSuffix(parts[1], ".git")
		}
	}
	return ""
}

// syncTfVarsGitHub updates github_owner and github_repository in an existing
// terraform.tfvars file so the trust policy stays in sync when fast-deploying.
func syncTfVarsGitHub(tfDir, githubOwner, exactRepoName string) {
	if githubOwner == "" || exactRepoName == "" {
		return
	}
	tfVarsPath := filepath.Join(tfDir, "terraform.tfvars")
	tfVarsData, err := os.ReadFile(tfVarsPath)
	if err != nil {
		return
	}
	lines := strings.Split(string(tfVarsData), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "github_owner ") || strings.HasPrefix(line, "github_owner=") {
			lines[i] = fmt.Sprintf("github_owner = \"%s\"", githubOwner)
		} else if strings.HasPrefix(line, "github_repository ") || strings.HasPrefix(line, "github_repository=") {
			lines[i] = fmt.Sprintf("github_repository = \"%s\"", exactRepoName)
		}
	}
	os.WriteFile(tfVarsPath, []byte(strings.Join(lines, "\n")), 0644)
}
