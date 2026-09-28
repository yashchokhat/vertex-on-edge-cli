package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yashchokhat/vertex-on-edge/internal/ui"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current status of the deployed application",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintBanner()
		ui.PrintInfo("Checking deployment status...")
		// TODO: read local state, run terraform show / AWS API queries
		fmt.Println("  Infrastructure: Active")
		fmt.Println("  URL: http://example.com")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
