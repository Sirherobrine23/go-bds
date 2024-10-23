// Merge Folders same Linux Overlayfs directly from Golang
package mergefs

import (
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const SuffixToDeletedFile string = "__mergfsdelete_"

type Mergefs struct {
	TopLayer    string
	LowerLayers []string
}

// Return new Mergefs struct with TopLayer
func NewMergefs(top string, layers ...string) *Mergefs { return &Mergefs{top, layers} }

func (merge Mergefs) checkForDeleted(name string) (string, error) {
	if merge.TopLayer == "" {
		return "", nil
	}
	for name != "" {
		var base string
		name, base = filepath.Split(name)
		if _, err := os.Stat(filepath.Join(merge.TopLayer, name, SuffixToDeletedFile+base)); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		return filepath.Join(merge.TopLayer, name, SuffixToDeletedFile+base), nil
	}
	return "", nil
}

// Is Read And Write
func (merge Mergefs) Exists(name string) bool {
	_, err := merge.Stat(name)
	return err == nil
}

// ReadDir reads the named directory, returning all its directory entries. If an error occurs reading the directory, ReadDir returns the entries it was able to read before the error, along with the error.
func (merge Mergefs) ReadDir(name string) ([]fs.DirEntry, error) {
	name = filepath.Clean(name)
	deleteFile, err := merge.checkForDeleted(name)
	if err != nil {
		return nil, err
	} else if deleteFile != "" {
		return nil, &os.PathError{Err: fs.ErrNotExist, Path: name, Op: "mergefs"}
	}

	mapFile := make(map[string]fs.DirEntry)
	notExistsLen := len(append(merge.LowerLayers, merge.TopLayer))
	for _, folderPath := range append(merge.LowerLayers, merge.TopLayer) {
		if folderPath == "" {
			notExistsLen--
			continue
		}
		entrys, err := os.ReadDir(filepath.Join(folderPath, name))
		if err != nil {
			if os.IsNotExist(err) {
				notExistsLen--
				continue
			}
			return nil, err
		}
		for _, fsEntry := range entrys {
			if _, err := os.Stat(filepath.Join(folderPath, name, SuffixToDeletedFile+fsEntry.Name())); strings.HasSuffix(fsEntry.Name(), SuffixToDeletedFile) || !os.IsNotExist(err) {
				delete(mapFile, fsEntry.Name()[len(SuffixToDeletedFile):]) // Delete entry if exists
				continue
			}
			mapFile[fsEntry.Name()] = fsEntry // Set new entry
		}
	}
	if notExistsLen == 0 {
		return nil, &os.PathError{Err: fs.ErrNotExist, Path: name, Op: "mergefs"}
	}
	return slices.Collect(maps.Values(mapFile)), nil
}

// Mkdir creates a new directory with the specified name and permission bits (before umask). If there is an error, it will be of type *PathError.
func (merge Mergefs) Mkdir(name string, perm fs.FileMode) error {
	name = filepath.Clean(name)
	if merge.TopLayer == "" {
		return &os.PathError{Err: fs.ErrPermission, Op: "mergefs", Path: name}
	}
	if deleteFile, _ := merge.checkForDeleted(name); deleteFile != "" {
		os.RemoveAll(deleteFile)
	}
	return os.Mkdir(filepath.Join(merge.TopLayer, name), perm)
}

// MkdirAll creates a directory named path, along with any necessary parents, and returns nil, or else returns an error. The permission bits perm (before umask) are used for all directories that MkdirAll creates. If path is already a directory, MkdirAll does nothing and returns nil.
func (merge Mergefs) MkdirAll(name string, perm fs.FileMode) error {
	name = filepath.Clean(name)
	if merge.TopLayer == "" {
		return &os.PathError{Err: fs.ErrPermission, Op: "mergefs", Path: name}
	}
	if deleteFile, _ := merge.checkForDeleted(name); deleteFile != "" {
		os.RemoveAll(deleteFile)
	}
	return os.MkdirAll(filepath.Join(merge.TopLayer, name), perm)
}

// Stat returns a [FileInfo] describing the named file. If there is an error, it will be of type [*PathError].
func (merge Mergefs) Stat(name string) (os.FileInfo, error) {
	if name == "" || name == "." {
		return nil, &os.PathError{Err: fs.ErrNotExist, Path: name, Op: "mergefs"}
	} else if deleteFile, _ := merge.checkForDeleted(name); deleteFile != "" {
		return nil, &os.PathError{Err: fs.ErrNotExist, Path: name, Op: "mergefs"}
	}

	name = filepath.Clean(name)
	DirPath, FileName := filepath.Split(name)
	files, err := merge.ReadDir(DirPath)
	if err != nil {
		return nil, err
	}
	for _, kn := range files {
		if kn.Name() == FileName {
			return kn.Info()
		}
	}
	return nil, &os.PathError{Err: fs.ErrNotExist, Path: name, Op: "mergefs"}
}

// Create creates or truncates the named file. If the file already exists, it is truncated. If the file does not exist, it is created with mode 0o666 (before umask). If successful, methods on the returned File can be used for I/O; the associated file descriptor has mode O_RDWR. If there is an error, it will be of type *PathError.
func (merge Mergefs) Create(name string) (*os.File, error) {
	name = filepath.Clean(name)
	if merge.TopLayer == "" {
		return nil, &os.PathError{Err: fs.ErrPermission, Op: "mergefs", Path: name}
	}
	Dir, FileName := filepath.Split(name)
	if _, err := os.Stat(filepath.Join(merge.TopLayer, Dir, SuffixToDeletedFile+FileName)); err == nil {
		os.Remove(filepath.Join(merge.TopLayer, Dir, SuffixToDeletedFile+FileName))
	}
	return os.Create(filepath.Join(merge.TopLayer, name))
}

// OpenFile is the generalized open call; most users will use Open or Create instead. It opens the named file with specified flag (O_RDONLY etc.). If the file does not exist, and the O_CREATE flag is passed, it is created with mode perm (before umask). If successful, methods on the returned File can be used for I/O. If there is an error, it will be of type *PathError.
func (merge Mergefs) OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	name = filepath.Clean(name)
	if !(flag&os.O_CREATE == 0 || flag&os.O_APPEND == 0) {
		if merge.TopLayer == "" {
			return nil, &os.PathError{Err: fs.ErrPermission, Op: "mergefs", Path: name}
		}
		Dir, FileName := filepath.Split(name)
		if _, err := os.Stat(filepath.Join(merge.TopLayer, Dir, SuffixToDeletedFile+FileName)); err == nil {
			os.Remove(filepath.Join(merge.TopLayer, Dir, SuffixToDeletedFile+FileName))
		}

		// Copy file or folder from last layer
		if !(flag&os.O_APPEND == 0 || flag&os.O_RDWR == 0 || flag&os.O_WRONLY == 0) {
			if _, err := os.Stat(filepath.Join(merge.TopLayer, name)); os.IsNotExist(err) {
				// Copy layer to new location
				if _, err := os.Stat(filepath.Join(merge.TopLayer, name)); os.IsNotExist(err) {
					ns, err := fs.Sub(NewFS(&merge), name)
					if err != nil {
						return nil, err
					}
					if err = os.CopyFS(filepath.Join(merge.TopLayer, name), ns); err != nil {
						return nil, err
					}
				}
			}
		}

		return os.OpenFile(filepath.Join(merge.TopLayer, name), flag, perm)
	}

	if merge.TopLayer != "" {
		Dir, FileName := filepath.Split(name)
		if _, err := os.Stat(filepath.Join(merge.TopLayer, Dir, SuffixToDeletedFile+FileName)); err == nil {
			return nil, &os.PathError{Err: fs.ErrNotExist, Path: name, Op: "mergefs"}
		}

	}

	foldersTargets := append(merge.LowerLayers, merge.TopLayer)
	slices.Reverse(foldersTargets)
	for _, folderPath := range foldersTargets {
		f, err := os.OpenFile(filepath.Join(folderPath, name), flag, perm)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		return f, nil
	}
	return nil, &os.PathError{Err: fs.ErrNotExist, Path: name, Op: "mergefs"}
}

