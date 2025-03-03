package mclog

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"sirherobrine23.com.br/go-bds/go-bds/overlayfs"
)

var (
	_ FileSystem = Local("")
	_ FileSystem = &Mergefs{}
)

type File interface {
	fs.File
	io.Seeker
	io.Writer
}

type FileSystem interface {
	fs.FS
	Create(string) (File, error)
}

func CreateID(fss FileSystem, dir string) (string, File, error) {
	id := make([]byte, 16)
	for attemps := 8; attemps > 0; attemps-- {
		if _, err := rand.Read(id); err != nil {
			return "", nil, &fs.PathError{Op: "create", Path: dir, Err: err}
		}
		id[6] = (id[6] & 0x0f) | 0x40
		id[8] = (id[8] & 0x3f) | 0x80
		name := fmt.Sprintf("%x", id)
		if file, err := fss.Create(filepath.Join(dir, name)); err == nil {
			return name, file, nil
		}
	}
	return "", nil, &fs.PathError{Op: "create", Path: dir, Err: fmt.Errorf("cannot make file id")}
}

// Maneger log files in Local disk
type Local string

func (local Local) Open(name string) (fs.File, error) { return os.OpenInRoot(string(local), name) }
func (local Local) Create(name string) (File, error) {
	r, err := os.OpenRoot(string(local))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return r.Create(name)
}

type Mergefs overlayfs.FsMergeFs

func (fss Mergefs) Open(name string) (fs.File, error) { return overlayfs.FsMergeFs(fss).Open(name) }
func (fss Mergefs) Create(name string) (File, error) {
	if fss.MergedFS == nil {
		return nil, fs.ErrInvalid
	}
	return fss.MergedFS.Create(filepath.Join(fss.Subdir, filepath.Clean(name)))
}
