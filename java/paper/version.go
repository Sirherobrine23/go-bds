package paper

import (
	"fmt"
	"io"
	"time"

	"sirherobrine23.org/go-bds/go-bds/request"
)

type Version struct {
	BuildDate time.Time `json:"build"`
	FileURL   string    `json:"fileUrl"`
}

type Versions struct {
	Project  string                        `json:"project"`
	Versions map[string]map[string]Version `json:"versions"`
}

func (paper *Versions) Releases() error {
	paper.Versions = map[string]map[string]Version{}
	if !(paper.Project == "paper" || paper.Project == "folia" || paper.Project == "velocity") {
		return fmt.Errorf(`invalid project, accept only, "paper", "folia" or "velocity"`)
	}

	req := request.RequestOptions{HttpError: true, Url: fmt.Sprintf("https://api.papermc.io/v2/projects/%s", paper.Project)}
	var projectVersions struct {
		Versions []string `json:"versions"`
	}
	_, err := req.Do(&projectVersions)
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

		req.Url = fmt.Sprintf("https://api.papermc.io/v2/projects/%s/versions/%s/builds", paper.Project, version)
		_, err = req.Do(&builds)
		if err != nil {
			return err
		}

		buildInfo := builds.Builds[len(builds.Builds)-1]
		paper.Versions[version] = map[string]Version{}
		for k, v := range buildInfo.Downloads {
			paper.Versions[version][k] = Version{
				BuildDate: buildInfo.BuildTime,
				FileURL:   fmt.Sprintf("https://api.papermc.io/v2/projects/%s/versions/%s/builds/%d/downloads/%s", paper.Project, version, buildInfo.Build, v.Name),
			}
		}
	}

	return nil
}

func (w *Version) Get() (io.ReadCloser, error) {
	res, err := (&request.RequestOptions{HttpError: true, Url: w.FileURL}).Request()
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}
