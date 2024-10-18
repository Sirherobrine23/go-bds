package mojang

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

// Start server and mount overlayfs if version not exists localy download
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

		// Golang target
		osTarget := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

		rootVersion, ok := versions[server.Version]
		if !ok {
			return ErrVersionNotFound
		}

		sysVer, ok := rootVersion.Platforms[osTarget]
		if !ok {
			// Check if avaible to emulate server
			if slices.Contains([]string{"linux", "windows"}, runtime.GOOS) {
				if archok, _ := binfmt.FindByPlatform("amd64"); archok {
					sysVer, ok = rootVersion.Platforms[fmt.Sprintf("%s/amd64", runtime.GOOS)]
				}
			}

			// if not possible download to plaftorm and arch return 'ErrPlaftorm' because cannot emulate or another problems
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
	if runtime.GOOS == "windows" {
		if !(runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64") {
			return fmt.Errorf("run minecraft server in Windows with x64/amd64 or arm64")
		}
		serverExecOptions.Arguments = []string{"bedrock_server.exe"}
		server.ServerProc = &exec.Os{}
	} else if runtime.GOOS == "linux" {
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

		if server.RootfsConfig != nil {
			server.ServerProc = server.RootfsConfig
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
