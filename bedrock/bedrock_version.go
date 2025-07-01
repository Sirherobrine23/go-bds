package bedrock

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/binfmt"
	"sirherobrine23.com.br/go-bds/go-bds/utils/js_types"
	"sirherobrine23.com.br/go-bds/go-bds/utils/semver"
	"sirherobrine23.com.br/go-bds/request/v2"
)

// Version with targets servers
type Version struct {
	Version   string                      `json:"version"`          // Server version
	IsPreview bool                        `json:"preview"`          // Version is preview
	Docker    map[string]string           `json:"images,omitempty"` // Docker image
	Plaforms  map[string]*PlatformVersion `json:"platforms"`        // OS targets servers, <os>/<arch>(/<variant>) docker style
}

// Return serve [sirherobrine23.com.br/go-bds/go-bds/semver.Version]
func (version *Version) SemverVersion() semver.Version { return semver.New(version.Version) }

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
	semver.Sort(releasesVersions)
	return releasesVersions.At(-1)
}

// Get last preview version
func (versions Versions) LatestPreview() *Version {
	previewVersions := js_types.Slice[*Version](versions).Filter(func(v *Version) bool { return v.IsPreview })
	semver.Sort(previewVersions)
	return previewVersions.At(-1)
}

type MojangApiVersion struct {
	Type    string `json:"downloadType"`
	URL     string `json:"downloadUrl"`
	Version string
}

func (versions *Versions) versionProcess(newVersions <-chan MojangApiVersion, errChan chan<- error, wait *sync.WaitGroup) {
	defer wait.Done() // Workder done after newVersions close
	for value := range newVersions {
		zipFile, _, err := request.SaveTmp(value.URL, "", &request.Options{Method: "GET", Header: MojangHeaders})
		if err != nil {
			errChan <- err
			continue
		}

		// Make SHA1 to zip File
		sha1 := sha1.New()
		io.Copy(sha1, zipFile)

		// Start Platform struct with current UTC time
		platformVersion := &PlatformVersion{
			ZipFile:     value.URL,
			ZipSHA1:     hex.EncodeToString(sha1.Sum(nil)),
			ReleaseDate: time.Now().UTC(),
		}

		// Process Zip File
		stat, _ := zipFile.Stat()
		zipFiles, err := zip.NewReader(zipFile, stat.Size())
		if err != nil {
			zipFile.Close()
			os.Remove(zipFile.Name())
			errChan <- err
			continue
		}

		// Get platform string <os>/<arch>(/<variant>) docker style
		var platform string
		if fileIndex := slices.IndexFunc(zipFiles.File, func(file *zip.File) bool { return strings.HasPrefix(file.Name, "bedrock_server") }); fileIndex >= 0 {
			file := zipFiles.File[fileIndex]
			platformVersion.ReleaseDate = file.Modified

			// Openfile
			serveRead, err := file.Open()
			if err != nil {
				zipFile.Close()
				os.Remove(zipFile.Name())
				errChan <- err
				continue
			}

			serverFile, err := io.ReadAll(serveRead)
			serveRead.Close()
			if err != nil {
				zipFile.Close()
				os.Remove(zipFile.Name())
				errChan <- err
				continue
			}

			// Get platform
			bin, err := binfmt.GetBinary(bytes.NewReader(serverFile))
			if err != nil {
				zipFile.Close()
				os.Remove(zipFile.Name())
				errChan <- err
				continue
			}
			platform = binfmt.String(bin)
		}

		zipFile.Close()
		os.Remove(zipFile.Name())
		if version, _ := versions.Get(value.Version); version.Plaforms[platform] == nil {
			version.Plaforms[platform] = platformVersion
		}
	}
}

// Fetch versions from minecraft.net and append to [*Versions] if not exists.
// This make SHA1 to ZIP file and get server release Date/time
func (versions *Versions) FetchFromMinecraftDotNet() error {
	linkers, _, err := request.JSON[struct {
		Result struct {
			Links []MojangApiVersion `json:"links"`
		} `json:"result"`
	}]("https://net-secondary.web.minecraft-services.net/api/v1.0/download/links", nil)
	if err != nil {
		return fmt.Errorf("cannot get server versions: %s", err)
	}
	pageVersions := linkers.Result.Links

	newVersions := make(chan MojangApiVersion)
	errChan := make(chan error, len(pageVersions)-1)

	var wait sync.WaitGroup
	for range runtime.NumCPU() {
		wait.Add(1)
		go versions.versionProcess(newVersions, errChan, &wait)
	}

	for _, value := range pageVersions {
		if !strings.HasSuffix(value.URL, ".zip") {
			continue
		}

		// Get server version from file server
		// example: bedrock-server-1.6.1.0.zip => 1.6.1.0
		value.Version = strings.TrimSuffix(strings.TrimPrefix(path.Base(value.URL), "bedrock-server-"), ".zip")
		if _, exist := versions.Get(value.Version); exist == ErrNoVersion { // Create version to slice
			*versions = append(*versions, &Version{
				Version:   value.Version,
				IsPreview: strings.Contains(value.Type, "Preview"),
				Docker:    map[string]string{},
				Plaforms:  map[string]*PlatformVersion{},
			})
		}
		newVersions <- value // Add to worker
	}

	// Close channel and wait to worker's done
	close(newVersions)
	wait.Wait()
	if len(errChan) > 0 {
		return <-errChan
	}
	close(errChan)

	semver.Sort(*versions)
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

// Download server, convert zip file to tar with gzip compression and make SHA1 from result of compression
func (target PlatformVersion) ConvertTar(w io.Writer) (string, error) {
	// Server download
	zipFile, _, err := request.SaveTmp(target.ZipFile, "", &request.Options{Method: "GET", Header: MojangHeaders})
	if err != nil {
		return "", fmt.Errorf("cannot download server: %w", err)
	}
	defer os.Remove(zipFile.Name())
	defer zipFile.Close()

	stat, _ := zipFile.Stat()
	zipFiles, err := zip.NewReader(zipFile, stat.Size())
	if err != nil {
		return "", fmt.Errorf("cannot get files from insider zip: %w", err)
	}

	// Create new SHA1 sum and gzip compressor
	sha1Sum := sha1.New()
	gz := gzip.NewWriter(io.MultiWriter(w, sha1Sum))
	defer gz.Close()

	// Create new tarball
	tarWriter := tar.NewWriter(gz)
	defer tarWriter.Close()

	for _, file := range zipFiles.File {
		tarHead, err := tar.FileInfoHeader(file.FileInfo(), file.Name)
		if err != nil {
			return "", fmt.Errorf("cannot make tar header: %w", err)
		}
		tarHead.Name = file.Name                               // Set original file name
		if err := tarWriter.WriteHeader(tarHead); err != nil { // Write header to tarbal
			return "", fmt.Errorf("cannot write tar header: %w", err)
		} else if !file.FileInfo().Mode().IsRegular() { // skip content copy if not file
			continue
		}

		// Open file
		fileRead, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("cannot open file: %w", err)
		}
		defer fileRead.Close()

		// Copy content
		if _, err := io.Copy(tarWriter, fileRead); err != nil {
			return "", fmt.Errorf("cannot copy file: %w", err)
		}
		fileRead.Close() // Close file opened
	}

	// End file
	if err := tarWriter.Close(); err != nil {
		return "", fmt.Errorf("cannot close tar: %w", err)
	} else if err := gz.Close(); err != nil {
		return "", fmt.Errorf("cannot close gzip: %w", err)
	}

	// return SHA1 sum
	return hex.EncodeToString(sha1Sum.Sum(nil)), nil
}
