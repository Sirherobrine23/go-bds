package request

import (
	"archive/tar"
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/descompress"
)

// Extract options
type ExtractOptions struct {
	Cwd            string // Folder output
	Strip          int    // Remove n components from file extraction
	PreserveOwners bool   // Preserver user and group if avaible
}

func (opts ExtractOptions) StripPath(name string) string {
	if name = filepath.ToSlash(name); name[0] == '/' {
		name = name[1:]
	}
	return filepath.Join(strings.Split(name, "/")[max(0, opts.Strip):]...)
}

// Create request and extract to Cwd folder
func Tar(Url string, ExtractOptions ExtractOptions, RequestOption *Options) error {
	res, err := Request(Url, RequestOption)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	descompressed, err := descompress.NewDescompress(res.Body)
	if err != nil {
		return err
	}
	defer descompressed.Close()

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
		if _, err = os.Lstat(link[0]); err != nil && os.IsNotExist(err) {
			if err = os.Symlink(link[1], link[0]); err != nil {
				return err
			}
		}
	}

	return nil
}

// Create request save to temporary file and extract to Cwd folder
func Zip(Url string, ExtractOptions ExtractOptions, RequestOptions *Options) error {
	file, _, err := SaveTmp(Url, "", RequestOptions)
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

	zipFile, err := zip.NewReader(file, stat.Size())
	if err != nil {
		return err
	}

	// Files targets
	fileReader, fileWriter := io.ReadCloser(nil), io.WriteCloser(nil)
	for _, zipFile := range zipFile.File {
		filepathCwd, fsInfo := filepath.Join(ExtractOptions.Cwd, ExtractOptions.StripPath(zipFile.Name)), zipFile.FileInfo()
		if fsInfo.IsDir() {
			if err = os.MkdirAll(filepathCwd, fsInfo.Mode()); err != nil {
				return err
			}
			continue
		} else if fileReader, err = zipFile.Open(); err != nil {
			return err
		} else if fileWriter, err = os.OpenFile(filepathCwd, os.O_CREATE|os.O_WRONLY, fsInfo.Mode()); err != nil {
			_ = fileReader.Close()
			return err
		}

		_, err = io.CopyN(fileWriter, fileReader, fsInfo.Size())
		_ = fileReader.Close()
		fileReader = nil
		_ = fileWriter.Close()
		fileWriter = nil
		if err != nil {
			return err
		}
	}
	return nil
}
