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
		if err = os.MkdirAll(filepath.Dir(OnSave), 0600); err != nil {
			return res, err
		}
	}

	file, err := os.Create(OnSave)
	if err != nil {
		return res, err
	}
	defer file.Close()
	_, err = io.Copy(file, res.Body)
	return res, err
}

// Create request and save request file to local file in system temporary directory
func SaveTmp(Url string, Option *Options) (*os.File, *http.Response, error) {
	res, err := Request(Url, Option)
	if err != nil {
		return nil, res, err
	}
	defer res.Body.Close()

	// Create temp file in system temporary files directory
	local, err := os.CreateTemp(os.TempDir(), "gobds*requestfile")
	if err != nil {
		return nil, nil, err
	}

	// Copy body
	if _, err := io.Copy(local, res.Body); err != nil {
		return local, res, err
	} else if _, err = local.Seek(0, 0); err != nil {
		return local, res, err
	}

	return local, res, nil
}