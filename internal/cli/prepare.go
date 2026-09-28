package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yashchokhat/vertex-on-edge/internal/ui"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare the cloud environment without deploying",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintBanner()
		ui.PrintInfo("Preparing cloud environment...")
		// TODO: Just run terraform init and plan, do not apply
		fmt.Println("  (Environment preparation logic goes here)")
		ui.PrintSuccess("Environment ready for deployment.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(prepareCmd)
}
