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
)

var (
	ErrNotOverlayAvaible error = errors.New("overlayfs not avaible")                // Current system ou another variables cannot mount merge filesystem or similar
	ErrNoCGOAvaible      error = errors.New("cgo is disabled to process syscall's") // Cannot mount mergefs/overlayfs with cgo disabled
	ErrMounted           error = errors.New("current path is mounted")              // Path target current as mounted
	ErrUnmounted         error = errors.New("path not mounted")
)

// Implements a linux Overlayfs in Golang
//
// We maintain some so -like om of [the] but with
// modifications only in their defined structures,
// If you do not have a upper folder any writing will be fully blocked
//
//
//  * Notes to, Mount(), Unmount() this functions to mount mergefs or overlayfs in kernel/system level if avaible
type Overlayfs struct {
	Target  string   // Destination folder with all Upper and Lower layers merged
	Upper   string   // Folder to write modifications, blank to read-only
	Lower   []string // Folders layers, read-only
	Workdir string   // Folder to write temporary files, only linux required

	ProcessInternal any // Save any information to Mount and Unmount
}

// Return new Overlayfs
//
// Crate new *Overlayfs with values, examples:
//
//	NewOverlayFS("/Root", "/data", "/workdir", "/low1", "/low2") // To read-write target
//	NewOverlayFS("/Root", "", "", "/low1", "/low2")              // To read-only target
func NewOverlayFS(TargetFolder, TopLayer, WorkdirFolder string, LowLayers ...string) *Overlayfs {
	return &Overlayfs{
		Target:  TargetFolder,
		Lower:   LowLayers,
		Upper:   TopLayer,
		Workdir: WorkdirFolder,
	}
}
