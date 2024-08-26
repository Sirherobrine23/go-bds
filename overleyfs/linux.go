//go:build linux

// For non root user mount in namespace (unshare -rm)
package overleyfs

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

var kernelOverlay bool = false

func init() {
	root, err := os.MkdirTemp(os.TempDir(), "overlay_test_*")
	if err != nil {
		return
	}
	defer os.RemoveAll(root)
	var fs Overlayfs

	fs.Workdir = filepath.Join(root, "workdir")
	fs.Target = filepath.Join(root, "merged")
	fs.Upper = filepath.Join(root, "upper")
	fs.Lower = []string{filepath.Join(root, "down")}
	for _, k := range []string{fs.Workdir, fs.Target, fs.Upper, fs.Lower[0]} {
		os.MkdirAll(k, 0600)
	}

	textExample := "google is best\n"
	os.WriteFile(filepath.Join(fs.Lower[0], "test1.txt"), []byte(textExample), 0600)
	os.WriteFile(filepath.Join(fs.Upper, "test2.txt"), []byte(textExample), 0600)

	defer fs.Unmount()
	kernelOverlay = true
	if err := fs.Mount(); err != nil {
		kernelOverlay = false
		return
	}

	d1, _ := os.ReadFile(filepath.Join(fs.Target, "test1.txt"))
	d2, _ := os.ReadFile(filepath.Join(fs.Target, "test2.txt"))
	for _, k := range [][]byte{d1, d2} {
		if string(k) != textExample {
			kernelOverlay = false
			break
		}
	}
	fs.Unmount()
}

// Mount overlayfs same `mount -t overlay overlay`:
//
//   - The working directory (Workdir) needs to be an empty directory on the same filesystem as the Upper directory.
func (w *Overlayfs) Mount() error {
	if kernelOverlay {
		flags, err := w.makeFlags()
		if err != nil {
			return err
		}
		return unix.Mount("overlay", w.Target, "overlay", 0, flags)
	}
	return ErrNotOverlayAvaible
}

// Unmount overlayfs same `unmount`
func (w *Overlayfs) Unmount() error {
	if kernelOverlay {
		return unix.Unmount(w.Target, unix.MNT_DETACH)
	}
	return ErrNotOverlayAvaible
}
