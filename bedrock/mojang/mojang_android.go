//go:build linux || android

package mojang

import "sirherobrine23.org/go-bds/go-bds/exec"

// Run minecraft bedrock insider proot to run with rootfs
type MojangProot struct {
	VersionsFolder string        // Folder with versions
	Version        string        // Version to run server
	Path           string        // Run server at folder
	Config         *MojangConfig // Server config

	ServerProc exec.Proc
}