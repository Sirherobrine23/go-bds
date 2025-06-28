// Prebuild java binarys from adoptium or liberica
package javaprebuild

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/utils/file_checker"
)

var (
	ErrSystem             = errors.New("install package manualy")
	ErrInvaliClassHeader  = errors.New("class magic is invalid")
	ErrInvaliMajorVersion = errors.New("class header major version is invalid")
)

// Java class file version: https://en.wikipedia.org/wiki/Java_version_history#Release_table
type JavaVersion uint16

func (ver JavaVersion) MarshalText() (text []byte, err error) {
	if ver <= 44 {
		return nil, fmt.Errorf("invalid jar class version")
	}

	switch v := int(ver) - 44; v {
	case 1:
		return []byte("JDK 1.0/1.1"), nil
	case 2:
		return []byte("JDK 1.2"), nil
	case 3:
		return []byte("JDK 1.3"), nil
	case 4:
		return []byte("JDK 1.4"), nil
	default:
		return fmt.Appendf(nil, "Java %d", v), nil
	}
}

func (ver *JavaVersion) UnmarshalText(data []byte) error {
	switch v := string(data); v {
	case "JDK 1.0/1.1":
		*ver = 0x2D
	case "JDK 1.2":
		*ver = 0x2E
	case "JDK 1.3":
		*ver = 0x2F
	case "JDK 1.4":
		*ver = 0x30
	default:
		switch {
		case strings.HasPrefix(v, "Java "):
			version, err := strconv.Atoi(v[5:])
			if err != nil {
				return err
			}
			*ver = JavaVersion(version + 44)
		case strings.Count(v, ".") >= 2:
			verSplit := strings.SplitN(v, ".", 3)
			switch v := verSplit[0]; v {
			case "1":
				version, err := strconv.Atoi(verSplit[1])
				if err != nil {
					return err
				} else if version >= 0 && version <= 11 {
					if version == 0 {
						version = 1
					}
					*ver = JavaVersion(version + 44)
					return nil
				}
			default:
				version, err := strconv.Atoi(verSplit[0])
				if err != nil {
					return err
				} else if version > 11 {
					*ver = JavaVersion(version + 44)
					return nil
				}
			}
		}
		return fmt.Errorf("invalid java version: %s", v)
	}
	return nil
}

func (ver JavaVersion) String() string {
	if verStr, err := ver.MarshalText(); err == nil {
		return string(verStr)
	}
	return "Unknown java version"
}

// Install java if local version not satisfact version and return java bin path
func (ver JavaVersion) Install(installPath string) (string, error) {
	if javaBin, localVersion, err := LocalVersion(); err == nil {
		// If version is same or more return local
		if localVersion >= ver {
			return javaBin, nil
		}
	}

	binName := "bin/java"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	binPath, err := file_checker.FindFile(installPath, binName)
	if err == nil {
		return binPath, nil
	} else if err = ver.InstallLatest(installPath); err != nil {
		return "", err
	}

	return file_checker.FindFile(installPath, binName)
}

// Install last version of Java version if avaible
func (ver JavaVersion) InstallLatest(installPath string) error {
	if err := ver.InstallLatestAdoptium(installPath); err != nil && err != ErrSystem {
		return err
	} else if err = ver.InstallLatestLiberica(installPath); err != nil && err != ErrSystem {
		return err
	}
	return ErrSystem
}

// Return jar file java version
func JarMajor(r io.ReaderAt, size int64) (JavaVersion, error) {
	jarZip, err := zip.NewReader(r, size)
	if err != nil {
		return 0, err
	}

	for _, manifest := range jarZip.File {
		if !(strings.Contains(manifest.Name, "META-INF") && strings.Contains(manifest.Name, "MANIFEST.MF")) {
			continue
		}

		manifestRead, err := manifest.Open()
		if err != nil {
			return 0, fmt.Errorf("cannot get META-INF/MANIFEST.MF: %s", err)
		}
		defer manifestRead.Close()

		manifestLine := bufio.NewScanner(manifestRead)
		for manifestLine.Scan() {
			line := manifestLine.Text()
			switch {
			// case strings.HasPrefix(line, "Specification-Version:"):
			// 	specVersion, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Specification-Version:")))
			// 	if err != nil {
			// 		return 0, err
			// 	}
			// 	return JavaVersion(specVersion+44), nil
			case strings.HasPrefix(line, "Main-Class:"):
				mainClassName := strings.TrimSpace(strings.TrimPrefix(line, "Main-Class:"))
				for _, mainClass := range jarZip.File {
					ext := path.Ext(mainClass.Name)
					if mainClassName != strings.ReplaceAll(mainClass.Name[:len(mainClass.Name)-len(ext)], "/", ".") {
						continue
					}

					switch ext {
					case ".class":
						mainClassFile, err := mainClass.Open()
						if err == nil {
							defer mainClassFile.Close()
							header := make([]byte, 10)
							if _, err = io.ReadFull(mainClassFile, header); err == nil {
								return GetMajor(header)
							}
							return 0, fmt.Errorf("cannot get class heaader: %s", err)
						}
						return 0, fmt.Errorf("cannot open main-class file: %s", err)
					case ".jar":
						externalJar, err := mainClass.Open()
						if err != nil {
							return 0, fmt.Errorf("cannot open main-class file: %s", err)
						}
						defer externalJar.Close()

						tmpFile, err := os.CreateTemp("", "jarprebuild_*.jar")
						if err != nil {
							return 0, err
						} else if _, err = io.Copy(tmpFile, externalJar); err != nil {
							return 0, err
						}

						// Close and remove file
						defer func() {
							tmpFile.Close()
							os.Remove(tmpFile.Name())
						}()

						return JarMajor(tmpFile, mainClass.FileInfo().Size())
					default:
						return 0, fmt.Errorf("cannot process jar file")
					}
				}

				return 0, fmt.Errorf("cannot process main-class (%q)", line)
			}
		}
		if err := manifestLine.Err(); err != nil {
			return 0, err
		}
	}

	return 0, fmt.Errorf("cannot get Main-Class")
}

// Get Major Release from class
func GetMajor(header []byte) (JavaVersion, error) {
	switch {
	case len(header) < 10:
		return 0, ErrInvaliClassHeader
	case !bytes.Equal(header[:4], []byte{0xCA, 0xFE, 0xBA, 0xBE}):
		return 0, ErrInvaliClassHeader
	default:
		jvm := JavaVersion(binary.BigEndian.Uint16(header[6:8]))
		if jvm < 45 {
			return 0, ErrInvaliMajorVersion
		}
		return jvm, nil
	}
}

// Get local java if avaible, get [JavaVersion] and bin path
func LocalVersion() (string, JavaVersion, error) {
	javaBin, err := exec.LookPath("java")
	if err != nil {
		return "", 0, err
	}

	log, err := exec.Command(javaBin, "-version").CombinedOutput()
	if err != nil {
		return "", 0, err
	}

	if versionMatchStart := bytes.IndexByte(log, '"'); versionMatchStart != -1 {
		versionMatchStart++
		if versionMatchEnd := bytes.IndexByte(log[versionMatchStart:], '"'); versionMatchEnd != -1 {
			var newVersion JavaVersion
			return javaBin, newVersion, newVersion.UnmarshalText(log[versionMatchStart : versionMatchStart+versionMatchEnd])
		}
	}

	return "", 0, ErrSystem
}
