package pmmp

import "time"

// Build info
type BuildInfo struct {
	PHPVersion   string `json:"php_version"`
	GitCommit    string `json:"git_commit"`
	McpeVersion  string `json:"mcpe_version"`
	PharDownload string `json:"download_url"`
}

// Pocketmine Release info
type Version struct {
	Version       string               `json:"version"`        // Pocketmine version
	MCVersion     string               `json:"mc_version"`     // Minecraft Bedrock version
	Release       time.Time            `json:"release_date"`   // Pocketmine release
	Phar          string               `json:"phar"`           // Pocketmine url download
	PHPVersion    string               `json:"php"`            // PHP versions
	PHPExtensions map[string][2]string `json:"php_extensions"` // Extensions in PHP: name => [src, version]
	Tools         map[string][2]string `json:"tools"`          // Tools to install/build: name => [src, version]
}
