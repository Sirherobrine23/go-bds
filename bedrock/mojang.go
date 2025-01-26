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
	"sirherobrine23.com.br/go-bds/go-bds/overlayfs"
)

var (
	ErrVersionNotFound error = errors.New("version not found")                                    // Version not found in remote cache or another storages
	ErrNoUpgrade       error = errors.New("cannot upgrade server")                                // Cannot upgrade server
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

func checkCreate(folder string) error {
	if _, err := os.Stat(folder); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(folder, 0777)
		}
		return err
	}
	return nil
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

func (server *Mojang) Close() error {
	if server.ServerProc != nil {
		if _, err := server.ServerProc.Write([]byte("stop\n")); err != nil {
			return err
		} else if err := server.ServerProc.Close(); err != nil {
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
	} else if server.Path == "" {
		return fmt.Errorf("set Path to run minecraft server")
	}

	for _, folder := range []string{
		server.VersionsPath,
		server.WorkdirPath,
		server.UpperPath,
		server.Path,
	} {
		if err := checkCreate(folder); err != nil {
			return err
		}
	}

	// Version folder path
	versionRoot := filepath.Join(server.VersionsPath, server.Version)

	// Check if version folder is empty to delete
	downloadServer := false
	if entrys, _ := os.ReadDir(versionRoot); len(entrys) == 0 {
		downloadServer = true
		if err := os.RemoveAll(versionRoot); err != nil && !os.IsNotExist(err) {
			return err // Return if not exist folder
		}
	}

	// Check and Download version if not exists
	if downloadServer {
		versions, err := FromVersions()
		if err != nil {
			return err
		}

		// Golang target
		osTarget := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

		rootVersion, ok := versions.Get(server.Version)
		if !ok {
			return ErrVersionNotFound
		}

		sysVer, ok := rootVersion.Platforms[osTarget]
		if !ok {
			// Check if avaible to emulate server
			if slices.Contains([]string{"linux", "windows"}, runtime.GOOS) {
				if archok, _ := binfmt.Target("amd64"); archok {
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

	// Version root and mods folder if exists
	lowerLayer := []string{versionRoot}
	if _, err := os.Stat(filepath.Join(server.VersionsPath, "mods")); err == nil {
		lowerLayer = append(lowerLayer, filepath.Join(server.VersionsPath, "mods"))
	}

	// Prepare overlayfs/mergefs configuration
	server.OverlayConfig = &overlayfs.Overlayfs{
		Target:  server.Path,        // Target path to merged folder
		Workdir: server.WorkdirPath, // Path to linux overlayfs
		Upper:   server.UpperPath,   // Path to save Target diff
		Lower:   lowerLayer,         // Only low layer require to run server, base server
	}

	var bedrockConfigExec exec.ProcExec

	// Mount overlayfs if avaible
	switch err := server.OverlayConfig.Mount(); err {
	case nil:
		bedrockConfigExec.Cwd = server.OverlayConfig.Target
	case overlayfs.ErrNotOverlayAvaible:
		server.OverlayConfig = nil
		bedrockConfigExec.Cwd = server.Path
		exist := slices.ContainsFunc([]string{filepath.Join(server.Path, "bedrock_server"), filepath.Join(server.Path, "bedrock_server.exe")}, func(rpath string) bool { _, err := os.Stat(rpath); return err == nil })

		// Move to backup folder
		backupFolder := ""
		if exist {
			// List current files
			files, err := os.ReadDir(server.Path)
			if err != nil {
				return errors.Join(err, ErrNoUpgrade)
			}

			// Create backup dir
			if backupFolder, err = os.MkdirTemp(server.Path, "upgradeDir*"); err != nil {
				return errors.Join(err, ErrNoUpgrade)
			}

			// Move files to folder
			for _, entry := range files {
				if err := os.Rename(filepath.Join(server.Path, entry.Name()), filepath.Join(backupFolder, entry.Name())); err != nil {
					return errors.Join(err, ErrNoUpgrade)
				}
			}
		}

		// Copy server to new location
		if err = os.CopyFS(server.Path, os.DirFS(versionRoot)); err != nil {
			return err
		}

		// If is backup, move world, text, and more
		if exist {
			files, err := os.ReadDir(backupFolder)
			if err != nil {
				return errors.Join(err, ErrNoUpgrade)
			}
			filesToCopy := []string{
				"allowlist.json",
				"permissions.json",
				"server.properties",
				"worlds",
				"development_behavior_packs",
				"development_resource_packs",
			}
			files = slices.DeleteFunc(files, func(entry os.DirEntry) bool { return !slices.Contains(filesToCopy, entry.Name()) })
			for _, entry := range files {
				_ = os.RemoveAll(filepath.Join(server.Path, entry.Name()))
				if err := os.Rename(filepath.Join(backupFolder, entry.Name()), filepath.Join(server.Path, entry.Name())); err != nil {
					return errors.Join(err, ErrNoUpgrade)
				}
			}
			if err = os.RemoveAll(backupFolder); err != nil {
				return err
			}
		}
	default:
		return err
	}

	switch runtime.GOOS {
	case "windows":
		if !(runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64") {
			return fmt.Errorf("run minecraft server in Windows with x64/amd64 or arm64")
		}
		bedrockConfigExec.Arguments = []string{"bedrock_server.exe"}
		server.ServerProc = &exec.Os{}
	case "linux", "android":
		if requireEmulate, err := binfmt.RequireEmulate(filepath.Join(versionRoot, "bedrock_server")); err == nil && requireEmulate {
			var binfmtInfo binfmt.Binfmt
			if binfmtInfo, err = binfmt.ResolveBinfmt(filepath.Join(versionRoot, "bedrock_server")); err != nil {
				return err
			}

			switch v := server.ServerProc.(type) {
			case *exec.DockerContainer:
				// Require add Platform to Docker image
			case *exec.Proot:
				qemuArgs := binfmtInfo.ProgramArgs()
				v.Qemu = qemuArgs[0]
				bedrockConfigExec.Arguments = qemuArgs[1:]
			default:
				server.ServerProc = &exec.Os{}
				bedrockConfigExec.Arguments = binfmtInfo.ProgramArgs()
			}
		}

		bedrockConfigExec.Arguments = append(bedrockConfigExec.Arguments, "./bedrock_server")
		bedrockConfigExec.Environment = map[string]string{"LD_LIBRARY_PATH": "."}
	default:
		return ErrPlatform
	}

	if server.ServerProc == nil {
		server.ServerProc = &exec.Os{}
	}

	// Start server
	return server.ServerProc.Start(bedrockConfigExec)
}
