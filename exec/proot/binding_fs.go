package proot

import (
	"io"
	"io/fs"
	"path"
	"path/filepath"
)

var _ File = (*FSSFile)(nil)

type FSSFile struct {
	fs.File
	fileName string
}

func (fssFile FSSFile) Name() string { return path.Base(filepath.ToSlash(fssFile.fileName)) }
func (fssFile FSSFile) WriteTo(w io.Writer) (n int64, err error) {
	if WW, ok := fssFile.File.(io.WriterTo); ok {
		return WW.WriteTo(w)
	}
	return io.Copy(w, fssFile.File)
}

func (fssFile FSSFile) Write(b []byte) (n int, err error) {
	if w, ok := fssFile.File.(io.Writer); ok {
		return w.Write(b)
	}
	return 0, fs.ErrInvalid
}

func (fssFile FSSFile) ReadFrom(r io.Reader) (n int64, err error) {
	if RR, ok := fssFile.File.(io.ReaderFrom); ok {
		return RR.ReadFrom(r)
	}
	return io.Copy(fssFile, r)
}

func (fssFile FSSFile) Seek(offset int64, whence int) (ret int64, err error) {
	if s, ok := fssFile.File.(io.Seeker); ok {
		return s.Seek(offset, whence)
	}
	return 0, fs.ErrInvalid
}

func (fssFile FSSFile) ReadAt(b []byte, off int64) (n int, err error)  { return 0, fs.ErrInvalid }
func (fssFile FSSFile) ReadDir(n int) ([]fs.DirEntry, error)           { return nil, fs.ErrInvalid }
func (fssFile FSSFile) Readdir(n int) ([]fs.FileInfo, error)           { return nil, fs.ErrInvalid }
func (fssFile FSSFile) Sync() error                                    { return nil }
func (fssFile FSSFile) WriteAt(b []byte, off int64) (n int, err error) { return 0, fs.ErrInvalid }

var _ Binding = (*FS)(nil)

// Implement Binding to [io/fs.FS]
type FS struct {
	fs.FS
}

func (FS) ReadOnly() bool                               { return true }
func (FS) Mkdir(name string, perm fs.FileMode) error    { return fs.ErrInvalid }
func (FS) Symlink(oldname string, newname string) error { return fs.ErrInvalid }
func (FS) Chmod(name string, perm fs.FileMode) error    { return fs.ErrInvalid }
func (FS) Chown(name string, uid, gid int) error        { return fs.ErrInvalid }

func (fss FS) ReadDir(name string) ([]fs.DirEntry, error)     { return fs.ReadDir(fss.FS, name) }
func (fss FS) Stat(name string) (stat fs.FileInfo, err error) { return fs.Stat(fss.FS, name) }

func (fss FS) OpenFile(name string, flags int, perm fs.FileMode) (File, error) {
	openFlags := OsFlags(flags)
	if openFlags.IsWrite() || openFlags.CreateIfNotExist() {
		return nil, fs.ErrInvalid
	}
	file, err := fss.FS.Open(name)
	if err != nil {
		return nil, err
	}
	return &FSSFile{fileName: name, File: file}, nil
}