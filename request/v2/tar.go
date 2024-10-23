package request

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"compress/zlib"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

type ExtractOptions struct {
	Strip                 int    // Remove n components from file extraction
	Cwd                   string // Folder output
	PreserveOwners        bool   // Preserver user and group
	Gzip, Bzip2, Zlib, Xz bool   // Uncompress
}

func stripPath(fpath string, st int) string {
	return filepath.Join(strings.Split(filepath.ToSlash(fpath), "/")[int(math.Max(0, float64(st))):]...)
}

// Create request and extract to Cwd folder
func Tar(Url string, TarOption ExtractOptions, RequestOption *Options) error {
	request, err := MountRequest(Url, RequestOption)
	if err != nil {
		return err
	}

	res, err := request.MakeRequestWithStatus()
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var tarball io.Reader = res.Body
	if TarOption.Gzip || strings.Contains(res.Header.Get("Content-Type"), "application/gzip") {
		if tarball, err = gzip.NewReader(res.Body); err != nil {
			return err
		}
		defer tarball.(*gzip.Reader).Close()
	} else if TarOption.Xz || strings.Contains(res.Header.Get("Content-Type"), "application/x-xz") {
		if tarball, err = xz.NewReader(res.Body); err != nil {
			return err
		}
	} else if TarOption.Zlib || strings.Contains(res.Header.Get("Content-Type"), "application/zlib") {
		if tarball, err = zlib.NewReader(res.Body); err != nil {
			return err
		}
		defer tarball.(io.ReadCloser).Close()
	} else if TarOption.Bzip2 || strings.Contains(res.Header.Get("Content-Type"), "application/bzip2") || strings.Contains(res.Header.Get("Content-Type"), "application/x-bzip2") {
		tarball = bzip2.NewReader(res.Body)
	}

	tarReader := tar.NewReader(tarball)
	for {
		head, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		rootFile := filepath.Join(TarOption.Cwd, stripPath(head.Name, TarOption.Strip))
		if rootFile == TarOption.Cwd {
			continue
		}

		fsInfo := head.FileInfo()
		mode := fsInfo.Mode()
		if fsInfo.IsDir() {
			if err := os.MkdirAll(rootFile, mode.Perm()); err != nil {
				return err
			}
			continue
		} else if !mode.IsRegular() {
			continue
		}
		localFile, err := os.OpenFile(rootFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Perm())
		if err != nil {
			return err
		} else if _, err := io.CopyN(localFile, tarReader, head.Size); err != nil {
			localFile.Close() // Close file
			return err
		}
		localFile.Close() // Close file
	}
}
