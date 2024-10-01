//go:build windows

// Create Virtual filesystem with Windows ProjFS and merged folders
package overleyfs

import (
	"github.com/balazsgrill/potatodrive/win/projfs"
)

// Stop Windows Project Filesystem virtualization
func (w *Overlayfs) Unmount() error { return nil }

func (w *Overlayfs) Mount() error {
	_ = projfs.PRJ_CALLBACKS{}
	return nil
}
