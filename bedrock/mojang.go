package bedrock

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"sirherobrine23.com.br/go-bds/go-bds/binfmt"
	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/internal/fsdiff"
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
		} else { // Get diff from files and copy
			filesNode, err := fsdiff.Diff(os.DirFS(versionRoot), os.DirFS(server.Path))
			if err != nil {
				return err
			}

			for _, path := range filesNode {
				stat, _ := os.Stat(filepath.Join(versionRoot, path))
				os.MkdirAll(filepath.Join(server.Path, filepath.Dir(path)), stat.Mode()) // Create folder if not exists

				// Create
				v1File, _ := os.Open(filepath.Join(versionRoot, path))
				defer v1File.Close()

				// Open
				v2File, err := os.OpenFile(filepath.Join(server.Path, path), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, stat.Mode())
				if err != nil {
					return err
				}
				defer v2File.Close()

				// Copy content
				if _, err := io.Copy(v2File, v1File); err != nil {
					return err
				}
				v1File.Close()
				v2File.Close()
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
