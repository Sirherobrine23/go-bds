package nukkit

import (
	"fmt"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/exec"
)

type Nukkit struct {
	ServerPath string `json:"serverPath"`
	Version    string `json:"version"`
	MaxMemory  int64  `json:"maxMemory"` // Max memory allocate to Java heap
	MinMemory  int64  `json:"minMemory"` // Minimum memory allocate to initial Java heap
}

func (w *Nukkit) Download() error {
	files, err := ListFiles()
	if err != nil {
		return err
	}

	var build NukkitBuild
	if len(w.Version) > 0 {
		for _, k := range files {
			if k.CommitID == w.Version {
				build = k
				break
			}
		}
	} else {
		build = files[0]
	}

	if err = build.Download(w.ServerPath); err != nil {
		return err
	}

	return nil
}

func (w *Nukkit) Start() (exec.Server, error) {
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
