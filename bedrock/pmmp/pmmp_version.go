package pmmp

import (
	"encoding/json"
	"iter"
	"path"
	"slices"
	"strings"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/utils/semver"
	"sirherobrine23.com.br/go-bds/request/github"
	"sirherobrine23.com.br/go-bds/request/v2"
)

// Pocketmine Release info
type Version struct {
	Version   string    `json:"version"`              // Pocketmine version
	MCVersion string    `json:"mc_version,omitempty"` // Minecraft Bedrock version
	Release   time.Time `json:"release_date"`         // Pocketmine release
	Phar      string    `json:"phar"`                 // Pocketmine url download
	PHP       *PHP      `json:"php_info,omitempty"`   // PHP info
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
func (versions *Versions) GetVersionsFromGithub(phpBuilds PHPs) error {
	// Clean versions
	*versions = (*versions)[:0]

	// Pocketmine repo
	client := github.NewClient("pmmp", "PocketMine-MP", "")
	owners := []string{"PocketMine", "pmmp"}
	for _, owner := range owners {
		client.Username = owner
		for pocketRelease, err := range client.ReleaseSeq() {
			if err != nil {
				return err
			}

			// Start new struct
			markdow := pocketRelease.Body
			newVersion := &Version{
				Version: pocketRelease.TagName,
				Release: pocketRelease.PublishedAt,
			}

			for _, asset := range pocketRelease.Assets {
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
						newVersion.MCVersion = strings.Trim(string(mcpeVersion), "\"")
					}

					// PHP Version
					if phpVersion, ok := buildInfo["php_version"]; ok {
						phpVersion := strings.Trim(string(phpVersion), "\"")
						if newVersion.PHP = findPhp(phpVersion, slices.Backward(phpBuilds)); newVersion.PHP == nil {
							newVersion.PHP = &PHP{
								PHPVersion: phpVersion,
							}
						}
					}
				}
			}

			if newVersion.MCVersion == "" && markdow != "" {
				mdLines := strings.Split(strings.ReplaceAll(markdow, "\r\n", "\n"), "\n")
				switch {
				case len(mdLines) > 0 && strings.HasPrefix(mdLines[0], "**For Minecraft: PE "):
					mdLines[0] = strings.Trim(mdLines[0], "* ")
					newVersion.MCVersion = mdLines[0][18:]
				case len(mdLines) > 0 && strings.HasPrefix(mdLines[0], "**For Minecraft: Bedrock Edition "):
					mdLines[0] = strings.Trim(mdLines[0], "* ")
					newVersion.MCVersion = mdLines[0][31:]
				}
			}

			if newVersion.Phar != "" {
				*versions = append(*versions, newVersion)
			}
		}
	}

	slices.SortFunc(*versions, func(a, b *Version) int {
		return a.Release.Compare(b.Release)
	})

	return nil
}

func findPhp(version string, seq iter.Seq2[int, *PHP]) *PHP {
	for _, php := range seq {
		if strings.HasPrefix(php.PHPVersion, version) {
			return php
		}
	}
	return nil
}
