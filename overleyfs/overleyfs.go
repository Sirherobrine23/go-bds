// Mount overlayfs in compatible system
package overleyfs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"sirherobrine23.com.br/go-bds/go-bds/internal/mergefs"
)

var ErrNotOverlayAvaible error = errors.New("overlayfs not avaible")

type Overlayfs struct {
	Target  string   // Folder with merged another folder
	Workdir string   // Folder to write temporary files
	Upper   string   // Folder to write modifications, blank to read-only
	Lower   []string // Folders layers, read-only

	internalStruct any // Struct to save backend or syscall structs
}

// Get fs.FS from merged Lowers folders
func (w *Overlayfs) GoMerge() (fs.FS, error) {
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
