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
	"slices"

	"sirherobrine23.com.br/go-bds/go-bds/binfmt"
	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/overlayfs"
	"sirherobrine23.com.br/go-bds/go-bds/utils/file_checker"
	"sirherobrine23.com.br/go-bds/go-bds/utils/js_types"
)

// Make new bedrock config
func NewBedrock(version *Version, versionFolder, cwd, upper, workdir string) (*Bedrock, error) {
	if version == nil {
		return nil, ErrNoVersion
	}

	bedrockConfig := &Bedrock{
		PID:       &exec.Os{},
		Version:   version,
		Overlayfs: nil,
		ServerStart: exec.ProcExec{
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
		bedrockConfig.ServerStart.Arguments = []string{"bedrock_server.exe"}
		target, ok := version.Plaforms["windows/amd64"]
		if !ok {
			return nil, ErrNoVersion
		}

		if file_checker.FolderIsEmpty(versionFolder) {
			if err := target.Extract(versionFolder); err != nil {
				return nil, err
			}
		}
	case "linux", "android":
		bedrockConfig.ServerStart.Environment = exec.Env{"LD_LIBRARY_PATH": "."}
		bedrockConfig.ServerStart.Arguments = []string{"./bedrock_server"}

		target, ok := version.Plaforms[fmt.Sprintf("linux/%s", runtime.GOARCH)]
		if !ok {
			if target, ok = version.Plaforms["linux/amd64"]; !ok {
				return nil, ErrNoVersion
			}
		}

		if file_checker.FolderIsEmpty(versionFolder) {
			if err := target.Extract(versionFolder); err != nil {
				return nil, err
			}
		}
	default:
		return nil, ErrPlatform
	}

	// Copy server if
	if !overlayfs.OverlayfsAvaible() && !file_checker.FolderIsEmpty(cwd) {
		// Files to not delete
		filesToIgnore := []string{"allowlist.json", "permissions.json", "server.properties", "worlds"}

		// List current fileNames
		fileNames, err := os.ReadDir(cwd)
		if err != nil {
			return nil, err
		}
		fileNames = js_types.Slice[os.DirEntry](fileNames).Filter(func(input os.DirEntry) bool { return !slices.Contains(filesToIgnore, input.Name()) })

		for _, fileToDelete := range fileNames {
			if err := os.RemoveAll(filepath.Join(cwd, fileToDelete.Name())); err != nil && !os.IsNotExist(err) {
				return nil, err
			}
		}

		// Copy server folder
		if fileNames, err = os.ReadDir(versionFolder); err != nil {
			return nil, err
		}
		fileNames = js_types.Slice[os.DirEntry](fileNames).Filter(func(input os.DirEntry) bool { return !slices.Contains(filesToIgnore, input.Name()) })
		for _, fileToCopy := range fileNames {
			if err := os.CopyFS(filepath.Join(cwd, fileToCopy.Name()), os.DirFS(filepath.Join(versionFolder, fileToCopy.Name()))); err != nil {
				return nil, err
			}
		}
	} else if overlayfs.OverlayfsAvaible() {
		bedrockConfig.Overlayfs = &overlayfs.Overlayfs{
			Target:  cwd,
			Upper:   upper,
			Workdir: workdir,
			Lower: []string{
				versionFolder,
			},
		}
	} else if file_checker.FolderIsEmpty(cwd) {
		if err := os.CopyFS(cwd, os.DirFS(versionFolder)); err != nil {
			return nil, err
		}
	}

	return bedrockConfig, nil
}

type Bedrock struct {
	PID         exec.Proc            // process status
	ServerStart exec.ProcExec        // Server command
	Version     *Version             // Server version
	Overlayfs   *overlayfs.Overlayfs // Overlayfs mounted
}

// Make server backup with [*archive/tar.Writer]
//
// If server mounted with [*sirherobrine23.com.br/go-bds/go-bds/overlayfs.Overlayfs] backup only Upper layer
// else backup entire server folder
func (bed Bedrock) Tar(w io.Writer) error {
	tarball := tar.NewWriter(w)
	defer tarball.Close()
	if bed.Overlayfs == nil {
		return tarball.AddFS(os.DirFS(bed.ServerStart.Cwd))
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
		return wr.AddFS(os.DirFS(bed.ServerStart.Cwd))
	}
	return wr.AddFS(os.DirFS(bed.Overlayfs.Upper))
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
	fileInfo, err := binfmt.Open(filepath.Join(bed.ServerStart.Cwd, bed.ServerStart.Arguments[0]))
	if err != nil {
		return err
	} else if fileInfo.GoOs() != runtime.GOOS {
		return ErrPlatform
	}

	// Check if require emulator
	if fileInfo.GoArch() != runtime.GOARCH && runtime.GOOS != "windows" {
		emulator := binfmt.AsEmulator(fileInfo)
		if emulator == nil {
			return ErrPlatform
		}
		bed.ServerStart.Arguments = append(emulator, bed.ServerStart.Arguments...)
	}

	// Start server
	if err := bed.PID.Start(bed.ServerStart); err != nil {
		return err
	}
	return nil
}
