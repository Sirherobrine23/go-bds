package exec

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"sirherobrine23.com.br/go-bds/go-bds/request"
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

// Download ubuntu base to host arch if avaible
func (proc *Proot) DownloadUbuntuRootfs(Version, Arch string) error {
	res, err := (&request.RequestOptions{HttpError: true, Url: fmt.Sprintf("https://cdimage.ubuntu.com/ubuntu-base/releases/%s/release/ubuntu-base-%s-base-%s.tar.gz", Version, Version, Arch)}).Request()
	if err != nil {
		return err
	}
	defer res.Body.Close()

	gz, err := gzip.NewReader(res.Body)
	if err != nil {
		return err
	}
	defer gz.Close()

	os.MkdirAll(proc.Rootfs, 0700)
	tarball := tar.NewReader(gz)
	for {
		head, err := tarball.Next()
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return err
		}
		fullPath := filepath.Join(proc.Rootfs, head.Name)
		fileinfo := head.FileInfo()

		if fileinfo.IsDir() {
			if err := os.MkdirAll(fullPath, fileinfo.Mode()); err != nil {
				return err
			} else if err := os.Chtimes(fullPath, head.AccessTime, head.ModTime); err != nil {
				return err
			}
			continue
		}

		// Create folder if not exist to create file
		os.MkdirAll(filepath.Dir(fullPath), 0666)
		file, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileinfo.Mode())
		if err != nil {
			return err
		} else if err := os.Chtimes(fullPath, head.AccessTime, head.ModTime); err != nil {
			return err
		}

		// Copy file
		if _, err := io.CopyN(file, tarball, fileinfo.Size()); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				continue
			}
			return err
		}
	}
	return nil
}
