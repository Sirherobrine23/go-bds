package request

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func Zip(Url string, TarOption TarOptions, RequestOption *Options) error {
	request, err := MountRequest(Url, RequestOption)
	if err != nil {
		return err
	}

	res, err := request.MakeRequestWithStatus()
	if err != nil {
		return err
	}
	defer res.Body.Close()

	file, err := os.CreateTemp("", "tmpzipfile*.zip")
	if err != nil {
		return err
	}
	defer file.Close()
	defer os.Remove(file.Name())
	if _, err := io.Copy(file, res.Body); err != nil {
		return err
	} else if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	zipFile, err := zip.NewReader(file, stat.Size())
	if err != nil {
		return err
	}

	for _, file := range zipFile.File {
		fileInfo := file.FileInfo()
		rootFile := filepath.Join(TarOption.Cwd, stripPath(fileInfo.Name(), TarOption.Strip))
		if fileInfo.IsDir() {
			if err := os.MkdirAll(rootFile, fileInfo.Mode()); err != nil {
				return err
			} else if err := os.Chtimes(rootFile, file.Modified, file.Modified); err != nil {
				return err
			}
			continue
		}
		{
			fileExt, err := file.Open()
			if err != nil {
				return err
			}
			defer fileExt.Close()

			if err := os.MkdirAll(filepath.Dir(rootFile), fileInfo.Mode()); err != nil {
				return err
			}
			localFile, err := os.OpenFile(rootFile, os.O_CREATE|os.O_WRONLY, fileInfo.Mode())
			if err != nil {
				return err
			}
			defer localFile.Close()
			if _, err := io.CopyN(localFile, fileExt, fileInfo.Size()); err != nil {
				return err
			} else if err := os.Chtimes(rootFile, file.Modified, file.Modified); err != nil {
				return err
			}
			localFile.Close()
			fileExt.Close()
		}
	}

	return nil
}
