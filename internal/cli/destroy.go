package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yashchokhat/vertex-on-edge/internal/ui"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Tear down the deployed cloud infrastructure",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintBanner()
		
		if !ui.ConfirmAction("Are you sure you want to completely destroy the infrastructure? This cannot be undone.") {
			ui.PrintInfo("Destruction aborted.")
			os.Exit(0)
		}

		ui.PrintInfo("Initializing infrastructure teardown...")
		// TODO: locate terraform workspace, run terraform destroy
		fmt.Println("  (Terraform Destroy logic goes here)")
		
		ui.PrintSuccess("Infrastructure successfully destroyed.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
