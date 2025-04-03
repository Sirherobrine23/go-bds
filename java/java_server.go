package java

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	os_exec "os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/utils/file_checker"
	"sirherobrine23.com.br/go-bds/go-bds/utils/javaprebuild"
)

// Prepare folder to server
func NewServer(version Version, versionFolder, javaFolder, cwd string) (*Server, error) {
	if version == nil {
		return nil, ErrNoVersion
	}

	// Check if server exists
	serverFile := filepath.Join(versionFolder, version.Version(), "server.jar")
	if !file_checker.IsFile(serverFile) {
		if err := version.Install(filepath.Dir(serverFile)); err != nil {
			return nil, err
		}
	}

	// eula.txt
	eulaFile := filepath.Join(cwd, "eula.txt")
	if eulaBuff, _ := os.ReadFile(eulaFile); len(eulaBuff) == 0 || !bytes.Contains(eulaBuff, []byte("eula=true")) {
		if eulaBuff = bytes.ReplaceAll(eulaBuff, []byte("false"), []byte("true")); len(eulaBuff) == 0 {
			eulaBuff = []byte("eula=true")
		}
		if err := os.WriteFile(eulaFile, eulaBuff, 0755); err != nil {
			return nil, err
		}
	}

	// Java binary path
	javaPath := filepath.Join(javaFolder, strconv.Itoa(int(version.JavaVersion())))
	switch {
	case file_checker.IsDir(javaPath):
		if javaPath = filepath.Join(javaPath, "bin/java"); runtime.GOOS == "windows" {
			javaPath += ".exe"
		}
	default:
		err := version.JavaVersion().InstallLatest(javaPath)
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

	javaServer := &Server{
		PID:     &exec.Os{},
		Version: version,
		ServerStart: exec.ProcExec{
			Cwd: cwd,
			Arguments: []string{
				javaPath,
				"-jar",
				serverFile,
				"--nogui",
			},
		},
	}

	return javaServer, nil
}

type Server struct {
	PID         exec.Proc     // Struct to start server
	ServerStart exec.ProcExec // Process start
	Version     Version       // Server info

	stopCalled int8 // Server call stop command
}

// Make server backup with [*archive/tar.Writer]
func (javaServer Server) Tar(w io.Writer) error {
	tarball := tar.NewWriter(w)
	defer tarball.Close()
	return tarball.AddFS(os.DirFS(javaServer.ServerStart.Cwd))
}

// Make server backup with [*archive/zip.Writer]
func (javaServer Server) Zip(w io.Writer) error {
	wr := zip.NewWriter(w)
	defer wr.Close()
	return wr.AddFS(os.DirFS(javaServer.ServerStart.Cwd))
}

// Start server
func (javaServer *Server) Start() error {
	// if server not configured correctly return error
	if javaServer == nil || javaServer.PID == nil {
		return errors.New("cannot start server, server proc not defined")
	}

	// Start server
	if err := javaServer.PID.Start(javaServer.ServerStart); err != nil {
		return err
	}
	return nil
}

// Stop server
func (javaServer *Server) Stop() error {
	switch javaServer.stopCalled {
	case 0:
		javaServer.stopCalled = 1
		_, err := javaServer.PID.Write([]byte("stop\n"))
		return err
	case 1:
		javaServer.stopCalled = 2
		return javaServer.PID.Signal(syscall.SIGINT)
	default:
		javaServer.stopCalled = 3
		return javaServer.PID.Signal(syscall.SIGKILL)
	}
}
