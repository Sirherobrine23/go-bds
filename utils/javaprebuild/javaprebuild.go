package javaprebuild

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/utils/js_types"
)

var ErrSystem error = errors.New("install package manualy")

// Java class file version: https://en.wikipedia.org/wiki/Java_version_history#Release_table
type JavaVersion uint

func (ver JavaVersion) MarshalJSON() ([]byte, error) { return []byte(strconv.Itoa(int(ver))), nil }
func (ver *JavaVersion) UnmarshalJSON(data []byte) error {
	versionInt, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}
	*ver = JavaVersion(versionInt)
	return nil
}

func (ver JavaVersion) MarshalText() (text []byte, err error) { return []byte(ver.String()), nil }
func (ver JavaVersion) String() string {
	switch ver {
	case 0x2D:
		return "JDK 1.0/1.1"
	case 0x2E:
		return "JDK 1.2"
	case 0x2F:
		return "JDK 1.3"
	case 0x30:
		return "JDK 1.4"
	default:
		if ver >= 48 {
			return fmt.Sprintf("Java %d", int(ver)-44)
		}
		return "unknown java version"
	}
}

// Install prebuild java in path
func (ver JavaVersion) InstallLatest(installPath string) error {
	return InstallLatest(uint(ver-44), installPath)
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
		line = strings.ReplaceAll(line, ".", "/")
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
		if _, err := io.ReadFull(mainClassFile, header); err != nil {
			return 0, fmt.Errorf("cannot get class heaader: %s", err)
		}
		return GetMajor(header)
	}
	if err := manifestLine.Err(); err != nil {
		return 0, err
	}
	return 0, fmt.Errorf("cannot get Main-Class")
}

// Get Major Release from class
func GetMajor(header []byte) (JavaVersion, error) {
	if len(header) < 10 {
		return 0, fmt.Errorf("cannot parse class header, size: %d", len(header))
	} else if !bytes.Equal(header[:4], []byte{0xCA, 0xFE, 0xBA, 0xBE}) {
		return 0, fmt.Errorf("invalid header class magic: %s", header[:4])
	}
	return JavaVersion(binary.BigEndian.Uint16(header[6:])), nil
}
