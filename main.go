/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	//  "github.com/sunriselayer/sunrise-data/cmd"
	"github.com/sunriselayer/sunrise-data/api"
	"github.com/sunriselayer/sunrise-data/config"
)

func main() {
	// cmd.Execute()
	config, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	api.Handle(*config)
}
