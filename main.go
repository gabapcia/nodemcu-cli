package main

import (
	"log"
	"os"

	"github.com/GabrielBG99/nodemcu-cli/command"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "NodeMCU CLI",
		Usage:                "Manage your NodeMCU boards in your command line",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			command.Upload(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
