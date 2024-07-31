//go:build linux

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
	kernelOverlay bool = true
	fuseOverlay   bool = false

	ErrNotOverlayAvaible error = errors.New("overlayfs not avaible")
)

func init() {
	fuseOverlay = true
	if path, err := exec.LookPath("fuse-overlayfs"); err != nil || path == "" {
		fuseOverlay = false
	}

	root, err := os.MkdirTemp(os.TempDir(), "overlay_test_*")
	if err != nil {
		return
	}
	up := filepath.Join(root, "upper")
	down := filepath.Join(root, "down")
	workdir := filepath.Join(root, "workdir")
	merged := filepath.Join(root, "merged")

	for _, f := range []string{up, down, workdir, merged} {
		if os.MkdirAll(f, 0666) != nil {
			return
		}
		defer os.RemoveAll(f)
	}

	flags := fmt.Sprintf("userxattr,lowerdir=%s,upperdir=%s,workdir=%s", down, up, workdir)
	if err := unix.Mount("overlay", merged, "overlay", 0, flags); err != nil {
		kernelOverlay = false
		if !fuseOverlay {
			return
		} else if err := exec.Command("fuse-overlayfs", "-o", flags, merged).Run(); err != nil {
			fuseOverlay = false
			return
		} else {
			var unmounted bool
			// Attempt to unmount the FUSE mount using either fusermount or fusermount3.
			// If they fail, fallback to unix.Unmount
			for _, v := range []string{"fusermount3", "fusermount"} {
				err := exec.Command(v, "-u", merged).Run()
				if err == nil {
					unmounted = true
					break
				}
			}
			// If fusermount|fusermount3 failed to unmount the FUSE file system, make sure all
			// pending changes are propagated to the file system
			if !unmounted {
				fd, err := unix.Open(merged, unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
				if err == nil {
					unix.Close(fd)
				}
			}
		}
	} else {
		fuseOverlay = false
		unix.Unmount(merged, unix.MNT_DETACH)
	}
}

func (w *Overlayfs) Unmount() error {
	if fuseOverlay {
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
	return unix.Unmount(w.Target, unix.MNT_DETACH)
}

func (w *Overlayfs) Mount() error {
	if len(w.Lower) == 0 {
		return fmt.Errorf("set one lower dir")
	}

	if w.Workdir == "" && w.Upper != "" {
		return fmt.Errorf("set workdir to user Upperdir")
	}

	var flags string // Flags to mount overlay
	if w.Workdir != "" && w.Upper != "" {
		flags = fmt.Sprintf("userxattr,lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(w.Lower, ":"), w.Upper, w.Workdir)
	} else {
		flags = "userxattr,lowerdir=" + strings.Join(w.Lower, ":")
	}

	if kernelOverlay {
		return unix.Mount("overlay", w.Target, "overlay", 0, flags)
	} else if fuseOverlay {
		return exec.Command("fuse-overlayfs", "-o", flags, w.Target).Run()
	}
	return ErrNotOverlayAvaible
}
