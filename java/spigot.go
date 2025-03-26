package java

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/utils/javaprebuild"
	"sirherobrine23.com.br/go-bds/request/gohtml"
	"sirherobrine23.com.br/go-bds/request/v2"
	"sirherobrine23.com.br/go-bds/go-bds/semver"
)

var (
	SpigotBuildToolsURL string  = "https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar"
	_                   Version = &SpigotMC{}
)

type SpigotMC struct {
	Version      string `json:"version"`                // Server version
	Desc         string `json:"description"`            // Commit short description
	ToolVersion  int64  `json:"toolsVersion,omitempty"` // Spigot BuildTools version required
	JavaVersions []uint `json:"javaVersions,omitempty"` // Java major version range
	JavaFolder   string `json:"-"`                      // Java folder with prebuild binarys
	Ref          struct {
		Spigot      string `json:"Spigot"`      // https://hub.spigotmc.org/stash/projects/SPIGOT/repos/spigot/commits/{ID}
		Bukkit      string `json:"Bukkit"`      // https://hub.spigotmc.org/stash/projects/SPIGOT/repos/bukkit/commits/{ID}
		CraftBukkit string `json:"CraftBukkit"` // https://hub.spigotmc.org/stash/projects/SPIGOT/repos/Craftbukkit/commits/{ID}
	} `json:"refs"` // Repository Commit ID
}

// List all Spigot releases from hub.spigotmc.org
func ListSpigot(JavaBinaryFolder string) ListServer {
	return func() (Versions, error) {
		versionsURL := "https://hub.spigotmc.org/versions/"
		res, err := request.Request(versionsURL, nil)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		// Get all json files
		var urls struct {
			Files []string `html:"body > pre > a = href"`
		}
		if err := gohtml.NewDecode(res.Body, &urls); err != nil {
			return nil, err
		}

		urls.Files = slices.DeleteFunc(urls.Files, func(entry string) bool {
			return !(strings.HasPrefix(entry, "1.") && strings.HasSuffix(entry, ".json"))
		})

		Version := Versions{}
		for _, entry := range urls.Files {
			var spigotRelease SpigotMC
			if _, err := request.DoJSON(versionsURL+entry, &spigotRelease, nil); err != nil {
				return nil, err
			}
			spigotRelease.JavaFolder = JavaBinaryFolder
			spigotRelease.Version = strings.TrimSuffix(entry, ".json")
			if len(spigotRelease.JavaVersions) == 0 {
				spigotRelease.JavaVersions = []uint{52, 52}
			}
			Version = append(Version, spigotRelease)
		}
		semver.SortStruct(Version)
		return Version, nil
	}
}

func (ver SpigotMC) SemverVersion() *semver.Version { return semver.New(ver.Version) }

// Dirt function to get java version
func (ver SpigotMC) JVM() uint { return ver.JavaVersions[len(ver.JavaVersions)-1] - 44 }

// Build Spigot localy
//
// If InstallPath already has java in the path "InstallPath + bin/java", this version will be used
func (ver SpigotMC) Install(InstallPath string) error {
	if _, err := os.Stat(filepath.Join(InstallPath, ServerName)); err == nil {
		filesEntrys, _ := os.ReadDir(InstallPath)
		for _, file := range filesEntrys {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".jar") {
				os.Remove(filepath.Join(InstallPath, file.Name()))
			}
		}
	}

	// Folder path to build spigot server
	BuildDir := filepath.Join(InstallPath, "spigotbuild")
	if _, err := os.Stat(BuildDir); os.IsNotExist(err) {
		if err := os.MkdirAll(BuildDir, 0777); err != nil {
			return err
		}
	}
	defer os.RemoveAll(BuildDir) // Remove Before build

	// Download spigot build tools
	BuildTools := filepath.Join(BuildDir, "SpigotBuildTools.jar")
	if _, err := request.SaveAs(SpigotBuildToolsURL, BuildTools, nil); err != nil {
		return err
	}

	// Base to build
	execOpt := exec.ProcExec{
		Cwd:       BuildDir,
		Arguments: []string{"java", "-jar", "SpigotBuildTools.jar", "--rev", ver.Version, "--output-dir", InstallPath},
	}

	javaFromVersions := filepath.Join(ver.JavaFolder, fmt.Sprint(ver.JVM()), "bin/java")
	if runtime.GOOS == "windows" {
		javaFromVersions += ".exe"
	}

	if _, err := os.Stat(javaFromVersions); err == nil {
		execOpt.Arguments[0] = javaFromVersions // Set installed java
	} else if os.IsNotExist(err) {
		if err := javaprebuild.InstallLatest(ver.JVM(), filepath.Join(ver.JavaFolder, fmt.Sprint(ver.JVM()))); err != nil {
			if err == javaprebuild.ErrSystem {
				return fmt.Errorf("cannot build spigot server")
			}
			return err
		}
		execOpt.Arguments[0] = javaFromVersions // Set installed java
	} else if !exec.LocalBinExist(execOpt) {
		return fmt.Errorf("cannot build spigot server")
	}

	build := exec.Os{}
	if err := build.Start(execOpt); err != nil {
		return err
	}

	// Copy log to log file
	if stderr, err := build.StderrFork(); err == nil {
		defer stderr.Close()
		go io.Copy(io.Discard, stderr) //nolint:errcheck
	}
	if stdout, err := build.StdoutFork(); err == nil {
		defer stdout.Close()
		go io.Copy(io.Discard, stdout) //nolint:errcheck
	}

	// Wait process end
	if err := build.Wait(); err != nil {
		return err
	}

	if entrys, err := os.ReadDir(InstallPath); err == nil {
		if fileIndex := slices.IndexFunc(entrys, func(entry fs.DirEntry) bool { return strings.Contains(entry.Name(), "spigot") }); fileIndex >= 0 {
			if err := os.Rename(filepath.Join(InstallPath, entrys[fileIndex].Name()), filepath.Join(InstallPath, ServerName)); err != nil {
				return err
			}
		}
	}

	return nil
}
