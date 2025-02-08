package overlayfs

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func (over Overlayfs) Create(name string) (*os.File, error) {
	if !over.isRW() {
		return nil, &fs.PathError{Op: "create", Path: name, Err: fs.ErrPermission}
	}

	if ok, err := over.isFileMarkedDeleted(name); err != nil {
		return nil, err
	} else if ok {
		if err := over.removeIfDeleted(name); err != nil {
			return nil, rewriteName(name, err)
		}
	}

	return os.Create(filepath.Join(over.Upper, name))
}

func (over Overlayfs) Chmod(name string, perm fs.FileMode) error {
	if !over.isRW() {
		return &fs.PathError{Op: "chmod", Path: name, Err: fs.ErrPermission}
	}

	// Clone file to upper if not exists
	if err := over.copyFileToUpperIfAbsent(name); err != nil {
		return rewriteName(name, err)
	}
	return rewriteName(name, os.Chmod(filepath.Join(over.Upper, name), perm))
}

func (over Overlayfs) Chown(name string, uid, gid int) error {
	if !over.isRW() {
		return &fs.PathError{Op: "chmod", Path: name, Err: fs.ErrPermission}
	}
	// Clone file to upper if not exists
	if err := over.copyFileToUpperIfAbsent(name); err != nil {
		return rewriteName(name, err)
	}
	return rewriteName(name, os.Chown(filepath.Join(over.Upper, name), uid, gid))
}

func (over Overlayfs) Mkdir(name string, perm fs.FileMode) error {
	if !over.isRW() {
		return &fs.PathError{Op: "mkdir", Path: name, Err: fs.ErrPermission}
	}

	if name == "." {
		return nil
	} else if err := over.removeIfDeleted(name); err != nil {
		return rewriteName(name, err)
	}

	rootFS := filepath.Join(over.Upper, filepath.Dir(name))
	if _, err := os.Stat(rootFS); err != nil {
		if err := os.MkdirAll(rootFS, perm); err != nil {
			return rewriteName(name, err)
		}
	}
	return rewriteName(name, os.Mkdir(filepath.Join(over.Upper, name), perm))
}

func (over Overlayfs) Remove(name string) error {
	if !over.isRW() {
		return &fs.PathError{Op: "remove", Path: name, Err: fs.ErrPermission}
	} else if exist, err := over.isFileMarkedDeleted(name); err == nil && exist {
		return &fs.PathError{Op: "remove", Path: name, Err: fs.ErrNotExist}
	} else if err != nil {
		return err
	}

	fileUpper := filepath.Join(over.Upper, name)
	if _, err := os.Stat(fileUpper); err == nil {
		if err = os.Remove(fileUpper); err != nil {
			return rewriteName(name, err)
		}
	}

	// Write opaque file
	return rewriteName(name, over.makeFileDeleted(name))
}

// Extends

func (over Overlayfs) MkdirAll(name string, perm fs.FileMode) error {
	if !over.isRW() {
		return &fs.PathError{Op: "mkdir", Path: name, Err: fs.ErrPermission}
	} else if name == "." {
		return nil
	}

	current, pathS := "", strings.Split(filepath.ToSlash(name), "/")
	for _, name := range pathS {
		current = filepath.Join(current, name)

		if err := over.removeIfDeleted(current); err != nil {
			return err
		} else if _, err := over.Stat(current); errors.Is(err, fs.ErrNotExist) {
			if err = over.Mkdir(current, perm.Perm()); err != nil {
				return rewriteName(name, err)
			}
		}
	}

	// Change file chmod
	return over.Chmod(name, perm)
}

