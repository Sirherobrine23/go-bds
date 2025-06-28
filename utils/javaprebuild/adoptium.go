package javaprebuild

import (
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"sirherobrine23.com.br/go-bds/request/v2"
	"sirherobrine23.com.br/sirherobrine23/go-dpkg/dpkg"
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
func (ver JavaVersion) InstallLatestAdoptium(installPath string) error {
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

	opt := &request.Options{
		NotFollowRedirect: true,
		CodeProcess: request.MapCode{
			301: func(res *http.Response) (*http.Response, error) {
				res.StatusCode = 200
				return res, nil
			},
			302: func(res *http.Response) (*http.Response, error) {
				res.StatusCode = 200
				return res, nil
			},
		},
	}

	res, err := request.Request(fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/%d/ga/%s/%s/jdk/hotspot/normal/eclipse", featVersion, goos, arch), opt)
	if err != nil {
		return err
	}

	downloadUrl := res.Header.Get("Location")
	if downloadUrl == "" {
		return ErrSystem
	}

	switch strings.ToLower(filepath.Ext(path.Base(downloadUrl))) {
	case ".tar", ".tar.gz":
		return request.Tar(downloadUrl, request.ExtractOptions{Strip: 1, Cwd: installPath}, nil)
	case ".zip":
		return request.Zip(downloadUrl, request.ExtractOptions{Strip: 1, Cwd: installPath}, nil)
	case ".deb":
		res, err := request.Request(downloadUrl, nil)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		_, pkgData, err := dpkg.NewReader(res.Body)
		if err != nil {
			return err
		}
		return extractTar(pkgData, request.ExtractOptions{Strip: 0, Cwd: installPath})
	}

	return nil
}