// Open opens the named file for reading. If successful, methods on the returned file can be used for reading; the associated file descriptor has mode O_RDONLY. If there is an error, it will be of type *PathError.
func (merge Mergefs) Open(name string) (*os.File, error) {
	return merge.OpenFile(name, os.O_RDONLY, 0)
}

// Remove removes the named file or directory. If there is an error, it will be of type *PathError.
func (merge Mergefs) Remove(name string) error {
	name = filepath.Clean(name)
	if merge.TopLayer == "" {
		return &os.PathError{Err: fs.ErrPermission, Op: "mergefs", Path: name}
	} else if _, err := merge.Stat(filepath.Join(merge.TopLayer, name)); err == nil {
		if err = os.Remove(filepath.Join(merge.TopLayer, name)); err != nil {
			return err
		}
	}
	Dir, FileName := filepath.Split(name)
	deleteFile, err := os.Create(filepath.Join(merge.TopLayer, Dir, SuffixToDeletedFile+FileName))
	if err != nil {
		return err
	}
	return deleteFile.Close()
}

// RemoveAll removes path and any children it contains. It removes everything it can but returns the first error it encounters. If the path does not exist, RemoveAll returns nil (no error). If there is an error, it will be of type [*PathError].
func (merge Mergefs) RemoveAll(name string) error {
	name = filepath.Clean(name)
	if merge.TopLayer == "" {
		return &os.PathError{Err: fs.ErrPermission, Op: "mergefs", Path: name}
	} else if _, err := merge.Stat(filepath.Join(merge.TopLayer, name)); err == nil {
		if err = os.RemoveAll(filepath.Join(merge.TopLayer, name)); err != nil {
			return err
		}
	}
	Dir, FileName := filepath.Split(name)
	deleteFile, err := os.Create(filepath.Join(merge.TopLayer, Dir, SuffixToDeletedFile+FileName))
	if err != nil {
		return err
	}
	return deleteFile.Close()
}

