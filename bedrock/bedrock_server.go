package bedrock

import (
	"archive/tar"
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"sirherobrine23.com.br/go-bds/go-bds/binfmt"
	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/overlayfs"
)

func isClean(path string) bool {
	entrys, _ := os.ReadDir(path)
	return len(entrys) == 0
}

// Make new bedrock config
func NewBedrock(version *Version, versionFolder, cwd, upper, workdir string) (*Bedrock, error) {
	if version == nil {
		return nil, ErrNoVersion
	}

	bedrockConfig := &Bedrock{
		PID:       &exec.Os{},
		Version:   version,
		Overlayfs: nil,
		serverStart: exec.ProcExec{
			Cwd:         cwd,
			Arguments:   []string{},
			Environment: exec.Env{},
		},
	}

	// Folder path to storage server version
	versionFolder = filepath.Join(versionFolder, version.Version)

	// Correct config to GOOS
	switch runtime.GOOS {
	case "windows":
		bedrockConfig.serverStart.Arguments = []string{"bedrock_server.exe"}
		target, ok := version.Plaforms["windows/amd64"]
		if !ok {
			return nil, ErrNoVersion
		}

		if isClean(versionFolder) {
			if err := target.Extract(versionFolder); err != nil {
				return nil, err
			}
		}
	case "linux", "android":
		bedrockConfig.serverStart.Environment = exec.Env{"LD_LIBRARY_PATH": "."}
		bedrockConfig.serverStart.Arguments = []string{"./bedrock_server"}

		target, ok := version.Plaforms[fmt.Sprintf("linux/%s", runtime.GOARCH)]
		if !ok {
			if target, ok = version.Plaforms["linux/amd64"]; !ok {
				return nil, ErrNoVersion
			}
		}

		if isClean(versionFolder) {
			if err := target.Extract(versionFolder); err != nil {
				return nil, err
			}
		}
	default:
		return nil, ErrPlatform
	}

	// Check to overlayfs is avaible
	if overlayfs.OverlayfsAvaible() {
		bedrockConfig.Overlayfs = &overlayfs.Overlayfs{
			Target:  cwd,
			Upper:   upper,
			Workdir: workdir,
			Lower: []string{
				versionFolder,
			},
		}
	} else if isClean(cwd) {
		if err := os.CopyFS(cwd, os.DirFS(versionFolder)); err != nil {
			return nil, err
		}
	} else {
		// Compare versions
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
