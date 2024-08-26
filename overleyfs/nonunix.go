//go:build !linux

package overleyfs

func (w *Overlayfs) Mount() error {
	return ErrNotOverlayAvaible
}

func (w *Overlayfs) Unmount() error {
	return ErrNotOverlayAvaible
}