// WriteFile writes data to the named file, creating it if necessary. If the file does not exist, WriteFile creates it with permissions perm (before umask); otherwise WriteFile truncates it before writing, without changing permissions. Since WriteFile requires multiple system calls to complete, a failure mid-operation can leave the file in a partially written state.
func (merge Mergefs) WriteFile(name string, data []byte, perm fs.FileMode) error {
	f, err := merge.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// ReadFile reads the named file and returns the contents. A successful call returns err == nil, not err == EOF. Because ReadFile reads the whole file, it does not treat an EOF from Read as an error to be reported.
func (merge Mergefs) ReadFile(name string) ([]byte, error) {
	f, err := merge.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// Link creates newname as a hard link to the oldname file. If there is an error, it will be of type *LinkError.
func (merge Mergefs) Link(oldname string, newname string) error {
	if merge.TopLayer == "" {
		return &os.PathError{Err: fs.ErrPermission, Op: "mergefs"}
	}

	// Copy layer to new location
	if _, err := os.Stat(filepath.Join(merge.TopLayer, oldname)); os.IsNotExist(err) {
		ns, err := fs.Sub(NewFS(&merge), oldname)
		if err != nil {
			return err
		} else if err = os.CopyFS(filepath.Join(merge.TopLayer, oldname), ns); err != nil {
			return err
		}
	}

	// Create symlink in Top layer
	return os.Link(filepath.Join(merge.TopLayer, oldname), filepath.Join(merge.TopLayer, newname))
}

// Symlink creates newname as a symbolic link to oldname. On Windows, a symlink to a non-existent oldname creates a file symlink; if oldname is later created as a directory the symlink will not work. If there is an error, it will be of type *LinkError.
func (merge Mergefs) Symlink(oldname string, newname string) error {
	if merge.TopLayer == "" {
		return &os.PathError{Err: fs.ErrPermission, Op: "mergefs"}
	}

	// Copy layer to new location
	if _, err := os.Stat(filepath.Join(merge.TopLayer, oldname)); os.IsNotExist(err) {
		ns, err := fs.Sub(NewFS(&merge), oldname)
		if err != nil {
			return err
		} else if err = os.CopyFS(filepath.Join(merge.TopLayer, oldname), ns); err != nil {
			return err
		}
	}

	// Create symlink in Top layer
	return os.Symlink(filepath.Join(merge.TopLayer, oldname), filepath.Join(merge.TopLayer, newname))
}

func (merge Mergefs) Rename(oldpath string, newpath string) error {
	if merge.TopLayer == "" {
		return &os.PathError{Err: fs.ErrPermission, Op: "mergefs"}
	}
	return os.Rename(filepath.Join(merge.TopLayer, oldpath), filepath.Join(merge.TopLayer, newpath))
}
