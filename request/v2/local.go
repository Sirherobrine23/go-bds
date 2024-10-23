package request

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Save file to disk
func SaveAs(Url, OnSave string, Option *Options) (*http.Response, error) {
	res, err := Request(Url, Option)
	if err != nil {
		return res, err
	}
	defer res.Body.Close()

	// Create folder if not exists
	if _, err := os.Stat(filepath.Dir(OnSave)); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(OnSave), 0600)
	}

	file, err := os.Create(OnSave)
	if err != nil {
		return res, err
	}
	defer file.Close()
	_, err = io.Copy(file, res.Body)
	return res, err
}
