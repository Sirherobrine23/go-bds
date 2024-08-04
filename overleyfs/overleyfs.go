// Mount overlayfs in compatible system
package overleyfs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"sirherobrine23.org/go-bds/go-bds/internal/mergefs"
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

// Get go fs.FS from merged Upper and Lowers folders
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
