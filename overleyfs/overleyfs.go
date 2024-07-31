// Mount overlayfs in compatible system
package overleyfs

type MountUnmount interface {
	Mount() error   // Mount volumes in target
	UnMount() error // Unmount target
}

type Overlayfs struct {
	Target  string   // Folder with merged another folder
	Workdir string   // Folder to write temporary files
	Upper   string   // Folder to write modifications, blank to read-only
	Lower   []string // Folders layers, read-only
}
