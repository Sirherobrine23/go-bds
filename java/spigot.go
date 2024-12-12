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
	"sirherobrine23.com.br/go-bds/go-bds/java/javaprebuild"
	"sirherobrine23.com.br/go-bds/go-bds/request/gohtml"
	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
	"sirherobrine23.com.br/go-bds/go-bds/semver"
)

var (
	SpigotBuildToolsURL string        = "https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar"
	_                   VersionSearch = SpigotSearch{}
	_                   Version       = &SpigotMC{}
)

type SpigotSearch struct {
	Version []SpigotMC
}

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
func (versions *SpigotSearch) list() error {
	versionsURL := "https://hub.spigotmc.org/versions/"
	res, err := request.Request(versionsURL, nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Get all json files
	var urls struct {
		Files []string `html:"body > pre > a = href"`
	}
	if err := gohtml.NewParse(res.Body, &urls); err != nil {
		return err
	}

	urls.Files = slices.DeleteFunc(urls.Files, func(entry string) bool {
		return !(strings.HasPrefix(entry, "1.") && strings.HasSuffix(entry, ".json"))
	})
	for _, entry := range urls.Files {
		var spigotRelease SpigotMC
		if _, err := request.JSONDo(versionsURL+entry, &spigotRelease, nil); err != nil {
			return err
		}
		spigotRelease.Version = strings.TrimSuffix(entry, ".json")
		if len(spigotRelease.JavaVersions) == 0 {
			spigotRelease.JavaVersions = []uint{52, 52}
		}
		versions.Version = append(versions.Version, spigotRelease)
	}
	semver.SortStruct(versions.Version)
	return nil
}

func (versions SpigotSearch) Find(version string) (Version, error) {
	if err := versions.list(); err != nil {
		return nil, err
	}

	for _, ver := range versions.Version {
		if ver.Version == version {
			return ver, nil
		}
	}
	return nil, ErrNoFoundVersion
}

func (ver SpigotMC) SemverVersion() *semver.Version { return semver.New(ver.Version) }

// Dirt function to get java version
func (ver SpigotMC) JavaVersion() uint { return ver.JavaVersions[len(ver.JavaVersions)-1] - 44 }

// Build Spigot localy
func (ver SpigotMC) Install(InstallPath string) error {
	if _, err := os.Stat(filepath.Join(InstallPath, ServerMain)); err == nil {
		if err := os.Remove(filepath.Join(InstallPath, ServerMain)); err != nil {
			return err
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

	if !exec.LocalBinExist(execOpt) {
		javaFolder := filepath.Join(BuildDir, "java")
		if execOpt.Arguments[0] = filepath.Join(javaFolder, "bin/java"); runtime.GOOS == "windows" {
			execOpt.Arguments[0] += ".exe"
		}
		if _, err := os.Stat(execOpt.Arguments[0]); os.IsNotExist(err) {
			if err := javaprebuild.InstallLatest(ver.JavaVersion(), javaFolder); err != nil {
				if err == javaprebuild.ErrSystem {
					return fmt.Errorf("cannot build spigot server")
				}
				return err
			}
		}
	}

	build := exec.Os{}
	if err := build.Start(execOpt); err != nil {
		return err
	}

	// Copy log to log file
	if stderr, err := build.StderrFork(); err == nil {
		defer stderr.Close()
		go io.Copy(io.Discard, stderr)
	}
	if stdout, err := build.StdoutFork(); err == nil {
		defer stdout.Close()
		go io.Copy(io.Discard, stdout)
	}

	// Wait process end
	if err := build.Wait(); err != nil {
		return err
	}

	if entrys, err := os.ReadDir(InstallPath); err == nil {
		if fileIndex := slices.IndexFunc(entrys, func(entry fs.DirEntry) bool { return strings.Contains(entry.Name(), "spigot") }); fileIndex >= 0 {
			if err := os.Rename(filepath.Join(InstallPath, entrys[fileIndex].Name()), filepath.Join(InstallPath, ServerMain)); err != nil {
				return err
			}
		}
	}

	return nil
}
