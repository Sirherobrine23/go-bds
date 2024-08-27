//go:build linux

// For non root user mount in namespace (unshare -rm)
package overleyfs

import (
	"golang.org/x/sys/unix"
)

// Mount overlayfs same `mount -t overlay overlay`:
//
//   - The working directory (Workdir) needs to be an empty directory on the same filesystem as the Upper directory.
func (w *Overlayfs) Mount() error {
	flags, err := w.makeFlags()
	if err != nil {
		return err
	}
	return unix.Mount("overlay", w.Target, "overlay", 0, flags)
}

// Unmount overlayfs same `unmount`
func (w *Overlayfs) Unmount() error {
	return unix.Unmount(w.Target, unix.MNT_DETACH)
}
