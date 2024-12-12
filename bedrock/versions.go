package bedrock

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/internal/regex"
	"sirherobrine23.com.br/go-bds/go-bds/internal/semver"
	"sirherobrine23.com.br/go-bds/go-bds/request/gohtml"
	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

var (
	VersionsRemote string = "https://sirherobrine23.com.br/go-bds/BedrockFetch/raw/branch/main/versions.json" // Remote cached versions
	VersionMatch          = regex.MustCompile(`(?m)(\-|_)(?P<Version>[0-9\.]+)\.zip$`)

	ErrInvalidFileVersions error = errors.New("invalid versions file or url")            // Versions file invalid url schema
	ErrNoVersion           error = errors.New("cannot find version")                     // Version request not exists
	ErrNoPlatform          error = errors.New("platform not supported for this version") // Version request not exists

	MojangHeaders = map[string]string{
		// "Accept-Encoding":           "gzip, deflate",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"Accept-Language":           "en-US;q=0.9,en;q=0.8",
		"Sec-Ch-Ua":                 `"Google Chrome";v="123", "Not:A-Brand";v="8", "Chromium";v="123\"`,
		"Sec-Ch-Ua-Mobile":          "?0",
		"Sec-Ch-Ua-Platform":        `"Windows"`,
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	}
)

// Versions extracted from Minecraft website
type MojangHTML struct {
	Versions []struct {
		Version  string `json:"version"`                                                   // Server Version
		Preview  bool   `json:"isPreview"`                                                 // Server is preview
		Platform string `json:"platform" html:"div.card-footer > div > a = data-platform"` // golang target
		URL      string `json:"url" html:"div.card-footer > div > a = href"`               // File url
	} `json:"versions" html:"#main-content > div > div > div > div > div > div > div.server-card.aem-GridColumn.aem-GridColumn--default--12 > div > div > div > div"`
}

type VersionPlatform struct {
	ZipFile     string    `json:"zipFile"`     // Minecraft server url server
	ZipSHA1     string    `json:"zipSHA1"`     // SHA1 to verify integrety to zip file
	TarFile     string    `json:"tarFile"`     // Minecraft server url in tar type
	TarSHA1     string    `json:"tarSHA1"`     // SHA1 to verify integrety to tar file
	ReleaseDate time.Time `json:"releaseDate"` // Platform release/build day
}

type OldVersions map[string]Version
type Versions []Version
type Version struct {
	ServerVersion string                     `json:"version"`
	IsPreview     bool                       `json:"preview"`          // Preview server
	DockerImage   map[string]string          `json:"images,omitempty"` // Docker images
	Platforms     map[string]VersionPlatform `json:"platforms"`        // Golang platforms target
}

func (version Version) SemverVersion() *semver.Version { return semver.New(version.ServerVersion) }

func (old OldVersions) Migrate() (new Versions) {
	new = make(Versions, 0)
	for versionStr, versionStruct := range old {
		versionStruct.ServerVersion = versionStr // Add version to struct
		new = append(new, versionStruct)
	}
	return
}

// Check if version exists in slice
func (versions Versions) Has(ver string) (exit bool) {
	for _, versionStruct := range versions {
		if exit = (versionStruct.ServerVersion == ver); exit {
			break
		}
	}
	return
}

// Return version if exists in slice
func (versions Versions) Get(ver string) (version Version, ok bool) {
	for _, versionStruct := range versions {
		if ok = (versionStruct.ServerVersion == ver); ok {
			version = versionStruct
			break
		}
	}
	return
}

// Get versions from cached versions
func FromVersions() (Versions, error) {
	versions, _, err := request.JSON[Versions](VersionsRemote, nil)
	semver.SortStruct(versions)
	return versions, err
}

// Get latest stable release version
func (versions Versions) GetLatest() string {
	releasesVersions := slices.DeleteFunc(versions, func(v Version) bool { return !v.IsPreview })
	semver.SortStruct(releasesVersions)
	return releasesVersions[len(releasesVersions)-1].ServerVersion
}

