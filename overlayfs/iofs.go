package overlayfs

import (
	"io/fs"
	"path/filepath"
)

var (
	_ fs.FS         = &FsMergeFs{}
	_ fs.ReadFileFS = &FsMergeFs{}
	_ fs.ReadDirFS  = &FsMergeFs{}
	_ fs.StatFS     = &FsMergeFs{}
	_ fs.GlobFS     = &FsMergeFs{}
	_ fs.SubFS      = &FsMergeFs{}
)

type FsMergeFs struct {
	MergedFS *Overlayfs
	Subdir   string
}

// Pipe MergeFS to [io/fs.FS]
func NewFS(fs *Overlayfs) fs.FS {
	return &FsMergeFs{
		MergedFS: fs,
		Subdir:   "",
	}
}

func (fss FsMergeFs) Sub(dir string) (fs.FS, error) {
	return &FsMergeFs{MergedFS: fss.MergedFS, Subdir: filepath.Join(fss.Subdir, dir)}, nil
}

func (fss FsMergeFs) Open(name string) (fs.File, error) {
	if fss.MergedFS == nil {
		return nil, fs.ErrInvalid
	}
	return fss.MergedFS.Open(filepath.Join(fss.Subdir, name))
}

func (fss FsMergeFs) ReadDir(name string) ([]fs.DirEntry, error) {
	if fss.MergedFS == nil {
		return nil, fs.ErrInvalid
	}
	return fss.MergedFS.ReadDir(filepath.Join(fss.Subdir, name))
}

func (fss FsMergeFs) Stat(name string) (fs.FileInfo, error) {
	if fss.MergedFS == nil {
		return nil, fs.ErrInvalid
	}
	return fss.MergedFS.Stat(filepath.Join(fss.Subdir, name))
}

func (fss FsMergeFs) ReadFile(name string) ([]byte, error) {
	if fss.MergedFS == nil {
		return nil, fs.ErrInvalid
	}
	return fss.MergedFS.ReadFile(filepath.Join(fss.Subdir, name))
}

func (fss FsMergeFs) Glob(pattern string) ([]string, error) {
	if fss.MergedFS == nil {
		return nil, fs.ErrInvalid
	}
	return fs.Glob(fss, pattern)
}
