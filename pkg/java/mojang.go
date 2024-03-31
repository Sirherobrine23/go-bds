package java

import (
	"fmt"
	"time"

	"sirherobrine23.org/minecraft-server/go-bds/internal/request"
)

type MojangVersion struct {
	Version     string    `json:"id"`
	ReleaseType string    `json:"type"`
	ReleaseDate time.Time `json:"releaseTime"`
	UrlServer   string    `json:"url"`
}

type mojangPistonPackage struct {
	FilesDownloads map[string]struct {
		FileSize float64 `json:"size"`
		FileUrl  string  `json:"url"`
		Sha1     string  `json:"sha1"`
	} `json:"downloads"`
}

func GetMojangVersions() ([]MojangVersion, error) {
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
				fmt.Println(version.Version, v.FileUrl)
			}
		}
	}

	return data.Versions, nil
}
