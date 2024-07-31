//go:build linux

// For non root user mount in namespace (unshare -rm) or with fuse module
package overleyfs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

var (
	kernelOverlay bool = false
	fuseOverlay   bool = false

	ErrNotOverlayAvaible error = errors.New("overlayfs not avaible")
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

	defer fs.UnixUnmount()
	kernelOverlay = true
	if err := fs.UnixMount(); err != nil {
		kernelOverlay = false
		fuseOverlay = true
		defer fs.FuseUnmount()
		if err := fs.FuseMount(); err != nil {
			fuseOverlay = false
		}
	}

	d1, _ := os.ReadFile(filepath.Join(fs.Target, "test1.txt"))
	d2, _ := os.ReadFile(filepath.Join(fs.Target, "test2.txt"))
	for _, k := range [][]byte{d1, d2} {
		if string(k) != textExample {
			kernelOverlay = false
			fuseOverlay = false
			break
		}
	}
	fs.UnixUnmount()
	fs.FuseUnmount()
}

// Check if avaible overlayfs
func Avaible() bool {
	return kernelOverlay || fuseOverlay
}

func (w *Overlayfs) Mount() error {
	if len(w.Lower) == 0 {
		return fmt.Errorf("set one lower dir")
	} else if w.Workdir == "" && w.Upper != "" {
		return fmt.Errorf("set workdir to user Upperdir")
	} else if kernelOverlay {
		return w.UnixMount()
	} else if fuseOverlay {
		return w.FuseMount()
	}
	return ErrNotOverlayAvaible
}

func (w *Overlayfs) Unmount() error {
	if kernelOverlay {
		return w.UnixUnmount()
	} else if fuseOverlay {
		return w.FuseUnmount()
	}
	return ErrNotOverlayAvaible
}

// Mount overlayfs same `mount -t overlay overlay`
func (w *Overlayfs) UnixMount() error {
	var flags string // Flags to mount overlay
	if w.Workdir != "" && w.Upper != "" {
		flags = fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(w.Lower, ":"), w.Upper, w.Workdir)
	} else {
		flags = "lowerdir=" + strings.Join(w.Lower, ":")
	}
	return unix.Mount("overlay", w.Target, "overlay", 0, flags)
}

// Unmount overlayfs same `unmount`
func (w Overlayfs) UnixUnmount() error {
	return unix.Unmount(w.Target, unix.MNT_DETACH)
}

func (w *Overlayfs) FuseMount() error {
	var flags string // Flags to mount overlay
	if w.Workdir != "" && w.Upper != "" {
		flags = fmt.Sprintf("userxattr,lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(w.Lower, ":"), w.Upper, w.Workdir)
	} else {
		flags = "userxattr,lowerdir=" + strings.Join(w.Lower, ":")
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
