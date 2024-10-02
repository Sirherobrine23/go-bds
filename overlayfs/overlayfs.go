package overlayfs

import (
	"errors"
	"io/fs"

	"sirherobrine23.com.br/go-bds/go-bds/overlayfs/mergefs"
)

var (
	ErrNotOverlayAvaible error = errors.New("overlayfs not avaible")
	ErrNoCGOAvaible      error = errors.New("cgo is disabled to process syscall's")
)

type Overlayfs struct {
	Target  string   // Destination folder with all Upper and Lower layers merged
	Upper   string   // Folder to write modifications, blank to read-only
	Lower   []string // Folders layers, read-only
	Workdir string   // Folder to write temporary files, only linux required
}

// Return new Overlayfs
//
// Crate new *Overlayfs with values, examples:
//
// NewOverlayFS("/Root", "/data", "/workdir", "/low1", "/low2") // To read-write target
//
// NewOverlayFS("/Root", "", "", "/low1", "/low2") // To read-only target
func NewOverlayFS(TargetFolder, TopLayer, Workdir string, LowLayers ...string) *Overlayfs {
	return &Overlayfs{
		Target:  TargetFolder,
		Lower:   LowLayers,
		Upper:   TopLayer,
		Workdir: Workdir,
	}
}

// Get new [io/fs.FS] from layers with MergeFS style directly from Golang
func (w Overlayfs) MergeFS() fs.FS {
	return mergefs.NewFS(mergefs.NewMergefs(w.Upper, w.Lower...))
}
