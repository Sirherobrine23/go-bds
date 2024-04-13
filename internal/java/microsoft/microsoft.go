package microsoft

import (
	"encoding/json"
	"fmt"

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

func Releases() ([]globals.Version, error) {
	var versions []msVersion
	res, err := request.Request(request.RequestOptions{HttpError: true, Url: GithubVersions})
	if err != nil {
		return []globals.Version{}, err
	}
	defer res.Body.Close()
	if err = json.NewDecoder(res.Body).Decode(&versions); err != nil {
		return []globals.Version{}, err
	}

	mapVersion := []globals.Version{}
	for _, release := range versions {
		relVersion := globals.Version{
			Version: release.Version,
			Targets: map[string]string{},
		}

		for _, file := range release.Files {
			goarch := file.NodeArch
			goos := file.NodeOs

			switch file.NodeOs {
			case "sunos":
				goos = "solaris"
			case "win32":
				goos = "windows"
			}
			switch file.NodeArch {
			case "x64":
				goarch = "amd64"
			case "ia32":
				goarch = "386"
			}

			relVersion.Targets[fmt.Sprintf("%s/%s", goos, goarch)] = file.FileUrl
		}

		mapVersion = append(mapVersion, relVersion)
	}

	return mapVersion, nil
}
