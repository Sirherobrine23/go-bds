package adoptium

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"sirherobrine23.org/go-bds/go-bds/request"
)

var ErrNoAvaible = errors.New("host not avaible prebuild version")

type Versions []Version
type Version struct {
	Binaries []struct {
		Architecture  string `json:"architecture"` // x64, x86, x32, ppc64, ppc64le, s390x, aarch64, arm, sparcv9, riscv64
		Os            string `json:"os"`           // linux, windows, mac, solaris, aix, alpine-linux
		DownloadCount int    `json:"download_count"`
		HeapSize      string `json:"heap_size"`
		ImageType     string `json:"image_type"`
		JvmImpl       string `json:"jvm_impl"`
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

func download(path, url string) error {
	ext := filepath.Ext(filepath.Base(url))
	switch ext {
	case ".zip":
		localFile, err := os.CreateTemp(os.TempDir(), "java_*.zip")
		if err != nil {
			return err
		}
		defer localFile.Close()
		defer os.Remove(localFile.Name())

		res, err := (&request.RequestOptions{HttpError: true, Url: url}).Request()
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if _, err := io.Copy(localFile, res.Body); err != nil {
			return err
		} else if _, err := localFile.Seek(0, 0); err != nil {
			return err
		}

		s, _ := localFile.Stat()
		zr, err := zip.NewReader(localFile, s.Size())
		if err != nil {
			return err
		}

		for _, file := range zr.File {
			info := file.FileInfo()
			pathFixed := filepath.Join(path, strings.Join(filepath.SplitList(file.Name)[1:], "/"))
			if info.IsDir() {
				if err := os.MkdirAll(pathFixed, info.Mode()); err != nil {
					return err
				}
				continue
			}

			tfile, err := os.OpenFile(pathFixed, os.O_CREATE|os.O_EXCL|os.O_RDWR, info.Mode())
			if err != nil {
				return err
			}
			defer tfile.Close()
			zrf, err := file.Open()
			if err != nil {
				return err
			}
			defer zrf.Close()
			if _, err := io.Copy(tfile, zrf); err != nil {
				return err
			}
		}

		return nil
	case ".tgz", ".gz", ".tar.gz":
		res, err := (&request.RequestOptions{HttpError: true, Url: url}).Request()
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
				if err == io.EOF {
					break
				}
				return err
			}
			info := head.FileInfo()
			pathFixed := filepath.Join(path, strings.Join(filepath.SplitList(head.Name)[1:], "/"))
			if info.IsDir() {
				if err := os.MkdirAll(pathFixed, info.Mode()); err != nil {
					return err
				}
				continue
			}
			tfile, err := os.OpenFile(pathFixed, os.O_CREATE|os.O_EXCL|os.O_RDWR, info.Mode())
			if err != nil {
				return err
			}
			defer tfile.Close()
			if _, err := io.CopyN(tfile, tarball, info.Size()); err != nil {
				return err
			}
		}
		return nil
	}
	return ErrNoAvaible
}

func GetLatest(featVersion uint, path string) error {
	var reqOpt request.RequestOptions
	// architecture: x64, x86, x32, ppc64, ppc64le, s390x, aarch64, arm, sparcv9, riscv64
	var arch = runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "x64"
	case "386":
		arch = "x86"
	case "arm64":
		arch = "aarch64"
	}

	// os: linux, windows, mac, solaris, aix, alpine-linux
	var os = runtime.GOOS
	switch os {
	case "darwin":
		os = "mac"
	case "sunos":
		os = "solaris"
	}

	reqOpt.Url = fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/%d/ga/%s/%s/jdk/hotspot/normal/eclipse", featVersion, os, arch)
	fileUrl, err := reqOpt.GetRedirect()
	if err != nil {
		return err
	}
	return download(path, fileUrl)
}

func (version *Version) Download(path string) error {
	for _, bin := range version.Binaries {
		if !((bin.Os == "mac" && runtime.GOOS == "darwin") || bin.Os == runtime.GOOS) {
			continue
		} else if !((bin.Architecture == "x64" && runtime.GOARCH == "amd64") || (bin.Architecture == "x86" && runtime.GOARCH == "386") || (bin.Architecture == "aarch64" && runtime.GOARCH == "arm64") || bin.Architecture == runtime.GOARCH) {
			continue
		}
		if err := download(path, bin.Package.Link); err != nil {
			if err == ErrNoAvaible {
				continue
			}
			return err
		}
	}

	return ErrNoAvaible
}

func Releases() (Versions, error) {
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
	var concatedVersions Versions
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

		res, err := req.Request()
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

		var releases Versions
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

	return concatedVersions, nil
}
