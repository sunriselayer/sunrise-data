/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	//  "github.com/sunriselayer/sunrise-data/cmd"
	"github.com/sunriselayer/sunrise-data/api"
	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/tasks"
)

func main() {
	// cmd.Execute()
	config, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	context.GetContext(*config)

	tasks.RunTasks()

	api.Handle()
}
