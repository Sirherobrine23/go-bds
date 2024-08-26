//go:build linux || android

// Check if qemu or box64/box86 enabled in binfmt or ared exists in PATH's
package binfmt

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var lineSplit = regexp.MustCompile(`(\s|\t)+`)

var qemuArchs = map[string]struct{ Magic, Mask []byte }{
	"arm":     {[]byte("\x7fELF\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x28\x00"), []byte("\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff")},
	"aarch64": {[]byte("\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\xb7\x00"), []byte("\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff")},
	"ppc":     {[]byte("\x7fELF\x01\x02\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x14"), []byte("\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff")},
	"ppc64":   {[]byte("\x7fELF\x02\x02\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x15"), []byte("\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff")},
	"riscv32": {[]byte("\x7fELF\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\xf3\x00"), []byte("\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff")},
	"riscv64": {[]byte("\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\xf3\x00"), []byte("\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff")},
	"i386":    {[]byte("\x7fELF\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x03\x00"), []byte("\xff\xff\xff\xff\xff\xfe\xfe\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff")},
	"x86_64":  {[]byte("\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x3e\x00"), []byte("\xff\xff\xff\xff\xff\xfe\xfe\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff")},
}

func preapend[T any](input []T, rest ...T) []T {
	return append(rest, input...)
}

// List archs registred
func Archs() ([]*Fmt, error) {
	var arch []*Fmt
	entrys, err := os.ReadDir("/proc/sys/fs/binfmt_misc")
	if err != nil {
		if !(os.IsNotExist(err) || os.IsPermission(err)) {
			return nil, err
		}
		entrys = []fs.DirEntry{}
	}

	for _, file := range entrys {
		fileName := file.Name()
		if strings.HasPrefix(fileName, "qemu-") {
			qemuFmt, err := os.OpenFile(filepath.Join("/proc/sys/fs/binfmt_misc", fileName), os.O_RDONLY, 0)
			if err != nil {
				continue
			}
			defer qemuFmt.Close()

			var bin Fmt
			b := bufio.NewScanner(qemuFmt)
			for b.Scan() {
				line := lineSplit.Split(b.Text(), 2)

				switch strings.TrimSpace(line[0]) {
				case "enabled":
					bin.AutoEmulate = true
				case "interpreter":
					bin.Interpreter = strings.TrimSpace(line[1])
				case "flags:":
					bin.Flags = []string{}
					if line[1] != "" {
						for _, letter := range strings.Split(line[1], "") {
							switch strings.ToUpper(letter) {
							case "F", "f":
								bin.Flags = append(bin.Flags, "fix-binary")
							case "C", "c":
								bin.Flags = append(bin.Flags, "credentials")
							case "O", "o":
								bin.Flags = append(bin.Flags, "open-binary")
							case "P", "p":
								bin.Flags = append(bin.Flags, "preserve-argv[0]")
							default:
								bin.Flags = append(bin.Flags, letter)
							}
						}
					}
				case "offset":
					if _, err := fmt.Sscan(strings.TrimSpace(line[1]), &bin.Offset); err != nil {
						return arch, err
					}
				case "magic":
					if bin.Magic, err = hex.DecodeString(strings.TrimSpace(line[1])); err != nil {
						return arch, err
					}
				case "mask":
					if bin.Mask, err = hex.DecodeString(strings.TrimSpace(line[1])); err != nil {
						return arch, err
					}
				}
			}
			bin.Arch = strings.SplitN(fileName[5:], "-", 2)[0]
			arch = append(arch, &bin)
		}
	}

	if len(arch) == 0 {
		for Arch, qemuArch := range qemuArchs {
			var qemuPath string
			if qemuPath, _ = exec.LookPath(fmt.Sprintf("qemu-%s-static", Arch)); len(qemuPath) == 0 {
				if qemuPath, _ = exec.LookPath(fmt.Sprintf("qemu-%s", Arch)); len(qemuPath) == 0 {
					continue
				}
			}
			arch = append(arch, &Fmt{Arch: Arch, Interpreter: qemuPath, Magic: qemuArch.Magic, Mask: qemuArch.Mask})
		}

		// Add termux glibc repository bin
		if _, err := os.Stat(filepath.Join(os.Getenv("PREFIX"), "glibc/bin")); !os.IsNotExist(err) {
			if !strings.Contains(os.Getenv("PATH"), "glibc/bin") {
				os.Setenv("PATH", strings.Join([]string{filepath.Join(os.Getenv("PREFIX"), "glibc/bin"), os.Getenv("PATH")}, ":"))
			}
		}
	}

	// Box64 emulater
	if box86, _ := exec.LookPath("box86"); len(box86) > 0 {
		arch = preapend(arch, &Fmt{
			Arch:        "i386",
			Interpreter: box86,
			Magic:       qemuArchs["i386"].Magic,
			Mask:        qemuArchs["i386"].Mask,
		})
	}
	if box64, _ := exec.LookPath("box64"); len(box64) > 0 {
		arch = preapend(arch, &Fmt{
			Arch:        "x86_64",
			Interpreter: box64,
			Magic:       qemuArchs["x86_64"].Magic,
			Mask:        qemuArchs["x86_64"].Mask,
		})
	}

	return arch, nil
}
