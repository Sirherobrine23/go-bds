//go:build linux

package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"sirherobrine23.org/go-bds/go-bds/bedrock/mojang"
)

func init() {
	originalCommand := mojangCommand.Action
	mojangCommand.Flags = append(mojangCommand.Flags,
		&cli.BoolFlag{
			Name:    "overlay",
			Usage:   "Run server with server overlay and separete data from server",
			Aliases: []string{"o"},
		},
	)

	mojangCommand.Action = func(ctx *cli.Context) (err error) {
		if !ctx.Bool("overlay") {
			return originalCommand(ctx)
		}
		var mcserver = &mojang.MojangOverlayfs{}
		mcserver.VersionsFolder = filepath.Join(ctx.String("rootdir"), "bedrock/versions")

		mcserver.Version = ctx.String("version")
		mcserver.SavePath = ctx.String("data")
		mcserver.WorkdirPath = filepath.Join(ctx.String("data"), "../_mcworkdir")

		if mcserver.Path, err = os.MkdirTemp(os.TempDir(), "bdsmcrun_"); err != nil {
			return err
		}
		defer mcserver.Close()
		defer os.Remove(mcserver.Path)

		if err := mcserver.Start(); err != nil {
			return err
		}

		var stdout, stderr io.ReadCloser
		var stdin io.WriteCloser
		if stdout, err = mcserver.ServerProc.StdoutFork(); err != nil {
			return err
		} else if stderr, err = mcserver.ServerProc.StderrFork(); err != nil {
			stdout.Close()
			return err
		} else if stdin, err = mcserver.ServerProc.StdinFork(); err != nil {
			stdout.Close()
			stderr.Close()
			return err
		}
		defer stdout.Close()
		defer stderr.Close()
		defer stdin.Close()

		go io.Copy(ctx.App.ErrWriter, stderr)
		go io.Copy(ctx.App.Writer, stdout)
		go io.Copy(stdin, ctx.App.Reader)
		return mcserver.ServerProc.Wait()
	}
}
