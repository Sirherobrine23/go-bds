package mojang

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

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

func (version *VersionPlatform) Download(serverPath string) error {
	req := request.RequestOptions{Url: version.ZipFile}

	file, err := os.CreateTemp(os.TempDir(), "bdsserver")
	if err != nil {
		return err
	}
	defer file.Close()
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
