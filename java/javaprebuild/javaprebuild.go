package javaprebuild

import (
	"archive/zip"
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/utils/js_types"
)

var (
	ErrSystem error = errors.New("install package manualy")

	javaHeaderNames = map[uint]string{0x30: "JDK 1.4", 0x2F: "JDK 1.3", 0x2E: "JDK 1.2", 0x2D: "JDK 1.1"}
)

type JavaVersion uint

func (ver JavaVersion) MarshalText() (text []byte, err error) { return []byte(ver.String()), nil }
func (ver JavaVersion) String() string {
	if str, ok := javaHeaderNames[uint(ver)]; ok {
		return str
	} else if ver >= 25 {
		return "Java" + strconv.Itoa(int(ver-44))
	}
	return "unknown java version"
}

// Install prebuild java in path
func (ver JavaVersion) InstallLatest(installPath string) error {
	return InstallLatest(uint(ver), installPath)
}

// Return jar file java version
func JarMajor(r io.ReaderAt, size int64) (JavaVersion, error) {
	jarZip, err := zip.NewReader(r, size)
	if err != nil {
		return 0, err
	}
	fs := js_types.Slice[*zip.File](jarZip.File)

	manifest := fs.Find(func(file *zip.File) bool {
		return strings.Contains(file.Name, "META-INF") && strings.Contains(file.Name, "MANIFEST.MF")
	})

	manifestRead, err := manifest.Open()
	if err != nil {
		return 0, fmt.Errorf("cannot get META-INF/MANIFEST.MF: %s", err)
	}
	defer manifestRead.Close()

	manifestLine := bufio.NewScanner(manifestRead)
	for manifestLine.Scan() {
		line := manifestLine.Text()
		if !strings.HasPrefix(line, "Main-Class:") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "Main-Class:"))
		if !strings.HasSuffix(line, ".class") {
			line += ".class"
		}

		mainClass := fs.Find(func(file *zip.File) bool { return file.Name == line })
		if mainClass == nil {
			return 0, fmt.Errorf("cannot get main-class (%q)", line)
		}

		mainClassFile, err := mainClass.Open()
		if err != nil {
			return 0, fmt.Errorf("cannot open main-class file: %s", err)
		}
		defer mainClassFile.Close()

		header := make([]byte, 10)
		if n, err := io.ReadFull(mainClassFile, header); err != nil {
			return 0, fmt.Errorf("cannot get class heaader: %s", err)
		} else if n < 8 {
			return 0, fmt.Errorf("cannot get class header, invalid return offset")
		}

		return JavaVersion(binary.BigEndian.Uint16(header[6:])), nil
	}
	if err := manifestLine.Err(); err != nil {
		return 0, err
	}
	return 0, fmt.Errorf("cannot get Main-Class")
}
