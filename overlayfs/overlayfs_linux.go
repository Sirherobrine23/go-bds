//go:build linux

// For non root user mount in namespace (unshare -rm)
package overlayfs

import (
	"fmt"
	"path/filepath"
	"strings"

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

	var err error
	if overlay.Upper != "" {
		if overlay.Upper, err = filepath.Abs(overlay.Upper); err != nil {
			return err
		}
	}
	if overlay.Workdir != "" {
		if overlay.Workdir, err = filepath.Abs(overlay.Workdir); err != nil {
			return err
		}
	}
	for workIndex := range overlay.Lower {
		if overlay.Lower[workIndex], err = filepath.Abs(overlay.Lower[workIndex]); err != nil {
			return err
		}
	}

	var flags string // Flags to mount overlay
	if overlay.Workdir != "" && overlay.Upper != "" {
		flags = fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(overlay.Lower, ":"), overlay.Upper, overlay.Workdir)
	} else {
		flags = "lowerdir=" + strings.Join(overlay.Lower, ":")
	}
	return unix.Mount("overlay", overlay.Target, "overlay", 0, flags)
}

// Unmount overlayfs same `unmount`
func (overlay *Overlayfs) Unmount() error {
	return unix.Unmount(overlay.Target, unix.MNT_DETACH)
}
