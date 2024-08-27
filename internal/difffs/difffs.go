package difffs

import (
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

type Diff struct {
	Lowers []string // Layer of lowerdir
}

func (overlay Diff) ReadDir(fpath string) ([]fs.DirEntry, error) {
	overlayEntrys := make(map[string]fs.DirEntry)
	for _, layerLow := range overlay.Lowers {
		if _, err := os.Stat(filepath.Join(layerLow, fpath)); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		files, err := os.ReadDir(filepath.Join(layerLow, fpath))
		if err != nil {
			return nil, err
		}
		for _, entry := range files {
			overlayEntrys[entry.Name()] = entry
		}
	}

	return slices.Collect(maps.Values(overlayEntrys)), nil
}

func (overlay Diff) Stat(fpath string) (fs.FileInfo, error) {
	var f string
	for _, layerLow := range overlay.Lowers {
		if _, err := os.Stat(filepath.Join(layerLow, fpath)); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		f = filepath.Join(layerLow, fpath)
	}
	if f == "" {
		return nil, fs.ErrNotExist
	}
	return os.Stat(f)
}

func (overlay Diff) Open(fpath string) (fs.File, error) {
	var f string
	for _, layerLow := range overlay.Lowers {
		if _, err := os.Stat(filepath.Join(layerLow, fpath)); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		f = filepath.Join(layerLow, fpath)
	}
	if f == "" {
		return nil, fs.ErrNotExist
	}
	return os.Open(f)
}
