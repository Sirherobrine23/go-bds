package mojang

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"sirherobrine23.org/go-bds/go-bds/exec"
)

type Mojang struct {
	VersionsFolder string        // Folder with versions
	Version        string        // Version to run server
	Path           string        // Run server at folder
	Config         *MojangConfig // Server config

	ServerProc exec.Proc
}

// Start server and mount overlayfs if version not exists localy download
func (server *Mojang) Start() error {
	if server.Version == "latest" || server.Version == "" {
		versions, err := FromVersions()
		if err != nil {
			return fmt.Errorf("cannot get versions: %s", err.Error())
		}
		server.Version = GetLatest(versions)
	}

	versionRoot := filepath.Join(server.VersionsFolder, server.Version)
	if checkExist(versionRoot) {
		n, err := os.ReadDir(versionRoot)
		if err != nil {
			return err
		}
		if len(n) == 0 {
			if err := os.RemoveAll(versionRoot); err != nil {
				return err
			}
		}
	}

	if !checkExist(versionRoot) {
		versions, err := FromVersions()
		if err != nil {
			return fmt.Errorf("cannot get versions: %s", err.Error())
		}
		os.MkdirAll(versionRoot, 0700)

		var ok bool
		var version Version
		var target VersionPlatform
		if version, ok = versions[server.Version]; !ok {
			return fmt.Errorf("version not found in database")
		} else if target, ok = version.Platforms[fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)]; !ok {
			return fmt.Errorf("platform not supported")
		} else if err := target.Download(versionRoot); err != nil {
			return err
		}
	}

	if server.Path == "" {
		return fmt.Errorf("set Path to run minecraft server")
	}
	os.MkdirAll(server.Path, 0700)

	// Load config
	server.Config = &MojangConfig{}
	if err := server.Config.Load(server.Path); err != nil {
		return err
	}

	// Start server
	server.ServerProc = &exec.Os{}
	opt := exec.ProcExec{
		Cwd:         server.Path,
		Arguments:   []string{"./bedrock_server"},
		Environment: map[string]string{"LD_LIBRARY_PATH": "."},
	}

	if runtime.GOOS == "windows" {
		if !(runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64") {
			return fmt.Errorf("run minecraft server in Windows with x64/amd64 or arm64 emulation")
		}
		opt.Environment = make(map[string]string)
		opt.Arguments = []string{"bedrock_server.exe"}
	}

	if err := server.ServerProc.Start(opt); err != nil {
		return err
	}

	return nil
}
