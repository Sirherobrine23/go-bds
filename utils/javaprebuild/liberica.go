package javaprebuild

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"sirherobrine23.com.br/go-bds/request/v2"
	"sirherobrine23.com.br/sirherobrine23/go-dpkg/dpkg"
)

type libericaVersion struct {
	DownloadURL    string `json:"downloadUrl"`
	FeatureVersion int    `json:"featureVersion"`
	Os             string `json:"os"`
	Architecture   string `json:"architecture"`
	Version        string `json:"version"`
	Sha1           string `json:"sha1"`
	Filename       string `json:"filename"`
	Size           int    `json:"size"`
}

// Install latest liberica file
func (ver JavaVersion) InstallLatestLiberica(installPath string) error {
	requUrl, _ := url.Parse("https://api.bell-sw.com/v1/liberica/releases")
	query := requUrl.Query()
	query.Set("version-modifier", "latest")
	// query.Set("package-type", "tar.gz")
	query.Set("bundle-type", "jre")
	query.Set("version-feature", fmt.Sprint(uint(ver)-44))

	// os: "linux", "linux-musl", "macos", "solaris", "windows"
	switch runtime.GOOS {
	case "linux", "windows":
		query.Set("os", runtime.GOOS)
	case "darwin":
		query.Set("os", "macos")
	case "sunos":
		query.Set("os", "solaris")
	default:
		return ErrSystem
	}

	// bitness: "32", "64"
	// arch: "arm", "ppc", "riscv", "sparc", "x86"
	switch runtime.GOARCH {
	case "amd64":
		query.Set("arch", "x86")
		query.Set("bitness", "64")
	case "386":
		query.Set("arch", "x86")
		query.Set("bitness", "32")
	case "arm64":
		query.Set("arch", "arm")
		query.Set("bitness", "64")
	case "arm":
		query.Set("arch", "arm")
		query.Set("bitness", "32")
	case "ppc64":
		query.Set("arch", "ppc")
		query.Set("bitness", "64")
	default:
		return ErrSystem
	}

	requUrl.RawQuery = query.Encode()
	releases, _, err := request.MakeJSON[[]libericaVersion](requUrl, nil)
	if err != nil {
		return err
	}

	for _, release := range releases {
		downloadUrl := release.DownloadURL
		switch strings.ToLower(filepath.Ext(downloadUrl)) {
		case ".tar", ".tar.gz":
			return request.Tar(downloadUrl, request.ExtractOptions{Strip: 1, Cwd: installPath}, nil)
		case ".zip":
			return request.Zip(downloadUrl, request.ExtractOptions{Strip: 1, Cwd: installPath}, nil)
		case ".deb":
			res, err := request.Request(downloadUrl, nil)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			_, pkgData, err := dpkg.NewReader(res.Body)
			if err != nil {
				return err
			}
			return extractTar(pkgData, request.ExtractOptions{Strip: 0, Cwd: installPath})
		}
	}

	return ErrSystem
}

func extractTar(descompressed io.Reader, ExtractOptions request.ExtractOptions) error {
	linkes := [][2]string{}
	tarReader := tar.NewReader(descompressed)
	for {
		fileHeader, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		rootFile := filepath.Join(ExtractOptions.Cwd, ExtractOptions.StripPath(fileHeader.Name))
		if _, err := os.Stat(rootFile); err == nil {
			continue
		} else if rootFile == ExtractOptions.Cwd {
			continue
		} else if fileHeader.FileInfo().IsDir() {
			if err := os.MkdirAll(rootFile, fileHeader.FileInfo().Mode()); err != nil {
				return err
			}
			continue
		} else if fileHeader.FileInfo().Mode().Type() == fs.ModeSymlink {
			targetPath := filepath.Join(filepath.Dir(rootFile), fileHeader.Linkname)
			if filepath.IsAbs(fileHeader.Linkname) {
				targetPath = fileHeader.Linkname
			}
			linkes = append(linkes, [2]string{rootFile, targetPath})
			continue
		}

		if _, err := os.Stat(filepath.Dir(rootFile)); err != nil && os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(rootFile), 0777); err != nil {
				return err
			}
		}

		if fileHeader.FileInfo().Mode().IsRegular() {
			// Open file or create
			localFile, err := os.OpenFile(rootFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileHeader.FileInfo().Mode())
			if err != nil {
				return err
			}

			_, err = io.CopyN(localFile, tarReader, fileHeader.Size) // Copy data
			localFile.Close()                                        // Close file
			if err != nil {
				return err
			}
		}
	}

	// Create linkers
	for _, link := range linkes {
		if _, err := os.Lstat(link[0]); err != nil && os.IsNotExist(err) {
			if err = os.Symlink(link[1], link[0]); err != nil {
				return err
			}
		}
	}

	return nil
}
