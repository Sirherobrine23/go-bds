package bedrock

import (
	"archive/tar"
	"archive/zip"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/semver"
	"sirherobrine23.com.br/go-bds/go-bds/utils/js_types"
	"sirherobrine23.com.br/go-bds/request/v2"
)

// Version with targets servers
type Version struct {
	Version   string                      `json:"version"`          // Server version
	IsPreview bool                        `json:"preview"`          // Version is preview
	Docker    map[string]string           `json:"images,omitempty"` // Docker image
	Plaforms  map[string]*PlatformVersion `json:"platforms"`        // OS targets servers
}

// Return serve [sirherobrine23.com.br/go-bds/go-bds/semver.Version]
func (version *Version) SemverVersion() *semver.Version { return semver.New(version.Version) }

// Slice with versions
type Versions []*Version

// Check if versions exists in slice
func (versions Versions) HasVersion(ver string) bool {
	_, ok := versions.Get(ver)
	return ok == nil
}

// Return version if exists in slice
func (versions Versions) Get(ver string) (*Version, error) {
	for _, versionStruct := range versions {
		if versionStruct.Version == ver {
			return versionStruct, nil
		}
	}
	return nil, ErrNoVersion
}

// Get last version stable (oficial) release
func (versions Versions) LatestStable() *Version {
	releasesVersions := js_types.Slice[*Version](versions).Filter(func(v *Version) bool { return !v.IsPreview })
	semver.SortStruct(releasesVersions)
	return releasesVersions.At(-1)
}

// Get last preview version
func (versions Versions) LatestPreview() *Version {
	previewVersions := js_types.Slice[*Version](versions).Filter(func(v *Version) bool { return v.IsPreview })
	semver.SortStruct(previewVersions)
	return previewVersions.At(-1)
}

// Versions extracted from Minecraft website
type MojangHTML struct {
	Versions []struct {
		Version  string `json:"version"`                                                   // Server Version
		Preview  bool   `json:"isPreview"`                                                 // Server is preview
		Platform string `json:"platform" html:"div.card-footer > div > a = data-platform"` // golang target
		URL      string `json:"url" html:"div.card-footer > div > a = href"`               // File url
	} `json:"versions" html:"#main-content > div > div > div > div > div > div > div.server-card.aem-GridColumn.aem-GridColumn--default--12 > div > div > div > div"`
}

// Fetch versions from minecraft.net
func (versions *Versions) FetchFromMinecraftDotNet() error {
	pageVersions, _, err := request.GoHTML[MojangHTML]("https://minecraft.net/en-us/download/server/bedrock", &request.Options{Header: MojangHeaders})
	if err != nil {
		return err
	}

	// https://www.minecraft.net/bedrockdedicatedserver/bin-linux/bedrock-server-1.6.1.0.zip
	for _, value := range pageVersions.Versions {
		isPreview, platform, versionString := false, "", strings.TrimSuffix(strings.TrimPrefix(path.Base(value.URL), "bedrock-server-"), ".zip")

		// Set go platform
		switch value.Platform {
		case "serverBedrockLinux", "serverBedrockPreviewLinux":
			platform = "linux/amd64"
		case "serverBedrockWindows", "serverBedrockPreviewWindows":
			platform = "windows/amd64"
		default:
			return fmt.Errorf("cannot go target from %q", value.Platform)
		}

		// Check if is beta version
		switch value.Platform {
		case "serverBedrockPreviewWindows", "serverBedrockPreviewLinux":
			isPreview = true
		}

		version, exist := versions.Get(versionString)
		if exist != nil {
			*versions = append(*versions, &Version{
				Version:   versionString,
				IsPreview: isPreview,
				Docker:    map[string]string{},
				Plaforms:  map[string]*PlatformVersion{platform: {ReleaseDate: time.Unix(0, 0), ZipFile: value.URL}},
			})
			continue
		}
		version.Plaforms[platform] = &PlatformVersion{ReleaseDate: time.Unix(0, 0), ZipFile: value.URL}
	}
	semver.SortStruct(*versions)
	return nil
}

// File target to <os>/<arch>
type PlatformVersion struct {
	ReleaseDate time.Time `json:"releaseDate"` // Platform release/build day
	ZipFile     string    `json:"zipFile"`     // Minecraft server url server
	TarFile     string    `json:"tarFile"`     // Minecraft server url in tar type
	ZipSHA1     string    `json:"zipSHA1"`     // SHA1 to verify integrety to zip file
	TarSHA1     string    `json:"tarSHA1"`     // SHA1 to verify integrety to tar file
}

// Download server file and check file SHA1
func (target PlatformVersion) Download(w io.Writer) error {
	downloadUrl, fileSHA1 := target.ZipFile, target.ZipSHA1
	if target.TarFile != "" {
		downloadUrl = target.TarFile
		fileSHA1 = target.TarSHA1
	}

	// Request server file
	response, err := request.Request(downloadUrl, &request.Options{Method: "GET", Header: MojangHeaders})
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Dont check file SHA1
	if fileSHA1 == "" {
		_, err = io.Copy(w, response.Body)
		return err
	}

	sha1Sum := sha1.New()
	if _, err = io.Copy(io.MultiWriter(sha1Sum, w), response.Body); err != nil {
		return err
	} else if hex.EncodeToString(sha1Sum.Sum(nil)) != fileSHA1 {
		return errors.New("invalid file dowloaded")
	}
	return nil
}

// Extract server to folder path
func (target PlatformVersion) Extract(cwd string) error {
	switch {
	case target.TarFile != "":
		return request.Tar(target.TarFile, request.ExtractOptions{Cwd: cwd}, nil)
	case target.ZipFile != "":
		return request.Zip(target.ZipFile, request.ExtractOptions{Cwd: cwd}, &request.Options{Method: "GET", Header: MojangHeaders})
	default:
		return errors.New("cannot extract server target")
	}
}

// Download zip file and make SHA1 to Zip and Tar file
func (target *PlatformVersion) AppendSHA1() error {
	file, _, err := request.SaveTmp(target.ZipFile, "", &request.Options{Method: "GET", Header: MojangHeaders})
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	// SHA1
	zipSHA1, tarSHA1 := sha1.New(), sha1.New()
	if _, err = io.Copy(zipSHA1, file); err != nil {
		return err
	}
	target.ZipSHA1 = hex.EncodeToString(zipSHA1.Sum(nil))

	zipFile, err := zip.NewReader(file, stat.Size())
	if err != nil {
		return err
	}

	tarbal := tar.NewWriter(tarSHA1)
	defer tarbal.Close()
	for _, file := range zipFile.File {
		r, err := tar.FileInfoHeader(file.FileInfo(), "")
		if err != nil {
			return err
		}
		r.Name = path.Clean(strings.ReplaceAll(file.Name, "\\", "/"))
		if err = tarbal.WriteHeader(r); err != nil {
			return err
		}
		if file.FileInfo().IsDir() {
			continue
		}
		f, err := file.Open()
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err = io.CopyN(tarbal, f, file.FileInfo().Size()); err != nil {
			return err
		}
		f.Close()
	}
	tarbal.Close()

	target.TarSHA1 = hex.EncodeToString(tarSHA1.Sum(nil))
	return nil
}
