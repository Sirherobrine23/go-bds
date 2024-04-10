package mclog

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

const (
	FileLimitSize int64  = 10000000000 // Max allowed to Upload to mclo.gs
	MclogsApi     string = "https://api.mclo.gs"
	MclogsBase    string = "https://mclo.gs"
)

var (
	ErrNoId           error = errors.New("require mclo.gs id") // Require uploaded log to view
	ErrNoExists       error = errors.New("log no exists")      // id request not exists
	ErrDisabledUpload error = errors.New("upload log is disabled, enable fist")
)

type Mclog struct {
	Enable      bool   // Enable log upload
	MclogApi    string // URL API to mclo.gs, default: "https://api.mclo.gs"
	MclogBase   string // URL Base to mclo.gs, default: "https://mclo.gs"
	FileLogPath string // LOG file path location
	FileID      string // LOG file ID if success upload
}

func (Log *Mclog) Upload() error {
	if !Log.Enable {
		return ErrDisabledUpload
	} else if len(Log.MclogApi) == 0 {
		Log.MclogApi = MclogsApi
	} else if _, err := url.Parse(Log.MclogApi); err != nil {
		return err
	}

	// Open log file
	logFile, err := os.Open(Log.FileLogPath)
	if err != nil {
		return err
	}

	res, err := request.Request(request.RequestOptions{
		Url:    fmt.Sprintf("%s/1/log", Log.MclogApi),
		Method: "POST",
		Body:   logFile,
	})

	if err != nil {
		return err
	}

	var UploadStatus struct {
		Id           string `json:"id"`
		Success      bool   `json:"success"`
		ErrorMessage string `json:"error"`
	}

	defer res.Body.Close()
	if err = json.NewDecoder(res.Body).Decode(&UploadStatus); err != nil {
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

func (Log *Mclog) Insights() (Insights, error) {
	var logInsight Insights
	// Check to Valid url API
	if len(Log.MclogApi) == 0 {
		Log.MclogApi = MclogsApi
	} else if _, err := url.Parse(Log.MclogApi); err != nil {
		return logInsight, err
	} else if len(Log.FileID) == 0 {
		return logInsight, ErrNoId
	}

	res, err := request.Request(request.RequestOptions{
		Url:    fmt.Sprintf("%s/1/insights/%s", Log.MclogApi, Log.FileID),
		Method: "GET",
	})

	if err != nil {
		return logInsight, err
	}

	defer res.Body.Close()
	if err = json.NewDecoder(res.Body).Decode(&logInsight); err != nil {
		return logInsight, err
	} else if res.StatusCode == 404 {
		return logInsight, ErrNoExists
	}

	return logInsight, nil
}

func (Log *Mclog) Raw() (io.Reader, error) {
	// Check to Valid url API
	if len(Log.MclogApi) == 0 {
		Log.MclogApi = MclogsApi
	} else if _, err := url.Parse(Log.MclogApi); err != nil {
		return nil, err
	} else if len(Log.FileID) == 0 {
		return nil, ErrNoId
	}

	res, err := request.Request(request.RequestOptions{
		Url:    fmt.Sprintf("%s/1/raw/%s", Log.MclogApi, Log.FileID),
		Method: "GET",
	})

	if err != nil {
		return nil, err
	} else if res.StatusCode == 404 {
		return nil, ErrNoExists
	}

	return res.Body, nil
}
