package cmd

import (
	"github.com/spf13/cobra"

	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/tasks"
)

var validatorCmd = &cobra.Command{
	Use:   "validator",
	Short: "Run validator's proof tasks",
	Long:  `This command starts the validator's proof tasks.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.LoadConfig()
		if err != nil {
			panic(err)
		}

		// TODO check rpc is enabled
		context.GetProofContext(*config)

		tasks.RunTasks()
	},
}

func init() {
	rootCmd.AddCommand(validatorCmd)
}
