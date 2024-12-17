package mclog

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Storage interface {
	fs.FS
	Create(name string) (io.WriteCloser, error) // Save file in storage
	Remove(name string) error                   // Remove file from storage
}

func writeJSON(w io.Writer, obj any) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(obj)
}

func writeErr(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	if errors.Is(err, fs.ErrNotExist) {
		w.WriteHeader(404)
		writeJSON(w, MclogResponseStatus{Success: false, ErrorMessage: ErrNoExists.Error()})
		return
	}
	w.WriteHeader(500)
	writeJSON(w, MclogResponseStatus{Success: false, ErrorMessage: err.Error()})
}

func NewHandler(serverLimits Limits, storage Storage) *http.ServeMux {
	control := http.NewServeMux()

	// Return server limits to process logs
	control.HandleFunc("GET /1/limits", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		writeJSON(w, Limits{
			StorageTime: serverLimits.StorageTime / time.Second, // return only seconds
			MaxLength:   serverLimits.MaxLength,
			MaxLines:    serverLimits.MaxLines,
		})
	})

	// Get log stream
	control.HandleFunc("GET /1/raw/{id}", func(w http.ResponseWriter, r *http.Request) {
		logId := r.PathValue("id")
		body, err := storage.Open(logId)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				w.WriteHeader(404)
				writeJSON(w, MclogResponseStatus{Success: false, ErrorMessage: ErrNoExists.Error()})
				return
			}
			w.WriteHeader(500)
			writeJSON(w, MclogResponseStatus{Success: false, ErrorMessage: err.Error()})
			return
		}
		defer body.Close()

		w.Header().Set("Content-Type", "text/plain") // Set file stream
		if stat, err := body.Stat(); err == nil {
			w.Header().Set("content-length", strconv.FormatInt(stat.Size(), 10))
		}

		// Response
		w.WriteHeader(200)
		go io.Copy(w, body)
	})

	control.HandleFunc("GET /1/insights/{id}", func(w http.ResponseWriter, r *http.Request) {
		logId := r.PathValue("id") // get log id

		// Open log file
		logBody, err := storage.Open(logId)
		if err != nil {
			writeErr(w, err)
			return
		}
		defer logBody.Close()

		// Parse log body
		var logInsight Insights
		if err := logInsight.ParseStream(logBody); err != nil {
			writeErr(w, err)
			return
		}

		// return log insight
		w.WriteHeader(200)
		writeJSON(w, logInsight)
	})

	control.HandleFunc("POST /1/log", func(w http.ResponseWriter, r *http.Request) {
		// Close body after run function
		defer r.Body.Close()

		// Set only reader type
		logStream := io.Reader(r.Body)
		if serverLimits.MaxLength > 0 {
			logStream = io.LimitReader(r.Body, serverLimits.MaxLength) // Set max body lenght from limits
		}

		switch {
		case slices.Contains(r.Header.Values("Content-Type"), "application/octet-stream"):
			// noop, only reader body from logStream
		case slices.Contains(r.Header.Values("Content-Type"), "application/x-www-form-urlencoded"):
			// Read all body
			body, err := io.ReadAll(logStream)
			if err != nil {
				writeErr(w, err)
				return
			}

			// Parse body
			form, err := url.ParseQuery(string(body))
			if err != nil {
				writeErr(w, err)
				return
			}

			// Check if body have "content"
			if !form.Has("content") {
				w.WriteHeader(400)
				writeJSON(w, MclogResponseStatus{Success: false, ErrorMessage: "require the 'content' in body"})
				return
			}

			// Attemp write log file
			logStream = strings.NewReader(form.Get("content"))
		default:
			w.WriteHeader(400)
			writeJSON(w, MclogResponseStatus{Success: false, ErrorMessage: "Require 'application/x-www-form-urlencoded' or raw stream/'application/octet-stream'"})
			return
		}

		var randomId string
		for randomId == "" {
			buff := make([]byte, 8)
			if _, err := rand.Read(buff); err != nil {
				writeErr(w, err)
				return
			}
			randomId = hex.EncodeToString(buff)
			if f, err := storage.Open(randomId); err == nil {
				f.Close() // Close file
				randomId = ""
			}
		}

		// Attemp write log file
		file, err := storage.Create(randomId)
		if err != nil {
			writeErr(w, err)
			return
		}
		defer file.Close()
		if _, err = io.Copy(file, logStream); err != nil {
			writeErr(w, err)
			return
		}

		// Write success log id
		w.WriteHeader(201)
		writeJSON(w, MclogResponseStatus{Success: true, Id: randomId})
	})

	return control
}
