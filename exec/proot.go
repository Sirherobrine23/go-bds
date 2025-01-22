package exec

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"runtime"

	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
	"sirherobrine23.com.br/go-bds/go-bds/semver"
)

var (
	_ Proc = &Proot{}

	ErrNoExtractUbuntu error = errors.New("cannot extract Ubuntu base image")
)

// Mount rootfs and run command insider in proot
//
// if network not resolve names add nameserver to /etc/resolv.conf (`(echo 'nameserver 1.1.1.1'; echo 'nameserver 8.8.8.8') > /etc/resolv.conf`)
type Proot struct {
	Rootfs   string              // Rootfs to mount to run proot
	Qemu     string              // Execute guest programs through QEMU, exp: "qemu-x86_64" or "qemu-x86_64-static"
	GID, UID uint                // User and Group ID, default is root
	Binds    map[string][]string // Bind mount directories, example: "/dev": {"/dev", "/root/dev"} => /dev -> /root/dev and /dev
	Os                           // Extends from Os struct
}

// Append dns server to /etc/resolv.conf
//
//	Example: Proot.Proot(netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("1.1.1.1"))
func (pr Proot) AddNameservers(aadrs ...netip.Addr) error {
	file, err := os.OpenFile(filepath.Join(pr.Rootfs, "etc/resolv.conf"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Fprint(file, "\n")
	for _, addr := range aadrs {
		if _, err := fmt.Fprintf(file, "nameserver %s\n", addr.String()); err != nil {
			return err
		}
	}
	return nil
}

// Mount proot and Execute process
func (pr *Proot) Start(options ProcExec) error {
	exec := ProcExec{
		Environment: options.Environment,
		// proot -r ./rootfs -q qemu-x86_64 -0 -w / -b /dev -b /proc -b /sys
		Arguments: []string{
			"proot",
			"-r", pr.Rootfs,
			"-b", "/dev",
			"-b", "/proc",
			"-b", "/sys",
		},
	}
	for src, dsts := range pr.Binds {
		for _, dst := range dsts {
			exec.Arguments = append(exec.Arguments, "-b", fmt.Sprintf("%s:%s", src, dst))
		}
	}
	if pr.GID != 0 || pr.UID != 0 {
		exec.Arguments = append(exec.Arguments, "-i", fmt.Sprintf("%d:%d", pr.UID, pr.GID))
	} else {
		exec.Arguments = append(exec.Arguments, "-0")
	}
	if pr.Qemu != "" {
		exec.Arguments = append(exec.Arguments, "-q", pr.Qemu)
	}
	if options.Cwd != "" {
		exec.Arguments = append(exec.Arguments, "-w", options.Cwd)
	}
	exec.Arguments = append(exec.Arguments, options.Arguments...)
	if runtime.GOOS == "android" {
		if exec.Environment == nil {
			exec.Environment = Env{}
		}
		exec.Environment["LD_PRELOAD"] = "" // Remove to termux
	}

	return pr.Os.Start(exec)
}

type ubuntuSeries struct {
	TotalSize int            `json:"total_size,omitempty"`
	Entries   []ubuntuEntrie `json:"entries,omitempty"`
}

type ubuntuEntrie struct {
	Name          string `json:"name,omitempty"`
	Version       string `json:"version,omitempty"`
	Status        string `json:"status,omitempty"`
	Supported     bool   `json:"supported,omitempty"`
	Architectures string `json:"architectures_collection_link,omitempty"`
}

func (entry ubuntuEntrie) SemverVersion() *semver.Version { return semver.New(entry.Version) }

type ubuntuArch struct {
	Enabled   bool   `json:"enabled"`
	ChrootURL string `json:"chroot_url"`
	Arch      string `json:"architecture_tag"`
}

type ubuntuArchitectures struct {
	TotalSize int          `json:"total_size"`
	Entries   []ubuntuArch `json:"entries"`
}

// Download ubuntu base to host arch if avaible
//
//	example: Proot.DownloadUbuntuRootfs("", "") // Latest version to current arch
//	example: Proot.DownloadUbuntuRootfs("24.10", "amd64")
//	example: Proot.DownloadUbuntuRootfs("24.10", "arm64")
//	example: Proot.DownloadUbuntuRootfs("24.10", "riscv64")
func (proc Proot) DownloadUbuntuRootfs(Version, Arch string) error {
	data, _, err := request.JSON[ubuntuSeries]("https://api.launchpad.net/devel/ubuntu/series", nil)
	if err != nil {
		return errors.Join(ErrNoExtractUbuntu, err)
	}
	semver.SortStruct(data.Entries)
	versionSelect := ubuntuEntrie{}
	if Version == "" {
		versionSelect = data.Entries[len(data.Entries)-1]
	}

	// ArchitecturesCollectionLink
	archs, _, err := request.JSON[ubuntuArchitectures](versionSelect.Architectures, nil)
	if err != nil {
		return errors.Join(ErrNoExtractUbuntu, err)
	}

	// Select Arch
	selectedArch := ubuntuArch{}
	if Arch == "" {
		switch runtime.GOARCH {
		case "386":
			Arch = "i386"
		default:
			Arch = runtime.GOARCH
		}
	}

	// Check if exists
	for _, arch := range archs.Entries {
		if arch.Arch == Arch {
			selectedArch = arch
			break
		}
	}

	// Extract to rootfs
	if selectedArch.ChrootURL != "" {
		extract := request.ExtractOptions{
			Cwd:   proc.Rootfs,
			Strip: 1,
		}
		return request.Tar(selectedArch.ChrootURL, extract, nil)
	}
	return ErrNoExtractUbuntu
}
