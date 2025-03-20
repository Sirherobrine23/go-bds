package java

import (
	"fmt"
	"slices"
	"time"

	"sirherobrine23.com.br/go-bds/request/v2"
)

var (
	// Paper projects
	PaperProjects []string = []string{"paper", "folia", "velocity"}

	paperProjectURL          string = "https://api.papermc.io/v2/projects/%s"
	paperProjectBuildsURL    string = "https://api.papermc.io/v2/projects/%s/versions/%s/builds"
	paperProjectGetBuildsURL string = "https://api.papermc.io/v2/projects/%s/versions/%s/builds/%d/downloads/%s"
)

func ListPaper(ProjectTarget string) (ListServer, error) {
	if !slices.Contains(PaperProjects, ProjectTarget) {
		return nil, fmt.Errorf("")
	}
	return func() (Versions, error) {
		var projectVersions struct {
			Versions []string `json:"versions"`
		}
		if _, err := request.DoJSON(fmt.Sprintf(paperProjectURL, ProjectTarget), &projectVersions, nil); err != nil {
			return nil, err
		}

		Version := Versions{}
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

			if _, err := request.DoJSON(fmt.Sprintf(paperProjectBuildsURL, ProjectTarget, version), &builds, nil); err != nil {
				return nil, err
			}

			latestBuild := builds.Builds[len(builds.Builds)-1]
			root := GenericVersion{
				Version:    version,
				JVMVersion: 0,
				URLs: []struct {
					Name string
					URL  string
				}{},
			}

			for k, v := range latestBuild.Downloads {
				switch k {
				case "application":
					root.URLs = append(root.URLs, struct {
						Name string
						URL  string
					}{ServerName, fmt.Sprintf(paperProjectGetBuildsURL, ProjectTarget, version, latestBuild.Build, v.Name)})
				default:
					root.URLs = append(root.URLs, struct {
						Name string
						URL  string
					}{v.Name, fmt.Sprintf(paperProjectGetBuildsURL, ProjectTarget, version, latestBuild.Build, v.Name)})
				}
			}
			Version = append(Version, root)
		}

		return Version, nil
	}, nil
}
