package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/context"
)

var rollkitCmd = &cobra.Command{
	Use:   "rollkit",
	Short: "Start the Rollkit server",
	Long:  `This command starts the Rollkit server.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.LoadConfig()
		if err != nil {
			panic(err)
		}

		// TODO check rpc is enabled
		context.GetPublishContext(*config)

		Serve()
	},
}

func init() {
	rootCmd.AddCommand(rollkitCmd)
}
