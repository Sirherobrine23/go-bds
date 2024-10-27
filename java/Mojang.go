package java

import (
	"path/filepath"

	"sirherobrine23.com.br/go-bds/go-bds/internal/semver"
	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

var (
	_ VersionSearch = &MojangSearch{}
	_ Version       = &MojangVersion{}
)

type MojangVersion struct {
	Version   string // Server version
	ServerURL string // Server URL
	JVMVersion uint   // Java version
}

type MojangSearch struct {
	Version map[string]*MojangVersion
}

func (mojangSearch *MojangSearch) list() error {
	if len(mojangSearch.Version) != 0 {
		return nil
	}
	mojangSearch.Version = make(map[string]*MojangVersion)

	type pistonInfo struct {
		Latest   map[string]string `json:"latest"`
		Versions []struct {
			Url     string `json:"url"`
			Release string `json:"type"`
		} `json:"versions"`
	}

	type mojangPistonPackage struct {
		Version        string `json:"id"`
		Type           string `json:"type"`
		FilesDownloads map[string]struct {
			FileSize int64  `json:"size"`
			FileUrl  string `json:"url"`
			Sha1     string `json:"sha1"`
		} `json:"downloads"`
		Java struct {
			VersionMajor uint   `json:"majorVersion"`
			Component    string `json:"component"`
		} `json:"javaVersion"`
	}

	data, _, err := request.JSON[pistonInfo]("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json", nil)
	if err != nil {
		return err
	}

	for _, version := range data.Versions {
		if version.Release != "release" {
			continue
		}

		releaseInfo, _, err := request.JSON[mojangPistonPackage](version.Url, nil)
		if err != nil {
			return err
		} else if serverFile, ok := releaseInfo.FilesDownloads["server"]; ok {
			mojangSearch.Version[releaseInfo.Version] = &MojangVersion{
				Version:   releaseInfo.Version,
				JVMVersion: releaseInfo.Java.VersionMajor,
				ServerURL: serverFile.FileUrl,
			}
		}
	}

	return nil
}

func (mojangSearch MojangSearch) Find(version string) (Version, error) {
	if err := mojangSearch.list(); err != nil {
		return nil, err
	} else if ver, ok := mojangSearch.Version[version]; ok {
		return ver, nil
	}
	return nil, ErrNoFoundVersion
}

func (mojangVer MojangVersion) JavaVersion() uint              { return mojangVer.JVMVersion }
func (mojangVer MojangVersion) SemverVersion() *semver.Version { return semver.New(mojangVer.Version) }
func (mojangVer MojangVersion) Install(FolderPath string) error {
	if mojangVer.ServerURL == "" {
		return ErrNoFoundVersion
	} else if _, err := request.SaveAs(mojangVer.ServerURL, filepath.Join(FolderPath, ServerMain), nil); err != nil {
		return err
	}
	return nil
}
