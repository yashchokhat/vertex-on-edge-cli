package cli

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/yashchokhat/vertex-on-edge/internal/config"
	"github.com/yashchokhat/vertex-on-edge/internal/detector"
	"github.com/yashchokhat/vertex-on-edge/internal/ui"
	"github.com/yashchokhat/vertex-on-edge/pkg/models"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   config.CLIBinary,
	Short: config.AppName + " - " + config.AppTagline,
	Long: config.AppName + "\n" + config.AppTagline + "\n\n" +
		"Deploy your application to your own cloud infrastructure\n" +
		"without needing to manually learn Docker, CI/CD, or cloud infrastructure.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.RunStartupAnimation() // Clears screen
		ui.PrintBanner()
		ui.PrintWelcome()

		if !ui.PromptTermsAcceptance() {
			os.Exit(0)
		}

		ui.ClearScreen()

		// Detect project
		return runDetection()
	},
}

func init() {
	cfg = config.DefaultConfig()
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// runDetection performs project detection on the current directory.
func runDetection() error {
	path := ui.PromptProjectPath()
	
	sp := ui.SpinnerStart("Detecting project stack...")

	d := detector.New()
	info, err := d.Detect(path)

	if err != nil {
		sp.Stop(false)
		ui.PrintError("Detection failed.", err.Error())
		return err
	}

	sp.Stop(true)

	var finalStack models.DetectedStack

	if info.HasDetection() {
		ui.PrintDetectionResult(info)
		if ui.ConfirmDetection() {
			ui.PrintSuccess("Detected successfully.")
			finalStack = *info.Primary()
			ui.PromptMissingDetails(&finalStack)
		} else {
			manualStack := ui.SelectStackManual()
			finalStack = models.DetectedStack{Framework: manualStack}
			ui.PromptMissingDetails(&finalStack)
		}
	} else {
		ui.PrintNoDetection(info)
		manualStack := ui.SelectStackManual()
		finalStack = models.DetectedStack{Framework: manualStack}
		ui.PromptMissingDetails(&finalStack)
	}

	_ = finalStack // Used for future stages (e.g. Docker generation)

	return nil
}
