//go:build linux || android

package mojang

import (
	"fmt"
	"io"
	"net"
	"os"
	osExec "os/exec"
	"path/filepath"
	"runtime"
	"slices"

	"sirherobrine23.org/go-bds/go-bds/exec"
	"sirherobrine23.org/go-bds/go-bds/overleyfs"
)

var UbuntuVersion string = "24.04" // Ubuntu base version, default is latest lts

// Run minecraft bedrock insider proot to run with rootfs
type MojangProot struct {
	VersionsFolder string // Folder with versions
	Rootfs         string // Rootfs to run in proot
	Version        string // Version to run server
	Path           string // Run server at folder

	Handler *Handlers     // Server handlers
	Config  *MojangConfig // Server config

	ServerProc exec.Proc
}

func (server *MojangProot) Start() error {
	if server.Rootfs == "" {
		return fmt.Errorf("set rootfs")
	}

	if server.Version == "latest" || server.Version == "" {
		versions, err := FromVersions()
		if err != nil {
			return fmt.Errorf("cannot get versions: %s", err.Error())
		}
		server.Version = GetLatest(versions)
	}

	versionRoot := filepath.Join(server.VersionsFolder, server.Version)
	if checkExist(versionRoot) {
		n, err := os.ReadDir(versionRoot)
		if err != nil {
			return err
		}
		if len(n) == 0 {
			if err := os.RemoveAll(versionRoot); err != nil {
				return err
			}
		}
	}

	if !checkExist(versionRoot) {
		versions, err := FromVersions()
		if err != nil {
			return fmt.Errorf("cannot get versions: %s", err.Error())
		}
		os.MkdirAll(versionRoot, 0700)

		var ok bool
		var version Version
		var target VersionPlatform
		if version, ok = versions[server.Version]; !ok {
			return fmt.Errorf("version not found in database")
		} else if target, ok = version.Platforms[fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)]; !ok {
			return fmt.Errorf("platform not supported")
		} else if err := target.Download(versionRoot); err != nil {
			return err
		}
	}

	// Check if rootfs contains bash or sh, if not download ubuntu base and install curl
	if !slices.ContainsFunc([]string{filepath.Join(server.Rootfs, "bin/sh"), filepath.Join(server.Rootfs, "bin/bash"), filepath.Join(server.Rootfs, "usr/bin/sh"), filepath.Join(server.Rootfs, "usr/bin/bash")}, checkExist) {
		var proc = new(exec.Proot)
		if err := proc.DownloadUbuntuRootfs(UbuntuVersion); err != nil {
			return err
		} else if err := proc.AddNameservers(net.ParseIP("1.1.1.1"), net.ParseIP("2606:4700:4700::1111"), net.ParseIP("8.8.8.8"), net.ParseIP("2001:4860:4860::8888")); err != nil {
			return err
		} else if err := proc.Start(exec.ProcExec{Environment: map[string]string{"DEBIAN_FRONTEND": "noninteractive"}, Arguments: []string{"bash", "-c", "apt-get update && apt install -y curl"}}); err != nil {
			return err
		} else if err := proc.Wait(); err != nil {
			return err
		}
	}

	server.ServerProc = &exec.Proot{Rootfs: server.Rootfs}
	if runtime.GOARCH != "amd64" {
		var qemuEmus = []string{"qemu-x86_64", "qemu-x86_64-static"}
		for _, qemu := range qemuEmus {
			if ppath, _ := osExec.LookPath(qemu); len(ppath) > 0 {
				server.ServerProc.(*exec.Proot).Qemu = qemu // Add qemu amd64 emulation to proot, is run if rootfs is amd64 arch
				break
			}
		}
	}

	// Merge new version
	rootfsServer := filepath.Join(server.Rootfs, server.Path)
	os.RemoveAll(rootfsServer + "old")
	if err := os.Rename(rootfsServer, rootfsServer+"old"); err != nil {
		return err
	}
	fs, err := (&overleyfs.Overlayfs{Lower: []string{versionRoot, rootfsServer + "old"}}).GoMerge()
	if err != nil {
		return err
	}
	os.MkdirAll(rootfsServer, 0600)
	if err := copyToDisk(fs, ".", rootfsServer); err != nil {
		return err
	}

	recopy := []string{"server.properties", "allowlist.json", "permissions.json"}
	for _, file := range recopy {
		oldPath, newPath := filepath.Join(rootfsServer+"old", file), filepath.Join(rootfsServer, file)
		oldFile, err := os.Open(oldPath)
		if err != nil {
			return err
		}
		defer oldFile.Close()
		newFile, err := os.Create(newPath)
		if err != nil {
			return err
		}
		defer newFile.Close()
		if _, err := io.Copy(newFile, oldFile); err != nil {
			return err
		}
	}

	if err := server.ServerProc.Start(exec.ProcExec{Cwd: server.Path, Arguments: []string{"./bedrock_server"}, Environment: map[string]string{"LD_LIBRARY_PATH": "."}}); err != nil {
		return err
	}

	// Handler parse
	if server.Handler != nil {
		log, err := server.ServerProc.StdoutFork()
		if err == nil {
			go server.Handler.RegisterScan(log)
		}
	}

	return nil
}
