package pmmp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/utils/semver"
	"sirherobrine23.com.br/go-bds/request/github"
	"sirherobrine23.com.br/go-bds/request/v2"
)

// Pocketmine Release info
type Version struct {
	Version   string    `json:"version"`      // Pocketmine version
	MCVersion string    `json:"mc_version"`   // Minecraft Bedrock version
	Release   time.Time `json:"release_date"` // Pocketmine release
	Phar      string    `json:"phar"`         // Pocketmine url download
	PHP       *PHP      `json:"php_info"`     // PHP info
}

// Return semver version
func (ver Version) SemverVersion() semver.Version { return semver.New(ver.Version) }

func (ver Version) Download(path string) error {
	if ver.Phar != "" {
		_, err := request.SaveAs(ver.Phar, path, nil)
		return err
	}
	return ErrNoVersion
}

// All Pocketmine Releases
type Versions []*Version

// List all releases and PHP prebuilds from Github Releases to
// PocketMine/PocketMine-MP and pmmp/PocketMine-MP
func (versions *Versions) GetVersionsFromGithub() error {
	client := github.NewClient("pmmp", "PocketMine-MP", "")
	*versions = (*versions)[:0]
	for PMMPRelease, err := range client.ReleaseSeq() {
		if err != nil {
			return err
		}

		// Start new struct
		newVersion := &Version{
			Version: PMMPRelease.TagName,
			Release: PMMPRelease.PublishedAt,
			PHP: &PHP{
				Tools:         map[string][]*PHPSource{},
				Downloads:     map[string]string{},
			},
		}

		for _, asset := range PMMPRelease.Assets {
			switch path.Ext(asset.Name) {
			case ".phar":
				newVersion.Phar = asset.BrowserDownloadURL
				newVersion.Release = asset.UpdatedAt
			case ".json":
				buildInfo, _, err := request.JSON[map[string]json.RawMessage](asset.BrowserDownloadURL, nil)
				if err != nil {
					return err
				}

				// Minecraft Bedrock Version
				if mcpeVersion, ok := buildInfo["mcpe_version"]; ok {
					newVersion.MCVersion = string(mcpeVersion[1 : len(mcpeVersion)-2])
				}

				// PHP Version
				if phpVersion, ok := buildInfo["php_version"]; ok {
					newVersion.PHP = &PHP{
						PHPVersion: string(phpVersion[1 : len(phpVersion)-2]),
					}
				}

				// Prebuild PHP files
				if phpDownloadVersion, ok := buildInfo["php_download_url"]; ok {
					phpPrebuildUrl, err := url.Parse(string(phpDownloadVersion[1 : len(phpDownloadVersion)-2]))
					if err == nil {
						tagRelease := phpPrebuildUrl.Path[strings.LastIndex(phpPrebuildUrl.Path, "/"):]
						newVersion.PHP.UnixScript = fmt.Sprintf("https://raw.githubusercontent.com/pmmp/PHP-Binaries/%s/compile.sh", tagRelease)

						// Historically Windows has had several scripts to build PHP
						newVersion.PHP.WinScript = fmt.Sprintf("https://raw.githubusercontent.com/pmmp/PHP-Binaries/%s/windows-compile-vs.ps1", tagRelease)
						newVersion.PHP.WinScript = fmt.Sprintf("https://raw.githubusercontent.com/pmmp/PHP-Binaries/%s/windows-compile-vs.bat", tagRelease)
						newVersion.PHP.WinOldPs = fmt.Sprintf("https://raw.githubusercontent.com/pmmp/PHP-Binaries/%s/windows-binaries.ps1", tagRelease)
						newVersion.PHP.WinSh = fmt.Sprintf("https://raw.githubusercontent.com/pmmp/PHP-Binaries/%s/windows-binaries.sh", tagRelease)

						// Create new client to PHP binaries
						client := github.NewClient("pmmp", "PHP-Binaries", "")

						// Append to PHP Struct prebuild files
						if phpRelease, err := client.ReleaseTag(tagRelease); err == nil {
							for _, phpAsset := range phpRelease.Assets {
								name := strings.ToLower(phpAsset.Name)
								switch {
								case strings.Contains(name, "debug"):
									continue
								case (strings.Contains(name, "darwin") || strings.Contains(phpAsset.Name, "macos")) && strings.Contains(name, "arm64"):
									newVersion.PHP.Downloads["darwin/arm64"] = phpAsset.BrowserDownloadURL
								case (strings.Contains(name, "darwin") || strings.Contains(phpAsset.Name, "macos")):
									newVersion.PHP.Downloads["darwin/amd64"] = phpAsset.BrowserDownloadURL
								case strings.Contains(name, "android") && strings.Contains(name, "arm64"):
									newVersion.PHP.Downloads["android/arm64"] = phpAsset.BrowserDownloadURL
								case strings.Contains(name, "android"):
									newVersion.PHP.Downloads["android/amd64"] = phpAsset.BrowserDownloadURL
								case strings.Contains(name, "windows") && strings.Contains(name, "arm64"):
									newVersion.PHP.Downloads["windows/arm64"] = phpAsset.BrowserDownloadURL
								case strings.Contains(name, "windows"):
									newVersion.PHP.Downloads["windows/amd64"] = phpAsset.BrowserDownloadURL
								case strings.Contains(name, "linux") && strings.Contains(name, "arm64"):
									newVersion.PHP.Downloads["linux/arm64"] = phpAsset.BrowserDownloadURL
								case strings.Contains(name, "linux") && strings.Contains(name, "arm"):
									newVersion.PHP.Downloads["linux/arm"] = phpAsset.BrowserDownloadURL
								case strings.Contains(name, "linux"):
									newVersion.PHP.Downloads["linux/amd64"] = phpAsset.BrowserDownloadURL
								}
							}
						}
					}
				}
			}
		}

		// Skip append to versions
		if newVersion.Phar == "" {
			continue
		}
		*versions = append(*versions, newVersion)
	}

	// Fetch from old Release
	client.Username = "PocketMine"
	for PocketmineRelease, err := range client.ReleaseSeq() {
		if err != nil {
			return fmt.Errorf("cannot get release to PocketMine/PocketMine-MP: %s", err)
		}

		// Make new struct to old release
		newVersion := &Version{
			Version: PocketmineRelease.TagName,
			Release: PocketmineRelease.PublishedAt,
		}

		for _, asset := range PocketmineRelease.Assets {
			switch path.Ext(asset.Name) {
			case ".phar":
				newVersion.Phar = asset.BrowserDownloadURL
				newVersion.Release = asset.UpdatedAt
			}
		}

		// Skip append to versions
		if newVersion.Phar == "" {
			continue
		}
		*versions = append(*versions, newVersion)
	}

	semver.Sort(*versions)

	return nil
}
