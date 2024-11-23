package java

import (
	"os"
	"path/filepath"

	"sirherobrine23.com.br/go-bds/go-bds/bedrock"
	"sirherobrine23.com.br/go-bds/go-bds/internal/semver"
	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

// Default server name for file
const ServerName string = "server.jar"

// Version request not exists
var ErrVersionNotExist = bedrock.ErrNoVersion

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
	return nil
}

var _ Version = &GenericVersion{}
