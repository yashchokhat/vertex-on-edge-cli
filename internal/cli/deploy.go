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
	RunE:  runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func runDeploy(cmd *cobra.Command, args []string) error {
	ui.RunStartupAnimation()
	ui.PrintBanner()
	ui.PrintWelcome()

	// 1. Startup Check
	ui.PrintInfo("Checking environment...")

	// Check Git
	if _, err := exec.LookPath("git"); err != nil {
		ui.PrintError("Git is not installed", "Git is required for deployments.")
		ui.PrintInfo("Install guide: https://git-scm.com/downloads")
		os.Exit(1)
	}
	fmt.Println("  ✓ Git")

	// Check Docker
	if _, err := exec.LookPath("docker"); err != nil {
		ui.PrintError("Docker is not installed", "Docker is required to build containers.")
		ui.PrintInfo("Install guide: https://docs.docker.com/get-docker/")
		os.Exit(1)
	}
	fmt.Println("  ✓ Docker")

	// Check Terraform
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

	// 2. Authentication
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

	// 3. Project Selection
	projectPath := ui.PromptProjectPath()
	ui.PrintSuccess(fmt.Sprintf("Project selected: %s\n", projectPath))

	// 3.5 Check for existing configuration
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
		// --- CONFIGURATION WIZARD ---

		// 4. Framework Detection
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

		// 5. Cloud Selection
		selectedProviderID := ui.PromptCloudProvider()
		providerID = string(selectedProviderID)

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

			// Phase 1: Verify AWS Access
			if callerOut, err := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text").Output(); err == nil {
				ui.PrintInfo(fmt.Sprintf("AWS Account ID: %s", strings.TrimSpace(string(callerOut))))
			}
		}

		// 6. Target Selection
		selectedTargetID := ui.PromptDeploymentTarget(selectedProviderID)
		targetID = string(selectedTargetID)

		// 7. Testing Strategy
		testingFrameworks := ui.PromptTestingFramework(primary.Language, primary.Framework)

		// 8. Generate Project Files (Dockerfile & GitHub Actions) in parallel
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

		// 9. Cloud Configuration
		ui.PrintInfo(fmt.Sprintf("Configuring deployment for %s on %s...", primary.Framework, targetID))

		exactRepoName = info.Name
		safeProjectName = strings.ToLower(strings.ReplaceAll(info.Name, " ", "-"))

		// Phase 2: Detect GitHub Repository (do not hardcode owner or assume directory name matches repo)
		githubOwner := os.Getenv("USER") // Global Fallback if all GitHub APIs fail
		
		// Attempt to get the actual remote repository identity if it exists
		ghRepoViewCmd := exec.Command("gh", "repo", "view", "--json", "owner,name", "-q", ".owner.login + \"/\" + .name")
		ghRepoViewCmd.Dir = projectPath
		if repoInfo, err := ghRepoViewCmd.Output(); err == nil {
			parts := strings.Split(strings.TrimSpace(string(repoInfo)), "/")
			if len(parts) == 2 {
				githubOwner = parts[0]
				exactRepoName = parts[1] // Override the local directory name with the true remote repo name
			}
		} else {
			// Fallback for brand new projects that aren't on GitHub yet
			ghApiUserCmd := exec.Command("gh", "api", "user", "-q", ".login")
			if ghUser, err := ghApiUserCmd.Output(); err == nil {
				githubOwner = strings.TrimSpace(string(ghUser))
			}
			
			// Try to extract from git remote if gh repo view failed
			gitRemoteCmd := exec.Command("git", "-C", projectPath, "remote", "get-url", "origin")
			if remoteOut, err := gitRemoteCmd.Output(); err == nil {
				remoteStr := strings.TrimSpace(string(remoteOut))
				// Handle both HTTPS and SSH urls
				if strings.HasPrefix(remoteStr, "https://github.com/") {
					parts := strings.Split(strings.TrimPrefix(remoteStr, "https://github.com/"), "/")
					if len(parts) >= 2 {
						githubOwner = parts[0]
						exactRepoName = strings.TrimSuffix(parts[1], ".git")
					}
				} else if strings.HasPrefix(remoteStr, "git@github.com:") {
					parts := strings.Split(strings.TrimPrefix(remoteStr, "git@github.com:"), "/")
					if len(parts) >= 2 {
						githubOwner = parts[0]
						exactRepoName = strings.TrimSuffix(parts[1], ".git")
					}
				}
			}
		}

		vars := map[string]string{
			"aws_region":        awsRegion,
			"project_name":      safeProjectName,
			"instance_type":     "t3.micro",
			"app_port":          "3000",
			"github_owner":      githubOwner,
			"github_repository": exactRepoName,
		}



		tfDir, err = terraform.RenderTemplates(projectPath, providerID, targetID, vars)
		if err != nil {
			ui.PrintError("Failed to generate Terraform templates", err.Error())
			return nil
		}
		ui.PrintSuccess(fmt.Sprintf("Terraform templates generated at %s", tfDir))
	} else {
		// --- FAST DEPLOY ---
		ui.PrintInfo("Using existing Vertex-on-Edge configuration...")
		providerID = configuredProvider
		targetID = configuredTarget

		githubOwner := os.Getenv("USER")
		exactRepoName = filepath.Base(projectPath)
		fastRepoViewCmd := exec.Command("gh", "repo", "view", "--json", "owner,name", "-q", ".owner.login + \"/\" + .name")
		fastRepoViewCmd.Dir = projectPath
		if repoInfo, err := fastRepoViewCmd.Output(); err == nil {
			parts := strings.Split(strings.TrimSpace(string(repoInfo)), "/")
			if len(parts) == 2 {
				githubOwner = parts[0]
				exactRepoName = parts[1]
			}
		} else {
			// Fallback for brand new projects that aren't on GitHub yet
			ghApiUserCmd := exec.Command("gh", "api", "user", "-q", ".login")
			if ghUser, err := ghApiUserCmd.Output(); err == nil {
				githubOwner = strings.TrimSpace(string(ghUser))
			}
			
			// Try to extract from git remote if gh repo view failed
			gitRemoteCmd := exec.Command("git", "-C", projectPath, "remote", "get-url", "origin")
			if remoteOut, err := gitRemoteCmd.Output(); err == nil {
				remoteStr := strings.TrimSpace(string(remoteOut))
				// Handle both HTTPS and SSH urls
				if strings.HasPrefix(remoteStr, "https://github.com/") {
					parts := strings.Split(strings.TrimPrefix(remoteStr, "https://github.com/"), "/")
					if len(parts) >= 2 {
						githubOwner = parts[0]
						exactRepoName = strings.TrimSuffix(parts[1], ".git")
					}
				} else if strings.HasPrefix(remoteStr, "git@github.com:") {
					parts := strings.Split(strings.TrimPrefix(remoteStr, "git@github.com:"), "/")
					if len(parts) >= 2 {
						githubOwner = parts[0]
						exactRepoName = strings.TrimSuffix(parts[1], ".git")
					}
				}
			}
		}
		safeProjectName = strings.ToLower(strings.ReplaceAll(exactRepoName, " ", "-"))

		// Synchronize Terraform variables to the current GitHub repository context
		// This strictly guarantees the IAM Trust Policy will accept OIDC requests from the current repo
		if githubOwner != "" && exactRepoName != "" {
			tfVarsPath := filepath.Join(tfDir, "terraform.tfvars")
			if tfVarsData, err := os.ReadFile(tfVarsPath); err == nil {
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
		}
	}

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

	// 10. Extract Outputs & Configure CI/CD
	ui.PrintInfo("Extracting infrastructure outputs...")



	if providerID == string(platform.ProviderAWS) {
		// Validate OIDC Provider exists
		checkCmd := exec.Command("aws", "iam", "list-open-id-connect-providers", "--query", "OpenIDConnectProviderList[*].Arn", "--output", "text")
		if out, err := checkCmd.Output(); err == nil {
			outputStr := string(out)
			if !strings.Contains(outputStr, "token.actions.githubusercontent.com") {
				ui.PrintError("OIDC_PROVIDER_MISSING", "The GitHub Actions OIDC provider could not be found in AWS. Ensure your AWS account permits OIDC provider creation.")
				return nil
			}
		} else {
			ui.PrintError("OIDC_PROVIDER_MISSING", "Failed to query AWS for OIDC providers: " + err.Error())
			return nil
		}

		// Validate global IAM Role Trust Policy
		roleName := "vertexOnEdge-cli"
		trustCmd := exec.Command("aws", "iam", "get-role", "--role-name", roleName, "--query", "Role.AssumeRolePolicyDocument", "--output", "json")
		if out, err := trustCmd.Output(); err == nil {
			if !strings.Contains(string(out), "token.actions.githubusercontent.com") {
				ui.PrintError("OIDC_ROLE_TRUST_INVALID", fmt.Sprintf("The IAM role '%s' does not correctly reference the GitHub OIDC provider.", roleName))
				return nil
			}
		} else {
			ui.PrintError("OIDC_ROLE_MISSING", "Failed to retrieve global IAM role trust policy (vertexOnEdge-cli): " + err.Error())
			return nil
		}

		fmt.Println("    ✓  OIDC configured successfully")
	}

	appUrl, err := runner.Output("application_url")
	if err != nil {
		ui.PrintError("Failed to get Application URL from Terraform", err.Error())
	}
	
	instanceId, err := runner.Output("instance_id")
	if err != nil {
		// Non-fatal, just a warning if it doesn't exist yet
	}

	ui.PrintInfo("Securing GitHub Actions environment...")
	secrets := map[string]string{
		"AWS_REGION":          awsRegion,
		"ECR_REPOSITORY_NAME": safeProjectName,
	}
	
	if instanceId != "" {
		secrets["EC2_INSTANCE_ID"] = strings.TrimSpace(instanceId)
	}

	// Inject the global vertexOnEdge-cli IAM role ARN
	accountIDOut, err := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text").Output()
	if err == nil {
		accountID := strings.TrimSpace(string(accountIDOut))
		if accountID != "" {
			secrets["AWS_ROLE_ARN"] = "arn:aws:iam::" + accountID + ":role/vertexOnEdge-cli"
		}
	}



	spinner = ui.SpinnerStart("Initializing GitHub configuration...")
	err = ghactions.InitAndPush(projectPath, exactRepoName, secrets, func(status string) {
		spinner.Update(status)
	})
	if err != nil {
		spinner.Stop(false)
		ui.PrintError("GitHub repository synchronization failed", err.Error())
		fmt.Println("\n  ✓ AWS infrastructure deployed")
		fmt.Println("  ✓ EC2 instance ready")
		fmt.Println("  ✓ ECR registry ready")
		fmt.Println("  ✓ IAM OIDC role ready")
		fmt.Println("\n  ✕ GitHub repository synchronization failed")
		fmt.Println("\n  Reason:\n    GitHub authentication failed, missing 'workflow' token scope, or unresolved merge conflicts.")
		fmt.Println("\n  Nothing else needs to be provisioned on AWS.")
		// We could add a resume command here later
		return err
	} else {
		spinner.Stop(true)
		ui.PrintSuccess("Repository pushed to GitHub with secure CI/CD secrets!")
		ui.PrintSuccess("Vertex-on-Edge Deployment Handoff Complete!")
		if appUrl != "" {
			fmt.Printf("\n  Your application will be live at: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#00B4D8")).Render(strings.TrimSpace(appUrl)))
			fmt.Printf("  (Please allow 3-5 minutes for the first GitHub Actions pipeline to finish building and deploying your container.)\n\n")
		}
	}

	return nil
}
