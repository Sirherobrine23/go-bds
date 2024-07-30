//go:build !linux || (!bdsexperimental && linux)

package mojang

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gookit/properties"
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

func (w *Mojang) Start() (*exec.Server, error) {
	data, err := os.ReadFile(filepath.Join(w.ServerPath, "server.properties"))
	if err != nil {
		return nil, err
	} else if err = properties.Unmarshal(data, &w.Config); err != nil {
		return nil, err
	}

	filename := "./bedrock_server"
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}

	progStr := &exec.ServerOptions{
		Cwd:       w.ServerPath,
		Arguments: []string{filename},
	}
	exeProcess, err := progStr.Run()

	if err != nil {
		return exeProcess, err
	}

	return exeProcess, err
}
