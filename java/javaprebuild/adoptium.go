//go:build (aix && ppc64) || darwin || (windows && amd64) || (linux && (amd64 || arm64 || riscv64 || ppc64le || s390x)) || (solaris && (amd64 || sparcv9))

// Download java with builds from adoptium
package javaprebuild

import (
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

// Install latest version from adoptium
func InstallLatest(featVersion uint, installPath string) error {
	var reqOpt request.Options
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

	// os: linux, windows, mac, solaris, aix, alpine-linux
	os := runtime.GOOS
	switch os {
	case "darwin":
		os = "mac"
	case "sunos":
		os = "solaris"
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

	reqOpt.NotFollowRedirect = true
	reqOpt.CodeProcess = map[int]request.RequestStatusFunction{301: processRedirect, 302: processRedirect, 307: processRedirect}
	Url := fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/%d/ga/%s/%s/jdk/hotspot/normal/eclipse", featVersion, os, arch)
	if _, err := request.Request(Url, &reqOpt); err != nil {
		return err
	}
	return nil
}
