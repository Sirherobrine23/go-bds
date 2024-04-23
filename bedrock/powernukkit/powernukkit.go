package powernukkit

import (
	"fmt"
	"path/filepath"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/exec"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

type Powernukkit struct {
	ServerPath string `json:"serverPath"`
	Version    string `json:"version"`
	MaxMemory  int64  `json:"maxMemory"` // Max memory allocate to Java heap
	MinMemory  int64  `json:"minMemory"` // Minimum memory allocate to initial Java heap
}

func (w *Powernukkit) Download() error {
	releases, err := Releases()
	if err != nil {
		return err
	}

	serverRelease, ok := Version{}, false
	if serverRelease, ok = releases[w.Version]; !ok {
		return fmt.Errorf("%s not exists in release", w.Version)
	}

	req := request.RequestOptions{HttpError: true, Url: serverRelease.File}
	if err = req.File(filepath.Join(w.ServerPath, "server.jar")); err != nil {
		return err
	}

	return nil
}

func (w *Powernukkit) Start() (exec.Server, error) {
	// Check memory
	if w.MinMemory > w.MaxMemory || w.MaxMemory <= 0 || w.MinMemory <= 0 {
		w.MaxMemory, w.MinMemory = 1024, 1024
	}

	// Create exec command
	opts := exec.ServerOptions{
		Cwd: w.ServerPath,
		Arguments: []string{
			"java",
			"-jar",
			fmt.Sprintf("-Xmx%dM", w.MaxMemory),
			fmt.Sprintf("-Xms%dM", w.MaxMemory),
			"server.jar",
			"--nogui",
		},
	}

	// Start server
	run, err := opts.Run()
	if err != nil {
		return exec.Server{}, err
	}

	// add server struct here

	return run, nil
}
