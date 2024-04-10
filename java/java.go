package java

import (
	"errors"
	"os"
	"path/filepath"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/exec"
)

var (
	DefaultServerJarName = "server.jar"                             // Default name to save .jar files
	ErrVersionNotExist   = errors.New("version not exists to java") // If version not exist
	ErrNoJava            = errors.New("cannot find java")           // Cannot get java path to run server
	ErrInstallServer     = errors.New("install server fist")        // Server not installed
)

func JarStart(serverPath string) (exec.ServerRun, error) {
	if !exec.ProgrammExist("java") {
		return exec.ServerRun{}, ErrNoJava
	} else if _, err := os.Stat(filepath.Join(serverPath, DefaultServerJarName)); os.IsNotExist(err) {
		return exec.ServerRun{}, ErrInstallServer
	}

	// make server arguments
	run := exec.ServerRun{
		Cwd: serverPath,
		Arguments: []string{
			"java",
			"-jar",
			DefaultServerJarName,
			"-nogui",
		},
	}

	err := run.Start()
	return run, err
}
