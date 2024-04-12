package oracle

import (
	"errors"
	"fmt"
	"runtime"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal"
)

var (
	AvaibleGOOS   []string = []string{"windows", "darwin", "linux"}
	AvaibleGOARCH []string = []string{"arm64", "x64"}

	ErrGOARCHNoAvaible error = errors.New("current arch no avaible java")
	ErrGOOSNoAvaible error = errors.New("current system no avaible java")
)

type Oracle struct {
	// Java version, example: 17, 21, 22, etc...
	Version int
}

func (w *Oracle) Get() error {
	if _, ok := internal.ArrayStringIncludes(AvaibleGOOS, runtime.GOOS); !ok {
		return ErrGOOSNoAvaible
	} else if _, ok := internal.ArrayStringIncludes(AvaibleGOARCH, runtime.GOARCH); !ok {
		return ErrGOARCHNoAvaible
	}

	var fileUrl string
	GOARCH := runtime.GOARCH
	if runtime.GOOS == "linux" {
		// https://download.oracle.com/java/22/latest/jdk-22_linux-aarch64_bin.tar.gz
		// https://download.oracle.com/java/22/latest/jdk-22_linux-x64_bin.tar.gz
		if runtime.GOARCH == "arm64" {
			GOARCH = "aarch64"
		}
		fileUrl = fmt.Sprintf("https://download.oracle.com/java/%s/latest/jdk-%s_linux-%s_bin.tar.gz", w.Version, w.Version, GOARCH)
		} else if runtime.GOOS == "darwin" {
		// https://download.oracle.com/java/22/latest/jdk-22_macos-aarch64_bin.tar.gz
		// https://download.oracle.com/java/22/latest/jdk-22_macos-x64_bin.tar.gz
		fileUrl = fmt.Sprintf("https://download.oracle.com/java/%s/latest/jdk-%s_macos-%s_bin.tar.gz", w.Version, w.Version, GOARCH)
	} else if runtime.GOOS == "windows" {
		// https://download.oracle.com/java/22/latest/jdk-22_windows-x64_bin.zip
		fileUrl = "https://download.oracle.com/java/22/latest/jdk-22_windows-x64_bin.zip"
	}

	fmt.Println(fileUrl)
	return nil
}
