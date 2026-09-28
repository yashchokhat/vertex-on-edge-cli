package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yashchokhat/vertex-on-edge/internal/config"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of " + config.AppName,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s v%s\n", config.AppName, config.Version)
		if config.Commit != "dev" {
			fmt.Printf("Commit:     %s\n", config.Commit)
			fmt.Printf("Built:      %s\n", config.BuildDate)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
