package powernukkit

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

const (
	ProjectVersions string = "https://raw.githubusercontent.com/PowerNukkit/powernukkit-version-aggregator/master/powernukkit-versions.json"
)

type aggretor struct {
	Release   int64  `json:"releaseTime"`
	Minecraft string   `json:"minecraftVersion"`
	Version   string   `json:"version"`
	Commit    string   `json:"commitId"`
	Snapshot  int64    `json:"snapshotBuild"`
	Artefacts []string `json:"artefacts"`
}

type versionsAggretor map[string][]aggretor

type Version struct {
	ReleaseType string    `json:"releaseType"`
	Release     time.Time `json:"releaseTime"`
	Version     string    `json:"version"`
	Minecraft   string    `json:"minecraftVersion"`
	File        string    `json:"jarFile"`
}

func padStart(source string, size int, text string) string {
	n := source
	for len(n) < size {
		n = text + n
	}
	return n
}

func MapFiles() ([]Version, error) {
	newList := []Version{}
	res, err := request.Request(request.RequestOptions{HttpError: true, Url: ProjectVersions})
	if err != nil {
		return newList, err
	}

	defer res.Body.Close()
	var versions versionsAggretor
	if err = json.NewDecoder(res.Body).Decode(&versions); err != nil {
		return newList, err
	}

	for ReleaseType, Targets := range versions {
		for _, release := range Targets {

			if artefactId, exit := internal.ArrayStringIncludes(release.Artefacts, "REDUCED_JAR", "SHADED_JAR"); exit {
				artefactExtension := ""
				if artefactId == "REDUCED_JAR" {
					artefactExtension = ".jar"
				} else if artefactId == "REDUCED_SOURCES_JAR" {
					artefactExtension = "-sources.jar"
				} else if artefactId == "SHADED_JAR" {
					artefactExtension = "-shaded.jar"
				} else if artefactId == "SHADED_SOURCES_JAR" {
					artefactExtension = "-shaded-sources.jar"
				} else if artefactId == "JAVADOC_JAR" {
					artefactExtension = "-javadoc.jar"
				}

				utcTime := time.UnixMilli(int64(release.Release)).UTC()
				var fileUrl string
				if release.Snapshot > 0 {
					ver := release.Version
					if strings.Index(release.Version, "-SNAPSHOT") > 0 {
						ver = release.Version[:strings.Index(release.Version, "-SNAPSHOT")]
					}

					dateVer := strings.Join([]string{
						padStart(strconv.Itoa(utcTime.Year()), 4, "0"),
						padStart(strconv.Itoa(int(utcTime.Month())), 2, "0"),
						padStart(strconv.Itoa(utcTime.Day()), 2, "0"),
						".",
						padStart(strconv.Itoa(utcTime.Hour()), 2, "0"),
						padStart(strconv.Itoa(utcTime.Minute()), 2, "0"),
						padStart(strconv.Itoa(utcTime.Second()), 2, "0"),
					}, "")

					fileUrl = fmt.Sprintf("https://oss.sonatype.org/content/repositories/snapshots/org/powernukkit/powernukkit/%s-SNAPSHOT/powernukkit-%s-%s-%d%s", ver, ver, dateVer, release.Snapshot, artefactExtension)
				} else {
					fileUrl = fmt.Sprintf("https://search.maven.org/remotecontent?filepath=org/powernukkit/powernukkit/%s/powernukkit-%s%s", release.Version, release.Version, artefactExtension)
				}

				newList = append(newList, Version{
					ReleaseType,
					utcTime,
					release.Version,
					release.Minecraft,
					fileUrl,
				})
			}
		}
	}

	return newList, nil
}
