//go:build !linux

// Overlayfs not avaible to current platform
package overlayfs

// Current platform not supported to mount Overlayfs or Similar, returning ErrNotOverlayAvaible
func (w *Overlayfs) Mount() error { return ErrNotOverlayAvaible }

// Current platform not supported to unmount Overlayfs or Similar, returning ErrNotOverlayAvaible
func (w *Overlayfs) Unmount() error { return ErrNotOverlayAvaible }
