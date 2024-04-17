package microsoft

import (
	"runtime"
	"strconv"
	"strings"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/cache"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/java/globals"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

const (
	GithubVersions string = "https://github.com/actions/setup-java/raw/main/src/distributions/microsoft/microsoft-openjdk-versions.json"
)

type msVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []struct {
		File     string `json:"filename"`
		FileUrl  string `json:"download_url"`
		NodeArch string `json:"arch"`     // "arm" | "arm64" | "ia32" | "mips" | "mipsel" | "ppc" | "ppc64" | "riscv64" | "s390" | "s390x" | "x64"
		NodeOs   string `json:"platform"` // "aix" | "android" | "darwin" | "freebsd" | "haiku" | "linux" | "openbsd" | "sunos" | "win32" | "cygwin" | "netbsd"
	} `json:"files"`
}

func reverseReleases(arr []msVersion) {
	left := 0
	right := len(arr) - 1
	for left < right {
		arr[left], arr[right] = arr[right], arr[left]
		left++
		right--
	}
}

func Releases() (globals.Version, error) {
	if cache.Get("microsoft", "releases") != nil {
		value, ok := cache.Get("microsoft", "releases").(globals.Version)
		if ok {
			return value, nil
		}
	}

	var msVersions []msVersion
	req := request.RequestOptions{HttpError: true, Url: GithubVersions}
	_, err := req.Do(&msVersions)
	if err != nil {
		return nil, err
	}

	reverseReleases(msVersions)
	versions := globals.Version{}
	for _, release := range msVersions {
		for _, file := range release.Files {
			var ARCHTOGO, OSTOGO string
			switch file.NodeArch {
			case "x64":
				ARCHTOGO = "amd64"
			case "aarch64":
				ARCHTOGO = "arm64"
			default:
				continue
			}

			switch file.NodeOs {
			case "darwin":
				OSTOGO = "darwin"
			case "linux":
				OSTOGO = "linux"
			case "win32":
				OSTOGO = "windows"
			default:
				continue
			}

			if ARCHTOGO != runtime.GOARCH {
				continue
			} else if OSTOGO != runtime.GOOS {
				continue
			}

			versionFamily, err := strconv.Atoi(strings.Split(release.Version, ".")[0])
			if err != nil {
				return nil, err
			}

			versions[versionFamily] = globals.VersionBundle{
				FileUrl:  file.FileUrl,
				Checksum: "",
			}
		}
	}

	if len(versions) == 0 {
		return nil, globals.ErrNoTargets
	}

	cache.Set("microsoft", "releases", versions, globals.DefaultTime)
	return versions, nil
}
