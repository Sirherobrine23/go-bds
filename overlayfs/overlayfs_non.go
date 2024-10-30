//go:build !(linux || (windows && !winfspexp))

// Overlayfs not avaible to current platform
package overlayfs

// Current platform not supported to mount Overlayfs or Similar, returning ErrNotOverlayAvaible
func (overlay *Overlayfs) Mount() error { return ErrNotOverlayAvaible }

// Current platform not supported to unmount Overlayfs or Similar, returning ErrNotOverlayAvaible
func (overlay *Overlayfs) Unmount() error { return ErrNotOverlayAvaible }
