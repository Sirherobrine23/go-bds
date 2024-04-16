package host

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal"
)

// openjdk 21.0.2 2024-01-16
var OpenJDKVersionMatch *regexp.Regexp = regexp.MustCompile(`(?m)^openjdk (?P<Version>[0-9a-z\.\-_]+) (?P<Day>[0-9a-z\-]+)$`)

var JavaPath string // Host java path

func HostVersion() string {
	var err error
	if runtime.GOOS == "darwin" {
		pathsFind := []string{"/usr/local/opt/openjdk/bin", "/opt/local/opt/openjdk/bin"}
		for _, pathing := range pathsFind {
			_, err = os.Open(pathing)
			if os.IsNotExist(err) {
				continue
			} else if !os.IsExist(err) && err != nil {
				return ""
			}
			JavaPath = filepath.Join(pathing, "java")
			break
		}
	}

	if len(JavaPath) == 0 {
		JavaPath, err = exec.LookPath("java")
		if err != nil {
			return ""
		}
	}

	log, err := exec.Command(JavaPath, "--version").Output()
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(log), "\n") {
		line = strings.TrimSpace(line)
		if OpenJDKVersionMatch.MatchString(line) {
			info := internal.FindAllGroups(OpenJDKVersionMatch, line)
			return info["Version"]
		}
	}
	return ""
}
