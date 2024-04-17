package zulu

import (
	"runtime"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/cache"
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

func reverseReleases(arr []zuluBundle) {
	left := 0
	right := len(arr) - 1
	for left < right {
		arr[left], arr[right] = arr[right], arr[left]
		left++
		right--
	}
}

func Releases() (globals.Version, error) {
	if cache.Get("zulu", "releases") != nil {
		values, ok := cache.Get("zulu", "releases").(globals.Version)
		if ok {
			return values, nil
		}
	}

	opts := request.RequestOptions{
		HttpError: true,
		Url:       "https://api.azul.com/zulu/download/community/v1.0/bundles/",
		Querys: map[string]string{
			"bundle_type": "jdk",
		},
	}

	opts.Querys["os"] = runtime.GOOS
	if opts.Querys["os"] == "darwin" {
		opts.Querys["os"] = "macos"
	}

	opts.Querys["arch"] = runtime.GOARCH
	if opts.Querys["arch"] == "arm64" {
		opts.Querys["arch"] = "arm"
		opts.Querys["hw_bitness"] = "64"
	} else if opts.Querys["arch"] == "amd64" {
		opts.Querys["arch"] = "x86"
		opts.Querys["hw_bitness"] = "64"
	} else if opts.Querys["arch"] == "arm" {
		opts.Querys["arch"] = "arm"
		opts.Querys["hw_bitness"] = "32"
	} else if opts.Querys["arch"] == "386" {
		opts.Querys["arch"] = "x86"
		opts.Querys["hw_bitness"] = "32"
	} else if opts.Querys["arch"] == "ppc" {
		opts.Querys["arch"] = "ppc"
		opts.Querys["hw_bitness"] = "32"
	}

	versions := globals.Version{}
	var zuluReleases []zuluBundle
	opts.Do(&zuluReleases)

	reverseReleases(zuluReleases)
	for _, release := range zuluReleases {
		versions[release.JavaVersion[0]] = globals.VersionBundle{
			FileUrl:  release.URL,
			Checksum: "",
		}
	}

	cache.Set("zulu", "releases", versions, globals.DefaultTime)
	return versions, nil
}
