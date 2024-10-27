package java

import (
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/internal/semver"
	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

var (
	_ VersionSearch = &PaperSearch{}
	_ Version       = &PaperVersion{}

	paperProjects            []string = []string{"paper", "folia", "velocity"}
	paperProjectURL          string   = "https://api.papermc.io/v2/projects/%s"
	paperProjectBuildsURL    string   = "https://api.papermc.io/v2/projects/%s/versions/%s/builds"
	paperProjectGetBuildsURL string   = "https://api.papermc.io/v2/projects/%s/versions/%s/builds/%d/downloads/%s"
)

type PaperSearch struct {
	ProjectTarget string
	Versions      map[string]PaperVersion
}

type PaperVersion struct {
	MCTarget  string
	BuildDate time.Time
	JVM       uint
	FileURL   []struct {
		Type, Name, URL string
	}
}

func (paper PaperSearch) Find(version string) (Version, error) {
	if ver, ok := paper.Versions[version]; ok {
		return ver, nil
	}
	return nil, ErrNoFoundVersion
}

func (paper *PaperSearch) List() error {
	if !slices.Contains(paperProjects, paper.ProjectTarget) {
		return ErrNoFoundVersion
	}
	paper.Versions = make(map[string]PaperVersion)

	var projectVersions struct {
		Versions []string `json:"versions"`
	}
	_, err := request.JSONDo(fmt.Sprintf(paperProjectURL, paper.ProjectTarget), &projectVersions, nil)
	if err != nil {
		return err
	}

	for _, version := range projectVersions.Versions {
		var builds struct {
			Builds []struct {
				Build     int64     `json:"build"`
				BuildTime time.Time `json:"time"`
				Downloads map[string]struct {
					Name   string `json:"name"`
					SHA256 string `json:"sha256"`
				} `json:"downloads"`
			} `json:"builds"`
		}

		if _, err = request.JSONDo(fmt.Sprintf(paperProjectBuildsURL, paper.ProjectTarget, version), &builds, nil); err != nil {
			return err
		}

		latestBuild := builds.Builds[len(builds.Builds)-1]
		root := PaperVersion{
			MCTarget:  version,
			BuildDate: latestBuild.BuildTime,
			FileURL: []struct {
				Type string
				Name string
				URL  string
			}{},
		}
		for k, v := range latestBuild.Downloads {
			root.FileURL = append(root.FileURL, struct {
				Type string
				Name string
				URL  string
			}{k, v.Name, fmt.Sprintf(paperProjectGetBuildsURL, paper.ProjectTarget, version, latestBuild.Build, v.Name)})
		}
		paper.Versions[version] = root
	}

	return nil
}

func (ver PaperVersion) SemverVersion() *semver.Version { return semver.New(ver.MCTarget) }
func (ver PaperVersion) JavaVersion() uint {
	switch {
	case semver.New("1.12").LessThan(*ver.SemverVersion()):
		return 8
	case semver.New("1.16").LessThan(*ver.SemverVersion()):
		return 11
	case semver.New("1.17").LessThan(*ver.SemverVersion()):
		return 16
	case semver.New("1.20").LessThan(*ver.SemverVersion()):
		return 17
	default:
		return 21
	}
}
func (ver PaperVersion) Install(path string) error {
	for _, asset := range ver.FileURL {
		var name string
		if name = asset.Name; asset.Type == "application" {
			name = ServerMain
		}
		if _, err := request.SaveAs(asset.URL, filepath.Join(path, name), nil); err != nil {
			return err
		}
	}
	return nil
}
