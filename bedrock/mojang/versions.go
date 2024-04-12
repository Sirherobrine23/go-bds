package mojang

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	ErrInvalidFileVersions error = errors.New("invalid versions file or url") // Versions file invalid url schema
	ErrNoVersion           error = errors.New("cannot find version")          // Version request not exists
)

type VersionMinecraft struct {
	FileUrl, Target string
	Preview         bool
}

type VersionTarget struct {
	NodePlatform string `json:"Platform"` // Nodejs Platform string
	NodeArch     string `json:"Arch"`     // Nodejs Archs
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
	ZipFile      string `json:"zip"`     // Local file and http to remote
	ZipSHA1      string `json:"zipSHA1"` // File SHA1
	TarFile      string `json:"tar"`     // Local file and http to remote
	TarSHA1      string `json:"tarSHA1"` // File SHA1
}

type Version struct {
	Version     string          `json:"version"`
	DateRelease time.Time       `json:"releaseDate"`
	ReleaseType string          `json:"type"`
	Targets     []VersionTarget `json:"targets"`
}

// Get versions from minecraft server page
func Minecraft() ([]VersionMinecraft, error) {
	versionsTarget := []VersionMinecraft{}
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
		}

		if len(goTarget) > 5 {
			versionsTarget = append(versionsTarget, VersionMinecraft{
				Preview: isPreview,
				FileUrl: file.String(),
				Target:  goTarget,
			})
		}
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

func (w *VersionTarget) Download(serverPath string) error {
	if err := os.MkdirAll(serverPath, os.FileMode(0o666)); !(err == nil || os.IsExist(err)) {
		return err
	}

	res, err := request.Request(request.RequestOptions{HttpError: true, Method: "GET", Url: w.TarFile})
	if err != nil {
		return err
	}

	defer res.Body.Close()
	zip, err := gzip.NewReader(res.Body)
	if err != nil {
		return err
	}

	defer zip.Close()
	tarStream := tar.NewReader(zip)

	for {
		head, err := tarStream.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if head.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(filepath.Join(serverPath, head.Name), head.FileInfo().Mode()); !(err == nil || os.IsExist(err)) {
				return err
			}
		} else {
			var file *os.File
			if file, err = os.OpenFile(filepath.Join(serverPath, head.Name), os.O_CREATE, head.FileInfo().Mode()); err != nil {
				return err
			}

			defer file.Close()
			if _, err = io.CopyN(file, tarStream, head.Size); err != nil {
				return err
			}
		}
	}

	return nil
}
