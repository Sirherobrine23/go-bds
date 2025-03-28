package allaymc

import (
	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/overlayfs"
)

type AllayMC struct {
	PID         exec.Proc            // process status
	serverStart exec.ProcExec        // Server command
	Version     *Version             // Server version
	Overlayfs   *overlayfs.Overlayfs // Overlayfs mounted
}
