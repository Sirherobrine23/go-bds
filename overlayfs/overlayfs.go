// Implements Overlayfs/Mergefs mount
//
// for Linux use kernel overlayfs
//
// for Windows use winfsp + go-bds/Overlayfs to mount overlayfs
//
// another platforms return ErrNotOverlayAvaible
package overlayfs

import (
	"errors"
	"io/fs"
)

var (
	ErrNotOverlayAvaible error = errors.New("overlayfs not avaible")                // Current system ou another variables cannot mount merge filesystem or similar
	ErrNoCGOAvaible      error = errors.New("cgo is disabled to process syscall's") // Cannot mount mergefs/overlayfs with cgo disabled
	ErrMounted           error = errors.New("current path is mounted")              // Path target current as mounted
)

type Overlayfs struct {
	Target  string   // Destination folder with all Upper and Lower layers merged
	Upper   string   // Folder to write modifications, blank to read-only
	Lower   []string // Folders layers, read-only
	Workdir string   // Folder to write temporary files, only linux required

	ProcessInternal any
}

// Get new [io/fs.FS] from layers with MergeFS style directly from Golang
func (overlay *Overlayfs) MergeFS() fs.FS { return NewFS(overlay) }

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
