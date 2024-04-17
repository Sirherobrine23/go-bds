package liberica

import (
	"fmt"
	"runtime"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/cache"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/java/globals"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
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

func reverseReleases(arr []release) {
	left := 0
	right := len(arr) - 1
	for left < right {
		arr[left], arr[right] = arr[right], arr[left]
		left++
		right--
	}
}

func Releases() (globals.Version, error) {
	if cache.Get("liberica", "releases") != nil {
		value, ok := cache.Get("liberica", "releases").(globals.Version)
		if ok {
			return value, nil
		}
	}

	req := request.RequestOptions{
		HttpError: true,
		Url:       "https://api.bell-sw.com/v1/liberica/releases",
		Querys: map[string]string{
			"installation-type": "archive",
			"bundle-type":       "jre",
			"arch":              runtime.GOARCH,
			"os":                runtime.GOOS,
		},
	}

	// 'linux', 'linux-musl', 'macos', 'solaris', 'windows';
	switch runtime.GOOS {
	case "darwin":
		req.Querys["os"] = "macos"
	case "windows":
		req.Querys["os"] = "windows"
	case "android":
		req.Querys["os"] = "linux-musl"
	case "linux":
	case "solaris":
		req.Querys["os"] = runtime.GOOS
	default:
		return nil, globals.ErrNoSupportOs
	}

	switch runtime.GOARCH {
	case "arm64":
		req.Querys["bitness"] = "64"
		req.Querys["arch"] = "arm"
	case "arm":
		req.Querys["bitness"] = "32"
		req.Querys["arch"] = "arm"
	case "amd64":
		req.Querys["bitness"] = "64"
		req.Querys["arch"] = "x86"
	case "386":
		req.Querys["bitness"] = "32"
		req.Querys["arch"] = "x86"
	case "ppc64":
	case "ppc64le":
		req.Querys["bitness"] = "64"
		req.Querys["arch"] = "ppc"
	default:
		return nil, globals.ErrNoSupportArch
	}

	var rels []release
	_, err := req.Do(&rels)
	if err != nil {
		return nil, err
	}

	reverseReleases(rels)
	version := globals.Version{}
	for _, release := range rels {
		version[release.FeatureVersion] = globals.VersionBundle{
			FileUrl:  release.DownloadURL,
			Checksum: fmt.Sprintf("sha1:%s", release.Sha1),
		}
	}

	return version, nil
}
