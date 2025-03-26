//go:build !(linux || (windows && (amd64 || 386 || arm64)))

package overlayfs

func OverlayfsAvaible() bool { return false }

// Current platform not supported to mount Overlayfs or Similar, returning ErrNotOverlayAvaible
func (overlay *Overlayfs) Mount() error { return ErrNotOverlayAvaible }

// Current platform not supported to unmount Overlayfs or Similar, returning ErrNotOverlayAvaible
func (overlay *Overlayfs) Unmount() error { return ErrNotOverlayAvaible }
