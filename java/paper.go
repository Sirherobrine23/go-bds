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
	Version       map[string]PaperVersion
}

type PaperVersion struct {
	MCTarget string
	JVM      uint // Java version
	FileURL  []struct {
		Type, Name, URL string
	}
}

func listPaperProject(ProjectTarget string) (map[string]PaperVersion, error) {
	if !slices.Contains(paperProjects, ProjectTarget) {
		return nil, ErrNoFoundVersion
	}

	var projectVersions struct {
		Versions []string `json:"versions"`
	}
	_, err := request.JSONDo(fmt.Sprintf(paperProjectURL, ProjectTarget), &projectVersions, nil)
	if err != nil {
		return nil, err
	}

	mcVersion, err := mojangList()
	if err != nil {
		return nil, err
	}

	Version := make(map[string]PaperVersion)
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

		if _, err = request.JSONDo(fmt.Sprintf(paperProjectBuildsURL, ProjectTarget, version), &builds, nil); err != nil {
			return nil, err
		}

		latestBuild := builds.Builds[len(builds.Builds)-1]
		root := PaperVersion{
			MCTarget: version,
			FileURL: []struct {
				Type string
				Name string
				URL  string
			}{},
		}

		if mcVer, ok := mcVersion[version]; ok {
			root.JVM = mcVer.JVMVersion
		} else {
			switch {
			case semver.New("1.12").LessThan(*semver.New(version)):
				root.JVM = 8
			case semver.New("1.16").LessThan(*semver.New(version)):
				root.JVM = 11
			case semver.New("1.17").LessThan(*semver.New(version)):
				root.JVM = 16
			case semver.New("1.20").LessThan(*semver.New(version)):
				root.JVM = 17
			default:
				root.JVM = 21
			}
		}

		for k, v := range latestBuild.Downloads {
			root.FileURL = append(root.FileURL, struct {
				Type string
				Name string
				URL  string
			}{k, v.Name, fmt.Sprintf(paperProjectGetBuildsURL, ProjectTarget, version, latestBuild.Build, v.Name)})
		}
		Version[version] = root
	}

	return Version, nil
}

func (paper *PaperSearch) Find(version string) (_ Version, err error) {
	if len(paper.Version) == 0 {
		if paper.Version, err = listPaperProject(paper.ProjectTarget); err != nil {
			return nil, err
		}
	}

	if ver, ok := paper.Version[version]; ok {
		return ver, nil
	}
	return nil, ErrNoFoundVersion
}

func (ver PaperVersion) SemverVersion() *semver.Version { return semver.New(ver.MCTarget) }
func (ver PaperVersion) JavaVersion() uint              { return ver.JVM }
func (ver PaperVersion) Install(path string) error {
	for _, asset := range ver.FileURL {
		name := asset.Name
		if asset.Type == "application" {
			name = ServerMain
		}
		if _, err := request.SaveAs(asset.URL, filepath.Join(path, name), nil); err != nil {
			return err
		}
	}
	return nil
}
