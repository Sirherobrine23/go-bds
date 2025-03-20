//go:build linux && arm

package javaprebuild

import (
	"fmt"
	"net/url"
	"runtime"

	"sirherobrine23.com.br/go-bds/request/v2"
)

type libericaVersion struct {
	DownloadURL    string `json:"downloadUrl"`
	FeatureVersion int    `json:"featureVersion"`
	Os             string `json:"os"`
	Architecture   string `json:"architecture"`
	Version        string `json:"version"`
	Sha1           string `json:"sha1"`
	Filename       string `json:"filename"`
	Size           int    `json:"size"`
}

// Install latest liberica file
func InstallLatest(featVersion uint, installPath string) error {
	requUrl, _ := url.Parse("https://api.bell-sw.com/v1/liberica/releases")
	query := requUrl.Query()
	query.Set("version-modifier", "latest")
	query.Set("package-type", "tar.gz")
	query.Set("bundle-type", "jre")
	query.Set("version-feature", fmt.Sprint(featVersion))

	// os: "linux", "linux-musl", "macos", "solaris", "windows"
	switch runtime.GOOS {
	case "linux", "windows":
		query.Set("os", runtime.GOOS)
	case "darwin":
		query.Set("os", "macos")
	case "sunos":
		query.Set("os", "solaris")
	default:
		return ErrSystem
	}

	// bitness: "32", "64"
	// arch: "arm", "ppc", "riscv", "sparc", "x86"
	switch runtime.GOARCH {
	case "amd64":
		query.Set("arch", "x86")
		query.Set("bitness", "64")
	case "386":
		query.Set("arch", "x86")
		query.Set("bitness", "32")
	case "arm64":
		query.Set("arch", "arm")
		query.Set("bitness", "64")
	case "arm":
		query.Set("arch", "arm")
		query.Set("bitness", "32")
	case "ppc64":
		query.Set("arch", "ppc")
		query.Set("bitness", "64")
	default:
		return ErrSystem
	}

	requUrl.RawQuery = query.Encode()
	releases, _, err := request.JSON[[]libericaVersion](requUrl.String(), nil)
	if err != nil {
		return err
	} else if len(releases) == 0 {
		return ErrSystem
	}
	return request.Tar(releases[0].DownloadURL, request.ExtractOptions{Strip: 1, Cwd: installPath}, nil)
}
