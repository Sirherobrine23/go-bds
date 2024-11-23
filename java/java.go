package java

import (
	"sirherobrine23.com.br/go-bds/go-bds/exec"
	"sirherobrine23.com.br/go-bds/go-bds/overlayfs"
)

type Java struct {
	Version string // Version to run server
	Variant string // Server to run, example: mojang, spigot, paper, etc...

	VersionsPath string // Folder path to save extracted Minecraft versions
	JVMPath      string // Folder to storage java versions
	WorkdirPath  string // Path Workdir to Overlayfs
	UpperPath    string // Path to save diff changes files, only platforms required's and same filesystem to 'Path'
	Path         string // Server folder to run Minecraft server

	ListVersions  ListServer           // function to list server versions
	ServerProc    exec.Proc            // Server process
	OverlayConfig *overlayfs.Overlayfs // Config to overlayfs, go-bds replace necesarys configs
}
