package java

import (
	"fmt"
	"path/filepath"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

type paperProjectVersion struct {
	MCTarget  string
	BuildDate time.Time
	FileURL   []struct {
		Type, Name, URL string
	}
}

// Get all releases from paper project
func PaperReleases(Project string) (Releases, error) {
	if !(Project == "paper" || Project == "folia" || Project == "velocity") {
		return nil, fmt.Errorf(`invalid project, accept only, "paper", "folia" or "velocity"`)
	}

	var projectVersions struct {
		Versions []string `json:"versions"`
	}
	_, err := request.JSONDo(fmt.Sprintf("https://api.papermc.io/v2/projects/%s", Project), &projectVersions, nil)
	if err != nil {
		return nil, err
	}

	release := make(Releases)
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

		if _, err = request.JSONDo(fmt.Sprintf("https://api.papermc.io/v2/projects/%s/versions/%s/builds", Project, version), &builds, nil); err != nil {
			return release, err
		}

		latestBuild := builds.Builds[len(builds.Builds)-1]
		root := paperProjectVersion{
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
			}{k, v.Name, fmt.Sprintf("https://api.papermc.io/v2/projects/%s/versions/%s/builds/%d/downloads/%s", Project, version, latestBuild.Build, v.Name)})
		}
		release[version] = root
	}

	return release, nil
}

func (w paperProjectVersion) ReleaseType() string    { return "oficial" }
func (w paperProjectVersion) String() string         { return w.MCTarget }
func (w paperProjectVersion) ReleaseTime() time.Time { return w.BuildDate }
func (w paperProjectVersion) Download(folder string) error {
	for _, asset := range w.FileURL {
		var name string
		if name = asset.Name; asset.Type == "application" {
			name = ServerMain
		}
		if _, err := request.SaveAs(asset.URL, filepath.Join(folder, name), nil); err != nil {
			return err
		}
	}
	return nil
}
