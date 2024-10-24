package java

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/java/adoptium"
)

var (
	ErrInstallServer error = errors.New("install server fist")
	ErrNoServer      error = errors.New("cannot find server")
)

const ServerMain string = "server.jar"

// Global struct to Minecraft java server to run .jar
type JavaServer struct {
	JavaFolders string `json:"javaFolders"` // Java bins, if blank use local java
	JavaVersion uint   `json:"javaVersion"` // Java version to run
	SavePath    string `json:"savePath"`    // Folder path to save server run data

	SeverProc exec.Proc // Interface to process running
}

// Start server
//
// On run this function YOU auto accept minecraft EULA https://www.minecraft.net/en-us/eula
func (w *JavaServer) Start() error {
	if _, err := os.Stat(filepath.Join(w.SavePath, ServerMain)); os.IsNotExist(err) {
		return ErrInstallServer
	}
	w.SeverProc = &exec.Os{}
	var opts = exec.ProcExec{
		Arguments: []string{"java", "-jar", ServerMain, "--nogui"},
		Cwd:       w.SavePath,
	}

	if w.JavaFolders != "" {
		opts.Arguments[0] = w.JavaFolders
	} else {
		javaRootFolder := filepath.Join(w.JavaFolders, strconv.FormatInt(int64(w.JavaVersion), 10))
		if err := adoptium.InstallLatest(w.JavaVersion, javaRootFolder); err != nil {
			return err
		}
		opts.Arguments[0] = filepath.Join(javaRootFolder, "bin/java")
	}

	// Write eula=true
	os.WriteFile(filepath.Join(w.SavePath, "eula.txt"), []byte("eula=true"), 0600)
	if err := w.SeverProc.Start(opts); err != nil {
		return err
	}
	return nil
}
