//go:build linux

package overlayfs

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

type mountPoints []*mountPoint

func (mounts mountPoints) Get(target string) *mountPoint {
	for _, mount := range mounts {
		if mount.Path == target {
			return mount
		}
	}
	return nil
}

func (mounts mountPoints) Exist(target string) bool {
	return mounts.Get(target) != nil
}

type mountPoint struct {
	Device string
	Path   string
	Type   string
	Opts   []string // Opts may contain sensitive mount options (like passwords) and MUST be treated as such (e.g. not logged).
	Freq   int
	Pass   int
}

func parseProcMounts() (mountPoints, error) {
	expectedNumFieldsPerLine := 6 // Number of fields per line in /proc/mounts as per the fstab man page.
	mountProc, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}

	out := []*mountPoint{}
	buff := bufio.NewScanner(mountProc)
	for buff.Scan() {
		line := buff.Text()

		if line == "" {
			// the last split() item is empty string following the last \n
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != expectedNumFieldsPerLine {
			// Do not log line in case it contains sensitive Mount options
			return nil, fmt.Errorf("wrong number of fields (expected %d, got %d)", expectedNumFieldsPerLine, len(fields))
		}

		mp := &mountPoint{
			Device: fields[0],
			Path:   fields[1],
			Type:   fields[2],
			Opts:   strings.Split(fields[3], ","),
		}

		freq, err := strconv.Atoi(fields[4])
		if err != nil {
			return nil, err
		}
		mp.Freq = freq

		pass, err := strconv.Atoi(fields[5])
		if err != nil {
			return nil, err
		}
		mp.Pass = pass

		out = append(out, mp)
	}
	return out, nil
}

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

	for _, folderPath := range append(overlay.Lower, overlay.Target, overlay.Upper, overlay.Workdir) {
		if folderPath == "" {
			continue
		} else if _, err := os.Stat(folderPath); os.IsNotExist(err) {
			if err = os.MkdirAll(folderPath, 0777); err != nil {
				return err
			}
		}
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
	info := time.NewTicker(time.Nanosecond)
	defer info.Stop()
	for range info.C {
		mountedParts, err := parseProcMounts()
		if err != nil {
			return err
		} else if !mountedParts.Exist(overlay.Target) {
			break // Skip ared unmounted
		}

		err = unix.Unmount(overlay.Target, unix.MNT_DETACH)
		// Ignore
		if errors.Is(err, syscall.Errno(22)) {
			err = nil
		}

		// return error if exist
		if err != nil {
			return err
		}
	}
	return nil
}
