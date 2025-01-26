//go:build windows && (amd64 || 386 || arm64)

package overlayfs

import (
	"os"
	"path/filepath"

	"github.com/aegistudio/go-winfsp"
	"github.com/aegistudio/go-winfsp/gofs"
)

type winfspMerge struct{ *Overlayfs }

func (winfsp *winfspMerge) OpenFile(name string, flag int, perm os.FileMode) (gofs.File, error) {
	return winfsp.Overlayfs.OpenFile(name, flag, perm)
}

// Mount go-mergefs with winfsp
func (overlay *Overlayfs) Mount() error {
	if _, ok := overlay.ProcessInternal.(*winfsp.FileSystem); ok {
		return ErrMounted
	} else if !filepath.IsAbs(overlay.Target) {
		return os.ErrInvalid
	}

	sys, err := winfsp.Mount(gofs.New(&winfspMerge{Overlayfs: overlay}), overlay.Target, winfsp.FileSystemName("go-mergefs"))
	if err != nil {
		return err
	}
	overlay.ProcessInternal = sys
	return nil
}

// Umount winfsp filesystem if mounted ok
func (overlay *Overlayfs) Unmount() error {
	if sys, ok := overlay.ProcessInternal.(*winfsp.FileSystem); ok {
		sys.Unmount()
		overlay.ProcessInternal = nil // Remove winfsp filesystem from struct
	}

	return nil
}
