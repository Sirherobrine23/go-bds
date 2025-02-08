//go:build windows && (amd64 || 386 || arm64)

package overlayfs

import (
	"os"
	"path/filepath"

	"github.com/aegistudio/go-winfsp"
	"github.com/aegistudio/go-winfsp/gofs"
)

type winfspMerge struct{ *Overlayfs }

func (winfsp winfspMerge) OpenFile(name string, flag int, perm os.FileMode) (gofs.File, error) {
	return winfsp.Overlayfs.OpenFile(name, flag, perm)
}

func (fsp *Overlayfs) getWinfsp() (*winfsp.FileSystem, error) {
	if fssp, ok := fsp.ProcessInternal.(*winfsp.FileSystem); ok {
		return fssp, nil
	}
	return nil, ErrUnMounted
}

// Mount go-mergefs with winfsp
//
// This function supports drive letters (X:) or directories as mount points:
//
//   - Drive letters: Refer to the documentation of the DefineDosDevice Windows API to better understand how they are created.
//
//   - Directories: They can be used as mount points for disk based file systems. They cannot be used for network file systems. This is a limitation that Windows imposes on junctions.
func (overlay *Overlayfs) Mount() error {
	_, err := overlay.getWinfsp()
	if err != nil && err != ErrUnMounted {
		return err
	}

	// Clean folder path to mount
	if !filepath.IsAbs(overlay.Target) {
		if overlay.Target, err = filepath.Abs(overlay.Target); err != nil {
			return err
		}
	}

	overlay.ProcessInternal, err = winfsp.Mount(gofs.New(&winfspMerge{Overlayfs: overlay}), overlay.Target, winfsp.FileSystemName("go-mergefs"))
	return err
}

// Umount winfsp filesystem if mounted
func (overlay *Overlayfs) Unmount() error {
	fssp, err := overlay.getWinfsp()
	if err != nil {
		return err
	}
	fssp.Unmount()
	overlay.ProcessInternal = nil
	return nil
}
