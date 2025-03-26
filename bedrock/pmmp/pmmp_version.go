package pmmp

import (
	"fmt"
	"io"
	"os/exec"
	"path"
	"runtime"
	"time"

	"sirherobrine23.com.br/go-bds/request/v2"
)

// Build info
type BuildInfo struct {
	PHPVersion   string `json:"php_version"`
	GitCommit    string `json:"git_commit"`
	McpeVersion  string `json:"mcpe_version"`
	PharDownload string `json:"download_url"`
}

// Pocketmine Release info
type Version struct {
	Version   string    `json:"version"`      // Pocketmine version
	MCVersion string    `json:"mc_version"`   // Minecraft Bedrock version
	Release   time.Time `json:"release_date"` // Pocketmine release
	Phar      string    `json:"phar"`         // Pocketmine url download
	PHP       PHP       `json:"php_info"`     // PHP info
}

func (ver Version) Download(path string) error {
	if ver.Phar != "" {
		_, err := request.SaveAs(ver.Phar, path, nil)
		return err
	}
	return ErrNoVersion
}

// Pocketmine PHP required and tools to build
type PHP struct {
	PHPVersion    string               `json:"php"`            // PHP versions
	WinScript     string               `json:"win_script"`     // Windows script
	UnixScript    string               `json:"unix_script"`    // Unix script to build
	PHPExtensions map[string][2]string `json:"php_extensions"` // Extensions in PHP: name => [src, version]
	Tools         map[string][2]string `json:"tools"`          // Tools to install/build: name => [src, version]
	Downloads     map[string]string    `json:"downloads"`      // Prebuilds php files
}

// Install prebuild binary's
func (php PHP) Install(installPath string) error {
	if urlDownload, ok := php.Downloads[fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)]; ok {
		switch path.Ext(urlDownload) {
		case ".zip":
			return request.Zip(urlDownload, request.ExtractOptions{Cwd: installPath}, nil)
		default:
			return request.Tar(urlDownload, request.ExtractOptions{Cwd: installPath}, nil)
		}
	}
	return fmt.Errorf("prebuild to %s/%s not exists", runtime.GOOS, runtime.GOARCH)
}

// Build php localy
func (php PHP) Build(logWrite io.Writer) error {
	switch runtime.GOOS {
	case "aix", "plan9", "solaris", "js", "illumos", "ios", "dragonfly":
		return ErrPlatform
	case "windows":
		cmd := exec.Command("powershell", fmt.Sprintf("irm %q | iex", php.WinScript))
		cmd.Stderr = logWrite
		cmd.Stdout = logWrite
		return cmd.Run()
	default:
		res, err := request.Request(php.UnixScript, nil)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		cmd := exec.Command("sh", "-c", "-")
		cmd.Stdin = res.Body
		cmd.Stderr = logWrite
		cmd.Stdout = logWrite
		return cmd.Run()
	}
}
