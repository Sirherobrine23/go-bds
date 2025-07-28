package bedrock

import (
	"archive/tar"
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/binfmt"
	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/utils/file_checker"
	"sirherobrine23.com.br/go-bds/go-bds/utils/js_types"
	"sirherobrine23.com.br/go-bds/overlayfs"
)

var currentPlatform = func() string {
	bin, err := binfmt.Open(os.Args[0])
	if err != nil {
		return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	}
	return strings.ReplaceAll(binfmt.String(bin), "android", "linux")
}()

type Bedrock struct {
	PID            exec.Proc            // process status
	ServerStart    exec.ProcExec        // Server command
	Overlayfs      *overlayfs.Overlayfs // Overlayfs mounted
	Version        *Version             // Server version
	PlaformVersion *PlatformVersion     // Server version target
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
		ServerStart: exec.ProcExec{
			Cwd:         cwd,
			Arguments:   []string{},
			Environment: exec.Env{},
		},
	}

	// Folder path to storage server version
	versionFolder = filepath.Join(versionFolder, version.Version)
	if _, err := os.Stat(versionFolder); os.IsNotExist(err) {
		if err = os.MkdirAll(versionFolder, 0777); err != nil {
			return nil, err
		}
	}

	// Correct config to GOOS
	switch runtime.GOOS {
	default:
		return nil, ErrPlatform
	case "windows":
		bedrockConfig.ServerStart.Arguments = []string{"bedrock_server.exe"}
		target, ok := version.Plaforms["windows/amd64"]
		if !ok {
			return nil, ErrNoVersion
		}
		bedrockConfig.PlaformVersion = target

		if file_checker.FolderIsEmpty(versionFolder) {
			if err := target.Extract(versionFolder); err != nil {
				return nil, err
			}
		}
	case "linux", "android":
		bedrockConfig.ServerStart.Environment = exec.Env{"LD_LIBRARY_PATH": "."}
		bedrockConfig.ServerStart.Arguments = []string{"./bedrock_server"}

		target, ok := version.Plaforms[currentPlatform]
		if !ok {
			if target, ok = version.Plaforms["linux/amd64"]; !ok {
				return nil, ErrNoVersion
			}
		}
		bedrockConfig.PlaformVersion = target

		if file_checker.FolderIsEmpty(versionFolder) {
			if err := target.Extract(versionFolder); err != nil {
				return nil, err
			}
		}
	}

	switch {
	case overlayfs.OverlayfsAvaible(): // Attemp boot server with overlayfs if else copy files
		bedrockConfig.Overlayfs = overlayfs.NewOverlayFS(cwd, upper, workdir, versionFolder)
	case file_checker.FolderIsEmpty(cwd): // Copy server
		if err := os.CopyFS(cwd, os.DirFS(versionFolder)); err != nil {
			return nil, fmt.Errorf("cannot copy bedrock server to cwd: %s", err)
		}
	default: // Delete file from old server and copy
		// 1. Copy to cwd+".old"
		// 2. Remove all files
		// 3. Copy old files to new fresh copy
		oldCwd := cwd + "old"

		// Files to not delete
		FilesToCopy := []string{"allowlist.json", "permissions.json", "server.properties", "worlds"}

		remoteFile, _ := os.ReadDir(cwd)
		copyFiles := js_types.Slice[os.DirEntry](remoteFile).Filter(func(input os.DirEntry) bool { return slices.Contains(FilesToCopy, input.Name()) })


		// Remove old backup if exists
		if _, err := os.Stat(oldCwd); err == nil {
			if err = os.RemoveAll(oldCwd); err != nil {
				return nil, fmt.Errorf("cannot remove old server: %s", err)
			}
		}

		// Make server backup
		if err := os.CopyFS(oldCwd, os.DirFS(cwd)); err != nil { // Copy server to Cwd+"old"
			os.RemoveAll(oldCwd)
			return nil, fmt.Errorf("cannot move old server: %s", err)
		} else if err = file_checker.RemoveFiles(cwd, remoteFile); err != nil { // Delete files from cwd folder
			return nil, fmt.Errorf("cannot delete files inside on cwd: %s", err)
		} else if err = os.CopyFS(cwd, os.DirFS(versionFolder)); err != nil { // Copy new server version
			return nil, fmt.Errorf("cannot copy bedrock server to cwd: %s", err)
		} else if err = file_checker.ReplaceFiles(oldCwd, cwd, copyFiles); err != nil { // Replace files from old installation
			return nil, fmt.Errorf("cannot copy old files to new copy: %s", err)
		}
	}

	return bedrockConfig, nil
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
func (bed *Bedrock) Start(ctx context.Context) error {
	// if server not configured correctly return error
	if bed == nil || bed.PID == nil {
		return errors.New("cannot start server, server proc not defined")
	}

	// If overlayfs configured mount before start server
	if bed.Overlayfs != nil {
		if err := bed.Overlayfs.Mount(ctx); err != nil {
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
