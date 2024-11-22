package mclog

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

var (
	MclogsApi  string = "https://api.mclo.gs"
	MclogsBase string = "https://mclo.gs"

	ErrNoId     error = errors.New("require mclo.gs id") // Require uploaded log to view
	ErrNoExists error = errors.New("log no exists")      // id request not exists
)

type Insights struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Analysis map[string][]struct {
		Label   string `json:"label"`
		Value   string `json:"value"`
		Message string `json:"message"`
		Counter int64  `json:"counter"`
		Entry   struct {
			Level     int64     `json:"level"`
			Prefix    string    `json:"prefix"`
			EntryTime time.Time `json:"time"`
			Lines     []struct {
				Numbers int64  `json:"number"`
				Content string `json:"content"`
			} `json:"lines"`
		} `json:"level"`
	} `json:"analysis"`
}

type Mclog struct {
	MclogApi  string // URL API to mclo.gs, default: "https://api.mclo.gs"
	MclogBase string // URL Base to mclo.gs, default: "https://mclo.gs"
	FileID    string // LOG file ID if success upload
}

// Upload server log
func (Log *Mclog) Upload(logStream io.Reader) error {
	if len(Log.MclogApi) == 0 {
		Log.MclogApi = MclogsApi
	}

	type JSONResponseStatus struct {
		Id           string `json:"id"`
		Success      bool   `json:"success"`
		ErrorMessage string `json:"error"`
	}

	UploadStatus, res, err := request.JSON[JSONResponseStatus](fmt.Sprintf("%s/1/log", Log.MclogApi), &request.Options{Method: "POST", Body: logStream})
	if err != nil {
		return err
	} else if !UploadStatus.Success && len(UploadStatus.ErrorMessage) > 0 {
		return errors.New(UploadStatus.ErrorMessage)
	} else if !UploadStatus.Success {
		return fmt.Errorf("cannot upload file, http status code %d, message: %q", res.StatusCode, res.Status)
	}

	// Copy id to struct
	Log.FileID = UploadStatus.Id

	return nil
}

// Return log insights from API
func (Log *Mclog) Insights() (*Insights, error) {
	if len(Log.FileID) == 0 {
		return nil, ErrNoId
	} else if len(Log.MclogApi) == 0 {
		Log.MclogApi = MclogsApi
	} else if _, err := url.Parse(Log.MclogApi); err != nil {
		return nil, err
	}

	logInsight, res, err := request.JSON[Insights](fmt.Sprintf("%s/1/insights/%s", Log.MclogApi, Log.FileID), nil)
	if err != nil {
		return nil, err
	} else if res.StatusCode == 404 {
		return nil, ErrNoExists
	}
	return &logInsight, nil
}

// Return raw minecraft log
func (Log *Mclog) Raw() (io.ReadCloser, error) {
	// Check to Valid url API
	if len(Log.MclogApi) == 0 {
		Log.MclogApi = MclogsApi
	} else if len(Log.FileID) == 0 {
		return nil, ErrNoId
	}

	res, err := request.Request(fmt.Sprintf("%s/1/raw/%s", Log.MclogApi, Log.FileID), nil)
	if err != nil {
		return nil, err
	} else if res.StatusCode == 404 {
		return nil, ErrNoExists
	}

	return res.Body, nil
}
