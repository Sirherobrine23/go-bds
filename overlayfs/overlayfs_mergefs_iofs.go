package overlayfs

import (
	"io/fs"
	"path"
	"path/filepath"
)

var (
	_ fs.FS         = &FsMergeFs{}
	_ fs.ReadFileFS = &FsMergeFs{}
	_ fs.ReadDirFS  = &FsMergeFs{}
	_ fs.StatFS     = &FsMergeFs{}
	_ fs.GlobFS     = &FsMergeFs{}
	_ fs.SubFS      = &FsMergeFs{}
)

// Implementes [io/fs.FS] to Overlayfs
type FsMergeFs struct {
	MergedFS *Overlayfs // Overlayfs
	Subdir   string     // Sub dir to find files
}

// Return [io/fs.FS] from Overlayfs
func (fs *Overlayfs) Mergefs() fs.FS {
	return &FsMergeFs{
		MergedFS: fs,
		Subdir:   "",
	}
}

// Sub returns an [FS] corresponding to the subtree rooted at fsys's dir.
//
// If dir is ".", Sub returns fsys unchanged.
// Otherwise, if fs implements [SubFS], Sub returns fsys.Sub(dir).
// Otherwise, Sub returns a new [FS] implementation sub that,
// in effect, implements sub.Open(name) as fsys.Open(path.Join(dir, name)).
// The implementation also translates calls to ReadDir, ReadFile, and Glob appropriately.
//
// Note that Sub(os.DirFS("/"), "prefix") is equivalent to os.DirFS("/prefix")
// and that neither of them guarantees to avoid operating system
// accesses outside "/prefix", because the implementation of [os.DirFS]
// does not check for symbolic links inside "/prefix" that point to
// other directories. That is, [os.DirFS] is not a general substitute for a
// chroot-style security mechanism, and Sub does not change that fact.
func (fss FsMergeFs) Sub(dir string) (fs.FS, error) {
	return &FsMergeFs{MergedFS: fss.MergedFS, Subdir: filepath.Join(fss.Subdir, filepath.Clean(dir))}, nil
}

// Open opens the named file.
//
// When Open returns an error, it should be of type *PathError
// with the Op field set to "open", the Path field set to name,
// and the Err field describing the problem.
//
// Open should reject attempts to open names that do not satisfy
// ValidPath(name), returning a *PathError with Err set to
// ErrInvalid or ErrNotExist.
func (fss FsMergeFs) Open(name string) (fs.File, error) {
	if fss.MergedFS == nil {
		return nil, fs.ErrInvalid
	}
	return fss.MergedFS.Open(filepath.Join(fss.Subdir, filepath.Clean(name)))
}

// ReadDir reads the named directory
// and returns a list of directory entries sorted by filename.
func (fss FsMergeFs) ReadDir(name string) ([]fs.DirEntry, error) {
	if fss.MergedFS == nil {
		return nil, fs.ErrInvalid
	}
	return fss.MergedFS.ReadDir(filepath.Join(fss.Subdir, filepath.Clean(name)))
}

// Stat returns a FileInfo describing the file.
// If there is an error, it should be of type *PathError.
func (fss FsMergeFs) Stat(name string) (fs.FileInfo, error) {
	if fss.MergedFS == nil {
		return nil, fs.ErrInvalid
	}
	return fss.MergedFS.Stat(filepath.Join(fss.Subdir, filepath.Clean(name)))
}

// ReadFile reads the named file and returns its contents.
// A successful call returns a nil error, not io.EOF.
// (Because ReadFile reads the whole file, the expected EOF
// from the final Read is not treated as an error to be reported.)
//
// The caller is permitted to modify the returned byte slice.
// This method should return a copy of the underlying data.
func (fss FsMergeFs) ReadFile(name string) ([]byte, error) {
	if fss.MergedFS == nil {
		return nil, fs.ErrInvalid
	}
	return fss.MergedFS.ReadFile(filepath.Join(fss.Subdir, filepath.Clean(name)))
}

// Glob returns the names of all files matching pattern,
// providing an implementation of the top-level
// Glob function.
func (fss FsMergeFs) Glob(pattern string) ([]string, error) {
	if fss.MergedFS == nil {
		return nil, fs.ErrInvalid
	}
	return globWithLimit(fss, pattern, 0)
}

func globWithLimit(fsys fs.FS, pattern string, depth int) (matches []string, err error) {
	// This limit is added to prevent stack exhaustion issues. See
	// CVE-2022-30630.
	const pathSeparatorsLimit = 10000
	if depth > pathSeparatorsLimit {
		return nil, path.ErrBadPattern
	}

	// Check pattern is well-formed.
	if _, err := path.Match(pattern, ""); err != nil {
		return nil, err
	}
	if !hasMeta(pattern) {
		if _, err = fs.Stat(fsys, pattern); err != nil {
			return nil, nil
		}
		return []string{pattern}, nil
	}

	dir, file := path.Split(pattern)
	dir = cleanGlobPath(dir)

	if !hasMeta(dir) {
		return glob(fsys, dir, file, nil)
	}

	// Prevent infinite recursion. See issue 15879.
	if dir == pattern {
		return nil, path.ErrBadPattern
	}

	var m []string
	m, err = globWithLimit(fsys, dir, depth+1)
	if err != nil {
		return nil, err
	}
	for _, d := range m {
		matches, err = glob(fsys, d, file, matches)
		if err != nil {
			return
		}
	}
	return
}

// cleanGlobPath prepares path for glob matching.
func cleanGlobPath(path string) string {
	switch path {
	case "":
		return "."
	default:
		return path[0 : len(path)-1] // chop off trailing separator
	}
}

// glob searches for files matching pattern in the directory dir
// and appends them to matches, returning the updated slice.
// If the directory cannot be opened, glob returns the existing matches.
// New matches are added in lexicographical order.
func glob(fss fs.FS, dir, pattern string, matches []string) (m []string, e error) {
	m = matches
	infos, err := fs.ReadDir(fss, dir)
	if err != nil {
		return // ignore I/O error
	}

	for _, info := range infos {
		n := info.Name()
		matched, err := path.Match(pattern, n)
		if err != nil {
			return m, err
		}
		if matched {
			m = append(m, path.Join(dir, n))
		}
	}
	return
}

// hasMeta reports whether path contains any of the magic characters
// recognized by path.Match.
func hasMeta(path string) bool {
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '*', '?', '[', '\\':
			return true
		}
	}
	return false
}
