package mclog

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

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
	Mkdir(string, fs.FileMode) error
}

func MkdirAll(fss FileSystem, dir string, perm fs.FileMode) error {
	paths := strings.Split(filepath.ToSlash(dir), "/")
	for pathIndex := range paths {
		if err := fss.Mkdir(path.Join(paths[:pathIndex]...), perm); err != nil {
			return err
		}
	}
	return nil
}

func CreateID(fss FileSystem, dir string) (string, File, error) {
	dir = path.Clean(filepath.ToSlash(dir))
	if _, err := fs.Stat(fss, dir); err != nil {
		if err := MkdirAll(fss, dir, 0755); err != nil {
			return "", nil, err
		}
	}

	id := make([]byte, 16)
	for attemps := 8; attemps > 0; attemps-- {
		if _, err := rand.Read(id); err != nil {
			return "", nil, &fs.PathError{Op: "create", Path: dir, Err: err}
		}
		id[6] = (id[6] & 0x0f) | 0x40
		id[8] = (id[8] & 0x3f) | 0x80
		name := fmt.Sprintf("%x", id)
		if _, err := fs.Stat(fss, path.Join(dir, name)); err == nil {
			continue // Skip exists
		} else if file, err := fss.Create(path.Join(dir, name)); err == nil {
			return name, file, nil
		}
	}
	return "", nil, &fs.PathError{Op: "create", Path: dir, Err: fmt.Errorf("cannot make file id")}
}

// Maneger log files in Local disk
type Local string

func (local Local) Open(name string) (fs.File, error) { return os.OpenInRoot(string(local), name) }
func (local Local) Mkdir(name string, perm fs.FileMode) error {
	r, err := os.OpenRoot(string(local))
	if err != nil {
		return err
	}
	defer r.Close()
	return r.Mkdir(name, perm.Perm())
}
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
func (fss Mergefs) Mkdir(name string, perm fs.FileMode) error {
	return overlayfs.FsMergeFs(fss).MergedFS.Mkdir(filepath.Join(fss.Subdir, filepath.Clean(name)), perm.Perm())
}
func (fss Mergefs) Create(name string) (File, error) {
	if fss.MergedFS == nil {
		return nil, fs.ErrInvalid
	}
	return fss.MergedFS.Create(filepath.Join(fss.Subdir, filepath.Clean(name)))
}
