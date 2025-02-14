/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"flag"

	"github.com/rs/zerolog"

	"github.com/sunriselayer/sunrise-data/api"
	"github.com/sunriselayer/sunrise-data/cmd"
	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/tasks"
)

func main() {
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// cmd.Execute()
	config, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// TODO check rpc is enabled
	context.GetContext(*config)

	tasks.RunTasks()

	// start grpc server for rollkit
	go cmd.Serve()

	api.Handle()
}
