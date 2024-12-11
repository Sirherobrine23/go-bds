package java

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

var _ ListServer = ListPurpur

func ListPurpur() (Versions, error) {
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

	Version := make(map[string]Version)
	for _, version := range versions.Versions {
		buildInfo, _, err := request.JSON[projectBuilds](fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s", version), nil)
		if err != nil {
			return nil, err
		}
		resBuild, _, err := request.JSON[struct {
			MCStarget string `json:"version"`
			Result    string `json:"result"`
			Time      int64  `json:"timestamp"`
			Build     string `json:"build"`
		}](fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s", version, buildInfo.Builds.Latest), nil)
		if err != nil {
			return nil, err
		}

		if strings.ToUpper(resBuild.Result) != "SUCCESS" {
			for _, build := range buildInfo.Builds.All {
				if _, err = request.JSONDo(fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s", version, build), &resBuild, nil); err != nil {
					return nil, err
				}
				if strings.ToUpper(resBuild.Result) == "SUCCESS" {
					break
				}
			}
		}

		if strings.ToUpper(resBuild.Result) == "SUCCESS" {
			Version[version] = GenericVersion{
				Version: resBuild.MCStarget,
				URLs: []struct {
					Name string
					URL  string
				}{
					{
						ServerName,
						fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s/download", version, resBuild.Build),
					},
				},
			}
		}
	}

	return slices.Collect(maps.Values(Version)), nil
}
