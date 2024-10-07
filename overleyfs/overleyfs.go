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
	Lower   []string         // Folders layers, read-only
	Upper   string           // Folder to write modifications, blank to read-only
	Target  string           // Destination folder with all Upper and Lower layers merged
	Workdir string           // Folder to write temporary files, only linux required
	FS      *mergefs.Mergefs // Mergefs for internals process

	internalStruct any // Struct to save backend or syscall structs
}

// Get new fs.FS from layers
func (w *Overlayfs) GoMerge() fs.FS {
	return mergefs.NewFS(mergefs.NewMergefsWithTopLayer(w.Upper, w.Lower...))
}
