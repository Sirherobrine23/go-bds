package java

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/utils/javaprebuild"
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
		if _, err := javac.ServerProc.Write([]byte("stop\n")); err != nil && err != io.EOF {
			return err
		} else if err = javac.ServerProc.Wait(); err != nil {
			return err
		}
		javac.ServerProc = nil
	}

	if javac.OverlayConfig != nil {
		switch err := javac.OverlayConfig.Unmount(); err {
		case nil, overlayfs.ErrNoCGOAvaible, overlayfs.ErrNotOverlayAvaible:
			javac.OverlayConfig = nil
		default:
			return err
		}
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

	versionFolder, jvmFolder := filepath.Join(javac.VersionsPath, version.SemverVersion().String()), filepath.Join(javac.JVMPath, fmt.Sprint(version.JVM()))

	var processConfig exec.ProcExec
	processConfig.Cwd = javac.Path
	processConfig.Arguments = []string{javaprebuild.JavaBinName, "-jar", ServerName, "-nogui"}

	if _, err := os.Stat(javac.Path); os.IsNotExist(err) {
		if err = os.MkdirAll(javac.Path, 0777); err != nil {
			return err
		}
	}

	if _, err := os.Stat(filepath.Join(jvmFolder, "bin", javaprebuild.JavaBinName)); err == nil {
		processConfig.Arguments[0] = filepath.Join(jvmFolder, "bin", javaprebuild.JavaBinName)
	} else if os.IsNotExist(err) {
		if err = javaprebuild.InstallLatest(version.JVM(), jvmFolder); err == nil {
			processConfig.Arguments[0] = filepath.Join(jvmFolder, "bin", javaprebuild.JavaBinName)
		} else if err != javaprebuild.ErrSystem {
			return err
		} else {
			jvmFolder = ""
		}
	}

	if _, err := os.Stat(filepath.Join(versionFolder, ServerName)); os.IsNotExist(err) {
		if err = version.Install(versionFolder); err != nil {
			return err
		}
	}

	copyServer := func() error {
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
		if err := CopyFS(javac.Path, os.DirFS(versionFolder)); err != nil {
			return err
		}
		return nil
	}

	if jvmFolder == "" {
		if err := copyServer(); err != nil {
			return err
		}
	} else {
		javac.OverlayConfig = &overlayfs.Overlayfs{
			Target:  javac.Path,
			Workdir: javac.WorkdirPath,
			Upper:   javac.UpperPath,
			Lower: []string{
				jvmFolder,
				versionFolder,
			},
		}

		if err := javac.OverlayConfig.Mount(); err == nil {
			processConfig.Arguments[0] = filepath.Join("./bin", javaprebuild.JavaBinName)
		} else if err != javaprebuild.ErrSystem {
			return err
		} else {
			if err := copyServer(); err != nil {
				return err
			}
		}
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
