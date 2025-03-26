package bedrock

import (
	"archive/tar"
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"sirherobrine23.com.br/go-bds/go-bds/binfmt"
	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/overlayfs"
)

type VersionSearch func(version string) (*Version, error)

// Make new bedrock config
func NewBedrock(version, versionFolder, cwd string, search VersionSearch) (*Bedrock, error) {
	bedrockConfig := &Bedrock{
		PID:       &exec.Os{},
		Version:   nil,
		Overlayfs: nil,
		serverStart: exec.ProcExec{
			Cwd: cwd,
		},
	}

	bedrockConfig.serverStart.Cwd = cwd
	switch runtime.GOOS {
	case "windows":
		bedrockConfig.serverStart.Arguments = []string{"bedrock_server.exe"}
		search(version)
	case "linux", "android":
		bedrockConfig.serverStart.Environment = map[string]string{"LD_LIBRARY_PATH": "."}
		bedrockConfig.serverStart.Arguments = []string{"./bedrock_server"}
	default:
		return nil, ErrPlatform
	}
	return bedrockConfig, nil
}

type Bedrock struct {
	PID         exec.Proc            // process status
	serverStart exec.ProcExec        // Server command
	Version     *Version             // Server version
	Overlayfs   *overlayfs.Overlayfs // Overlayfs mounted
}

// Start server
func (bed *Bedrock) Start() error {
	// if server not configured correctly return error
	if bed == nil || bed.PID == nil {
		return errors.New("cannot start server, server proc not defined")
	}

	// If overlayfs configured mount before start server
	if bed.Overlayfs != nil {
		if err := bed.Overlayfs.Mount(); err != nil {
			return err
		}
	}

	// Open file and check file format
	fileInfo, err := binfmt.Open(filepath.Join(bed.serverStart.Cwd, bed.serverStart.Arguments[0]))
	if err != nil {
		return err
	} else if fileInfo.GoOs() != runtime.GOOS {
		return ErrPlatform
	}

	// Check if require emulator
	if fileInfo.GoArch() != runtime.GOARCH {
		emulator := binfmt.AsEmulator(fileInfo)
		if emulator == nil {
			return ErrPlatform
		}
		bed.serverStart.Arguments = append(emulator, bed.serverStart.Arguments...)
	}

	// Start server
	if err := bed.PID.Start(bed.serverStart); err != nil {
		return err
	}
	return nil
}

// Make server backup with [*archive/tar.Writer]
//
// If server mounted with [*sirherobrine23.com.br/go-bds/go-bds/overlayfs.Overlayfs] backup only Upper layer
// else backup entire server folder
func (bed Bedrock) Tar(w io.Writer) error {
	tarball := tar.NewWriter(w)
	defer tarball.Close()
	if bed.Overlayfs == nil {
		return tarball.AddFS(os.DirFS(bed.serverStart.Cwd))
	}
	return tarball.AddFS(os.DirFS(bed.Overlayfs.Upper))
}

// Make server backup with [*archive/zip.Writer]
//
// If server mounted with [*sirherobrine23.com.br/go-bds/go-bds/overlayfs.Overlayfs] backup only Upper layer
// else backup entire server folder
func (bed Bedrock) Zip(w io.Writer) error {
	wr := zip.NewWriter(w)
	defer wr.Close()
	if bed.Overlayfs == nil {
		return wr.AddFS(os.DirFS(bed.serverStart.Cwd))
	}
	return wr.AddFS(os.DirFS(bed.Overlayfs.Upper))
}
