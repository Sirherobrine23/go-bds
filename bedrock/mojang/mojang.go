package mojang

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/binfmt"
	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/internal"
	"sirherobrine23.com.br/go-bds/go-bds/overleyfs"
)

var ErrPlatform error = errors.New("current platform no supported or cannot emulate arch") // Cannot run server in platform or cannot emulate arch

type Mojang struct {
	VersionsFolder string // Folder with versions
	Version        string // Version to run server
	Path           string // Server folder to run

	RunInOverlayfs bool // Run server with overlayfs, require set Path and Workdir (RunInConfig) to run server
	RunInRootfs    bool // Run server with proot (chroot), recomends run if running in android or non root/sudo users
	RunInConfig    any  // Extra configs to run Overlayfs or proot

	Handler    *Handlers     // Server handlers
	Config     *MojangConfig // Server config
	ServerProc exec.Proc     // Server process

	overlayfs *overleyfs.Overlayfs
}

func (server *Mojang) Close() error {
	if server.ServerProc != nil {
		server.ServerProc.Write([]byte("stop\n"))
		if err := server.ServerProc.Close(); err != nil {
			return err
		}
	}

	if server.overlayfs != nil {
		if err := server.overlayfs.Unmount(); err != nil {
			return err
		}
	}

	return nil
}

// Start server and mount overlayfs if version not exists localy download
func (server *Mojang) Start() error {
	// Get latest version if empty or `latest`
	if server.Version == "" || strings.ToLower(server.Version) == "latest" {
		versions, err := FromVersions()
		if err != nil {
			return err
		}
		server.Version = GetLatest(versions)
	}

	// Version folder
	versionRoot := filepath.Join(server.VersionsFolder, server.Version)

	// Clear version folder if empty
	if versionEmpty, err := internal.EmptyFolder(versionRoot); err != nil {
		return err
	} else if !versionEmpty {
		if err := os.RemoveAll(versionRoot); err != nil {
			return err
		}
	}

	if !internal.ExistPath(versionRoot) {
		versions, err := FromVersions()
		if err != nil {
			return err
		} else if err := os.MkdirAll(versionRoot, 0700); err != nil {
			return err
		}

		var target VersionPlatform
		if version, ok := versions[server.Version]; !ok {
			return fmt.Errorf("version not found in database")
		} else if target, ok = version.Platforms[fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)]; !ok {
			if ok, err = binfmt.FindByPlatform("linux/amd64"); err != nil {
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
	if runtime.GOOS == "windows" {
		if !(runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64") {
			return fmt.Errorf("run minecraft server in Windows with x64/amd64 or arm64 emulation")
		}
		serverExecOptions.Arguments = []string{"bedrock_server.exe"}
		server.ServerProc = &exec.Os{}
	} else if runtime.GOOS == "linux" {
		if server.RunInOverlayfs {
			if _, ok := server.RunInConfig.(string); !ok {
				return fmt.Errorf("require workdir seted in RunInConfig to mount overlayfs")
			}
			runDir, err := os.MkdirTemp(os.TempDir(), "bdsserver_*")
			if err != nil {
				return err
			}

			server.overlayfs = &overleyfs.Overlayfs{}
			server.overlayfs.Upper = server.Path
			server.overlayfs.Workdir = server.RunInConfig.(string)

			server.overlayfs.Target = filepath.Join(runDir, "merged")
			server.overlayfs.Lower = []string{versionRoot}
			if err := server.overlayfs.Mount(); err != nil {
				return err
			}
			serverExecOptions.Cwd = server.overlayfs.Target
		}

		if server.RunInRootfs {
			if _, ok := server.RunInConfig.(string); !ok {
				return fmt.Errorf("require rootfs seted in RunInConfig to run proot")
			}
			server.ServerProc = &exec.Proot{Rootfs: server.RunInConfig.(string)}
		}

		requireQemu, err := binfmt.CheckEmulate(filepath.Join(versionRoot, "bedrock_server"))
		if err != nil {
			return err
		} else if requireQemu {
			binfmt, err := binfmt.GetBinfmtEmulater(filepath.Join(versionRoot, "bedrock_server"))
			if err != nil {
				return err
			}
			if proot, ok := server.ServerProc.(*exec.Proot); ok {
				proot.Qemu = binfmt.Interpreter
			} else {
				server.ServerProc = &exec.Os{}
				serverExecOptions.Arguments = []string{binfmt.Interpreter}
			}
		}
		serverExecOptions.Arguments = append(serverExecOptions.Arguments, "./bedrock_server")
		serverExecOptions.Environment = map[string]string{"LD_LIBRARY_PATH": versionRoot}
	} else {
		return ErrPlatform
	}

	// Load config
	server.Config = &MojangConfig{}
	if err := server.Config.Load(server.Path); err != nil {
		return err
	}

	if err := server.ServerProc.Start(serverExecOptions); err != nil {
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
