//go:build android && (arm64 || amd64)

package main

import (
	"github.com/urfave/cli/v2"
)

func init() {
	mojangCommand.Action = func(ctx *cli.Context) error {
		// var mcserver mojang.MojangProot
		return nil
	}
}
