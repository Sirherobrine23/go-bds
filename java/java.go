package java

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/client"
	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/internal/ccopy"
	javadown "sirherobrine23.com.br/go-bds/go-bds/java/adoptium"
)

//go:embed javac/*
var javac embed.FS

// Global struct to Minecraft java server to run .jar
type JavaServer struct {
	JavaFolders string `json:"javaFolders"` // Java bins, if blank use local java or `docker:` to run insider container
	JavaVersion uint   `json:"javaVersion"` // Java version to run
	SavePath    string `json:"savePath"`    // Folder path to save server run data

	SeverProc exec.Proc // Interface to process running
}

func (w *JavaServer) Start() error {
	w.SeverProc = &exec.Os{}
	var opts = exec.ProcExec{
		Arguments: []string{"java", "-jar", "server.jar", "--nogui"},
		Cwd:       w.SavePath,
	}

	if w.JavaFolders != "" {
		if strings.HasPrefix(w.JavaFolders, "docker:") {
			cli, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				return err
			}
			if w.JavaFolders[7:] == "" {
				w.JavaFolders = w.JavaFolders[7:] + fmt.Sprintf("docker.io/eclipse-temurin:%d-jre", w.JavaVersion)
			}
			opts.Cwd = "/data/mcjava"
			w.SeverProc = &exec.Docker{
				DockerClient:      cli,
				DockerImage:       w.JavaFolders[7:],
				Network:           "host",
				ReplaceEntrypoint: true,
				LocalFolders: []string{
					fmt.Sprintf("%s:/data/mcjava:rw", w.SavePath),
				},
			}
		} else {
			opts.Arguments[0] = w.JavaFolders
		}
	} else {
		javaRootFolder := filepath.Join(w.JavaFolders, strconv.FormatInt(int64(w.JavaVersion), 10))
		javacEmbed := fmt.Sprintf("javac/%d", w.JavaVersion)
		if ccopy.FSExists(javac, javacEmbed) {
			if err := ccopy.FSCopyDirectory(javac, javacEmbed, javaRootFolder); err != nil {
				return err
			}
		} else if err := javadown.InstallLatest(w.JavaVersion, javaRootFolder); err != nil {
			return err
		}
		opts.Arguments[0] = filepath.Join(javaRootFolder, "bin/java")
	}
	os.WriteFile(filepath.Join(w.SavePath, "eula.txt"), []byte("eula=true"), 0600)
	if err := w.SeverProc.Start(opts); err != nil {
		return err
	}
	return nil
}
