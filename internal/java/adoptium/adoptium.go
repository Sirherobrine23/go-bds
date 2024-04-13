package adoptium

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"golang.org/x/mod/semver"
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

func Releases() ([]globals.Version, error) {
	versionsMap := map[string]globals.Version{}

	// %5B1.0%2C100.0%5D == [1.0,100.0]
	p := request.RequestOptions{
		Url: `https://api.adoptium.net/v3/assets/version/%5B1.0%2C100.0%5D`,
		Querys: map[string]string{
			"project":     "jdk",
			"image_type":  "jdk",
			"semver":      "true",
			"page_size":   "20",
			"page":        "0",
			"heap_size":   "normal",
			"sort_method": "DEFAULT",
			"sort_order":  "DESC",
		},
	}
	pageApi, _ := p.String()

	for {
		if len(pageApi) == 0 {
			break
		}

		res, err := request.Request(request.RequestOptions{HttpError: true, Url: pageApi})
		if err != nil {
			return []globals.Version{}, err
		}

		pageApi = ""
		var releases apiVersions
		defer res.Body.Close()
		if err = json.NewDecoder(res.Body).Decode(&releases); err != nil {
			return []globals.Version{}, err
		}

		if len(res.Header["Link"]) > 0 {
			links := request.ParseMultipleLinks(res.Header["Link"]...)
			for _, k := range links {
				if _, ok := k.HasKeyValue("rel", "next", "Next"); ok {
					pageApi = k.URL
					break
				} else if _, ok := k.HasKeyValue("Rel", "next", "Next"); ok {
					pageApi = k.URL
					break
				}
			}
		}

		for _, rel := range releases {
			versionStruct := globals.Version{
				Version: strings.ReplaceAll(rel.VersionData.Semver, "-beta+", "+"),
				Targets: map[string]string{},
			}

			for _, file := range rel.Binaries {
				if file.Os == "alpine-linux" {
					continue
				}

				goarch := file.Architecture
				goos := file.Os
				switch goos {
				case "sunos":
					goos = "solaris"
				case "win32":
					goos = "windows"
				case "macos":
					goos = "darwin"
				}
				switch goarch {
				case "x64":
					goarch = "amd64"
				case "ia32":
					goarch = "386"
				}

				versionStruct.Targets[fmt.Sprintf("%s/%s", goos, goarch)] = file.Package.Link
			}

			if _, ok := versionsMap[versionStruct.Version]; !ok {
				versionsMap[versionStruct.Version] = versionStruct
			} else {
				for key, value := range versionStruct.Targets {
					versionsMap[versionStruct.Version].Targets[key] = value
				}
			}
		}
	}

	versionsArr := []globals.Version{}
	for _, v := range versionsMap {
		versionsArr = append(versionsArr, v)
	}
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
