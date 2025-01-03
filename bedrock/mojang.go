package bedrock

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/binfmt"
	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/overlayfs"
)

var ErrPlatform error = errors.New("current platform no supported or cannot emulate arch") // Cannot run server in platform or cannot emulate arch

type Mojang struct {
	VersionsFolder string // Folder with versions
	Version        string // Version to run server
	Path           string // Server folder to run

	OverlayConfig *overlayfs.Overlayfs // Config to overlayfs, go-bds replace necesarys configs

	ServerProc exec.Proc // Server process
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
	// Get latest version if empty or `latest`
	if server.Version == "" || strings.ToLower(server.Version) == "latest" {
		versions, err := FromVersions()
		if err != nil {
			return err
		}
		server.Version = versions.GetLatest()
	}

	// Version folder
	versionRoot := filepath.Join(server.VersionsFolder, server.Version)

	// Clear version folder if empty
	if entrys, _ := os.ReadDir(versionRoot); len(entrys) == 0 {
		if err := os.RemoveAll(versionRoot); err != nil {
			return err
		}
	}

	// Check and Download version if not exists
	if _, err := os.Stat(versionRoot); os.IsNotExist(err) {
		versions, err := FromVersions()
		if err != nil {
			return err
		} else if err := os.MkdirAll(versionRoot, 0666); err != nil {
			return err
		}

		var target VersionPlatform
		if version, ok := versions.Get(server.Version); !ok {
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
		if err := target.Download(versionRoot); err != nil {
			return err
		}
	}

	if server.Path == "" {
		return fmt.Errorf("set Path to run minecraft server")
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
