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

		var oidcProviderArn string
		var oidcWg sync.WaitGroup

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

			oidcWg.Add(1)
			go func() {
				defer oidcWg.Done()
				checkCmd := exec.Command("aws", "iam", "list-open-id-connect-providers", "--query", "OpenIDConnectProviderList[*].Arn", "--output", "text")
				if out, err := checkCmd.Output(); err == nil {
					outputStr := string(out)
					parts := strings.Fields(outputStr)
					for _, arn := range parts {
						if strings.Contains(arn, "token.actions.githubusercontent.com") {
							oidcProviderArn = arn
							break
						}
					}
				}
			}()
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

		safeProjectName = strings.ToLower(strings.ReplaceAll(info.Name, " ", "-"))

		vars := map[string]string{
			"aws_region":    awsRegion,
			"project_name":  safeProjectName,
			"instance_type": "t3.micro",
			"app_port":      "3000",
			"github_repo":   "yashchokhat/" + safeProjectName,
		}

		if selectedProviderID == platform.ProviderAWS {
			oidcWg.Wait()
			if oidcProviderArn != "" {
				vars["create_oidc_provider"] = "false"
				vars["oidc_provider_arn"] = oidcProviderArn
				// Patch the existing OIDC provider to ensure it has the latest GitHub thumbprints
				// This fixes "Not authorized to perform sts:AssumeRoleWithWebIdentity" if the user has an old provider
				exec.Command("aws", "iam", "update-open-id-connect-provider-thumbprint", "--open-id-connect-provider-arn", oidcProviderArn, "--thumbprint-list", "6938fd4d98bab03faadb97b34396831e3780aea1", "1c58a3a8518e8759bf075b76b750d4f2df264fcd", "1b511abead59c6ce207077c0bf0e0043b1382612").Run()
			}
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
		safeProjectName = strings.ToLower(strings.ReplaceAll(filepath.Base(projectPath), " ", "-"))
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

	roleArn, err := runner.Output("github_actions_role_arn")
	if err != nil {
		ui.PrintError("Failed to get IAM Role ARN from Terraform", err.Error())
	}

	appUrl, err := runner.Output("application_url")
	if err != nil {
		ui.PrintError("Failed to get Application URL from Terraform", err.Error())
	}

	ui.PrintInfo("Securing GitHub Actions environment...")
	secrets := map[string]string{
		"AWS_REGION":          awsRegion,
		"ECR_REPOSITORY_NAME": safeProjectName,
	}

	if roleArn != "" {
		secrets["AWS_ROLE_ARN"] = strings.TrimSpace(roleArn)
	}

	spinner = ui.SpinnerStart("Initializing GitHub configuration...")
	err = ghactions.InitAndPush(projectPath, safeProjectName, secrets, func(status string) {
		spinner.Update(status)
	})
	if err != nil {
		spinner.Stop(false)
		ui.PrintError("Failed to configure GitHub repository", err.Error())
	} else {
		spinner.Stop(true)
		ui.PrintSuccess("Repository pushed to GitHub with secure CI/CD secrets!")
	}

	ui.PrintSuccess("Vertex-on-Edge Deployment Handoff Complete!")
	if appUrl != "" {
		fmt.Printf("\n  Your application will be live at: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#00B4D8")).Render(strings.TrimSpace(appUrl)))
		fmt.Printf("  (Please allow 3-5 minutes for the first GitHub Actions pipeline to finish building and deploying your container.)\n\n")
	}

	return nil
}
