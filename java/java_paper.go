package java

import (
	"fmt"
	"maps"
	"os"
	"runtime"
	"slices"
	"sync"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/utils/javaprebuild"
	"sirherobrine23.com.br/go-bds/go-bds/utils/semver"
	"sirherobrine23.com.br/go-bds/request/v2"
)

var (
	// Paper projects
	PaperProjects []string = []string{"paper", "folia", "velocity"}

	paperProjectURL          string = "https://api.papermc.io/v2/projects/%s"
	paperProjectBuildsURL    string = "https://api.papermc.io/v2/projects/%s/versions/%s/builds"
	paperProjectGetBuildsURL string = "https://api.papermc.io/v2/projects/%s/versions/%s/builds/%d/downloads/%s"
)

// Fetch versions to Paper Server
func (vers *Versions) FetchPaperVersions() error { return vers.fetchPaperProject("paper") }

// Fetch versions to Folia Server
func (vers *Versions) FetchFoliaVersions() error { return vers.fetchPaperProject("folia") }

// Fetch versions to Velocity Server
func (vers *Versions) FetchVelocityVersions() error { return vers.fetchPaperProject("velocity") }

type paperBuilds struct {
	Version   string
	Build     int64     `json:"build"`
	BuildTime time.Time `json:"time"`
	Downloads map[string]struct {
		Name   string `json:"name"`
		SHA256 string `json:"sha256"`
	} `json:"downloads"`
}

func paperWorkder(vers *Versions, ProjectTarget string, job <-chan paperBuilds, errPtr *error, wg *sync.WaitGroup) {
	defer wg.Done()
	for latestBuild := range job {
		if !slices.Contains(slices.Collect(maps.Keys(latestBuild.Downloads)), "application") {
			continue
		}

		downloadUrl := fmt.Sprintf(paperProjectGetBuildsURL, ProjectTarget, latestBuild.Version, latestBuild.Build, latestBuild.Downloads["application"].Name)
		jarFile, _, err := request.SaveTmp(downloadUrl, "", nil)
		if err != nil {
			*errPtr = err
			continue
		}
		defer os.Remove(jarFile.Name())
		defer jarFile.Close()
		stat, _ := jarFile.Stat()
		jvm, err := javaprebuild.JarMajor(jarFile, stat.Size())
		if err != nil {
			jarFile.Close()
			os.Remove(jarFile.Name())
			*errPtr = err
			continue
		}
		jarFile.Close()
		os.Remove(jarFile.Name())
		*vers = append(*vers, GenericVersion{ServerVersion: latestBuild.Version, DownloadURL: downloadUrl, JVM: jvm})
	}
}

// Generic fetch to Paper Project
func (vers *Versions) fetchPaperProject(ProjectTarget string) (err error) {
	if !slices.Contains(PaperProjects, ProjectTarget) {
		return fmt.Errorf("invalid paper project name: %s", ProjectTarget)
	}

	var projectVersions struct {
		Versions []string `json:"versions"`
	}
	if _, err := request.DoJSON(fmt.Sprintf(paperProjectURL, ProjectTarget), &projectVersions, nil); err != nil {
		return err
	}

	// Clean versions slice
	*vers = (*vers)[:0]
	var wg sync.WaitGroup
	jobs := make(chan paperBuilds)
	for range runtime.NumCPU() * 2 {
		wg.Add(1)
		go paperWorkder(vers, ProjectTarget, jobs, &err, &wg)
	}

	for _, version := range projectVersions.Versions {
		var builds struct {
			Builds []paperBuilds `json:"builds"`
		}
		if _, err := request.DoJSON(fmt.Sprintf(paperProjectBuildsURL, ProjectTarget, version), &builds, nil); err != nil {
			close(jobs)
			return err
		}

		latestBuild := builds.Builds[len(builds.Builds)-1]
		latestBuild.Version = version
		jobs <- latestBuild
	}

	close(jobs) // Done jobs
	wg.Wait()   // Wait workers finish
	semver.Sort(*vers)
	return
}
