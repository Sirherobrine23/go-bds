//go:build linux

package overleyfs

import (
	"errors"
	"fmt"
	"io"
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
	if err := exec.Command("fuse-overlayfs", "-h").Run(); err != nil {
		fuseOverlay = false
		return
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
		if os.MkdirAll(f, os.FileMode(7777)) != nil {
			return
		}
		defer os.RemoveAll(f)
	}

	flags := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", down, up, workdir)
	if kernelOverlay {
		if err := unix.Mount("overlay", merged, "overlay", 0, flags); err != nil {
			kernelOverlay = false
			return
		}
	} else if fuseOverlay {
		if err := exec.Command("fuse-overlayfs", "-o", flags, merged).Run(); err != nil {
			fuseOverlay = false
			return
		}
	} else {
		return
	}
	file, _ := os.Create(filepath.Join(merged, "google.txt"))
	file.Write([]byte("ok google"))
	file.Close()

	file, err = os.Create(filepath.Join(up, "google.txt"))
	if err != nil {
		unix.Unmount(merged, unix.MNT_DETACH)
		return
	}
	if buff, _ := io.ReadAll(file); string(buff) != "ok google" {
		kernelOverlay = false
		fuseOverlay = false
	}
	file.Close()
	unix.Unmount(merged, unix.MNT_DETACH)
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
		flags = fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(w.Lower, ":"), w.Upper, w.Workdir)
	} else {
		flags = "lowerdir=" + strings.Join(w.Lower, ":")
	}

	if kernelOverlay {
		return unix.Mount("overlay", w.Target, "overlay", 0, flags)
	} else if fuseOverlay {
		flags += ",userxattr"
		return exec.Command("fuse-overlayfs", "-o", flags, w.Target).Run()
	}
	return ErrNotOverlayAvaible
}
