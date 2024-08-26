package purpur

import (
	"fmt"
	"io"
	"strings"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/request"
)

type projectVersions struct {
	Versions []string `json:"versions"`
}

type projectBuilds struct {
	Builds struct {
		Latest string   `json:"latest"`
		All    []string `json:"all"`
	} `json:"builds"`
}

type Version struct {
	BuildDate time.Time `json:"build"`
	FileURL   string    `json:"fileUrl"`
}

func Releases() (map[string]Version, error) {
	mapInfo := map[string]Version{}
	req := request.RequestOptions{Url: "https://api.purpurmc.org/v2/purpur", HttpError: true}
	var versions projectVersions
	_, err := req.Do(&versions)
	if err != nil {
		return mapInfo, err
	}

	for _, version := range versions.Versions {
		req.Url = fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s", version)
		var buildInfo projectBuilds
		if _, err = req.Do(&buildInfo); err != nil {
			return mapInfo, err
		}

		var resBuild struct {
			Result string `json:"result"`
			Time   int64  `json:"timestamp"`
			Build  string `json:"build"`
		}
		req.Url = fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s", version, buildInfo.Builds.Latest)
		if _, err = req.Do(&resBuild); err != nil {
			return mapInfo, err
		}

		if strings.ToUpper(resBuild.Result) != "SUCCESS" {
			for _, build := range buildInfo.Builds.All {
				req.Url = fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s", version, build)
				if _, err = req.Do(&resBuild); err != nil {
					return mapInfo, err
				}
				if strings.ToUpper(resBuild.Result) == "SUCCESS" {
					break
				}
			}
		}

		if strings.ToUpper(resBuild.Result) == "SUCCESS" {
			mapInfo[version] = Version{
				BuildDate: time.UnixMilli(resBuild.Time),
				FileURL:   fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s/download", version, resBuild.Build),
			}
		}
	}
	return mapInfo, nil
}

func (w *Version) Get() (io.ReadCloser, error) {
	res, err := (&request.RequestOptions{HttpError: true, Url: w.FileURL}).Request()
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}
