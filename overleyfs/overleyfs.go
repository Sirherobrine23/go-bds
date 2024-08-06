// Mount overlayfs in compatible system
package overleyfs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"sirherobrine23.org/go-bds/go-bds/internal/mergefs"
)

var (
	ErrNotOverlayAvaible error = errors.New("overlayfs not avaible")

	fuseOverlay   bool = false
	kernelOverlay bool = false
)

type MountUnmount interface {
	Mount() error   // Mount volumes in target
	UnMount() error // Unmount target
}

type Overlayfs struct {
	Target  string   // Folder with merged another folder
	Workdir string   // Folder to write temporary files
	Upper   string   // Folder to write modifications, blank to read-only
	Lower   []string // Folders layers, read-only
}

func (w *Overlayfs) makeFlags() (_ string, err error) {
	// overlay on /var/lib/docker/overlay2/5e7aff79cd206c6672c453913df640bf73f075981366fd2c3b81780b5cb776e9/merged
	// workdir=/var/lib/docker/overlay2/5e7aff79cd206c6672c453913df640bf73f075981366fd2c3b81780b5cb776e9/work
	// upperdir=/var/lib/docker/overlay2/5e7aff79cd206c6672c453913df640bf73f075981366fd2c3b81780b5cb776e9/diff
	// lowerdir=/var/lib/docker/overlay2/l/4UKYKDRRHSYV7T6FMWQV7XGOJU
	//          /var/lib/docker/overlay2/l/X4HBSZ4R5V7LFSZYXQ5T7V3Q2Q

	if len(w.Lower) == 0 {
		return "", fmt.Errorf("set one lower dir")
	} else if w.Workdir == "" && w.Upper != "" {
		return "", fmt.Errorf("set workdir to user Upperdir")
	}

	if w.Upper != "" {
		if w.Upper, err = filepath.Abs(w.Upper); err != nil {
			return "", err
		}
	}
	if w.Workdir != "" {
		if w.Workdir, err = filepath.Abs(w.Workdir); err != nil {
			return "", err
		}
	}
	for workIndex := range w.Lower {
		if w.Lower[workIndex], err = filepath.Abs(w.Lower[workIndex]); err != nil {
			return "", err
		}
	}

	var flags string // Flags to mount overlay
	if w.Workdir != "" && w.Upper != "" {
		flags = fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(w.Lower, ":"), w.Upper, w.Workdir)
	} else {
		flags = "lowerdir=" + strings.Join(w.Lower, ":")
	}

	return flags, nil
}

// Get fs.FS from merged Lowers folders
func (w *Overlayfs) GoMerge() (*mergefs.MergedFS, error) {
	var fss []fs.FS
	for _, folderPath := range w.Lower {
		fpath, err := filepath.Abs(folderPath)
		if err != nil {
			return nil, err
		}

		_, err = os.Stat(folderPath)
		if err != nil {
			return nil, fmt.Errorf("mergefs: %s", err.Error())
		}

		fss = append(fss, os.DirFS(fpath))
	}
	return mergefs.Merge(fss...), nil
}
