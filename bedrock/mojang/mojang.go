package mojang

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"sirherobrine23.org/go-bds/go-bds/exec"
	ccopy "sirherobrine23.org/go-bds/go-bds/internal/copy"
	"sirherobrine23.org/go-bds/go-bds/overleyfs"
)

type Mojang struct {
	VersionsFolder string // Folder with versions
	Version        string // Version to run server
	Path           string // Run server at folder

	Handler *Handlers     // Server handlers
	Config  *MojangConfig // Server config

	ServerProc exec.Proc
}

func (server *Mojang) Close() error {
	if server.ServerProc != nil {
		server.ServerProc.Write([]byte("stop\n"))
		return server.ServerProc.Close()
	}
	return nil
}

// Start server and mount overlayfs if version not exists localy download
func (server *Mojang) Start() error {
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
		os.MkdirAll(versionRoot, 0700)

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

	// Merge new version
	if !checkExist(server.Path) {
		os.MkdirAll(server.Path, 0700)
		if err := ccopy.CopyDirectory(versionRoot, server.Path); err != nil {
			return err
		}
	} else {
		os.RemoveAll(server.Path + "old")
		if err := os.Rename(server.Path, server.Path+"old"); err != nil {
			return err
		}
		fs, err := (&overleyfs.Overlayfs{Lower: []string{versionRoot, server.Path + "old"}}).GoMerge()
		if err != nil {
			return err
		}
		os.MkdirAll(server.Path, 0600)
		if err := copyToDisk(fs, ".", server.Path); err != nil {
			return err
		}

		recopy := []string{"server.properties", "allowlist.json", "permissions.json"}
		for _, file := range recopy {
			oldPath, newPath := filepath.Join(server.Path+"old", file), filepath.Join(server.Path, file)
			oldFile, err := os.Open(oldPath)
			if err != nil {
				return err
			}
			defer oldFile.Close()
			newFile, err := os.Create(newPath)
			if err != nil {
				return err
			}
			defer newFile.Close()
			if _, err := io.Copy(newFile, oldFile); err != nil {
				return err
			}
		}
	}

	// Start server
	server.ServerProc = &exec.Os{}
	opt := exec.ProcExec{
		Cwd:         server.Path,
		Arguments:   []string{"./bedrock_server"},
		Environment: map[string]string{"LD_LIBRARY_PATH": "."},
	}

	if runtime.GOOS == "windows" {
		if !(runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64") {
			return fmt.Errorf("run minecraft server in Windows with x64/amd64 or arm64 emulation")
		}
		opt.Environment = make(map[string]string)
		opt.Arguments = []string{"bedrock_server.exe"}
	}

	// Load config
	server.Config = &MojangConfig{}
	if err := server.Config.Load(server.Path); err != nil {
		return err
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

func copyToDisk(fsys fs.FS, root, target string) error {
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		fmt.Printf("%q ==> %q\n", path, filepath.Join(target, path))
		if err != nil {
			return err
		} else if d.IsDir() {
			return os.MkdirAll(filepath.Join(target, path), d.Type())
		}
		fullPath := filepath.Join(target, path)
		os.MkdirAll(filepath.Dir(fullPath), 0600)

		fsFile, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer fsFile.Close()

		file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, d.Type())
		if err != nil {
			return err
		}
		defer file.Close()

		// Copy data
		_, err = io.Copy(file, fsFile)
		return err
	})
}
