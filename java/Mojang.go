package java

import (
	"path/filepath"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

type pistonInfo struct {
	Latest   map[string]string `json:"latest"`
	Versions []struct {
		Url string `json:"url"`
	} `json:"versions"`
}

type mojangPistonPackage struct {
	Version        string    `json:"id"`
	Type           string    `json:"type"`
	ReleaseDate    time.Time `json:"releaseTime"`
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

type Releases map[string]Release
type Release interface {
	String() string               // Version string
	ReleaseType() string          // Release type: ("oficial", "alpha", "snapshot")
	ReleaseTime() time.Time       // Release date
	Download(folder string) error // Download server to folder
}

// Get all releases from mojang servers
func MojangReleases() (Releases, error) {
	pistonVersions, _, err := request.JSON[pistonInfo]("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json", nil)
	if err != nil {
		return nil, err
	}

	versions := make(Releases)
	for _, release := range pistonVersions.Versions {
		ReleaseInfo, _, err := request.JSON[mojangPistonPackage](release.Url, nil)
		if err != nil {
			return nil, err
		} else if _, ok := ReleaseInfo.FilesDownloads["server"]; ok {
			versions[ReleaseInfo.Version] = ReleaseInfo
		}
	}

	return versions, nil
}

func (rel mojangPistonPackage) String() string         { return rel.Version }
func (rel mojangPistonPackage) ReleaseType() string    { return rel.Type }
func (rel mojangPistonPackage) ReleaseTime() time.Time { return rel.ReleaseDate }
func (rel mojangPistonPackage) Download(folder string) error {
	if server, ok := rel.FilesDownloads["server"]; ok {
		_, err := request.SaveAs(server.FileUrl, filepath.Join(folder, ServerMain), nil)
		return err
	}
	return ErrNoServer
}
