package pmmp

import (
	"archive/tar"
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/overlayfs"
	"sirherobrine23.com.br/go-bds/go-bds/utils/file_checker"
)

// Create and setup basics info to start Pocketmine-PMMP
func NewPocketmine(version *Version, versionFolder, cwd, upper, workdir string) (*Pocketmine, error) {
	if version == nil {
		return nil, ErrNoVersion
	}

	// Server file
	pocketmineFile := filepath.Join(versionFolder, "pocketmine", version.Version, "server.phar")
	if !file_checker.IsFile(pocketmineFile) {
		// Download pocketmine if not exists
		if err := version.Download(pocketmineFile); err != nil {
			return nil, err
		}
	}

	// PHP binary file
	phpFolder := filepath.Join(versionFolder, "php", version.PHP.PHPVersion)
	switch runtime.GOOS {
	case "windows":
		if !file_checker.IsFile(filepath.Join(phpFolder, "bin/php/php.exe")) {
			if err := version.PHP.Install(phpFolder); err != nil {
				return nil, err
			}
		}
		phpFolder = filepath.Join(phpFolder, "bin/php/php.exe")
	default:
		if !file_checker.IsFile(filepath.Join(phpFolder, "bin/php")) {
			if err := version.PHP.Install(phpFolder); err != nil {
				return nil, err
			}
		}
		phpFolder = filepath.Join(phpFolder, "bin/php")
	}

	// Config to Pocketmine
	pmmpConfig := &Pocketmine{
		PID:       &exec.Os{},
		Version:   version,
		Overlayfs: nil,
		ServerStart: exec.ProcExec{
			Cwd: cwd,
			Arguments: []string{
				phpFolder,
				pocketmineFile,
				"--no-wizard",
			},
			Environment: exec.Env{},
		},
	}

	if overlayfs.OverlayfsAvaible() {
		pmmpConfig.Overlayfs = &overlayfs.Overlayfs{
			Target:  cwd,
			Upper:   upper,
			Workdir: workdir,
			Lower: []string{
				filepath.Dir(pocketmineFile),
				filepath.Join(versionFolder, "php", version.PHP.PHPVersion),
			},
		}

		pmmpConfig.ServerStart.Arguments[1] = filepath.Base(pocketmineFile)
		switch runtime.GOOS {
		case "windows":
			pmmpConfig.ServerStart.Arguments[0] = "./bin/php/php.exe"
		default:
			pmmpConfig.ServerStart.Arguments[0] = "./bin/php"
		}
	}

	return pmmpConfig, nil
}

type Pocketmine struct {
	PID         exec.Proc            // process status
	ServerStart exec.ProcExec        // Server command
	Version     *Version             // Server version
	Overlayfs   *overlayfs.Overlayfs // Overlayfs mounted
}

// Make server backup with [*archive/tar.Writer]
//
// If server mounted with [*sirherobrine23.com.br/go-bds/go-bds/overlayfs.Overlayfs] backup only Upper layer
// else backup entire server folder
func (pmmp Pocketmine) Tar(w io.Writer) error {
	tarball := tar.NewWriter(w)
	defer tarball.Close()
	if pmmp.Overlayfs == nil {
		return tarball.AddFS(os.DirFS(pmmp.ServerStart.Cwd))
	}
	return tarball.AddFS(os.DirFS(pmmp.Overlayfs.Upper))
}

// Make server backup with [*archive/zip.Writer]
//
// If server mounted with [*sirherobrine23.com.br/go-bds/go-bds/overlayfs.Overlayfs] backup only Upper layer
// else backup entire server folder
func (pmmp Pocketmine) Zip(w io.Writer) error {
	wr := zip.NewWriter(w)
	defer wr.Close()
	if pmmp.Overlayfs == nil {
		return wr.AddFS(os.DirFS(pmmp.ServerStart.Cwd))
	}
	return wr.AddFS(os.DirFS(pmmp.Overlayfs.Upper))
}

func (pmmp *Pocketmine) Start() error {
	if pmmp == nil || pmmp.PID == nil {
		return errors.New("cannot start server, server proc not defined")
	}

	// If overlayfs configured mount before start server
	if pmmp.Overlayfs != nil {
		if err := pmmp.Overlayfs.Mount(); err != nil {
			return err
		}
	}

	// Start server
	if err := pmmp.PID.Start(pmmp.ServerStart); err != nil {
		return err
	}
	return nil
}
