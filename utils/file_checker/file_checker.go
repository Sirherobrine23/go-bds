package file_checker

import (
	"fmt"
	"io"
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
	return len(fileEntrys) == 0
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

func RemoveFiles(dir string, entrys []os.DirEntry) error {
	for _, entry := range entrys {
		if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func ReplaceFiles(from, to string, entrys []os.DirEntry) error {
	for _, entry := range entrys {
		to, from := filepath.Join(to, entry.Name()), filepath.Join(from, entry.Name())
		if entry.IsDir() {
			if _, err := os.Stat(to); err == nil {
				if err = os.RemoveAll(to); err != nil {
					return err
				}
			}
			if err := os.CopyFS(to, os.DirFS(from)); err != nil {
				return err
			}
		} else {
			if entry.Type().IsRegular() {
				// Openfile to read
				fromFile, err := os.Open(from)
				if err != nil {
					return fmt.Errorf("cannot open file: %s", err)
				}
				defer fromFile.Close()

				// Create new file in target
				toFile, err := os.OpenFile(to, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, entry.Type().Perm())
				if err != nil {
					return fmt.Errorf("cannot make file target: %s", err)
				}
				defer toFile.Close()

				// Copy file
				if _, err = io.Copy(toFile, fromFile); err != nil {
					return err
				}

				// Close files R/W
				fromFile.Close()
				toFile.Close()
			}
		}
	}
	return nil
}
