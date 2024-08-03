package mojang

import (
	"fmt"
	"runtime"

	"sirherobrine23.org/go-bds/go-bds/exec"
)

type Mojang struct {
	ServerPath string       // Server path to download, run server
	Version    string       // Server version
	Config     MojangConfig // Config server file
}

func (w *Mojang) Download() (*VersionPlatform, error) {
	versions, err := FromVersions()
	if err != nil {
		return nil, err
	}

	goTarget := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	for version, ver := range versions {
		if version == w.Version {
			if target, ok := ver.Platforms[goTarget]; ok {
				return &target, target.Download(w.ServerPath)
			}
			return nil, ErrNoPlatform
		}
	}

	return nil, ErrNoVersion
}

func (w *Mojang) Start() (exec.Proc, error) {
	w.Config = MojangConfig{}
	w.Config.Load(w.ServerPath)

	filename := "./bedrock_server"
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}

	progStr := exec.ProcExec{
		Cwd:       w.ServerPath,
		Arguments: []string{filename},
	}

	var exeProcess exec.Os
	if err := exeProcess.Start(progStr); err != nil {
		return nil, err
	}

	return &exeProcess, nil
}
