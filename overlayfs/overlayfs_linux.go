//go:build linux

// For non root user mount in namespace (unshare -rm)
package overlayfs

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

// Mount overlayfs same `mount -t overlay overlay`:
//
//   - The working directory (Workdir) needs to be an empty directory on the same filesystem as the Upper directory.
func (overlay *Overlayfs) Mount() error {
	// overlay on /var/lib/docker/overlay2/5e7aff79cd206c6672c453913df640bf73f075981366fd2c3b81780b5cb776e9/merged
	//    workdir=/var/lib/docker/overlay2/5e7aff79cd206c6672c453913df640bf73f075981366fd2c3b81780b5cb776e9/work
	//   upperdir=/var/lib/docker/overlay2/5e7aff79cd206c6672c453913df640bf73f075981366fd2c3b81780b5cb776e9/diff
	//  lowerdir=/var/lib/docker/overlay2/l/4UKYKDRRHSYV7T6FMWQV7XGOJU
	//           /var/lib/docker/overlay2/l/X4HBSZ4R5V7LFSZYXQ5T7V3Q2Q
	if len(overlay.Lower) == 0 {
		return fmt.Errorf("set one lower dir")
	} else if overlay.Workdir == "" && overlay.Upper != "" {
		return fmt.Errorf("set workdir to user Upperdir")
	}

	// Flags to mount overlay
	flags := "lowerdir=" + strings.Join(overlay.Lower, ":")
	if overlay.Workdir != "" && overlay.Upper != "" {
		flags = fmt.Sprintf("upperdir=%s,workdir=%s,lowerdir=%s", overlay.Upper, overlay.Workdir, strings.Join(overlay.Lower, ":"))
	}
	err := unix.Mount("overlay", overlay.Target, "overlay", 0, flags)
	if errors.Is(err, syscall.Errno(1)) {
		err = fs.ErrPermission
	}
	return err
}

// Unmount overlayfs same `unmount`
func (overlay *Overlayfs) Unmount() error {
	return unix.Unmount(overlay.Target, unix.MNT_DETACH)
}
