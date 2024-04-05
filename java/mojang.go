package java

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"golang.org/x/mod/semver"
	"sirherobrine23.org/minecraft-server/go-bds/internal/request"
)

type MojangVersion struct {
	Version     string    `json:"id"`
	ReleaseType string    `json:"type"`
	ReleaseDate time.Time `json:"releaseTime"`
	UrlServer   string    `json:"url"`
}

func (PistonRelease *MojangVersion) Download(serverPath string) error {
	res, err := request.Request(request.RequestOptions{Method: "GET", HttpError: true, Url: PistonRelease.UrlServer})
	if err != nil {
		return err
	}

	// Create directory if not exists
	defer res.Body.Close()
	os.MkdirAll(serverPath, os.FileMode(0o666))
	serverJarFile, err := os.Create(filepath.Join(serverPath, DefaultServerJarName))
	if err != nil {
		return err
	}

	defer serverJarFile.Close()
	if _, err = io.Copy(serverJarFile, res.Body); err != nil {
		return err
	}

	// Accept eula
	err = os.WriteFile(filepath.Join(serverPath, "eula.txt"), []byte("eula=true\n"), os.FileMode(0o666))

	// Set file to exec in unix system, windows is ignored
	serverJarFile.Chmod(os.FileMode(0x775))
	return err
}

type mojangPistonPackage struct {
	FilesDownloads map[string]struct {
		FileSize int64  `json:"size"`
		FileUrl  string `json:"url"`
		Sha1     string `json:"sha1"`
	} `json:"downloads"`
	Java struct {
		VersionMajor int32  `json:"majorVersion"`
		Component    string `json:"component"`
	} `json:"javaVersion"`
}

func MojangListVersions() ([]MojangVersion, error) {
	var data struct {
		Latest   map[string]string `json:"latest"`
		Versions []MojangVersion   `json:"versions"`
	}

	err := request.GetJson("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json", &data)
	if err != nil {
		return nil, err
	}

	for versionIndex, version := range data.Versions {
		var pistonInfo mojangPistonPackage
		err = request.GetJson(version.UrlServer, &pistonInfo)
		if err != nil {
			return nil, err
		}

		data.Versions[versionIndex].UrlServer = ""
		for name, v := range pistonInfo.FilesDownloads {
			if name == "server" {
				data.Versions[versionIndex].UrlServer = v.FileUrl
			}
		}
	}

	sort.Slice(data.Versions, func(i, j int) bool {
		return semver.Compare(data.Versions[i].Version, data.Versions[j].Version) == 1
	})
	return data.Versions, nil
}

func MojangDownload(version, serverPath string) (MojangVersion, error) {
	var VersionManifest struct {
		Latest   map[string]string `json:"latest"`
		Versions []MojangVersion   `json:"versions"`
	}

	err := request.GetJson("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json", &VersionManifest)
	if err != nil {
		return MojangVersion{}, err
	}

	for _, versionInfo := range VersionManifest.Versions {
		if versionInfo.Version == version || version == "latest" && versionInfo.Version == VersionManifest.Latest["release"] {
			var pistonInfo mojangPistonPackage
			err = request.GetJson(versionInfo.UrlServer, &pistonInfo)
			if err != nil {
				return MojangVersion{}, err
			}

			for name, file := range pistonInfo.FilesDownloads {
				if name == "server" {
					versionInfo.UrlServer = file.FileUrl
					err := versionInfo.Download(serverPath)
					return versionInfo, err
				}
			}
		}
	}

	return MojangVersion{}, ErrVersionNotExist
}
