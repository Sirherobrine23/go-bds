// Mount overlayfs in compatible system
package overleyfs

import (
	"errors"
	"io/fs"

	"sirherobrine23.com.br/go-bds/go-bds/overleyfs/mergefs"
)

var (
	ErrNotOverlayAvaible error = errors.New("overlayfs not avaible")
	ErrNoCGOAvaible      error = errors.New("cgo is disabled to process syscall's")
)

type Overlayfs struct {
	Target  string           // Folder with merged another folder
	Workdir string           // Folder to write temporary files
	Upper   string           // Folder to write modifications, blank to read-only
	Lower   []string         // Folders layers, read-only
	FS      *mergefs.Mergefs // Mergefs for internals process

	internalStruct any // Struct to save backend or syscall structs
}

// Get new fs.FS from layers
func (w *Overlayfs) GoMerge() fs.FS {
	return mergefs.NewFS(mergefs.NewMergefsWithTopLayer(w.Upper, w.Lower...))
}