// Get latest preview release version
func (versions Versions) GetLatestPreview() string {
	previewVersions := slices.DeleteFunc(versions, func(v Version) bool { return v.IsPreview })
	semver.SortStruct(previewVersions)
	return previewVersions[len(previewVersions)-1].ServerVersion
}

// Download and Extract server to folder
func (version VersionPlatform) Download(serverPath string) error {
	if version.TarFile == "" && version.ZipFile == "" {
		return fmt.Errorf("invalid system target to download, cannot download server")
	}

	extractOptions := request.ExtractOptions{Cwd: serverPath}
	if version.TarFile != "" { // Not check file signature
		return request.Tar(version.TarFile, extractOptions, nil)
	}
	return request.Zip(version.ZipFile, extractOptions, nil)
}

// Get new versions from minecraft.net/en-us/download/server/bedrock
func FetchFromWebsite() (*MojangHTML, error) {
	res, err := request.Request("https://minecraft.net/en-us/download/server/bedrock", &request.Options{Header: MojangHeaders})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var body MojangHTML
	if err := gohtml.NewParse(res.Body, &body); err != nil {
		return nil, err
	}

	for index, value := range body.Versions {
		body.Versions[index].Version = VersionMatch.FindAllGroups(value.URL)["Version"]

		// Set go platform
		switch value.Platform {
		case "serverBedrockLinux", "serverBedrockPreviewLinux":
			body.Versions[index].Platform = "linux/amd64"
		case "serverBedrockWindows", "serverBedrockPreviewWindows":
			body.Versions[index].Platform = "windows/amd64"
		default:
			return nil, fmt.Errorf("cannot go target from %q", value.Platform)
		}

		// Check if is beta version
		switch value.Platform {
		case "serverBedrockPreviewWindows", "serverBedrockPreviewLinux":
			body.Versions[index].Preview = true
		}
	}
	return &body, nil
}

// Convert to versions and fill ReleaseDate, ZipFile and ZipSHA1
func (mojangWeb MojangHTML) ConvertToVersions() (Versions, error) {
	versions := make(Versions, 0)
	for _, WebVersion := range mojangWeb.Versions {
		if !versions.Has(WebVersion.Version) {
			versions = append(versions, Version{
				ServerVersion: WebVersion.Version,
				IsPreview:     WebVersion.Preview,
				DockerImage:   make(map[string]string),
				Platforms:     make(map[string]VersionPlatform),
			})
		}

		for appendedIndex := range versions {
			if versions[appendedIndex].ServerVersion != WebVersion.Version {
				continue
			}

			// Save file localy
			localFile, _, err := request.SaveTmp(WebVersion.URL, nil)
			if err != nil {
				if localFile != nil {
					os.Remove(localFile.Name())
				}
				return versions, err
			}

			stat, _ := localFile.Stat()
			zipFile, err := zip.NewReader(localFile, stat.Size())
			if err != nil {
				os.Remove(localFile.Name())
				return versions, err
			}

			// Create new struct
			plaftormRelease := VersionPlatform{ZipFile: WebVersion.URL}

			// Find file server to get build date
			for _, file := range zipFile.File {
				if strings.HasPrefix(file.FileInfo().Name(), "bedrock_server") {
					plaftormRelease.ReleaseDate = file.Modified
					break
				}
			}

			// Create sha1 from zip file
			if _, err = localFile.Seek(0, 0); err != nil {
				os.Remove(localFile.Name())
				return versions, err
			}
			sha1 := sha1.New()
			go io.Copy(sha1, localFile)
			plaftormRelease.ZipSHA1 = hex.EncodeToString(sha1.Sum(nil))

			// Delete zip file
			if err = os.Remove(localFile.Name()); err != nil {
				return versions, err
			}

			// Append to versions again
			ver := versions[appendedIndex]
			ver.Platforms[WebVersion.Platform] = plaftormRelease
			versions[appendedIndex] = ver
		}
	}
	semver.SortStruct(versions)
	return versions, nil
}
