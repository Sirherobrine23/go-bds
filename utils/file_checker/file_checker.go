package file_checker

import "os"

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
