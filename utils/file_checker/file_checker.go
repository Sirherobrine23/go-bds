package file_checker

import (
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"strings"
)

// Check if target path is file
func IsFile(path string) bool {
	if s, err := os.Stat(path); err == nil && !s.IsDir() {
		return true
	}
	return false
}

// Check if target path is dir
func IsDir(path string) bool {
	if s, err := os.Stat(path); err == nil && s.IsDir() {
		return true
	}
	return false
}

// Check if folder contains file
func FolderIsEmpty(path string) bool {
	fileEntrys, _ := os.ReadDir(path)
	return len(fileEntrys) > 0
}

func FindFile(folder, file string) (string, error) {
	for path, d := range walkSeq(folder) {
		if !d.IsDir() && strings.HasSuffix(filepath.ToSlash(path), filepath.ToSlash(file)) {
			return path, nil
		}
	}
	return "", fs.ErrNotExist
}

func walkSeq(folder string) iter.Seq2[string, fs.DirEntry] {
	return func(yield func(string, fs.DirEntry) bool) {
		filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
			if err == nil {
				if !yield(path, d) {
					return filepath.SkipAll
				}
			}
			return err
		})
	}
}
