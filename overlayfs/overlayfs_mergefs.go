package overlayfs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var ErrNotImplemented = errors.New("not implemented")

const (
	OpaqueWhiteout  string = ".wh."                                 // File to ignore
	fileCreateFlags        = os.O_CREATE | os.O_TRUNC | os.O_WRONLY // Flags to create file, if exists replace data
)

type flages int

func (f flages) Flags() (flags []int) {
	if int(f)&os.O_RDONLY != 0 {
		flags = append(flags, os.O_RDONLY)
	}
	if int(f)&os.O_WRONLY != 0 {
		flags = append(flags, os.O_WRONLY)
	}
	if int(f)&os.O_RDWR != 0 {
		flags = append(flags, os.O_RDWR)
	}
	if int(f)&os.O_APPEND != 0 {
		flags = append(flags, os.O_APPEND)
	}
	if int(f)&os.O_CREATE != 0 {
		flags = append(flags, os.O_CREATE)
	}
	if int(f)&os.O_EXCL != 0 {
		flags = append(flags, os.O_EXCL)
	}
	if int(f)&os.O_SYNC != 0 {
		flags = append(flags, os.O_SYNC)
	}
	if int(f)&os.O_TRUNC != 0 {
		flags = append(flags, os.O_TRUNC)
	}
	return
}

// Check flag have write permissions
func (f flages) IsRead() bool {
	flags := f.Flags()
	return slices.Contains(flags, os.O_RDWR) ||
		slices.Contains(flags, os.O_RDONLY)
}

// Check flag have write permissions
func (f flages) IsWrite() bool {
	flags := f.Flags()
	return slices.Contains(flags, os.O_RDWR) ||
		slices.Contains(flags, os.O_WRONLY) ||
		slices.Contains(flags, os.O_CREATE) ||
		slices.Contains(flags, os.O_TRUNC)
}

func (f flages) CreateIfNotExist() bool {
	flags := f.Flags()
	return slices.Contains(flags, os.O_CREATE)
}

func rewriteName(name string, err error) error {
	if err == nil {
		return err
	} else if e, ok := err.(*fs.PathError); ok {
		e.Path = name
	} else if e, ok := err.(*os.LinkError); ok {
		e.Old = name
	}
	return err
}

// Check if rw overlayfs
func (over Overlayfs) isRW() bool { return over.Upper != "" }

type fileLayer struct {
	Layer     string      // layer Path
	Stat      fs.FileInfo // File info
	FromUpper bool        // is file present on Upper dir
}

//	func (over Overlayfs) lowLayers() []string {
//		layers := slices.Clone(over.Lower)
//		slices.Reverse(layers)
//		return layers
//	}
func (over Overlayfs) allLayers() []string {
	layers := append(slices.Clone(over.Lower), over.Upper)
	slices.Reverse(layers)
	return slices.DeleteFunc(layers, func(content string) bool { return strings.TrimSpace(content) == "" })
}

func (over Overlayfs) retrieveFileFromLayers(name string) ([]fileLayer, error) {
	content, layers := []fileLayer{}, over.allLayers()
	for _, layer := range layers {
		if layer == "" {
			continue
		}
		stat, err := os.Stat(filepath.Join(layer, name))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, rewriteName(name, err)
		}
		content = append(content, fileLayer{
			FromUpper: layer == over.Upper,
			Layer:     layer,
			Stat:      stat,
		})
	}
	return content, nil
}

func toMarkedToDeleteFilename(name string) string {
	dir, name := filepath.Split(name)
	return filepath.Join(dir, OpaqueWhiteout+name)
}

// Remove opaque file if exists
func (over Overlayfs) removeIfDeleted(name string) error {
	if over.isRW() {
		dir, name := filepath.Split(name)
		name = filepath.Join(over.Upper, dir, OpaqueWhiteout+name)
		if _, err := os.Stat(name); err == nil {
			return os.Remove(name)
		}
	}
	return nil
}

func (over Overlayfs) makeFileDeleted(name string) error {
	if over.isRW() {
		fullPath := filepath.Join(over.Upper, toMarkedToDeleteFilename(name))
		if _, err := os.Stat(filepath.Dir(fullPath)); err != nil {
			if err = os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return err
			}
		}
		return os.WriteFile(fullPath, nil, 0666)
	}
	return nil
}

// Check if file marked deleted
func (over Overlayfs) isFileMarkedDeleted(name string) (bool, error) {
	if over.isRW() && !(name == "." || name == "") {
		name = toMarkedToDeleteFilename(name)
		dir, name := filepath.Split(name)
		if dir != "" {
			if _, err := os.Stat(filepath.Join(over.Upper, dir)); err != nil && errors.Is(err, fs.ErrNotExist) {
				return over.isFileMarkedDeleted(dir) // recursive down check if deleted
			}
		}

		if _, err := os.Stat(filepath.Join(over.Upper, name)); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				err = nil
			}
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// Clone from lower layer to upper to modifications
func (over Overlayfs) copyFileToUpperIfAbsent(name string) error {
	files, err := over.retrieveFileFromLayers(name)
	if err != nil {
		return rewriteName(name, err)
	}

	// Locate layers
	fromUpperIndex, fromFistLowIndex := slices.IndexFunc(files, func(entry fileLayer) bool { return entry.FromUpper }), slices.IndexFunc(files, func(entry fileLayer) bool { return !entry.FromUpper })
	if (fromUpperIndex == -1 && fromFistLowIndex == -1) || fromUpperIndex != -1 {
		return nil // Ignore copy
	} else if fromFistLowIndex == -1 {
		return &fs.PathError{Op: "copy", Path: name, Err: fs.ErrNotExist}
	}

	// File stats
	fileInfo := files[fromFistLowIndex]
	lowPath, uperPath := filepath.Join(fileInfo.Layer, name), filepath.Join(over.Upper, name)

	// Remove opaque file
	if err := over.removeIfDeleted(name); err != nil {
		return err
	}

	return rewriteName(name, over.copyFromTo(fileInfo, name, lowPath, uperPath))
}

func (over Overlayfs) copyFromTo(fileInfo fileLayer, name, lowPath, uperPath string) error {
	switch fileInfo.Stat.Mode().Type() {
	case fs.ModeAppend, fs.ModeExclusive, fs.ModeTemporary, fs.ModeDevice, fs.ModeNamedPipe, fs.ModeSocket, fs.ModeSetuid, fs.ModeSetgid, fs.ModeCharDevice, fs.ModeSticky, fs.ModeIrregular:
		return &fs.PathError{Op: "copy", Path: name, Err: fs.ErrInvalid}
	case fs.ModeDir:
		return os.CopyFS(uperPath, os.DirFS(lowPath)) // Copy dir full dir
	case fs.ModeSymlink:
		target, err := os.Readlink(filepath.Join(lowPath))
		if err == nil {
			err = os.Symlink(target, uperPath)
		}
		return rewriteName(name, err)
	default:
		// Open file
		fromReadFile, err := os.Open(lowPath)
		if err != nil {
			return rewriteName(name, err)
		}
		defer fromReadFile.Close() // Close file

		// Create target file
		targetWriteFile, err := os.OpenFile(filepath.Join(over.Upper, name), fileCreateFlags, fileInfo.Stat.Mode().Perm())
		if err != nil {
			return rewriteName(name, err)
		}
		defer targetWriteFile.Close()

		// Copy file
		_, err = io.Copy(targetWriteFile, fromReadFile)
		return rewriteName(name, err)
	}
}
