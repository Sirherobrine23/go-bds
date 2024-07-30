package pocketmine

import (
	"time"

	"sirherobrine23.org/go-bds/go-bds/exec"
)

type Pocketmine struct {
	ServerPath string    `json:"serverPath"`    // Server path to download, run server
	Version    string    `json:"serverVersion"` // Server version
	Started    time.Time `json:"startedTime"`   // Server started date
	Ports      []int     `json:"ports"`         // Server ports
}

func (w *Pocketmine) Download() error {
	return nil
}

func (server *Pocketmine) Start() (*exec.Server, error) {
	opts := exec.ServerOptions{
		Arguments: []string{
			"php",
			"pocketmine.php",
			"--no-wizard",
			"--enable-ansi",
		},
	}
	return opts.Run()
}
