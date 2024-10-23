package java

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/internal/semver"
	"sirherobrine23.com.br/go-bds/go-bds/java/adoptium"
	"sirherobrine23.com.br/go-bds/go-bds/request/gohtml"
	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

var SpigotBuildToolsURL string = "https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar"

type SpigotMC struct {
	Version      string `json:"version"`                // Server version
	Desc         string `json:"description"`            // Commit short description
	ToolVersion  int64  `json:"toolsVersion,omitempty"` // Spigot BuildTools version required
	JavaVersions []uint `json:"javaVersions,omitempty"` // Java major version range
	Ref          struct {
		Spigot      string `json:"Spigot"`      // https://hub.spigotmc.org/stash/projects/SPIGOT/repos/spigot/commits/{ID}
		Bukkit      string `json:"Bukkit"`      // https://hub.spigotmc.org/stash/projects/SPIGOT/repos/bukkit/commits/{ID}
		CraftBukkit string `json:"CraftBukkit"` // https://hub.spigotmc.org/stash/projects/SPIGOT/repos/Craftbukkit/commits/{ID}
	} `json:"refs"` // Repository Commit ID
}

// List all Spigot releases from hub.spigotmc.org
func ListFromSpigotmc() ([]SpigotMC, error) {
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
	if err := gohtml.NewParse(res.Body, &urls); err != nil {
		return nil, err
	}

	urls.Files = slices.DeleteFunc(urls.Files, func(entry string) bool {
		return !(strings.HasPrefix(entry, "1.") && strings.HasSuffix(entry, ".json"))
	})
	SpigotReleases := []SpigotMC{}
	for _, entry := range urls.Files {
		var spigotRelease SpigotMC
		if _, err := request.JSONDo(versionsURL+entry, &spigotRelease, nil); err != nil {
			return SpigotReleases, err
		}
		spigotRelease.Version = strings.TrimSuffix(entry, ".json")
		if len(spigotRelease.JavaVersions) == 0 {
			spigotRelease.JavaVersions = []uint{52, 52}
		}
		SpigotReleases = append(SpigotReleases, spigotRelease)
	}
	semver.SortStruct(SpigotReleases)
	return SpigotReleases, nil
}

func (ver SpigotMC) SemverVersion() *semver.Version { return semver.New(ver.Version) }

// Dirt function to get java version
func (ver SpigotMC) JavaVersion() uint { return ver.JavaVersions[len(ver.JavaVersions)-1] - 44 }

// Build Spigot localy
func (ver SpigotMC) Build(BuildFolder, OutputFolder string, logStream io.Writer) error {
	BuildTools := filepath.Join(BuildFolder, "SpigotBuildTools.jar")

	// If not exists, Download latest version
	if _, err := os.Stat(BuildTools); os.IsNotExist(err) {
		// Download spigot build tools
		if _, err := request.SaveAs(SpigotBuildToolsURL, BuildTools, nil); err != nil {
			return err
		}
		defer os.Remove(BuildTools) // Remove Before build
	}

	// Base to build
	execOpt := exec.ProcExec{
		Cwd:       BuildFolder,
		Arguments: []string{"java", "-jar", "SpigotBuildTools.jar", "--rev", ver.Version, "--output-dir", OutputFolder},
	}

	if !exec.LocalBinExist(execOpt.Arguments[0]) {
		javaFolder := filepath.Join(BuildFolder, "java")
		if execOpt.Arguments[0] = filepath.Join(javaFolder, "bin/java"); runtime.GOOS == "windows" {
			execOpt.Arguments[0] += ".exe"
		}
		if _, err := os.Stat(execOpt.Arguments[0]); os.IsNotExist(err) {
			if err := adoptium.InstallLatest(ver.JavaVersion(), javaFolder); err != nil {
				if err == adoptium.ErrSystem {
					return fmt.Errorf("cannot build spigot server")
				}
				return err
			}
		}
	}

	d, _ := json.MarshalIndent(execOpt, "", "  ")
	fmt.Println(string(d))

	build := exec.Os{}
	if err := build.Start(execOpt); err != nil {
		return err
	}

	// Copy log to log file
	if stderr, err := build.StderrFork(); err == nil {
		defer stderr.Close()
		go io.Copy(logStream, stderr)
	}
	if stdout, err := build.StdoutFork(); err == nil {
		defer stdout.Close()
		go io.Copy(logStream, stdout)
	}

	return build.Wait() // Wait process end
}
