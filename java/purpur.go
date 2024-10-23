package java

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

type purpurVersion struct {
	MCStarget string    `json:"version"`
	BuildDate time.Time `json:"build"`
	FileURL   string    `json:"fileUrl"`
}

// Get all releases from purpur server
func PurpurReleases() (Releases, error) {
	type projectVersions struct {
		Versions []string `json:"versions"`
	}

	type projectBuilds struct {
		Builds struct {
			Latest string   `json:"latest"`
			All    []string `json:"all"`
		} `json:"builds"`
	}

	versions, _, err := request.JSON[projectVersions]("https://api.purpurmc.org/v2/purpur", nil)
	if err != nil {
		return nil, err
	}

	mapInfo := make(Releases)
	for _, version := range versions.Versions {
		buildInfo, _, err := request.JSON[projectBuilds](fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s", version), nil)
		if err != nil {
			return mapInfo, err
		}
		resBuild, _, err := request.JSON[struct {
			MCStarget string `json:"version"`
			Result    string `json:"result"`
			Time      int64  `json:"timestamp"`
			Build     string `json:"build"`
		}](fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s", version, buildInfo.Builds.Latest), nil)
		if err != nil {
			return mapInfo, err
		}

		if strings.ToUpper(resBuild.Result) != "SUCCESS" {
			for _, build := range buildInfo.Builds.All {
				if _, err = request.JSONDo(fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s", version, build), &resBuild, nil); err != nil {
					return mapInfo, err
				}
				if strings.ToUpper(resBuild.Result) == "SUCCESS" {
					break
				}
			}
		}

		if strings.ToUpper(resBuild.Result) == "SUCCESS" {
			mapInfo[version] = purpurVersion{
				MCStarget: resBuild.MCStarget,
				BuildDate: time.UnixMilli(resBuild.Time),
				FileURL:   fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s/download", version, resBuild.Build),
			}
		}
	}
	return mapInfo, nil
}

func (w purpurVersion) ReleaseType() string    { return "oficial" }
func (w purpurVersion) String() string         { return w.MCStarget }
func (w purpurVersion) ReleaseTime() time.Time { return w.BuildDate }
func (w purpurVersion) Download(folder string) error {
	_, err := request.SaveAs(w.FileURL, filepath.Join(folder, ServerMain), nil)
	return err
}
