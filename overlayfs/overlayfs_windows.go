//go:build windows && (amd64 || 386 || arm64)

// Mount virtual filesystem with winfsp
package overlayfs

import (
	"os"
	"path/filepath"

	"github.com/aegistudio/go-winfsp"
	"github.com/aegistudio/go-winfsp/gofs"

	"sirherobrine23.com.br/go-bds/go-bds/overlayfs/mergefs"
)

type winfspMerge struct{ *mergefs.Mergefs }

func (winfsp *winfspMerge) OpenFile(name string, flag int, perm os.FileMode) (gofs.File, error) {
	return winfsp.Mergefs.OpenFile(name, flag, perm)
}

// Mount go-mergefs with winfsp
func (w *Overlayfs) Mount() error {
	if _, ok := w.internal.(*winfsp.FileSystem); ok {
		return ErrMounted
	} else if !filepath.IsAbs(w.Target) {
		return os.ErrInvalid
	}

	sys, err := winfsp.Mount(gofs.New(&winfspMerge{mergefs.NewMergefs(w.Upper, w.Lower...)}), w.Target, winfsp.FileSystemName("go-mergefs"))
	if err != nil {
		return err
	}
	w.internal = sys
	return nil
}

// Umount winfsp filesystem if mounted ok
func (w *Overlayfs) Unmount() error {
	if sys, ok := w.internal.(*winfsp.FileSystem); ok {
		sys.Unmount()
		w.internal = nil // Remove winfsp filesystem from struct
	}

	return nil
}
