package request

import (
	"archive/tar"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/descompress"
)

type ExtractOptions struct {
	Strip          int    // Remove n components from file extraction
	Cwd            string // Folder output
	PreserveOwners bool   // Preserver user and group
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

	descompressed, err := descompress.NewDescompress(res.Body)
	if err != nil {
		return err
	} else if closer, ok := descompressed.(io.Closer); ok {
		defer closer.Close()
	}

	linkes := [][2]string{}
	tarReader := tar.NewReader(descompressed)
	for {
		head, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		rootFile := filepath.Join(TarOption.Cwd, stripPath(head.Name, TarOption.Strip))
		if _, err := os.Stat(rootFile); err == nil {
			continue
		} else if rootFile == TarOption.Cwd {
			continue
		} else if head.FileInfo().IsDir() {
			if err := os.MkdirAll(rootFile, head.FileInfo().Mode()); err != nil {
				return err
			}
			continue
		} else if head.FileInfo().Mode().Type() == fs.ModeSymlink {
			targetPath := filepath.Join(filepath.Dir(rootFile), head.Linkname)
			if filepath.IsAbs(head.Linkname) {
				targetPath = head.Linkname
			}
			linkes = append(linkes, [2]string{rootFile, targetPath})
			continue
		}

		if _, err := os.Stat(filepath.Dir(rootFile)); err != nil && os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(rootFile), 0777); err != nil {
				return err
			}
		}

		if head.FileInfo().Mode().IsRegular() {
			// Open file or create
			localFile, err := os.OpenFile(rootFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, head.FileInfo().Mode())
			if err != nil {
				return err
			}

			_, err = io.CopyN(localFile, tarReader, head.Size) // Copy data
			localFile.Close()                                  // Close file
			if err != nil {
				return err
			}
		}
	}

	// Create linkers
	for _, link := range linkes {
		if _, err := os.Lstat(link[0]); err != nil && os.IsNotExist(err) {
			if err := os.Symlink(link[1], link[0]); err != nil {
				return err
			}
		}
	}

	return nil
}
