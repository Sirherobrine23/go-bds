//go:build !(windows || linux)

// Overlayfs not avaible to current platform
package overleyfs

// Current platform not supported to mount Overlayfs or Similar, returning ErrNotOverlayAvaible
func (w *Overlayfs) Mount() error { return ErrNotOverlayAvaible }

// Current platform not supported to unmount Overlayfs or Similar, returning ErrNotOverlayAvaible
func (w *Overlayfs) Unmount() error { return ErrNotOverlayAvaible }