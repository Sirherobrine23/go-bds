//go:build linux

package overleyfs

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	kernelOverlay bool = false
	fuseOverlay   bool = false
)

func init() {
	kernelOverlay = true
	if err := exec.Command("modprobe", "overlay").Run(); err != nil {
		kernelOverlay = false
		fuseOverlay = true
		if err := exec.Command("fuse-overlayfs", "-h").Run(); err != nil {
			fuseOverlay = false
			return
		}
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

	defaultFlags := []string{fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", down, up, workdir), merged}
	if kernelOverlay {
		if err := exec.Command("mount", append([]string{"-t", "overlay", "overlay", "-o"}, defaultFlags...)...).Run(); err != nil {
			kernelOverlay = false
			return
		}
	} else if fuseOverlay {
		if err := exec.Command("fuse-overlayfs", append([]string{"-o"}, defaultFlags...)...).Run(); err != nil {
			fuseOverlay = false
			return
		}
	} else {
		return
	}
	defer exec.Command("umount", merged)
	file, _ := os.Create(filepath.Join(merged, "google.txt"))
	file.Write([]byte("ok google"))
	file.Close()

	file, err = os.Create(filepath.Join(up, "google.txt"))
	if err != nil {
		return
	}
	if buff, _ := io.ReadAll(file); string(buff) != "ok google" {
		kernelOverlay = false
		fuseOverlay = false
	}
	file.Close()
}

func (w *Overlayfs) Mount() error {
	if len(w.Lower) == 0 {
		return fmt.Errorf("set one lower dir")
	}
	defaultFlags := []string{fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(w.Lower, ":"), w.Upper, w.Workdir), w.Root}
	if w.Workdir == "" && w.Upper == "" {
		defaultFlags[0] = "lowerdir=" + strings.Join(w.Lower, ":")
	} else if w.Upper == "" && w.Workdir != "" {
		defaultFlags[0] = fmt.Sprintf("lowerdir=%s,workdir=%s", strings.Join(w.Lower, ":"), w.Workdir)
	}
	if kernelOverlay {
		if err := exec.Command("mount", append([]string{"-t", "overlay", "overlay", "-o"}, defaultFlags...)...).Run(); err != nil {
			return err
		}
	} else if fuseOverlay {
		if err := exec.Command("fuse-overlayfs", append([]string{"-o"}, defaultFlags...)...).Run(); err != nil {
			return err
		}
	}
	return nil
}