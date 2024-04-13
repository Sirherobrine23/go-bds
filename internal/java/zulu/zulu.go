package zulu

import (
	"encoding/json"
	"sort"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/mod/semver"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/java/globals"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

const (
	APIReleases string = "https://api.azul.com/zulu/download/community/v1.0/bundles/?bundle_type=jdk"
)

type zuluBundle struct {
	JavaVersion        []int  `json:"java_version"`
	JdkVersion         []int  `json:"jdk_version"`
	Name               string `json:"name"`
	OpenjdkBuildNumber int    `json:"openjdk_build_number"`
	URL                string `json:"url"`
	ZuluVersion        []int  `json:"zulu_version"`
}

func joinVersion(i []int) string {
	rel := []string{}
	for _, k := range i {
		rel = append(rel, strconv.Itoa(k))
	}
	return strings.Join(rel, ".")
}

var oss [][]string = [][]string{
	{
		"linux", "amd64",
	},
	{
		"linux", "arm64",
	},
	{
		"linux", "386",
	},
	{
		"darwin", "amd64",
	},
	{
		"darwin", "arm64",
	},
	{
		"windows", "386",
	},
	{
		"windows", "amd64",
	},
	{
		"windows", "arm64",
	},
}

func Releases() ([]globals.Version, error) {
	versionsGlobal := map[string]globals.Version{}
	for _, goTarget := range oss {
		goos := goTarget[0]
		goarch := goTarget[1]
		opts := request.RequestOptions{
			HttpError: true,
			Url:       "https://api.azul.com/zulu/download/community/v1.0/bundles/",
			Querys: map[string]string{
				"bundle_type": "jdk",
			},
		}

		opts.Querys["os"] = goos
		if goos == "darwin" {
			opts.Querys["os"] = "macos"
		}
		if goarch == "arm64" {
			opts.Querys["arch"] = "arm"
			opts.Querys["hw_bitness"] = "64"
		} else if goarch == "amd64" {
			opts.Querys["arch"] = "x86"
			opts.Querys["hw_bitness"] = "64"
		} else if goarch == "arm" {
			opts.Querys["arch"] = "arm"
			opts.Querys["hw_bitness"] = "32"
		} else if goarch == "386" {
			opts.Querys["arch"] = "x86"
			opts.Querys["hw_bitness"] = "32"
		} else if goarch == "ppc" {
			opts.Querys["arch"] = "ppc"
			opts.Querys["hw_bitness"] = "32"
		}

		res, err := request.Request(opts)
		var versions []zuluBundle
		if err != nil {
			return []globals.Version{}, err
		}
		defer res.Body.Close()
		if err = json.NewDecoder(res.Body).Decode(&versions); err != nil {
			return []globals.Version{}, err
		}

		for _, release := range versions {
			if !(strings.HasSuffix(release.Name, ".zip") || strings.HasSuffix(release.Name, ".tar.gz")) {
				continue
			}

			version := joinVersion(release.JdkVersion)
			if _, ok := versionsGlobal[version]; !ok {
				versionsGlobal[version] = globals.Version{
					Version: version,
					Targets: map[string]string{},
				}
			}
			versionsGlobal[version].Targets[fmt.Sprintf("%s/%s", goos, goarch)] = release.URL
		}
	}

	versionsArr := []globals.Version{}
	for _, k := range versionsGlobal { versionsArr = append(versionsArr, k) }
	sort.Slice(versionsArr, func(i, j int) bool {
		n := versionsArr[i].Version
		b := versionsArr[j].Version
		if !semver.IsValid(n) { n = fmt.Sprintf("v%s", n) }
		if !semver.IsValid(b) { b = fmt.Sprintf("v%s", b) }
		n = strings.Join(strings.Split(n, ".")[0:3], ".")
		b = strings.Join(strings.Split(b, ".")[0:3], ".")
		return semver.Compare(n, b) == 1
	})
	return versionsArr, nil
}
