//go:build (aix && ppc64) || darwin || (windows && amd64) || (linux && (amd64 || arm64 || riscv64 || ppc64le || s390x)) || (solaris && (amd64 || sparcv9))

// Download java with builds from adoptium
package javaprebuild

import (
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"sirherobrine23.com.br/go-bds/request/v2"
)

type adoptiumReleases struct {
	MostRecentRelease uint   `json:"most_recent_feature_release"`
	MostRecentLts     uint   `json:"most_recent_lts"`
	MostRecent        uint   `json:"most_recent_feature_version"`
	TIP               uint   `json:"tip_version"`
	LTSRelease        []uint `json:"available_lts_releases"`
	Release           []uint `json:"available_releases"`
}

// Install latest version from adoptium
func (ver JavaVersion) InstallLatest(installPath string) error {
	featVersion := uint(ver) - 44
	releases, _, _ := request.JSON[adoptiumReleases]("https://api.adoptium.net/v3/info/available_releases", nil)
	if !slices.Contains(releases.Release, featVersion) {
		return ErrSystem
	}

	// architecture: x64, x86, x32, ppc64, ppc64le, s390x, aarch64, arm, sparcv9, riscv64
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "x64"
	case "386":
		arch = "x86"
	case "arm64":
		arch = "aarch64"
	}

	// goos: linux, windows, mac, solaris, aix, alpine-linux
	goos := runtime.GOOS
	switch goos {
	case "darwin":
		goos = "mac"
	case "sunos":
		goos = "solaris"
	}

	processRedirect := func(res *http.Response) (*http.Response, error) {
		defer res.Body.Close()
		var RequestURL string
		if RequestURL = res.Header.Get("Location"); RequestURL == "" {
			if RequestURL = res.Header.Get("location"); RequestURL == "" {
				return res, ErrSystem
			}
		}
		extractOptions := request.ExtractOptions{Cwd: installPath, Strip: 1}
		if strings.ToLower(filepath.Ext(RequestURL)) == ".zip" {
			return res, request.Zip(RequestURL, extractOptions, nil)
		}
		return res, request.Tar(RequestURL, extractOptions, nil)
	}

	_, err := request.Request(fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/%d/ga/%s/%s/jdk/hotspot/normal/eclipse", featVersion, goos, arch), &request.Options{
		NotFollowRedirect: true,
		CodeProcess: request.MapCode{
			301: processRedirect,
			302: processRedirect,
			307: processRedirect,
			404: func(res *http.Response) (*http.Response, error) {
				if res != nil {
					res.Body.Close()
				}
				return nil, ErrSystem
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
