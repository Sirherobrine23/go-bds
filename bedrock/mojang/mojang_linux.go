//go:build linux

package mojang

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"sirherobrine23.org/go-bds/go-bds/exec"
	"sirherobrine23.org/go-bds/go-bds/overleyfs"
)

type MojangOverlayfs struct {
	VersionsFolder string        // Folder with versions
	Version        string        // Version to run server
	Path           string        // Run server at folder
	SavePath       string        // Folder path to save server run data
	WorkdirPath    string        // Workdir folder to overlayfs
	Handler        *Handlers     // Server handlers
	ServerProc     exec.Proc     // Server process

	Config         *MojangConfig // Server config
	overlayfs *overleyfs.Overlayfs // Overlayfs
}

func checkExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Create backup in tar format
func (server *MojangOverlayfs) TarBackup(w io.Writer) error {
	t := tar.NewWriter(w)
	defer t.Close()
	return t.AddFS(os.DirFS(server.SavePath))
}

// Create backup in zip format
func (server *MojangOverlayfs) ZipBackup(w io.Writer) error {
	z := zip.NewWriter(w)
	defer z.Close()
	return z.AddFS(os.DirFS(server.SavePath))
}

// Stop server if running and umount overlayfs
func (server *MojangOverlayfs) Close() error {
	if server.ServerProc != nil {
		server.ServerProc.Write([]byte("stop\n")) // Stop Server if running
	}

	if server.overlayfs != nil {
		server.overlayfs.Unmount()
		os.RemoveAll(server.overlayfs.Workdir)
	}
	return nil
}

// Start server and mount overlayfs if version not exists localy download
func (server *MojangOverlayfs) Start() error {
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
		os.MkdirAll(versionRoot, os.FileMode(0700))

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
	if server.SavePath == "" {
		return fmt.Errorf("set SavePath to save server data")
	}
	if server.WorkdirPath == "" {
		return fmt.Errorf("set WorkdirPath to make overlayfs temp files on change")
	}

	os.MkdirAll(server.Path, 0700)
	os.MkdirAll(server.SavePath, 0700)
	os.MkdirAll(server.WorkdirPath, 0700)

	// Mount overlayfs to preserver original server and make backups easy
	server.overlayfs = &overleyfs.Overlayfs{
		Target:  server.Path,     // Merged directorys to run server
		Upper:   server.SavePath, // Server run save data
		Workdir: server.WorkdirPath,
		Lower:   []string{versionRoot}, // Original minecraft server extracted folder
	}
	if err := server.overlayfs.Mount(); err != nil {
		return err
	}

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

	if err := server.ServerProc.Start(opt); err != nil {
		return err
	}

	// Handler parse
	if server.Handler != nil {
		log, err := server.ServerProc.StdoutFork()
		if err == nil {
			go server.Handler.RegisterScan(log)
		}
	}

	return nil
}
