package mojang

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"golang.org/x/mod/semver"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

var (
	DefaultServerJarName = "server.jar"                             // Default name to save .jar files
	ErrVersionNotExist   = errors.New("version not exists to java") // If version not exist
	ErrNoJava            = errors.New("cannot find java")           // Cannot get java path to run server
	ErrInstallServer     = errors.New("install server fist")        // Server not installed
)

type pistonInfo struct {
	Latest   map[string]string `json:"latest"`
	Versions []MojangVersion   `json:"versions"`
}

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
	var pistonVersions pistonInfo

	res, err := request.Request(request.RequestOptions{HttpError: true, Method: "GET", Url: "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"})
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if err = json.NewDecoder(res.Body).Decode(&pistonVersions); err != nil {
		return nil, err
	}

	for versionIndex, version := range pistonVersions.Versions {
		var pistonInfo mojangPistonPackage
		res, err := request.Request(request.RequestOptions{HttpError: true, Method: "GET", Url: version.UrlServer})
		if err != nil {
			return nil, err
		}

		defer res.Body.Close()
		if err = json.NewDecoder(res.Body).Decode(&pistonInfo); err != nil {
			return nil, err
		}

		pistonVersions.Versions[versionIndex].UrlServer = ""
		for name, v := range pistonInfo.FilesDownloads {
			if name == "server" {
				pistonVersions.Versions[versionIndex].UrlServer = v.FileUrl
			}
		}
	}

	sort.Slice(pistonVersions.Versions, func(i, j int) bool {
		return semver.Compare(pistonVersions.Versions[i].Version, pistonVersions.Versions[j].Version) == 1
	})
	return pistonVersions.Versions, nil
}

func MojangDownload(version, serverPath string) (MojangVersion, error) {
	var pistonVersions pistonInfo

	res, err := request.Request(request.RequestOptions{HttpError: true, Method: "GET", Url: "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"})
	if err != nil {
		return MojangVersion{}, err
	}

	defer res.Body.Close()
	if err = json.NewDecoder(res.Body).Decode(&pistonVersions); err != nil {
		return MojangVersion{}, err
	}

	for _, versionInfo := range pistonVersions.Versions {
		if versionInfo.Version == version || version == "latest" && versionInfo.Version == pistonVersions.Latest["release"] {
			var pistonInfo mojangPistonPackage

			res, err := request.Request(request.RequestOptions{HttpError: true, Method: "GET", Url: versionInfo.UrlServer})
			if err != nil {
				return MojangVersion{}, err
			}

			defer res.Body.Close()
			if err = json.NewDecoder(res.Body).Decode(&pistonInfo); err != nil {
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
