//go:build (aix && ppc64) || darwin || (windows && amd64) || (linux && (amd64 || arm64 || riscv64 || ppc64le || s390x)) || (solaris && (amd64 || sparcv9))

// Pre build openjdk
package adoptium

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/request"
)

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
	return ErrSystem
}

func InstallLatest(featVersion uint, installPath string) error {
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
	return download(installPath, fileUrl)
}
