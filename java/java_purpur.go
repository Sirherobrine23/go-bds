package java

import (
	"bytes"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"sync"

	"sirherobrine23.com.br/go-bds/go-bds/utils/javaprebuild"
	"sirherobrine23.com.br/go-bds/go-bds/utils/semver"
	"sirherobrine23.com.br/go-bds/request/v2"
)

func purpurWorkder(vers *Versions, job <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	type buildTargetInfo struct {
		MCStarget string `json:"version"`
		Result    string `json:"result"`
		Time      int64  `json:"timestamp"`
		Build     string `json:"build"`
	}

	type purpurBuild struct {
		Builds struct {
			Latest string   `json:"latest"`
			All    []string `json:"all"`
		} `json:"builds"`
	}

	for Version := range job {
		buildInfo, _, err := request.JSON[purpurBuild](fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s", Version), nil)
		if err != nil {
			continue
		}
		var resBuild buildTargetInfo
		slices.Reverse(buildInfo.Builds.All)
		for _, build := range buildInfo.Builds.All {
			_, err = request.DoJSON(fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s", Version, build), &resBuild, nil)
			if err != nil || strings.ToUpper(resBuild.Result) == "SUCCESS" {
				break
			}
		}

		if strings.ToUpper(resBuild.Result) != "SUCCESS" {
			continue
		}

		downloadUrl := fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s/download", Version, resBuild.Build)
		jarFile, _, err := request.Buffer(downloadUrl, nil)
		if err != nil {
			continue
		}

		jvm, err := javaprebuild.JarMajor(bytes.NewReader(jarFile), int64(len(jarFile)))
		if err != nil {
			continue
		}

		*vers = append(*vers, GenericVersion{
			ServerVersion: resBuild.MCStarget,
			DownloadURL:   downloadUrl,
			JVM:           jvm,
		})
	}
}

// Fetch versions from purpur API
func (vers *Versions) FetchPurpurVersions() error {
	type projectVersions struct {
		Versions []string `json:"versions"`
	}
	versions, _, err := request.JSON[projectVersions]("https://api.purpurmc.org/v2/purpur", nil)
	if err != nil {
		return err
	}

	*vers = (*vers)[:0]
	jobs := make(chan string)
	var wg sync.WaitGroup
	for range runtime.NumCPU() * 4 {
		wg.Add(1)
		go purpurWorkder(vers, jobs, &wg)
	}

	for _, version := range versions.Versions {
		jobs <- version
	}

	close(jobs) // Close jobs channel
	wg.Wait()   // Wait to workers finish
	semver.Sort(*vers)
	return nil
}
