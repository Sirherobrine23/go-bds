package java

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/java/javaprebuild"
	"sirherobrine23.com.br/go-bds/go-bds/overlayfs"
)

type Java struct {
	Version string // Version to run server
	Variant string // Server to run, example: mojang, spigot, paper, etc...

	VersionsPath string // Folder path to save extracted Minecraft versions
	JVMPath      string // Folder to storage java versions
	WorkdirPath  string // Path Workdir to Overlayfs
	UpperPath    string // Path to save diff changes files, only platforms required's and same filesystem to 'Path'
	Path         string // Server folder to run Minecraft server

	ListVersions  ListServer           // function to list server versions
	ServerProc    exec.Proc            // Server process
	OverlayConfig *overlayfs.Overlayfs // Config to overlayfs, go-bds replace necesarys configs
}

// Prepare new struct with basic setup
func New(Variant, JVMPath string) (javac *Java, err error) {
	javac = &Java{JVMPath: JVMPath}
	switch Variant {
	case "paper", "folia", "velocity":
		javac.ListVersions, err = ListPaper(Variant)
	case "spigot":
		javac.ListVersions = ListSpigot(JVMPath)
	case "purpur":
		javac.ListVersions = ListPurpur
	default:
		javac.ListVersions = ListMojang
	}
	return
}

func (javac *Java) Close() error {
	if javac.ServerProc != nil {
		if _, err := javac.ServerProc.Write([]byte("stop\n")); err != nil {
			return err
		} else if err = javac.ServerProc.Wait(); err != nil {
			return err
		}
		javac.ServerProc = nil
	}
	if javac.OverlayConfig != nil {
		if err := javac.OverlayConfig.Unmount(); err != nil {
			return err
		}
		javac.OverlayConfig = nil
	}
	return nil
}

func (javac *Java) Start() error {
	versionList, err := javac.ListVersions()
	if err != nil {
		return err
	} else if len(versionList) == 0 {
		return ErrVersionNotExist
	}

	version := versionList.Find(javac.Version)
	if version == nil {
		return ErrVersionNotExist
	}

	var processConfig exec.ProcExec
	processConfig.Cwd = javac.Path
	processConfig.Arguments = []string{"java", "-jar", ServerName, "-nogui"}

	prebuildJavaPath := filepath.Join(javac.JVMPath, fmt.Sprint(version.JVM()))
	if _, err := os.Stat(prebuildJavaPath); os.IsNotExist(err) {
		err = javaprebuild.InstallLatest(version.JVM(), prebuildJavaPath)
		if err != nil && err != javaprebuild.ErrSystem {
			return err
		}
	}

	if processConfig.Arguments[0] = filepath.Join(prebuildJavaPath, "bin/java"); runtime.GOOS == "windows" {
		processConfig.Arguments[0] += ".exe"
	} else if !exec.LocalBinExist(processConfig) {
		processConfig.Arguments[0] = "java"
		prebuildJavaPath = ""
	}

	versionFolder := filepath.Join(javac.VersionsPath, version.SemverVersion().String())
	if _, err := os.Stat(versionFolder); os.IsNotExist(err) {
		if err = version.Install(versionFolder); err != nil {
			return err
		}
	}

	javac.OverlayConfig = &overlayfs.Overlayfs{
		Target:  javac.Path,
		Upper:   javac.UpperPath,
		Workdir: javac.WorkdirPath,
		Lower: []string{
			prebuildJavaPath,
			versionFolder,
		},
	}

	if err := javac.OverlayConfig.Mount(); err != nil && (err == overlayfs.ErrNotOverlayAvaible || errors.Is(err, fs.ErrPermission)) {
		newJavaPath := filepath.Join(javac.Path, "java")
		versionFiles, _ := os.ReadDir(versionFolder)
		for _, file := range versionFiles {
			if err = os.RemoveAll(filepath.Join(javac.Path, file.Name())); err != nil && !os.IsNotExist(err) {
				return err
			}
		}

		if err = os.CopyFS(javac.Path, os.DirFS(versionFolder)); err != nil {
			return err
		}

		if prebuildJavaPath != "" {
			if err = os.CopyFS(newJavaPath, os.DirFS(prebuildJavaPath)); err != nil {
				return err
			}
			if processConfig.Arguments[0] = filepath.Join(newJavaPath, "bin/java"); runtime.GOOS == "windows" {
				processConfig.Arguments[0] += ".exe"
			}
		}
	} else if err != nil {
		return err
	}

	javac.ServerProc = &exec.Os{}
	return javac.ServerProc.Start(processConfig)
}

// Create backup from server
//
// if running in overlafs backup only Upper folder else backup full server
func (javac Java) Tar(w io.Writer) error {
	tarball := tar.NewWriter(w)
	defer tarball.Close()
	if javac.OverlayConfig != nil {
		return tarball.AddFS(os.DirFS(javac.OverlayConfig.Upper))
	}
	return tarball.AddFS(os.DirFS(javac.Path))
}
