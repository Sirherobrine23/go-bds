package mojang

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

const (
	VersionsRemote    string = "https://sirherobrine23.org/Minecraft-Server/BedrockFetch/raw/branch/main/versions.json" // Remote cached versions
	MinecraftPage     string = "https://www.minecraft.net/en-us/download/server/bedrock"                                // Minecraft page to find versions
	WindowsUrl        string = "https://minecraft.azureedge.net/bin-win/bedrock-server-%s.zip"                          // Windows x64/amd64 url file
	LinuxUrl          string = "https://minecraft.azureedge.net/bin-linux/bedrock-server-%s.zip"                        // Linux x64/amd64 url file
	WindowsPreviewUrl string = "https://minecraft.azureedge.net/bin-win-preview/bedrock-server-%s.zip"                  // Windows x64/amd64 preview url file
	LinuxPreviewUrl   string = "https://minecraft.azureedge.net/bin-linux-preview/bedrock-server-%s.zip"                // Linux x64/amd64 url file
)

var (
	ErrInvalidFileVersions error          = errors.New("invalid versions file or url") // Versions file invalid url schema
	ErrNoVersion           error          = errors.New("cannot find version")          // Version request not exists
	MatchVersion           *regexp.Regexp = regexp.MustCompile(`bedrock-server-(P<Version>[0-9\.\-_]+).zip$`)
)

type VersionTarget struct {
	Target  string `json:"goTarget"`
	ZipFile string `json:"zip"`     // Local file and http to remote
	ZipSHA1 string `json:"zipSHA1"` // File SHA1
}

type Version struct {
	Version     string          `json:"version"`
	DateRelease time.Time       `json:"releaseDate"`
	IsPreview   bool            `json:"isPreview"`
	Targets     []VersionTarget `json:"targets"`
}

// Get versions from minecraft server page
func Minecraft() (map[string]Version, error) {
	versionsTarget := map[string]Version{}
	doc, _, err := request.HtmlNode(request.RequestOptions{HttpError: true, Method: "GET", Url: MinecraftPage})
	if err != nil {
		return versionsTarget, err
	}

	zipFiles := []url.URL{}
	targets := doc.Find("#main-content > div > div > div > div > div > div > div.server-card.aem-GridColumn.aem-GridColumn--default--12 > div > div > div > div")
	targets.Each(func(i int, s *goquery.Selection) {
		// div.card-footer > div > a
		href := s.Find("div.card-footer > div > a").AttrOr("href", "")
		if len(href) > 5 {
			if parsedFile, err := url.Parse(href); err == nil {
				zipFiles = append(zipFiles, *parsedFile)
			}
		}
	})

	for _, file := range zipFiles {
		version := internal.FindAllGroups(MatchVersion, file.Path)["Version"]
		isPreview := false
		var goTarget string
		if strings.HasPrefix(file.Path, "/bin-win-preview") {
			isPreview = true
			goTarget = "windows/amd64"
		} else if strings.HasPrefix(file.Path, "/bin-linux-preview") {
			isPreview = true
			goTarget = "linux/amd64"
		} else if strings.HasPrefix(file.Path, "/bin-win") {
			goTarget = "windows/amd64"
		} else if strings.HasPrefix(file.Path, "/bin-linux") {
			goTarget = "linux/amd64"
		} else {
			continue
		}

		req := request.RequestOptions{Url: file.String()}

		var ok bool
		var versionRelease Version
		if versionRelease, ok = versionsTarget[version]; !ok {
			versionRelease = Version{
				Version:     version,
				IsPreview:   isPreview,
				DateRelease: time.Time{},
				Targets:     []VersionTarget{},
			}

			file, err := os.CreateTemp(os.TempDir(), "bdsserver")
			if err != nil {
				return nil, err
			}

			defer file.Close()
			if err = req.WriteStream(file); err != nil {
				return nil, err
			}

			stats, err := file.Stat()
			if err != nil {
				return nil, err
			}

			zipReader, err := zip.NewReader(file, stats.Size())
			if err != nil {
				return nil, err
			}

			for _, file := range zipReader.File {
				if strings.HasPrefix(file.Name, "bedrock_server") {
					versionRelease.DateRelease = file.Modified
					break
				}
			}

			file.Close()
			os.Remove(file.Name())
		}

		sha256, err := req.SHA256()
		if err != nil {
			return nil, err
		}

		versionRelease.Targets = append(versionsTarget[version].Targets, VersionTarget{
			Target:  goTarget,
			ZipFile: file.String(),
			ZipSHA1: sha256,
		})
		versionsTarget[version] = versionRelease
	}

	return versionsTarget, nil
}

// Get versions from cached versions
// remoteFileFetch set custom cache versions for load versions
func FromVersions(remoteFileFetch ...string) ([]Version, error) {
	fileFatch := VersionsRemote
	if len(remoteFileFetch) == 1 && len(remoteFileFetch[0]) > 2 {
		fileFatch = remoteFileFetch[0]
	}

	file, err := url.Parse(fileFatch)
	if err != nil {
		return []Version{}, err
	}

	versions := []Version{}
	if file.Scheme == "http" || file.Scheme == "https" {
		res, err := request.Request(request.RequestOptions{Method: "GET", HttpError: true, Url: fileFatch})
		if err != nil {
			return versions, err
		}

		defer res.Body.Close()
		if err = json.NewDecoder(res.Body).Decode(&versions); err != nil {
			return versions, err
		}
	} else if file.Scheme == "file" {
		osFile, err := os.Open(file.Path)
		if err != nil {
			return versions, err
		}

		defer osFile.Close()
		if err = json.NewDecoder(osFile).Decode(&versions); err != nil {
			return versions, err
		}
	} else {
		return versions, ErrInvalidFileVersions
	}

	return versions, nil
}

func (version *VersionTarget) Download(serverPath string) error {
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