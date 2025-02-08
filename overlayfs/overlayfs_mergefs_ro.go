package overlayfs

import (
	"errors"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func (over Overlayfs) Open(name string) (file *os.File, err error) {
	if ok, err := over.isFileMarkedDeleted(name); err != nil {
		return nil, err
	} else if ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	layers := over.allLayers()
	for _, folder := range layers {
		if file, err = os.Open(filepath.Join(folder, name)); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return
		}
		return // Return file
	}

	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

func (over Overlayfs) Stat(name string) (stat fs.FileInfo, err error) {
	if ok, err := over.isFileMarkedDeleted(name); err != nil {
		return nil, err
	} else if ok {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
	}

	layers := over.allLayers()
	for _, folder := range layers {
		if stat, err = os.Stat(filepath.Join(folder, name)); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return
		}
		return // Return file
	}

	return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
}

func (over Overlayfs) Lstat(name string) (stat fs.FileInfo, err error) {
	if ok, err := over.isFileMarkedDeleted(name); err != nil {
		return nil, err
	} else if ok {
		return nil, &fs.PathError{Op: "lstat", Path: name, Err: fs.ErrNotExist}
	}

	layers := over.allLayers()
	for _, folder := range layers {
		if stat, err = os.Lstat(filepath.Join(folder, name)); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			err = rewriteName(name, err)
			return
		}
		return
	}

	return nil, &fs.PathError{Op: "lstat", Path: name, Err: fs.ErrNotExist}
}

func (over Overlayfs) ReadDir(name string) ([]fs.DirEntry, error) {
	if ok, err := over.isFileMarkedDeleted(name); err != nil {
		return nil, err
	} else if ok {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}

	layers := over.allLayers()
	fileMap := map[string]fs.DirEntry{}

	for _, folder := range layers {
		entrys, err := os.ReadDir(filepath.Join(folder, name))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			if e, ok := err.(*fs.PathError); ok {
				e.Path = name
			}
			return nil, err
		}
		for _, entry := range entrys {
			fileMap[entry.Name()] = entry // Replace if exists
		}
	}

	if len(fileMap) == 0 {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}

	// Delete opaque files
	for key := range fileMap {
		if strings.HasPrefix(key, OpaqueWhiteout) {
			delete(fileMap, key)
			name := key[len(OpaqueWhiteout):]
			delete(fileMap, name)
		}
	}

	// Return slice from map values
	return slices.Collect(maps.Values(fileMap)), nil
}

// RO + Extends
func (over Overlayfs) ReadFile(name string) ([]byte, error) {
	file, err := over.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func (over Overlayfs) Readlink(name string) (string, error) {
	layers := over.allLayers()
	for _, folder := range layers {
		if folder == "" {
			continue
		}
		target, err := os.Readlink(filepath.Join(folder, name))
		if err != nil && errors.Is(err, fs.ErrNotExist) {
			continue
		}
		return target, err
	}
	return "", &fs.PathError{Op: "readlink", Path: name, Err: fs.ErrNotExist}
}
