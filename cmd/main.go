package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

var (
	Version string = "devel"
)

func main() {
	app := cli.NewApp()
	app.Name = "go-bds"
	app.Usage = "making minecraft servers easy and powerfull from one command"
	app.EnableBashCompletion = true
	app.HideHelpCommand = true
	app.HideVersion = true
	app.Version = Version
	app.AllowExtFlags = true

	userDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name: "rootdir",
			Value: filepath.Join(userDir, ".bds"),
			Aliases: []string{
				"root",
				"R",
			},
		},
	}

	app.Commands = []*cli.Command{
		&CommandBedrock,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(app.ErrWriter, err.Error())
		os.Exit(1)
	}
}
