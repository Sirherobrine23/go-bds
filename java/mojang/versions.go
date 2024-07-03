package mojang

import (
	"time"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

type pistonInfo struct {
	Latest   map[string]string `json:"latest"`
	Versions []struct {
		ID          string    `json:"id"`
		ReleaseType string    `json:"type"`
		ReleaseDate time.Time `json:"releaseTime"`
		Url         string    `json:"url"`
	} `json:"versions"`
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

type Version struct {
	ReleaseType string    `json:"type"`
	ReleaseDate time.Time `json:"releaseTime"`
	UrlServer   string    `json:"url"`
}

type Versions map[string]Version

func Releases() (Versions, error) {
	req := request.RequestOptions{HttpError: true, Url: "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"}
	var pistonVersions pistonInfo
	if _, err := req.Do(&pistonVersions); err != nil {
		return nil, err
	}

	versions := Versions{}
	for _, release := range pistonVersions.Versions {
		var ReleaseInfo mojangPistonPackage
		req := request.RequestOptions{HttpError: true, Url: release.Url}
		if _, err := req.Do(&ReleaseInfo); err != nil {
			return nil, err
		}
		if info, ok := ReleaseInfo.FilesDownloads["server"]; ok {
			versions[release.ID] = Version{
				ReleaseType: release.ReleaseType,
				ReleaseDate: release.ReleaseDate,
				UrlServer:   info.FileUrl,
			}
		}
	}

	return versions, nil
}
