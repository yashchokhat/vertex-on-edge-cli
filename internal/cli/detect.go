package cli

import (
	"github.com/spf13/cobra"
	"github.com/yashchokhat/vertex-on-edge/internal/detector"
	"github.com/yashchokhat/vertex-on-edge/internal/ui"
	"github.com/yashchokhat/vertex-on-edge/pkg/models"
)

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect the current project's technology stack",
	Long:  "Analyze the current directory to identify the programming language, framework, runtime, and package manager.",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := ui.PromptProjectPath()

		ui.PrintInfo("Analyzing project...")

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

		_ = finalStack // Used for future stages

		return nil
	},
}

func init() {
	rootCmd.AddCommand(detectCmd)
}
