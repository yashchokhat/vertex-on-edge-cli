package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/yashchokhat/vertex-on-edge/internal/platform"
)

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
