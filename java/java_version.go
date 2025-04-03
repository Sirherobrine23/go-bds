package java

import (
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/utils/javaprebuild"
	"sirherobrine23.com.br/go-bds/go-bds/utils/semver"
	"sirherobrine23.com.br/go-bds/request/v2"
)

// Version info to server
type Version interface {
	Version() string                       // Server version string
	Install(folder string) error           // Install server to folder path
	JavaVersion() javaprebuild.JavaVersion // Return java version
}

type GenericVersion struct {
	ServerVersion string
	JVM           javaprebuild.JavaVersion
	DownloadURL   string
}

func (v GenericVersion) Version() string                       { return v.ServerVersion }
func (v GenericVersion) JavaVersion() javaprebuild.JavaVersion { return v.JVM }
func (v GenericVersion) Install(folder string) error {
	_, err := request.SaveAs(v.DownloadURL, filepath.Join(folder, "server.jar"), nil)
	return err
}

// Versions is a list of Version
type Versions []Version

func pistonInfoWorker(versions *Versions, job <-chan *mojangVersion, errPtr *error, wg *sync.WaitGroup) {
	defer wg.Done()
	for data := range job {
		if semver.New(data.ID) == nil {
			continue
		}

		_, err := request.DoJSON(data.VersionURL, &data.ReleaseInfo, nil)
		if err != nil {
			*errPtr = err
			continue
		}
		if serverURL, ok := data.ReleaseInfo.Downloads["server"]; ok {
			*versions = append(*versions, GenericVersion{
				ServerVersion: data.ID,
				DownloadURL:   serverURL.URL,
				JVM:           javaprebuild.JavaVersion(data.ReleaseInfo.JavaVersion.MajorVersion + 44),
			})
		}
	}
}

// Fetch servers from Mojang servers
func (versions *Versions) FetchMojang() (err error) {
	data, _, err := request.JSON[mojangPistonVersion]("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json", nil)
	if err != nil {
		return err
	}

	// Clean slice
	*versions = (*versions)[:0]
	var wg sync.WaitGroup
	jobs := make(chan *mojangVersion)
	for range runtime.NumCPU() * 2 {
		wg.Add(1)
		go pistonInfoWorker(versions, jobs, &err, &wg)
	}

	// Send job to workers
	for _, versions := range data.Versions {
		jobs <- versions
	}

	// Done sending jobs
	close(jobs)
	wg.Wait() // Wait for all workers to finish
	semver.Sort(*versions)
	return
}

type mojangVersion struct {
	ID          string    `json:"id"`          // Server version
	Type        string    `json:"type"`        // Release type
	VersionURL  string    `json:"url"`         // ReleaseInfo URL
	ReleaseTime time.Time `json:"releaseTime"` // Release date
	UploadTime  time.Time `json:"time"`        // Upload date

	ReleaseInfo *struct {
		ID                     string `json:"id"`
		MainClass              string `json:"mainClass"`
		MinimumLauncherVersion int    `json:"minimumLauncherVersion"`
		Type                   string `json:"type"`
		Assets                 string `json:"assets"`
		Downloads              map[string]struct {
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			URL  string `json:"url"`
		} `json:"downloads"`
		JavaVersion struct {
			Component    string `json:"component"`
			MajorVersion int    `json:"majorVersion"`
		} `json:"javaVersion"`
	}
}

type mojangPistonVersion struct {
	Latest   map[string]string `json:"latest"`
	Versions []*mojangVersion  `json:"versions"`
}
