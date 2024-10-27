package java

import (
	"fmt"
	"path/filepath"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/internal/semver"
	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

var (
	_ VersionSearch = PurpurSearch{}
	_ Version       = PurpurVersion{}
)

type PurpurSearch struct {
	Version map[string]PurpurVersion
}

type PurpurVersion struct {
	MCStarget string `json:"version"`
	FileURL   string `json:"fileUrl"`
}

func (purpur PurpurSearch) list() error {
	purpur.Version = make(map[string]PurpurVersion)
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
		return err
	}

	for _, version := range versions.Versions {
		buildInfo, _, err := request.JSON[projectBuilds](fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s", version), nil)
		if err != nil {
			return err
		}
		resBuild, _, err := request.JSON[struct {
			MCStarget string `json:"version"`
			Result    string `json:"result"`
			Time      int64  `json:"timestamp"`
			Build     string `json:"build"`
		}](fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s", version, buildInfo.Builds.Latest), nil)
		if err != nil {
			return err
		}

		if strings.ToUpper(resBuild.Result) != "SUCCESS" {
			for _, build := range buildInfo.Builds.All {
				if _, err = request.JSONDo(fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s", version, build), &resBuild, nil); err != nil {
					return err
				}
				if strings.ToUpper(resBuild.Result) == "SUCCESS" {
					break
				}
			}
		}

		if strings.ToUpper(resBuild.Result) == "SUCCESS" {
			purpur.Version[version] = PurpurVersion{
				MCStarget: resBuild.MCStarget,
				FileURL:   fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s/download", version, resBuild.Build),
			}
		}
	}

	return nil
}

func (purpur PurpurSearch) Find(Version string) (Version, error) {
	if err := purpur.list(); err != nil {
		return nil, err
	}

	if ver, ok := purpur.Version[Version]; ok {
		return ver, nil
	}
	return nil, ErrNoFoundVersion
}

func (ver PurpurVersion) JavaVersion() uint {
	return PaperVersion{MCTarget: ver.MCStarget}.JavaVersion()
}
func (ver PurpurVersion) SemverVersion() *semver.Version { return semver.New(ver.MCStarget) }
func (ver PurpurVersion) Install(path string) error {
	_, err := request.SaveAs(ver.FileURL, filepath.Join(path, ServerMain), nil)
	return err
}
