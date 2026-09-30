package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/yashchokhat/vertex-on-edge/internal/platform"
)

// PromptCloudProvider asks the user to select a cloud provider.
func PromptCloudProvider() platform.ProviderID {
	var selected platform.ProviderID

	options := []huh.Option[platform.ProviderID]{
		huh.NewOption("AWS", platform.ProviderAWS),
		huh.NewOption("Google Cloud", platform.ProviderGCP),
		huh.NewOption("Microsoft Azure", platform.ProviderAzure),
	}

	err := huh.NewSelect[platform.ProviderID]().
		Title("Select cloud provider").
		Options(options...).
		Value(&selected).
		Run()

	if err == huh.ErrUserAborted {
		os.Exit(0)
	}

	return selected
}

// PromptDeploymentTarget asks the user to select a deployment target for the given provider.
func PromptDeploymentTarget(provider platform.ProviderID) platform.TargetID {
	targets := platform.GetTargetsForProvider(provider)
	var options []huh.Option[platform.TargetID]

	for _, t := range targets {
		name := t.Name
		if !t.Available {
			name += " (Coming soon)"
		}
		opt := huh.NewOption(name, t.ID)
		if !t.Available {
			// In huh, you can't easily disable an option, but we can handle it after selection
		}
		options = append(options, opt)
	}

	var selected platform.TargetID
	for {
		err := huh.NewSelect[platform.TargetID]().
			Title("Select deployment target").
			Options(options...).
			Value(&selected).
			Run()

		if err == huh.ErrUserAborted {
			os.Exit(0)
		}

		// Verify availability
		for _, t := range targets {
			if t.ID == selected && !t.Available {
				PrintError("Target not available", fmt.Sprintf("%s is coming soon.", t.Name))
				continue
			}
			if t.ID == selected && t.Available {
				return selected
			}
		}
	}
}

// AWSInfraChoices holds the user's infrastructure preferences for AWS deployment.
type AWSInfraChoices struct {
	CreateVPC           bool
	CreateSubnet        bool
	ExistingKeyPairName string
	AWSRegion           string
}

// PromptAWSRegion asks the user which AWS region to deploy to.
func PromptAWSRegion() string {
	var selected string

	// Fetch the currently configured region as the default suggestion.
	currentRegion := ""
	regionCmd := exec.Command("aws", "configure", "get", "region")
	if out, err := regionCmd.Output(); err == nil {
		currentRegion = strings.TrimSpace(string(out))
	}
	if currentRegion == "" {
		currentRegion = "ap-south-1"
	}

	options := []huh.Option[string]{
		huh.NewOption("ap-south-1  (Mumbai)", "ap-south-1"),
		huh.NewOption("us-east-1   (N. Virginia)", "us-east-1"),
		huh.NewOption("us-west-2   (Oregon)", "us-west-2"),
		huh.NewOption("eu-west-1   (Ireland)", "eu-west-1"),
		huh.NewOption("ap-southeast-1 (Singapore)", "ap-southeast-1"),
	}

	// Put the current region first if it is in the list.
	for i, opt := range options {
		if opt.Value == currentRegion && i != 0 {
			options[0], options[i] = options[i], options[0]
			break
		}
	}

	err := huh.NewSelect[string]().
		Title("Select AWS region").
		Options(options...).
		Value(&selected).
		Run()

	if err == huh.ErrUserAborted {
		os.Exit(0)
	}

	return selected
}

// PromptAWSInfrastructure guides the user through VPC, Subnet, and Key Pair choices.
func PromptAWSInfrastructure() AWSInfraChoices {
	choices := AWSInfraChoices{}

	// --- VPC ---
	var vpcChoice string
	err := huh.NewSelect[string]().
		Title("VPC Configuration").
		Description("A VPC is the network boundary for your EC2 instance.").
		Options(
			huh.NewOption("Use the default VPC (recommended for getting started)", "default"),
			huh.NewOption("Create a new dedicated VPC", "create"),
		).
		Value(&vpcChoice).
		Run()

	if err == huh.ErrUserAborted {
		os.Exit(0)
	}
	choices.CreateVPC = vpcChoice == "create"

	// --- Subnet ---
	if choices.CreateVPC {
		// If we are creating a new VPC, we always need a new subnet.
		choices.CreateSubnet = true
		PrintInfo("A new public subnet will be created inside the new VPC.")
	} else {
		var subnetChoice string
		err = huh.NewSelect[string]().
			Title("Subnet Configuration").
			Description("A subnet determines the availability zone for the instance.").
			Options(
				huh.NewOption("Use a default subnet (recommended)", "default"),
				huh.NewOption("Create a new public subnet in the default VPC", "create"),
			).
			Value(&subnetChoice).
			Run()

		if err == huh.ErrUserAborted {
			os.Exit(0)
		}
		choices.CreateSubnet = subnetChoice == "create"
	}

	// --- Key Pair ---
	// We no longer prompt for existing key pairs because GitHub Actions REQUIRES 
	// the raw private key to SSH into the instance for GHCR deployment.
	// Auto-generating a Terraform-managed key pair is mandatory.
	choices.ExistingKeyPairName = ""

	return choices
}
