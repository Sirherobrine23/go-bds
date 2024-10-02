package request

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"compress/zlib"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ExtractOptions struct {
	Strip             int    // Remove n components from file extraction
	Cwd               string // Folder output
	PreserveOwners    bool   // Preserver user and group
	Gzip, Bzip2, Zlib bool   // Uncompress
}

func stripPath(fpath string, st int) string {
	return filepath.Join(strings.Split(filepath.ToSlash(fpath), "/")[st:]...)
}

func updateFile(rootFile string, head *tar.Header, PreserveOwners bool) error {
	if err := os.Chtimes(rootFile, head.AccessTime, head.ModTime); err != nil {
		return err
	} else if PreserveOwners {
		if err := os.Chown(rootFile, head.Uid, head.Gid); err != nil {
			return err
		}
	}
	return nil
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
		defer tarball.(io.ReadCloser).Close()
	} else if TarOption.Zlib || strings.Contains(res.Header.Get("Content-Type"), "application/zlib") {
		if tarball, err = zlib.NewReader(res.Body); err != nil {
			return err
		}
		defer tarball.(io.ReadCloser).Close()
	} else if TarOption.Bzip2 || strings.Contains(res.Header.Get("Content-Type"), "application/bzip2") || strings.Contains(res.Header.Get("Content-Type"), "application/x-bzip2") {
		tarball = bzip2.NewReader(res.Body)
	}

	tarExt := tar.NewReader(tarball)
	for {
		head, err := tarExt.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fileInfo := head.FileInfo()
		rootFile := filepath.Join(TarOption.Cwd, stripPath(fileInfo.Name(), TarOption.Strip))
		if fileInfo.IsDir() {
			if err := os.MkdirAll(rootFile, fileInfo.Mode()); err != nil {
				return err
			} else if err := updateFile(rootFile, head, TarOption.PreserveOwners); err != nil {
				return err
			}
			continue
		}
		switch head.Typeflag {
		case tar.TypeBlock, tar.TypeChar, tar.TypeFifo:
			continue
		case tar.TypeSymlink:
			if err := os.Symlink(head.Linkname, rootFile); err != nil {
				return err
			}
		default:
			{
				if err := os.MkdirAll(filepath.Dir(rootFile), fileInfo.Mode()); err != nil {
					return err
				}
				localFile, err := os.OpenFile(rootFile, os.O_CREATE|os.O_WRONLY, fileInfo.Mode())
				if err != nil {
					return err
				}
				defer localFile.Close()
				if _, err := io.CopyN(localFile, tarExt, head.Size); err != nil {
					return err
				} else if err := updateFile(rootFile, head, TarOption.PreserveOwners); err != nil {
					return err
				}
				localFile.Close()
			}
		}
	}

	return nil
}
