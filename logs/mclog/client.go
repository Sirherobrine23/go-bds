package mclog

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"sirherobrine23.com.br/go-bds/request/v2"
)

// Client to mclog
type Mclog struct {
	MclogApi  string // URL API to mclo.gs, default: "https://api.mclo.gs"
	MclogBase string // URL Base to mclo.gs, default: "https://mclo.gs"
	FileID    string // LOG file ID if success upload
}

// Upload server log
func (Log *Mclog) Upload(logStream io.Reader) error {
	if Log.MclogApi == "" {
		Log.MclogApi = MclogsApi
	}

	limits, err := Log.Limits()
	if err != nil {
		return err
	}

	if limits.MaxLength != -1 {
		if limits.MaxLength > 0 {
			logStream = io.LimitReader(logStream, limits.MaxLength)
		} else {
			logStream = io.LimitReader(logStream, DefaultMaxLength)
		}
	}

	logBody, err := io.ReadAll(logStream)
	if err != nil && err != io.EOF {
		return err
	}

	data := url.Values{}
	lines := strings.Split(string(logBody), "\n")
	if limits.MaxLines > 0 && len(lines) > int(limits.MaxLines) {
		lines = lines[:limits.MaxLines]
	}
	data.Set("content", strings.Join(lines, "\n"))

	UploadStatus, res, err := request.JSON[MclogResponseStatus](fmt.Sprintf("%s/1/log", Log.MclogApi), &request.Options{
		Method: "POST",
		Header: map[string]string{"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8"},
		Body:   strings.NewReader(data.Encode()),
	})

	if err != nil {
		return err
	} else if !UploadStatus.Success && UploadStatus.ErrorMessage != "" {
		return errors.New(UploadStatus.ErrorMessage)
	} else if !UploadStatus.Success {
		return fmt.Errorf("cannot upload file, http status code %d, message: %q", res.StatusCode, res.Status)
	}

	// Copy id to struct
	Log.FileID = UploadStatus.Id

	return nil
}

// Get limits to server API
func (Log *Mclog) Limits() (*Limits, error) {
	if Log.MclogApi == "" {
		Log.MclogApi = MclogsApi
	}

	ApiLimit, _, err := request.JSON[Limits](fmt.Sprintf("%s/1/limits", Log.MclogApi), nil)
	if err != nil {
		return nil, err
	}

	// fix time from second to ns
	ApiLimit.StorageTime = ApiLimit.StorageTime * time.Second

	return &ApiLimit, nil
}

// Return log insights from API
func (Log *Mclog) Insights() (*Insights, error) {
	if Log.FileID == "" {
		return nil, ErrNoId
	} else if Log.MclogApi == "" {
		Log.MclogApi = MclogsApi
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
	if Log.MclogApi == "" {
		Log.MclogApi = MclogsApi
	} else if Log.FileID == "" {
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
