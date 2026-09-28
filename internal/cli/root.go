package cli

import (
	"github.com/spf13/cobra"
	"github.com/yashchokhat/vertex-on-edge/internal/config"
	"github.com/yashchokhat/vertex-on-edge/internal/ui"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "vertex-on-edge",
	Short: "Deploy your application to your cloud",
	Long: `Vertex-on-Edge
Infrastructure without the DevOps overhead.

Deploy your application to your cloud.
GitHub → github.com/yashchokhat/vertex-on-edge`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no arguments provided, just show the banner and help
		ui.RunStartupAnimation()
		ui.PrintBanner()
		ui.PrintWelcome()
		return cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cfg = config.DefaultConfig()
	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&cfg.NoAnimation, "no-animation", false, "Disable terminal animations")
	rootCmd.PersistentFlags().BoolVar(&cfg.NoColor, "no-color", false, "Disable colored output")
}
