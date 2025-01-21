package exec

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/request/gohtml"
	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
	"sirherobrine23.com.br/go-bds/go-bds/semver"
)

var _ Proc = &Proot{}

// Mount rootfs and run command insider in proot
//
// if network not resolve names add nameserver to /etc/resolv.conf (`(echo 'nameserver 1.1.1.1'; echo 'nameserver 8.8.8.8') > /etc/resolv.conf`)
type Proot struct {
	Rootfs string // Rootfs to mount to run proot
	Qemu   string // Execute guest programs through QEMU, exp: "qemu-x86_64" or "qemu-x86_64-static"
	*Os
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
	pr.Os = &Os{}

	exec := ProcExec{
		Environment: options.Environment,
		// proot -r ./rootfs -q qemu-x86_64 -0 -w / -b /dev -b /proc -b /sys
		Arguments: []string{
			"proot",
			fmt.Sprintf("--rootfs=%s", pr.Rootfs),
			"--bind=/dev",
			"--bind=/proc",
			"--bind=/sys",
			"-0", // Root ID
		},
	}

	if exec.Environment == nil {
		exec.Environment = Env{}
	}
	exec.Environment["LD_PRELOAD"] = ""

	if pr.Qemu != "" {
		exec.Arguments = append(exec.Arguments, "-q", pr.Qemu)
	}

	if options.Cwd != "" {
		exec.Arguments = append(exec.Arguments, "-w", options.Cwd)
	}

	exec.Arguments = append(exec.Arguments, options.Arguments...)
	return pr.Os.Start(exec)
}

// Download ubuntu base to host arch if avaible
//
//	example: Proot.DownloadUbuntuRootfs("24.10", "amd64")
//	example: Proot.DownloadUbuntuRootfs("24.10", "arm64")
//	example: Proot.DownloadUbuntuRootfs("24.10", "riscv64")
func (proc *Proot) DownloadUbuntuRootfs(Version, Arch string) error {
	if Version == "" {
		res, err := request.Request("https://cdimage.ubuntu.com/ubuntu-base/releases", nil)
		if err != nil {
			return errors.Join(err, errors.New("cannot gt ubuntu images"))
		}
		defer res.Body.Close()

		type UbuntuVersions struct {
			Versions []struct {
				Version string `html:"a"`
			} `html:"body > ul > li"`
		}

		var vers UbuntuVersions
		if err := gohtml.NewParse(res.Body, &vers); err != nil {
			return err
		}

		versions := []*semver.Version{}
		for _, data := range vers.Versions {
			if strings.Contains(data.Version, ".") {
				versions = append(versions, semver.New(data.Version[:len(data.Version)-1]))
			}
		}
		semver.Sort(versions)
		Version = versions[len(versions)-1].String()
	}

	if Arch == "" {
		switch runtime.GOARCH {
		case "386":
			Arch = "i386"
		default:
			Arch = runtime.GOARCH
		}
	}

	baseSystemUrl := fmt.Sprintf("https://cdimage.ubuntu.com/ubuntu-base/releases/%s/release/ubuntu-base-%s-base-%s.tar.gz", Version, Version, Arch)
	return request.Tar(baseSystemUrl, request.ExtractOptions{Cwd: proc.Rootfs}, nil)
}
