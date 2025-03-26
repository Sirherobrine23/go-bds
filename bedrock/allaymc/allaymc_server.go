package allaymc

import (
	"archive/tar"
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	os_exec "os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/utils/file_checker"
	"sirherobrine23.com.br/go-bds/go-bds/utils/javaprebuild"
)

// Prepare AllayMC with basic setup to struct
//
// This server not require Overlayfs
func NewAllayMC(version *Version, versionFolder, javaFolder, cwd string) (*AllayMC, error) {
	if version == nil {
		return nil, ErrNoVersion
	}

	serverFile := filepath.Join(versionFolder, version.Version, "server.jar")
	if !file_checker.IsFile(serverFile) {
		if err := version.Dowload(serverFile); err != nil {
			return nil, err
		}
	}

	javaPath := filepath.Join(javaFolder, strconv.Itoa(int(version.JavaVersion)))
	if !file_checker.IsDir(javaPath) {
		err := version.JavaVersion.InstallLatest(javaPath)
		switch err {
		case nil:
			if javaPath = filepath.Join(javaPath, "bin/java"); runtime.GOOS == "windows" {
				javaPath += ".exe"
			}
		case javaprebuild.ErrSystem:
			if javaPath, err = os_exec.LookPath("java"); err != nil {
				return nil, fmt.Errorf("install java in u system: %s", err)
			}
		default:
			return nil, err
		}
	}

	allayMcConfig := &AllayMC{
		PID:     &exec.Os{},
		Version: version,
		ServerStart: exec.ProcExec{
			Cwd: cwd,
			Arguments: []string{
				javaPath,
				"-jar",
				serverFile,
			},
		},
	}

	return allayMcConfig, nil
}

type AllayMC struct {
	PID         exec.Proc     // process status
	ServerStart exec.ProcExec // Server command
	Version     *Version      // Server version
}

// Make server backup with [*archive/tar.Writer]
func (allay AllayMC) Tar(w io.Writer) error {
	tarball := tar.NewWriter(w)
	defer tarball.Close()
	return tarball.AddFS(os.DirFS(allay.ServerStart.Cwd))
}

// Make server backup with [*archive/zip.Writer]
func (allay AllayMC) Zip(w io.Writer) error {
	wr := zip.NewWriter(w)
	defer wr.Close()
	return wr.AddFS(os.DirFS(allay.ServerStart.Cwd))
}

func (allay *AllayMC) Start() error {
	// if server not configured correctly return error
	if allay == nil || allay.PID == nil {
		return errors.New("cannot start server, server proc not defined")
	}
	return allay.PID.Start(allay.ServerStart)
}
