// Create difference from [io/fs.FS]
package fsdiff

import (
	"bytes"
	"crypto/sha512"
	"io"
	"io/fs"
	"os"
)

// Return files differences from [io/fs.FS]
func Diff(v1, v2 fs.FS) ([]string, error) {
	files := []string{}
	err := fs.WalkDir(v1, "/", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.Type().IsDir() {
			return err
		}

		v2file, err := v2.Open(path)
		if os.IsNotExist(err) {
			files = append(files, path)
		} else if err != nil {
			return err
		}
		defer v2file.Close()

		v1stat, _ := d.Info()
		v2stat, _ := v2file.Stat()
		if v1stat.ModTime().Equal(v2stat.ModTime()) {
			return nil
		}

		v1file, err := v1.Open(path)
		if err != nil {
			return err
		}
		defer v1file.Close()

		v1SHA := sha512.New()
		io.Copy(v1SHA, v1file)
		v2SHA := sha512.New()
		io.Copy(v2SHA, v2file)

		// Add file diff files
		if !bytes.Equal(v1SHA.Sum(nil), v2SHA.Sum(nil)) {
			files = append(files, path)
		}

		return nil
	})
	return files, err
}
