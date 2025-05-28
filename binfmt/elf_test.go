package binfmt

import (
	"archive/tar"
	"bytes"
	"debug/elf"
	"io"
	"net/url"
	"strings"
	"testing"

	"sirherobrine23.com.br/sirherobrine23/go-dpkg/apt"
	"sirherobrine23.com.br/sirherobrine23/go-dpkg/dpkg"
)

var DebianRepository = &apt.AptSource{
	Suites: []string{"sid"},
	URIs: []*url.URL{
		{
			Scheme: "http",
			Host:   "ftp.debian.org",
			Path:   "/debian",
		},
	},
}

func TestCurl(t *testing.T) {
	release, err := DebianRepository.Release()
	if err != nil {
		t.Skipf("cannot get release files: %s", err)
		return
	}
	main := release["sid"]
	DebianRepository.Suites = main.Components

sums:
	for _, sums := range main.MD5 {
		if !sums[0].IsBinaryPackage() {
			continue
		}
		for _, sum := range sums {
			pkgs := sum.Pkgs("sid", DebianRepository)
			for pkgDown, err := range pkgs {
				if err != nil {
					if _, ok := err.(*apt.HttpError); ok {
						break
					}
					t.Error(err)
					break
				}
				if pkgDown.Name == "curl" {
					deb, err := pkgDown.Download(DebianRepository.URIs[0])
					if err != nil {
						return
					}
					defer deb.Close()
					_, tarRead, err := dpkg.NewReader(deb)
					if err != nil {
						return
					}
					defer tarRead.Close()

					var head *tar.Header
					pkgData := tar.NewReader(tarRead)
					for head, err = pkgData.Next(); err == nil; head, err = pkgData.Next() {
						if !strings.HasSuffix(head.Name, "/curl") {
							io.Copy(io.Discard, pkgData)
							continue
						}
						curlBody, err := io.ReadAll(pkgData)
						if err != nil {
							break
						}

						elfFile, err := elf.NewFile(bytes.NewReader(curlBody))
						if err != nil {
							t.Errorf("cannot get curl elf: %s", err)
						}
						newElf := (*Elf)(elfFile)
						space := make([]byte, 12)
						for i := range space {
							space[i] = 0x20
						}
						copy(space, pkgDown.Architecture)
						t.Logf("Dpkg arch: %s Binfmt arch: %s", space, String(newElf))
					}

					tarRead.Close()
					deb.Close()
					continue sums
				}
			}
		}
	}
}
