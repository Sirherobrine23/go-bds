package java

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/internal/semver"
	"sirherobrine23.com.br/go-bds/go-bds/java/javaprebuild"
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
	Version           string        // Version to run server
	Path              string        // Server folder to run

	OverlayConfig *overlayfs.Overlayfs // Config to overlayfs, go-bds replace necesarys configs

	ServerProc exec.Proc // Interface to process running
}

// Start server
func (w *JavaServer) Start() error {
	if w.VersionsSearch == nil {
		w.VersionsSearch = &MojangSearch{}
	}

	versionPath := filepath.Join(w.VersionsFolder, w.Version)
	ver, err := w.VersionsSearch.Find(w.Version)
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(versionPath, ServerMain)); os.IsNotExist(err) {
		if err = ver.Install(versionPath); err != nil {
			return err
		}
	}

	var processConfig exec.ProcExec
	processConfig.Cwd = w.Path
	processConfig.Arguments = []string{"java", "-jar", ServerMain, "-nogui"}

	if !exec.LocalBinExist(processConfig) {
		javaRoot := filepath.Join(w.JVMVersionsFolder, fmt.Sprint(ver.JavaVersion()))
		if processConfig.Arguments[0] = filepath.Join(javaRoot, "bin/java"); runtime.GOOS == "windows" {
			processConfig.Arguments[0] += ".exe"
		}

		if _, err := os.Stat(processConfig.Arguments[0]); os.IsNotExist(err) {
			if err := javaprebuild.InstallLatest(ver.JavaVersion(), javaRoot); err != nil {
				return err
			}
		}
	}

	w.ServerProc = &exec.Os{}
	if err := w.ServerProc.Start(processConfig); err != nil {
		return err
	}

	return nil
}
