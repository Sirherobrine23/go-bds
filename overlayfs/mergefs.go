package overlayfs

import (
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const OpaqueWhiteout string = ".wh."

// Mkdir creates a new directory with the specified name and permission bits (before umask). If there is an error, it will be of type *PathError.
func (merge Overlayfs) Mkdir(name string, perm fs.FileMode) error {
	if merge.Upper == "" {
		return &fs.PathError{Op: "mkdir", Path: filepath.Clean(name), Err: fs.ErrPermission}
	}
	return os.Mkdir(filepath.Join(merge.Upper, name), perm)
}

// MkdirAll creates a directory named path, along with any necessary parents, and returns nil, or else returns an error. The permission bits perm (before umask) are used for all directories that MkdirAll creates. If path is already a directory, MkdirAll does nothing and returns nil.
func (merge Overlayfs) MkdirAll(name string, perm fs.FileMode) error {
	if merge.Upper == "" {
		return &fs.PathError{Op: "mkdir", Path: filepath.Clean(name), Err: fs.ErrPermission}
	}
	return os.MkdirAll(filepath.Join(merge.Upper, name), perm)
}

// Stat returns a [FileInfo] describing the named file. If there is an error, it will be of type [*PathError].
func (merge Overlayfs) Stat(name string) (os.FileInfo, error) {
	if name == "" || name == "." || name == "/" {
		if merge.Upper == "" {
			return nil, &fs.PathError{Op: "readdir", Path: filepath.Clean(name), Err: fs.ErrPermission}
		}
		return os.Stat(merge.Upper)
	}

	folders := append(merge.Lower, merge.Upper)
	slices.Reverse(folders)
	for _, folderPath := range folders {
		if folderPath == "" {
			continue
		}

		s, err := os.Stat(filepath.Join(folderPath, name))
		if err != nil {
			continue
		}
		return s, nil
	}
	return nil, &fs.PathError{Op: "readdir", Path: filepath.Clean(name), Err: fs.ErrNotExist}
}

// Create creates or truncates the named file. If the file already exists, it is truncated. If the file does not exist, it is created with mode 0o666 (before umask). If successful, methods on the returned File can be used for I/O; the associated file descriptor has mode O_RDWR. If there is an error, it will be of type *PathError.
func (merge Overlayfs) Create(name string) (*os.File, error) {
	return merge.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Open opens the named file for reading. If successful, methods on the returned file can be used for reading; the associated file descriptor has mode O_RDONLY. If there is an error, it will be of type *PathError.
func (merge Overlayfs) Open(name string) (*os.File, error) {
	return merge.OpenFile(name, os.O_RDONLY, 0)
}

// Remove removes the named file or directory. If there is an error, it will be of type *PathError.
func (merge Overlayfs) Remove(name string) error {
	if merge.Upper == "" {
		return &fs.PathError{Op: "remove", Path: filepath.Clean(name), Err: fs.ErrPermission}
	}
	if err := os.Remove(filepath.Join(merge.Upper, name)); err != nil {
		return err
	}
	dir, name := filepath.Split(name)
	return merge.WriteFile(filepath.Join(merge.Upper, dir, OpaqueWhiteout+name), nil, 0666) // Write null Head
}

// RemoveAll removes path and any children it contains. It removes everything it can but returns the first error it encounters. If the path does not exist, RemoveAll returns nil (no error). If there is an error, it will be of type [*PathError].
func (merge Overlayfs) RemoveAll(name string) error {
	if merge.Upper == "" {
		return &fs.PathError{Op: "remove", Path: filepath.Clean(name), Err: fs.ErrPermission}
	}
	if err := os.RemoveAll(filepath.Join(merge.Upper, name)); err != nil {
		return err
	}
	dir, name := filepath.Split(name)
	return merge.WriteFile(filepath.Join(merge.Upper, dir, OpaqueWhiteout+name), nil, 0666) // Write null Head
}

// WriteFile writes data to the named file, creating it if necessary. If the file does not exist, WriteFile creates it with permissions perm (before umask); otherwise WriteFile truncates it before writing, without changing permissions. Since WriteFile requires multiple system calls to complete, a failure mid-operation can leave the file in a partially written state.
func (merge Overlayfs) WriteFile(name string, data []byte, perm fs.FileMode) error {
	f, err := merge.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

// ReadFile reads the named file and returns the contents. A successful call returns err == nil, not err == EOF. Because ReadFile reads the whole file, it does not treat an EOF from Read as an error to be reported.
func (merge Overlayfs) ReadFile(name string) ([]byte, error) {
	f, err := merge.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// ReadDir reads the named directory, returning all its directory entries. If an error occurs reading the directory, ReadDir returns the entries it was able to read before the error, along with the error.
func (merge Overlayfs) ReadDir(name string) ([]fs.DirEntry, error) {
	if merge.Upper != "" {
		dir, name := filepath.Split(name)
		if _, err := merge.Stat(filepath.Join(merge.Upper, dir, OpaqueWhiteout+name)); err == nil {
			return nil, &fs.PathError{Op: "readdir", Path: filepath.Clean(name), Err: fs.ErrNotExist}
		}
	}

	toDelete, fileMap := []string{}, make(map[string]fs.DirEntry)
	for _, folderPath := range append(merge.Lower, merge.Upper) {
		folderEntrys, err := os.ReadDir(filepath.Join(folderPath, name))
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		skipCheck := folderPath == merge.Upper
		for _, file := range folderEntrys {
			if skipCheck {
				fileMap[file.Name()] = file
				continue
			} else if strings.HasPrefix(file.Name(), OpaqueWhiteout) {
				toDelete = append(toDelete, file.Name()[len(OpaqueWhiteout):])
				continue
			}
			fileMap[file.Name()] = file
		}
	}
	for _, fileDel := range toDelete {
		delete(fileMap, fileDel)
	}
	return slices.Collect(maps.Values(fileMap)), nil
}

// OpenFile is the generalized open call; most users will use Open or Create instead. It opens the named file with specified flag (O_RDONLY etc.). If the file does not exist, and the O_CREATE flag is passed, it is created with mode perm (before umask). If successful, methods on the returned File can be used for I/O. If there is an error, it will be of type *PathError.
func (merge Overlayfs) OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	if merge.Upper == "" {
		if flag&os.O_CREATE > 0 || flag&os.O_TRUNC > 0 || flag&os.O_RDWR > 0 || flag&os.O_WRONLY > 0 || flag&os.O_APPEND > 0 {
			return nil, &fs.PathError{Op: "open", Path: filepath.Clean(name), Err: fs.ErrPermission}
		}
		for _, folderPath := range merge.Lower {
			file, err := os.OpenFile(filepath.Join(folderPath, name), flag, perm)
			if os.IsNotExist(err) {
				continue
			} else if err != nil && !os.IsNotExist(err) {
				return file, err
			}
			return file, err
		}
		return nil, &fs.PathError{Op: "open", Path: filepath.Clean(name), Err: fs.ErrNotExist}
	}

	dir, name := filepath.Split(name)
	if flag&os.O_CREATE == 0 {
		if _, err := merge.Stat(filepath.Join(merge.Upper, dir, OpaqueWhiteout+name)); err == nil {
			return nil, &fs.PathError{Op: "open", Path: filepath.Clean(name), Err: fs.ErrNotExist}
		}
	}

	// Remove opaque file
	if _, err := merge.Stat(filepath.Join(merge.Upper, dir, OpaqueWhiteout+name)); err == nil {
		os.Remove(filepath.Join(merge.Upper, dir, OpaqueWhiteout+name))
	}

	// Copy and return with flags
	if flag&os.O_APPEND > 0 || flag&os.O_RDWR > 0 || flag&os.O_WRONLY > 0 {
		if _, err := os.Stat(filepath.Join(merge.Upper, name)); os.IsNotExist(err) {
			for _, folder := range merge.Lower {
				f1, err := os.Open(filepath.Join(folder, name))
				if os.IsNotExist(err) {
					continue // Next folder
				} else if err != nil {
					break
				}
				f2, err := os.OpenFile(filepath.Join(merge.Upper, name), os.O_WRONLY|os.O_CREATE, perm)
				if err != nil {
					f1.Close()
					return nil, err
				} else if _, err = io.Copy(f2, f1); err != nil {
					f1.Close()
					f2.Close()
					return nil, err
				}
				f1.Close()
				f2.Close()
			}
		}
		return os.OpenFile(filepath.Join(merge.Upper, name), flag, perm)
	}

	if flag&os.O_CREATE > 0 || flag&os.O_TRUNC > 0 {
		return os.OpenFile(filepath.Join(merge.Upper, name), flag, perm)
	}

	foldersToOpen := append(merge.Lower, merge.Upper)
	slices.Reverse(foldersToOpen)
	for _, folderPath := range foldersToOpen {
		f, err := os.OpenFile(filepath.Join(folderPath, name), flag, perm)
		if os.IsNotExist(err) {
			continue
		}
		return f, err
	}

	return nil, &fs.PathError{Op: "open", Path: filepath.Clean(name), Err: fs.ErrNotExist}
}
