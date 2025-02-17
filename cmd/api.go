package cmd

import (
	"github.com/spf13/cobra"

	"github.com/sunriselayer/sunrise-data/api"
	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/context"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the API server",
	Long:  `This command starts the API server.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.LoadConfig()
		if err != nil {
			panic(err)
		}

		// TODO check rpc is enabled
		context.GetPublishContext(*config)

		api.Handle()
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)
}
