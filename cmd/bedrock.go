package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"sirherobrine23.org/go-bds/go-bds/bedrock/mojang"
)

var mojangCommand = cli.Command{
	Name:        "mojang",
	Description: "Run offical minecraft server",
	Aliases:     []string{"m", "oficial"},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "version",
			Usage:   "Minecraft bedrock version to run",
			Aliases: []string{"v"},
			Value:   "latest",
		},
		&cli.StringFlag{
			Name:  "data",
			Usage: "On server run/save worlds and another files",
			Value: "bedrockdata",
		},
	},
	Action: func(ctx *cli.Context) error {
		var mcServer mojang.Mojang
		mcServer.VersionsFolder = filepath.Join(ctx.String("rootdir"), "bedrock/versions")
		mcServer.Version = ctx.String("verison")
		mcServer.Path = ctx.String("data")

		return mcServer.Start()
	},
}

var CommandBedrock = cli.Command{
	Name:        "bedrock",
	Aliases:     []string{"b"},
	Category:    "bedrock",
	Description: "Run minecraft bedrock Server in you system insider docker, proot or host",
	Usage:       "bedrock <server>",
	Subcommands: []*cli.Command{
		&mojangCommand,
		{
			Name:    "versions",
			Aliases: []string{"vv"},
			Usage:   "print versions",
			Action: func(ctx *cli.Context) error {
				versions, err := mojang.FromVersions()
				if err != nil {
					return err
				}
				d, _ := json.MarshalIndent(mojang.GetLatest(versions), "", "  ")
				fmt.Println(string(d))
				return nil
			},
		},
	},
}
