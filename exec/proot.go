//go:build android || linux

package exec

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

// Mount rootfs and run command insider in proot
//
// if network not resolve names add nameserver to /etc/resolv.conf (`(echo 'nameserver 1.1.1.1'; echo 'nameserver 8.8.8.8') > /etc/resolv.conf`)
type Proot struct {
	Rootfs string `json:"rootfs"` // Rootfs to mount to run proot
	Qemu   string `json:"qemu"`   // Execute guest programs through QEMU, exp: "qemu-x86_64" or "qemu-x86_64-static"
	*Os
}

// Append dns server to /etc/resolv.conf
func (pr *Proot) AddNameservers(aadrs ...net.IP) error {
	file, err := os.Open(filepath.Join(pr.Rootfs, "etc/resolv.conf"))
	if err != nil {
		return err
	}
	defer file.Close()
	for _, addr := range aadrs {
		if _, err := file.Write([]byte(fmt.Sprintf("nameserver %s\n", addr.String()))); err != nil {
			return err
		}
	}
	return nil
}

// Mount proot and Execute process
func (pr *Proot) Start(options ProcExec) error {
	pr.Os = new(Os)
	var exec ProcExec
	exec.Environment = options.Environment

	// proot -r ./rootfs -q qemu-x86_64 -0 -w / -b /dev -b /proc -b /sys
	exec.Arguments = []string{
		"proot",
		fmt.Sprintf("--rootfs=%s", pr.Rootfs),
		"--bind=/dev",
		"--bind=/proc",
		"--bind=/sys",
		"-0", // Root ID
	}

	if pr.Qemu != "" {
		exec.Arguments = append(exec.Arguments, "-q", pr.Qemu)
	}

	if options.Cwd != "" {
		exec.Arguments = append(exec.Arguments, "-w", options.Cwd)
	}

	exec.Arguments = append(exec.Arguments, options.Arguments...)
	return pr.Os.Start(exec)
}
