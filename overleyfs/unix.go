//go:build darwin || linux

package overleyfs

import (
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/unix"
)



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

	fuseOverlay = true
	defer fs.FuseUnmount()
	if err := fs.FuseMount(); err != nil {
		fuseOverlay = false
		return
	}
	fs.FuseUnmount()
}

// Attemp mount with fuse
func (w *Overlayfs) FuseMount() error {
	flags, err := w.makeFlags()
	if err != nil {
		return err
	}
	return exec.Command("fuse-overlayfs", "-o", flags, w.Target).Run()
}

// Try unmount fuse overlay with `fusermount3` or `fusermount`
func (w Overlayfs) FuseUnmount() error {
	var unmounted bool
	// Attempt to unmount the FUSE mount using either fusermount or fusermount3.
	// If they fail, fallback to unix.Unmount
	for _, v := range []string{"fusermount3", "fusermount"} {
		err := exec.Command(v, "-u", w.Target).Run()
		if err == nil {
			unmounted = true
			break
		}
	}
	// If fusermount|fusermount3 failed to unmount the FUSE file system, make sure all
	// pending changes are propagated to the file system
	if !unmounted {
		fd, err := unix.Open(w.Target, unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
		if err == nil {
			unix.Close(fd)
		}
	}
	return nil
}
