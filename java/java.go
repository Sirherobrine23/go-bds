package java

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/client"
	"sirherobrine23.org/go-bds/go-bds/exec"
	javadown "sirherobrine23.org/go-bds/go-bds/java/downloads"
)

// Run Server with docker image: Version -> image -> go platforms
var DockerImages = map[uint]map[string][]string{
	8: {
		"docker.io/eclipse-temurin:8-jre": []string{"windows/amd64", "linux/amd64", "linux/arm/v7", "linux/arm64", "linux/ppc64le"},
	},
	11: {
		"docker.io/eclipse-temurin:11-jre": []string{"windows/amd64", "linux/amd64", "linux/arm/v7", "linux/arm64", "linux/ppc64le"},
	},
	17: {
		"docker.io/eclipse-temurin:17-jre": []string{"windows/amd64", "linux/amd64", "linux/arm/v7", "linux/arm64", "linux/ppc64le"},
	},
	21: {
		"docker.io/eclipse-temurin:21-jre": []string{"windows/amd64", "linux/amd64", "linux/arm/v7", "linux/arm64", "linux/ppc64le"},
	},
	22: {
		"docker.io/eclipse-temurin:22-jre": []string{"windows/amd64", "linux/amd64", "linux/arm/v7", "linux/arm64", "linux/ppc64le"},
	},
	23: {
		"docker.io/openjdk:23-jdk": []string{"windows/amd64", "linux/amd64", "linux/arm64"},
	},
	24: {
		"docker.io/openjdk:24-jdk": []string{"windows/amd64", "linux/amd64", "linux/arm64"},
	},
}

//go:embed javac/*
var javac embed.FS

// Global struct to Minecraft java server to run .jar
type JavaServer struct {
	JavaVersionsPath string `json:"javaVersionsFolder"` // Java bins, if blank use local java or `docker:` to run insider container
	JavaVersion      uint   `json:"javaVersion"`        // Java version to run
	SavePath         string // Folder path to save server run data

	SeverProc exec.Proc // Interface to process running
}

func checkExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (w *JavaServer) Start() error {
	w.SeverProc = &exec.Os{}
	var opts = exec.ProcExec{
		Arguments: []string{"java", "-jar", "server.jar", "--nogui"},
		Cwd:       w.SavePath,
	}

	if w.JavaVersionsPath != "" {
		if strings.HasPrefix(w.JavaVersionsPath, "docker:") {
			cli, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				return err
			}
			if w.JavaVersionsPath[7:] == "" {
				w.JavaVersionsPath = w.JavaVersionsPath[7:] + fmt.Sprintf("docker.io/eclipse-temurin:%d-jre", w.JavaVersion)
			}
			opts.Cwd = "/data/mcjava"
			w.SeverProc = &exec.Docker{
				DockerClient:      cli,
				DockerImage:       w.JavaVersionsPath[7:],
				Network:           "host",
				ReplaceEntrypoint: true,
				LocalFolders:      []string{
					fmt.Sprintf("%s:/data/mcjava:rw", w.SavePath),
				},
			}
		} else {
			opts.Arguments[0] = w.JavaVersionsPath
		}
	} else {
		var avaibleInternal bool
		var majorVersion = strconv.Itoa(int(w.JavaVersion))
		if _, err := javac.ReadDir(fmt.Sprintf("javac/%d", w.JavaVersion)); err == nil {
			avaibleInternal = true
		}
		javaRootFolder := filepath.Join(w.JavaVersionsPath, majorVersion)
		if !checkExist(javaRootFolder) {
			if avaibleInternal {
				err := fs.WalkDir(javac, fmt.Sprintf("javac/%d", w.JavaVersion), func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					fixedPath := filepath.Join(filepath.SplitList(path)[2:]...)
					if d.IsDir() {
						return os.MkdirAll(filepath.Join(javaRootFolder, fixedPath), d.Type())
					}

					file, err := os.OpenFile(filepath.Join(javaRootFolder, fixedPath), os.O_CREATE|os.O_RDWR|os.O_EXCL, d.Type())
					if err != nil {
						return err
					}
					defer file.Close()

					gfile, _ := javac.Open(path)
					defer file.Close()
					if _, err := io.Copy(file, gfile); err != nil {
						return err
					}

					return nil
				})
				if err != nil {
					return err
				}
			} else {
				if err := javadown.InstallLatest(w.JavaVersion, javaRootFolder); err != nil {
					return err
				}
			}
		}
	}

	os.WriteFile(filepath.Join(w.SavePath, "eula.txt"), []byte("eula=true"), 0600)

	if err := w.SeverProc.Start(opts); err != nil {
		return err
	}
	return nil
}
