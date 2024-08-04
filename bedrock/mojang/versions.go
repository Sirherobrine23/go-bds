package mojang

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"sirherobrine23.org/go-bds/go-bds/internal"
	"sirherobrine23.org/go-bds/go-bds/internal/gohtml"
	"sirherobrine23.org/go-bds/go-bds/internal/semver"
	"sirherobrine23.org/go-bds/go-bds/request"
)

var (
	VersionsRemote string = "https://sirherobrine23.org/go-bds/BedrockFetch/raw/branch/main/versions.json" // Remote cached versions
)

const (
	MinecraftPage     string = "https://www.minecraft.net/en-us/download/server/bedrock"                 // Minecraft page to find versions
	WindowsUrl        string = "https://minecraft.azureedge.net/bin-win/bedrock-server-%s.zip"           // Windows x64/amd64 url file
	LinuxUrl          string = "https://minecraft.azureedge.net/bin-linux/bedrock-server-%s.zip"         // Linux x64/amd64 url file
	WindowsPreviewUrl string = "https://minecraft.azureedge.net/bin-win-preview/bedrock-server-%s.zip"   // Windows x64/amd64 preview url file
	LinuxPreviewUrl   string = "https://minecraft.azureedge.net/bin-linux-preview/bedrock-server-%s.zip" // Linux x64/amd64 url file
)

var (
	ErrInvalidFileVersions error          = errors.New("invalid versions file or url")            // Versions file invalid url schema
	ErrNoVersion           error          = errors.New("cannot find version")                     // Version request not exists
	ErrNoPlatform          error          = errors.New("platform not supported for this version") // Version request not exists
	MatchVersion           *regexp.Regexp = regexp.MustCompile(`bedrock-server-(P<Version>[0-9\.\-_]+).zip$`)
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

type Version struct {
	IsPreview   bool                       `json:"preview"`          // Preview server
	DockerImage map[string]string          `json:"images,omitempty"` // Docker images
	Platforms   map[string]VersionPlatform `json:"platforms"`        // Golang platforms target
}
type Versions map[string]Version

// Get versions from cached versions
func FromVersions() (Versions, error) {
	var versions Versions
	res, err := request.Request(request.RequestOptions{Method: "GET", HttpError: true, Url: VersionsRemote})
	if err != nil {
		return versions, err
	}

	defer res.Body.Close()
	if err = json.NewDecoder(res.Body).Decode(&versions); err != nil {
		return versions, err
	}

	return versions, nil
}

// Extract server to folder
func (version *VersionPlatform) Download(serverPath string) error {
	var req request.RequestOptions
	if version.TarSHA1 != "" && version.TarFile != "" {
		req.Url = version.TarFile
		res, err := req.Request()
		if err != nil {
			return err
		}
		defer res.Body.Close()
		gz, err := gzip.NewReader(res.Body)
		if err != nil {
			return err
		}
		defer gz.Close()
		tarball := tar.NewReader(gz)

		for {
			head, err := tarball.Next()
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					break
				}
				return err
			}
			fileinfo := head.FileInfo()
			fullPath := filepath.Join(serverPath, head.Name)
			if fileinfo.IsDir() {
				if err := os.MkdirAll(fullPath, fileinfo.Mode()); err != nil {
					return err
				} else if err := os.Chtimes(fullPath, head.AccessTime, head.ModTime); err != nil {
					return err
				}
				continue
			}

			// Create folder if not exist to create file
			os.MkdirAll(filepath.Dir(fullPath), 0666)
			file, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileinfo.Mode())
			if err != nil {
				return err
			} else if err := os.Chtimes(fullPath, head.AccessTime, head.ModTime); err != nil {
				return err
			}

			// Copy file
			if _, err := io.CopyN(file, tarball, fileinfo.Size()); err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					continue
				}
				return err
			}
		}

		return nil
	}

	file, err := os.CreateTemp(os.TempDir(), "bdsserver")
	if err != nil {
		return err
	}
	defer file.Close()
	defer os.Remove(file.Name())

	req.Url = version.ZipFile
	if err = req.WriteStream(file); err != nil {
		return err
	}

	stats, err := file.Stat()
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(file, stats.Size())
	if err != nil {
		return err
	}

	for _, file := range zipReader.File {
		fileInfo := file.FileInfo()
		filePath := filepath.Join(serverPath, file.Name)
		if fileInfo.IsDir() {
			err = os.MkdirAll(filePath, file.Mode())
			if err != nil {
				return err
			}
			continue
		}

		fileOs, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer fileOs.Close()

		zipFile, err := file.Open()
		if err != nil {
			return err
		}
		defer zipFile.Close()

		_, err = io.Copy(fileOs, zipFile)
		if err != nil {
			return err
		}
	}
	return nil
}

// Get new versions from minecraft.net/en-us/download/server/bedrock
func FetchFromWebsite() (*MojangHTML, error) {
	var req = request.RequestOptions{
		Url:       "https://minecraft.net/en-us/download/server/bedrock",
		HttpError: true,
	}
	res, err := req.Request()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var body MojangHTML
	if err := gohtml.NewParse(res.Body, &body); err != nil {
		return nil, err
	}

	var fileMatch = regexp.MustCompile(`(?m)(\-|_)(?P<Version>[0-9\.]+)\.zip$`)
	for index, value := range body.Versions {
		body.Versions[index].Version = internal.FindAllGroups(fileMatch, value.URL)["Version"]
		switch value.Platform {
		case "serverBedrockLinux":
			body.Versions[index].Platform = "linux/amd64"
		case "serverBedrockPreviewLinux":
			body.Versions[index].Platform = "linux/amd64"
			body.Versions[index].Preview = true
		case "serverBedrockWindows":
			body.Versions[index].Platform = "windows/amd64"
		case "serverBedrockPreviewWindows":
			body.Versions[index].Platform = "windows/amd64"
			body.Versions[index].Preview = true
		}
	}
	return &body, nil
}

func GetLatest(a Versions) string {
	var k []*semver.Version
	for key, v := range a {
		if v.IsPreview {
			continue
		}
		k = append(k, semver.New(key))
	}
	semver.Sort(k)
	return k[len(k)-1].String()
}

func GetLatestPreview(a Versions) string {
	var k []*semver.Version
	for key, v := range a {
		if !v.IsPreview {
			continue
		}
		k = append(k, semver.New(key))
	}
	semver.Sort(k)
	return k[len(k)-1].String()
}