func (over Overlayfs) RemoveAll(name string) error {
	if !over.isRW() {
		return &fs.PathError{Op: "remove", Path: name, Err: fs.ErrPermission}
	} else if _, err := over.Stat(name); err != nil {
		return rewriteName(name, err)
	}

	fileList := []string{}
	if err := fs.WalkDir(over.Mergefs(), name, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return rewriteName(name, err)
		} else if d.IsDir() {
			return nil
		}
		fileList = append(fileList, path)
		return nil
	}); err != nil {
		return rewriteName(name, err)
	}
	slices.Reverse(fileList) // Reverse list to get last files
	for _, name := range fileList {
		if err := over.Remove(name); err != nil {
			return rewriteName(name, err)
		}
	}

	// Write opaque file
	dir, name := filepath.Split(name)
	if err := over.WriteFile(filepath.Join(over.Upper, dir, OpaqueWhiteout+name), nil, 0666); err != nil {
		return rewriteName(name, err)
	}

	return nil
}

func (over Overlayfs) WriteFile(name string, data []byte, perm fs.FileMode) error {
	f, err := over.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return rewriteName(name, err)
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return rewriteName(name, err)
}

// RW + Especial

func (over Overlayfs) Rename(oldpath, newpath string) error {
	if !over.isRW() {
		return &fs.PathError{Op: "rename", Path: oldpath, Err: fs.ErrPermission}
	} else if !filepath.IsAbs(newpath) {
		newpath = filepath.Join(over.Upper, newpath)
	}

	oldFiles, err := over.retrieveFileFromLayers(oldpath)
	if err != nil {
		return err
	} else if len(oldFiles) == 0 {
		return &fs.PathError{Op: "rename", Path: oldpath, Err: fs.ErrNotExist}
	}
	fileNode := oldFiles[0]

	if fileNode.FromUpper {
		if err := os.Rename(filepath.Join(fileNode.Layer, oldpath), newpath); err != nil {
			return err
		}
	} else {
		if err := over.copyFromTo(fileNode, oldpath, filepath.Join(fileNode.Layer, oldpath), newpath); err != nil {
			return err
		}
	}

	return rewriteName(oldpath, over.makeFileDeleted(oldpath))
}

func (over Overlayfs) OpenFile(name string, flags int, perm fs.FileMode) (*os.File, error) {
	flaged := flages(flags)
	if flaged.IsWrite() && !over.isRW() {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrPermission}
	}

	if flaged.CreateIfNotExist() {
		if err := over.removeIfDeleted(name); err != nil {
			return nil, err
		}
	} else {
		if ok, err := over.isFileMarkedDeleted(name); err != nil {
			return nil, err
		} else if ok {
			return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
		}
	}

	if flaged.IsWrite() {
		// Clone file to upper if not exists
		if err := over.copyFileToUpperIfAbsent(name); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, rewriteName(name, err)
		}

		fromUpperPath := filepath.Join(over.Upper, name)
		if _, err := os.Stat(filepath.Dir(fromUpperPath)); err != nil {
			if err := os.MkdirAll(filepath.Dir(fromUpperPath), perm.Perm()); err != nil {
				return nil, rewriteName(name, err)
			}
		}

		file, err := os.OpenFile(fromUpperPath, flags, perm)
		if err != nil {
			err = rewriteName(name, err)
		}
		return file, err
	}

	layers, err := over.retrieveFileFromLayers(name)
	if err != nil {
		return nil, err
	} else if len(layers) == 0 {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	return os.OpenFile(filepath.Join(layers[0].Layer, name), flags, perm)
}

func (over Overlayfs) Symlink(oldname string, newname string) error {
	if !over.isRW() {
		return &os.LinkError{Op: "symlink", Err: fs.ErrPermission, Old: oldname, New: newname}
	} else if err := over.copyFileToUpperIfAbsent(oldname); err != nil {
		return rewriteName(oldname, err)
	}
	return os.Symlink(filepath.Join(over.Upper, oldname), filepath.Join(over.Upper, newname))
}

func (over Overlayfs) Truncate(name string, size int64) error {
	file, err := over.OpenFile(name, os.O_TRUNC | os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()
	return rewriteName(name, file.Truncate(size))
}
