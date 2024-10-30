package java

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/internal/semver"
	"sirherobrine23.com.br/go-bds/go-bds/java/adoptium"
	"sirherobrine23.com.br/go-bds/go-bds/overlayfs"
)

var (
	ErrInstallServer  error = errors.New("install server fist")
	ErrNoServer       error = errors.New("cannot find server")
	ErrNoFoundVersion error = errors.New("version not found")
)

const ServerMain string = "server.jar"

type Version interface {
	Install(path string) error      // Install server in path
	JavaVersion() uint              // Java version to Run server
	SemverVersion() *semver.Version // Platform version
}

type VersionSearch interface {
	Find(version string) (Version, error) // Find version and return Installer, if not exists return Version not found
}

// Global struct to Minecraft java server to run .jar
type JavaServer struct {
	VersionsFolder    string        // Folder with versions
	JVMVersionsFolder string        // Folder with versions
	VersionsSearch    VersionSearch // Struct Find server, default is from Mojang
	WorkdirPath       string        // Path Workdir to Overlayfs
	UpperPath         string        // Path to save diff changes files, only platforms required's and same filesystem to 'Path'
	Version           string        // Version to run server
	Path              string        // Server folder to run

	OverlayConfig *overlayfs.Overlayfs // Config to overlayfs, go-bds replace necesarys configs

	ServerProc exec.Proc // Interface to process running
}

// Start server
func (server *JavaServer) Start() error {
	if server.VersionsSearch == nil {
		server.VersionsSearch = &MojangSearch{}
	}

	versionRoot := filepath.Join(server.VersionsFolder, server.Version)
	ver, err := server.VersionsSearch.Find(server.Version)
	if err != nil {
		return err
	}

	var processConfig exec.ProcExec
	processConfig.Cwd = server.Path
	processConfig.Arguments = []string{"java", "-jar", ServerMain, "-nogui"}

	// Prepare overlayfs/mergefs configuration
	server.OverlayConfig = &overlayfs.Overlayfs{
		Target:  server.Path,        // Target path to merged folder
		Workdir: server.WorkdirPath, // Path to linux overlayfs
		Upper:   server.UpperPath,   // Path to save Target diff
		Lower:   []string{versionRoot},
	}

	javaRoot := filepath.Join(server.JVMVersionsFolder, fmt.Sprint(ver.JavaVersion()))
	if _, err := os.Stat(processConfig.Arguments[0]); os.IsNotExist(err) {
		if err := adoptium.InstallLatest(ver.JavaVersion(), javaRoot); err != nil && err != adoptium.ErrSystem {
			return err
		}
	}

	if _, err := os.Stat(javaRoot); !os.IsNotExist(err) {
		if processConfig.Arguments[0] = filepath.Clean("./bin/java"); runtime.GOOS == "windows" {
			processConfig.Arguments[0] += ".exe"
		}
		server.OverlayConfig.Lower = append(server.OverlayConfig.Lower, javaRoot)
	}

	// Mount overlayfs if avaible
	if err := server.OverlayConfig.Mount(); err != nil && err != overlayfs.ErrNotOverlayAvaible {
		return err // Return if another any error
	} else if err == overlayfs.ErrNotOverlayAvaible {
		CopyFS := func(dir string, fsys fs.FS) error {
			return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				fpath, err := filepath.Localize(path)
				if err != nil {
					return err
				}
				newPath := filepath.Join(dir, fpath)
				if d.IsDir() {
					return os.MkdirAll(newPath, 0777)
				} else if !d.Type().IsRegular() {
					return nil
				}

				r, err := fsys.Open(path)
				if err != nil {
					return err
				}
				defer r.Close()
				info, err := r.Stat()
				if err != nil {
					return err
				}
				w, err := os.OpenFile(newPath, os.O_CREATE|os.O_EXCL|os.O_TRUNC|os.O_WRONLY, 0666|info.Mode()&0777)
				if err != nil {
					return err
				}

				if _, err := io.Copy(w, r); err != nil {
					w.Close()
					return &fs.PathError{Op: "Copy", Path: newPath, Err: err}
				}
				return w.Close()
			})
		}

		if _, err := os.Stat(filepath.Join(versionRoot, ServerMain)); os.IsNotExist(err) {
			if err = ver.Install(versionRoot); err != nil {
				return err
			}
		}

		if err := CopyFS(server.Path, server.OverlayConfig.MergeFS()); err != nil {
			return err
		}
	} else {
		// Install server before mount java bins
		if _, err := os.Stat(filepath.Join(versionRoot, ServerMain)); os.IsNotExist(err) {
			if err = ver.Install(versionRoot); err != nil {
				return err
			}
		}
	}

	server.ServerProc = &exec.Os{}
	if err := server.ServerProc.Start(processConfig); err != nil {
		return err
	}

	return nil
}
