package adoptium

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/cache"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/java/globals"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

type apiVersions []struct {
	Binaries []struct {
		Architecture  string `json:"architecture"`
		DownloadCount int    `json:"download_count"`
		HeapSize      string `json:"heap_size"`
		ImageType     string `json:"image_type"`
		JvmImpl       string `json:"jvm_impl"`
		Os            string `json:"os"`
		Package       struct {
			Checksum      string `json:"checksum"`
			ChecksumLink  string `json:"checksum_link"`
			DownloadCount int    `json:"download_count"`
			Link          string `json:"link"`
			MetadataLink  string `json:"metadata_link"`
			Name          string `json:"name"`
			SignatureLink string `json:"signature_link"`
			Size          int    `json:"size"`
		} `json:"package"`
		Project   string    `json:"project"`
		ScmRef    string    `json:"scm_ref"`
		UpdatedAt time.Time `json:"updated_at"`
		Installer struct {
			Checksum      string `json:"checksum"`
			ChecksumLink  string `json:"checksum_link"`
			DownloadCount int    `json:"download_count"`
			Link          string `json:"link"`
			MetadataLink  string `json:"metadata_link"`
			Name          string `json:"name"`
			SignatureLink string `json:"signature_link"`
			Size          int    `json:"size"`
		} `json:"installer,omitempty"`
	} `json:"binaries"`
	DownloadCount int    `json:"download_count"`
	ID            string `json:"id"`
	ReleaseLink   string `json:"release_link"`
	ReleaseName   string `json:"release_name"`
	ReleaseType   string `json:"release_type"`
	Source        struct {
		Link string `json:"link"`
		Name string `json:"name"`
		Size int    `json:"size"`
	} `json:"source"`
	Timestamp   time.Time `json:"timestamp"`
	UpdatedAt   time.Time `json:"updated_at"`
	Vendor      string    `json:"vendor"`
	VersionData struct {
		Build          int    `json:"build"`
		Major          int    `json:"major"`
		Minor          int    `json:"minor"`
		OpenjdkVersion string `json:"openjdk_version"`
		Optional       string `json:"optional"`
		Pre            string `json:"pre"`
		Security       int    `json:"security"`
		Semver         string `json:"semver"`
	} `json:"version_data"`
}

func reverseReleases(arr apiVersions) {
	left := 0
	right := len(arr) - 1
	for left < right {
		arr[left], arr[right] = arr[right], arr[left]
		left++
		right--
	}
}

func Releases() (globals.Version, error) {
	if cache.Get("adoptium", "releases") != nil {
		cached, ok := cache.Get("adoptium", "releases").(globals.Version)
		if ok {
			return cached, nil
		}
	}

	// %5B1.0%2C100.0%5D == [1.0,100.0]
	req := request.RequestOptions{
		HttpError:   true,
		Url:         `https://api.adoptium.net/v3/assets/version/%5B1.0%2C100.0%5D`,
		CodesRetrys: []int{400},
		Querys: map[string]string{
			"project":     "jdk",
			"image_type":  "jre",
			"page_size":   "20",
			"sort_method": "DEFAULT",
			"sort_order":  "DESC",
			"semver":      "true",
		},
	}

	// architecture: x64, x86, x32, ppc64, ppc64le, s390x, aarch64, arm, sparcv9, riscv64
	goarch := runtime.GOARCH
	switch goarch {
	case "amd64":
		goarch = "x64"
	case "386":
		goarch = "x86"
	case "arm64":
		goarch = "aarch64"
	}
	req.Querys["architecture"] = goarch

	// os: linux, windows, mac, solaris, aix, alpine-linux
	goos := runtime.GOOS
	switch goos {
	case "darwin":
		goos = "mac"
	case "sunos":
		goos = "solaris"
	}
	req.Querys["os"] = goos

	dones, isErr, wait := 0, false, make(chan error)
	var concatedVersions apiVersions
	var requestMake func(pageUrl string)
	requestMake = func(pageUrl string) {
		if isErr {
			return
		}

		dones++
		var err error
		defer func() {
			dones--
			if err != nil {
				isErr = true
			}
			wait <- err
		}()

		if len(pageUrl) > 0 {
			req.Url = pageUrl
		}

		var res http.Response
		res, err = req.Request()
		if err != nil {
			return
		}

		defer res.Body.Close()
		if len(res.Header["Link"]) > 0 {
			links := request.ParseMultipleLinks(res.Header["Link"]...)
			for _, k := range links {
				if _, ok := k.HasKeyValue("rel", "next", "Next"); ok {
					go requestMake(k.URL)
					break
				} else if _, ok := k.HasKeyValue("Rel", "next", "Next"); ok {
					go requestMake(k.URL)
					break
				}
			}
		}

		var releases apiVersions
		if err = json.NewDecoder(res.Body).Decode(&releases); err != nil {
			return
		}

		concatedVersions = append(concatedVersions, releases...)
	}
	go requestMake(req.Url)

	for {
		err := <-wait
		if err != nil {
			isErr = true
			return nil, err
		}
		if dones == 0 {
			close(wait)
			break
		}
	}

	versions := globals.Version{}
	reverseReleases(concatedVersions)
	for _, release := range concatedVersions {
		if len(release.Binaries) == 0 {
			continue
		}
		versions[release.VersionData.Major] = globals.VersionBundle{
			FileUrl:  release.Binaries[0].Package.Link,
			Checksum: fmt.Sprintf("sha256:%s", release.Binaries[0].Package.Checksum),
		}
	}

	return versions, nil
}
