// Mount overlayfs in compatible system
package overleyfs

type Overlayfs struct {
	Root    string   // Folder with merged another folder
	Workdir string   // Folder to write temporary files
	Upper   string   // Folder to write modifications, blank to read-only
	Lower   []string // Folders layers, read-only
}
