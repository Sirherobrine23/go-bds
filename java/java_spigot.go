package java

import (
	"fmt"
	"os"
	os_exec "os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/utils/file_checker"
	"sirherobrine23.com.br/go-bds/go-bds/utils/javaprebuild"
	"sirherobrine23.com.br/go-bds/go-bds/utils/semver"
	"sirherobrine23.com.br/go-bds/request/v2"
)

// SpigotTool url
var SpigotBuildToolsURL string = "https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar"

// Fetch spigot versions from hub.spigotmc.org
func (vers *Versions) FetchSpigotVersions() error {
	// Get all json files
	var fileURLs struct {
		Files []string `html:"body > pre > a = href"`
	}

	if _, err := request.DoGoHTML("https://hub.spigotmc.org/versions/", &fileURLs, nil); err != nil {
		return err
	}

	*vers = (*vers)[:0]
	for _, file := range fileURLs.Files {
		if !(strings.HasPrefix(file, "1.") && strings.HasSuffix(file, ".json")) {
			continue
		}
		spigotInfo, _, err := request.JSON[*SpigotMC]("https://hub.spigotmc.org/versions/"+file, nil)
		if err != nil {
			return err
		}
		spigotInfo.MCVersion = strings.TrimSuffix(file, ".json")
		*vers = append(*vers, spigotInfo)
	}
	semver.Sort(*vers)
	return nil
}

// SpigotMC cannot have the public server files because of DCMA (https://www.spigotmc.org/wiki/unofficial-explanation-about-the-dmca/),
// so the server has to be compiled locally,
// so have all the tools for its use https://www.spigotmc.org/wiki/buildtools/#prerequisites
type SpigotMC struct {
	MCVersion    string `json:"version"`                // Server version
	Desc         string `json:"description"`            // Commit short description
	ToolVersion  int64  `json:"toolsVersion,omitempty"` // Spigot BuildTools version required
	JavaVersions []uint `json:"javaVersions,omitempty"` // Java major version range
	Ref          struct {
		Spigot      string `json:"Spigot"`      // https://hub.spigotmc.org/stash/projects/SPIGOT/repos/spigot/commits/{ID}
		Bukkit      string `json:"Bukkit"`      // https://hub.spigotmc.org/stash/projects/SPIGOT/repos/bukkit/commits/{ID}
		CraftBukkit string `json:"CraftBukkit"` // https://hub.spigotmc.org/stash/projects/SPIGOT/repos/Craftbukkit/commits/{ID}
	} `json:"refs"` // Repository Commit ID
}

// Server version
func (version SpigotMC) Version() string { return version.MCVersion }

// Return last java version to run server
func (version SpigotMC) JavaVersion() javaprebuild.JavaVersion {
	return javaprebuild.JavaVersion(version.JavaVersions[len(version.JavaVersions)-1])
}

// Build Spigot localy
//
// If InstallPath already has java in the path "InstallPath + bin/java", this version will be used
func (ver *SpigotMC) Install(InstallPath string) error {
	if file_checker.IsFile(filepath.Join(InstallPath, "server.jar")) {
		return nil
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

	javaPath := filepath.Join(BuildDir, "java")
	switch err := ver.JavaVersion().InstallLatest(javaPath); err {
	case nil:
		if javaPath = filepath.Join(javaPath, "bin/java"); runtime.GOOS == "windows" {
			javaPath += ".exe"
		}
	case javaprebuild.ErrSystem:
		if javaPath, err = os_exec.LookPath("java"); err != nil {
			return fmt.Errorf("install java in u system: %s", err)
		}
	default:
		return err
	}

	// Make build struct
	build := &exec.Os{}

	// --compile <[NONE,CRAFTBUKKIT,SPIGOT]>  Software to compile
	for _, buildTarget := range []string{"spigot", "craftbukkit"} {
		buildFlag := exec.ProcExec{
			Cwd: BuildDir,
			Arguments: []string{
				javaPath, "-jar", "SpigotBuildTools.jar",
				"--rev", ver.MCVersion,
				"--output-dir", InstallPath,
			},
		}

		// Append compile target
		switch buildTarget {
		case "spigot":
			buildFlag.Arguments = append(buildFlag.Arguments, "--compile", buildTarget, "--final-name", "server.jar")
		default:
			buildFlag.Arguments = append(buildFlag.Arguments, "--compile", buildTarget)
		}

		// Wait process end
		build.Start(buildFlag)
		if err := build.Wait(); err != nil {
			return err
		}
	}
	os.RemoveAll(BuildDir)
	os.WriteFile(filepath.Join(InstallPath, "server.jar"), []byte("eula=true"), 0755)
	return nil
}
