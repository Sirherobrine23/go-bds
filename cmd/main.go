package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"

	"sirherobrine23.org/go-bds/go-bds/bedrock/mojang"
)

func start() error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	var server mojang.MojangOverlayfs
	flag.StringVar(&server.VersionsFolder, "versionsroot", filepath.Join(homePath, ".bds/bedrock/versions"), "versions to save root")
	flag.StringVar(&server.Path, "path", filepath.Join(homePath, ".bds/bedrock/bdsserver"), "Path to run server")
	flag.StringVar(&server.SavePath, "save", filepath.Join(dir, "bdsdata"), "data save on")
	flag.StringVar(&server.WorkdirPath, "workdir", filepath.Join(dir, "workdir"), "overlayfs workdir")
	flag.StringVar(&server.Version, "version", "", "Server version")
	flag.Parse()
	defer server.Close()

	server.Handler = &mojang.Handlers{
		Ports:   make([]uint16, 0),
		Players: make([]mojang.PlayerConnection, 0),
	}

	if err := server.Start(); err != nil {
		return err
	}

	go io.Copy(server.ServerProc, os.Stdin)
	server.ServerProc.Stdlog.AddWriter(os.Stdout)

	server.ServerProc.Process.Wait()
	return nil
}

func main() {
	if err := start(); err != nil {
		panic(err)
	}
}
