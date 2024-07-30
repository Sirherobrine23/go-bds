// Mount overlayfs in compatible system
package overleyfs

type Overlayfs struct {
	Target  string   // Folder with merged another folder
	Workdir string   // Folder to write temporary files, if blank and Upper seted create in "$TMP/overlayfs_workdir_*"
	Upper   string   // Folder to write modifications, blank to read-only
	Lower   []string // Folders layers, read-only
}
