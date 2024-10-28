package bedrock

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"sirherobrine23.com.br/go-bds/go-bds/binfmt"
	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/overlayfs"
)

var (
	ErrVersionNotFound error = errors.New("version not found")                                    // Version not found in remote cache or another storages
	ErrPlatform        error = errors.New("current platform no supported or cannot emulate arch") // Cannot run server in platform or cannot emulate arch
)

type Mojang struct {
	Version string // Version to run server

	VersionsPath string // Folder path to save extracted Minecraft versions
	WorkdirPath  string // Path Workdir to Overlayfs
	UpperPath    string // Path to save diff changes files, only platforms required's and same filesystem to 'Path'
	Path         string // Server folder to run Minecraft server

	RootfsConfig *exec.Proot // Config to run in proot
	ServerProc   exec.Proc   // Server process

	OverlayConfig *overlayfs.Overlayfs // Config to overlayfs, go-bds replace necesarys configs

	ServerProc exec.Proc // Server process
}

func EmptyFolder(fpath string) (bool, error) {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return true, nil
	} else if err != nil {
		return false, err
	}
	entrys, err := os.ReadDir(fpath)
	if err != nil {
		return false, err
	}
	return len(entrys) == 0, nil
}

func (server *Mojang) Close() error {
	if server.ServerProc != nil {
		server.ServerProc.Write([]byte("stop\n"))
		if err := server.ServerProc.Close(); err != nil {
			return err
		}
	}

	if server.OverlayConfig != nil {
		if err := server.OverlayConfig.Unmount(); err != nil {
			return err
		}
	}

	return nil
}

// Prepare the server and if it doesn't exist it will be downloaded and configured, and if overlayfs is available it will be used.
//
// If the server requires emulation, it will be checked whether any program is available on the system, such as qemu or box64
func (server *Mojang) Start() error {
	if server.Version == "" || server.Version == "latest" {
		return fmt.Errorf("set valid Minecraft Bedrock version to run")
	}

	// Version folder path
	versionRoot := filepath.Join(server.VersionsPath, server.Version)

	// Check if version folder is empty to delete
	downloadServer := false
	if entrys, _ := os.ReadDir(versionRoot); len(entrys) == 0 {
		downloadServer = true
		if err := os.RemoveAll(versionRoot); !os.IsNotExist(err) {
			return err // Return if not exist folder
		}
	}

	// Check and Download version if not exists
	if downloadServer {
		versions, err := FromVersions()
		if err != nil {
			return err
		} else if err := os.MkdirAll(versionRoot, 0666); err != nil {
			return err
		}

		var target VersionPlatform
		if version, ok := versions[server.Version]; !ok {
			return fmt.Errorf("version not found in database")
		} else if target, ok = version.Platforms[fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)]; !ok {
			if ok, err = binfmt.Target("linux/amd64"); err != nil {
				return err
			} else if ok {
				target, ok = version.Platforms["linux/amd64"]
			}
			if !ok {
				return ErrPlatform
			}
		}

		if err := sysVer.Download(versionRoot); err != nil {
			return err
		}
	}

	if server.Path == "" {
		return fmt.Errorf("set Path to run minecraft server")
	}

	// Prepare overlayfs/mergefs configuration
	server.OverlayConfig = &overlayfs.Overlayfs{
		Target:  server.Path,           // Target path to merged folder
		Workdir: server.WorkdirPath,    // Path to linux overlayfs
		Upper:   server.UpperPath,      // Path to save Target diff
		Lower:   []string{versionRoot}, // Only low layer require to run server, base server
	}

	// Mount overlayfs if avaible
	if err := server.OverlayConfig.Mount(); err != nil {
		if err != overlayfs.ErrNotOverlayAvaible {
			return err // Return if another any error
		}

		exist := slices.ContainsFunc([]string{filepath.Join(server.Path, "bedrock_server"), filepath.Join(server.Path, "bedrock_server.exe")}, func(rpath string) bool { _, err := os.Stat(rpath); return err == nil })
		if !exist { // Copy full server
			if err := os.CopyFS(server.Path, os.DirFS(versionRoot)); err != nil {
				return err
			}
		} else {
			type fileBackup struct {
				name    string
				content []byte
			}

			files := []fileBackup{
				{"server.properties", nil},
				{"permissions.json", nil},
				{"allowlist.json", nil},
			}
			for fileIndex := range files {
				files[fileIndex].content, _ = os.ReadFile(filepath.Join(server.Path, files[fileIndex].name))
			}

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

			if err := CopyFS(server.Path, os.DirFS(versionRoot)); err != nil {
				return err
			}

			for fileIndex := range files {
				if len(files[fileIndex].content) > 0 {
					if err := os.WriteFile(filepath.Join(server.Path, files[fileIndex].name), files[fileIndex].content, 0); err != nil {
						return err
					}
				}
			}
		}
	}

	var serverExecOptions exec.ProcExec
	serverExecOptions.Cwd = server.Path

	switch runtime.GOOS {
	case "windows":
		if !(runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64") {
			return fmt.Errorf("run minecraft server in Windows with x64/amd64 or arm64")
		}
		serverExecOptions.Arguments = []string{"bedrock_server.exe"}
		server.ServerProc = &exec.Os{}
	case "linux", "android":
		if server.OverlayConfig != nil {
			if server.OverlayConfig.Upper == "" || server.OverlayConfig.Workdir == "" {
				return fmt.Errorf("bedrock,overlayfs: require Upper and Workdir to run with overlayfs")
			}

			server.OverlayConfig.Lower = append(server.OverlayConfig.Lower, versionRoot)
			if err := server.OverlayConfig.Mount(); err != nil {
				return err
			}
			serverExecOptions.Cwd = server.OverlayConfig.Target
		}

		useQemu, err := false, error(nil)
		if useQemu, err = binfmt.RequireEmulate(filepath.Join(versionRoot, "bedrock_server")); err != nil {
			return err
		}

		if useQemu {
			var emulater binfmt.Binfmt
			if emulater, err = binfmt.ResolveBinfmt(filepath.Join(versionRoot, "bedrock_server")); err != nil {
				return err
			}

			if proot, ok := server.ServerProc.(*exec.Proot); ok {
				emulaterArgs := emulater.ProgramArgs()
				proot.Qemu = emulaterArgs[0]
				serverExecOptions.Arguments = emulaterArgs[1:]
			} else {
				server.ServerProc = &exec.Os{}
				serverExecOptions.Arguments = emulater.ProgramArgs()
			}
		}

		serverExecOptions.Arguments = append(serverExecOptions.Arguments, "./bedrock_server")
		serverExecOptions.Environment = map[string]string{"LD_LIBRARY_PATH": versionRoot}
	default:
		return ErrPlatform
	}

	// Start server
	return server.ServerProc.Start(serverExecOptions)
}

// Create backup from server
//
// if running in overlafs backup only Upper folder else backup full server
func (server Mojang) Tar(w io.Writer) error {
	tarball := tar.NewWriter(w)
	defer tarball.Close()
	if server.OverlayConfig != nil {
		return tarball.AddFS(os.DirFS(server.OverlayConfig.Upper))
	}
	return tarball.AddFS(os.DirFS(server.Path))
}
