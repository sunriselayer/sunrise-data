/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	//  "github.com/sunriselayer/sunrise-data/cmd"
	"github.com/sunriselayer/sunrise-data/api"
	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/tasks"
)

func main() {
	context.GetContext()

	// cmd.Execute()
	tasks.RunTasks()

	api.Handle()
}
