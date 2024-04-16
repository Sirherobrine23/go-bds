package liberica

import (
	"encoding/json"
	"fmt"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/cache"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/java/globals"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

const (
	APIReleases string = "https://api.bell-sw.com/v1/liberica/releases?bundle-type=jre&installation-type=archive"
	APIFile     string = "https://api.bell-sw.com/v1/liberica/releases/%s"
)

type release struct {
	Bitness                int    `json:"bitness"`
	LatestLTS              bool   `json:"latestLTS"`
	UpdateVersion          int    `json:"updateVersion"`
	DownloadURL            string `json:"downloadUrl"`
	LatestInFeatureVersion bool   `json:"latestInFeatureVersion"`
	LTS                    bool   `json:"LTS"`
	BundleType             string `json:"bundleType"`
	FeatureVersion         int    `json:"featureVersion"`
	PackageType            string `json:"packageType"`
	FX                     bool   `json:"FX"`
	GA                     bool   `json:"GA"`
	Architecture           string `json:"architecture"`
	Latest                 bool   `json:"latest"`
	ExtraVersion           int    `json:"extraVersion"`
	BuildVersion           int    `json:"buildVersion"`
	EOL                    bool   `json:"EOL"`
	Os                     string `json:"os"`
	InterimVersion         int    `json:"interimVersion"`
	Version                string `json:"version"`
	Sha1                   string `json:"sha1"`
	Filename               string `json:"filename"`
	InstallationType       string `json:"installationType"`
	Size                   int    `json:"size"`
	PatchVersion           int    `json:"patchVersion"`
	TCK                    bool   `json:"TCK"`
	UpdateType             string `json:"updateType"`
}

func Releases() ([]globals.Version, error) {
	if cache.Get("liberica", "releases") != nil {
		value, ok := cache.Get("liberica", "releases").([]globals.Version)
		if ok {
			return value, nil
		}
	}

	res, err := request.Request(request.RequestOptions{HttpError: true, Url: APIReleases})
	if err != nil {
		return []globals.Version{}, err
	}

	defer res.Body.Close()
	var versionReleases []release
	if err = json.NewDecoder(res.Body).Decode(&versionReleases); err != nil {
		return []globals.Version{}, err
	}

	releases := map[string]globals.Version{}
	for _, release := range versionReleases {
		if release.Os == "linux-musl" || !(release.PackageType == "zip" || release.PackageType == "tar.gz") {
			continue
		}

		goos := release.Os
		goarch := release.Architecture
		if release.Bitness == 64 {
			if release.Architecture == "arm" {
				goarch = "arm64"
			} else if release.Architecture == "x86" {
				goarch = "amd64"
			} else if release.Architecture == "ppc" {
				goarch = "ppc64"
			}
		} else {
			if release.Architecture == "x86" {
				goarch = "386"
			}
		}

		if _, exist := releases[release.Version]; !exist {
			releases[release.Version] = globals.Version{
				Version: release.Version,
				Targets: map[string]string{},
			}
		}

		releases[release.Version].Targets[fmt.Sprintf("%s/%s", goos, goarch)] = fmt.Sprintf(APIFile, release.Filename)
	}

	arrVersions := []globals.Version{}
	for _, v := range releases {
		arrVersions = append(arrVersions, v)
	}

	cache.Set("liberica", "releases", arrVersions, globals.DefaultTime)
	return arrVersions, nil
}
