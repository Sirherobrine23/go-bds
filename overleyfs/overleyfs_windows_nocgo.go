//go:build windows && !cgo

// Create Virtual filesystem with Windows ProjFS and merged folders
package overleyfs

// Current platform not supported to mount Overlayfs or Similar, returning ErrNoCGOAvaible
func (w *Overlayfs) Mount() error { return ErrNoCGOAvaible }

// Current platform not supported to unmount Overlayfs or Similar, returning ErrNoCGOAvaible
func (w *Overlayfs) Unmount() error { return ErrNoCGOAvaible }