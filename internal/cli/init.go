package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yashchokhat/vertex-on-edge/internal/detector"
	"github.com/yashchokhat/vertex-on-edge/internal/ui"
	"github.com/yashchokhat/vertex-on-edge/pkg/models"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Vertex-on-Edge configuration in a project",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	ui.PrintBanner()
	
	projectPath := ui.PromptProjectPath()
	
	d := detector.New()
	spinner := ui.SpinnerStart("Detecting project stack...")
	info, err := d.Detect(projectPath)
	if err != nil {
		spinner.Stop(false)
		ui.PrintError("Detection failed", err.Error())
		os.Exit(1)
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

	ui.PrintSuccess(fmt.Sprintf("Initialized Vertex-on-Edge for %s", projectPath))
	return nil
}
