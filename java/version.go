package java

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/semver"
	"sirherobrine23.com.br/go-bds/request/v2"
)

// Default server name for file
const (
	ServerName string = "server.jar"
	EulaFile   string = "#This file was changed by go-bds\neula=true"
)

// Version not exists
var ErrVersionNotExist error = errors.New("cannot find version")

type Version interface {
	JVM() uint                      // Java version to Run server
	SemverVersion() *semver.Version // Platform version
	Install(path string) error      // Install server in path
}

type GenericVersion struct {
	Version    string // Server semver version
	JVMVersion uint   // Java version required
	URLs       []struct {
		Name string // File name
		URL  string // File URL
	}
}

// Generic function to get server versions
type ListServer func() (Versions, error)

// Add function to find server version on array
type Versions []Version

// Generic interface to find Version in array list
func (versions Versions) Find(version string) Version {
	semver.SortStruct(versions)
	slices.Reverse(versions)
	for _, ver := range versions {
		if ver.SemverVersion().String() == version {
			return ver
		}
	}
	return nil
}

func (generic GenericVersion) SemverVersion() *semver.Version { return semver.New(generic.Version) }

func (generic GenericVersion) JVM() uint {
	ver := generic.SemverVersion()
	if generic.JVMVersion == 0 {
		switch {
		case semver.New("1.12").LessThan(*ver):
			return 8
		case semver.New("1.16").LessThan(*ver):
			return 11
		case semver.New("1.17").LessThan(*ver):
			return 16
		case semver.New("1.20").LessThan(*ver):
			return 17
		default:
			return 21
		}
	}
	return generic.JVMVersion
}

func (generic GenericVersion) Install(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, 0777); err != nil {
			return err
		}
	}
	for _, downloadInfo := range generic.URLs {
		if _, err := request.SaveAs(downloadInfo.URL, filepath.Join(path, downloadInfo.Name), nil); err != nil {
			return err
		}
	}

	if file, err := os.Open(filepath.Join(path, ServerName)); err == nil {
		stat, _ := file.Stat()
		zipFile, err := zip.NewReader(file, stat.Size())
		if err == nil {
			for _, fileInfo := range zipFile.File {
				if !fileInfo.FileInfo().IsDir() && strings.HasPrefix(fileInfo.Name, "META-INF/versions") && fileInfo.Name[9:] != "versions.list" {
					localFilePath := filepath.Join(path, fileInfo.Name[9:])
					if _, err := os.Stat(filepath.Dir(localFilePath)); os.IsNotExist(err) {
						if err = os.MkdirAll(filepath.Dir(localFilePath), 0777); err != nil {
							return err
						}
					}
					newLocal, err := os.OpenFile(localFilePath, os.O_CREATE|os.O_TRUNC|os.O_TRUNC|os.O_WRONLY, 0777)
					if err != nil {
						return err
					}
					defer newLocal.Close()
					rFile, err := fileInfo.Open()
					if err != nil {
						newLocal.Close()
						return err
					}
					defer rFile.Close()
					_, err = io.Copy(newLocal, rFile)
					rFile.Close()
					if err != nil {
						return err
					}
				}
			}
		}
	}

	if err := os.WriteFile(filepath.Join(path, "eula.txt"), []byte(EulaFile), 0777); err != nil {
		return err
	}
	return nil
}

var _ Version = &GenericVersion{}
