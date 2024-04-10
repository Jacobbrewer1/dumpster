package dataaccess

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type localImpl struct{}

func NewLocal() Storage {
	return &localImpl{}
}

func (s *localImpl) SaveFile(_ context.Context, filePath string, file []byte) error {
	// Start the prometheus timer.
	t := prometheus.NewTimer(StorageLatency.With(prometheus.Labels{"query": "save_file"}))
	defer t.ObserveDuration()

	// Create a new file in the working directory with all directories.
	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}

	w, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	// Write the file to the bucket.
	_, err = w.Write(file)
	if err != nil {
		return fmt.Errorf("error writing file to bucket: %w", err)
	}

	// Close the file.
	err = w.Close()
	if err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}

	return nil
}

func (s *localImpl) DownloadFile(_ context.Context, filePath string) ([]byte, error) {
	// Start the prometheus timer.
	t := prometheus.NewTimer(StorageLatency.With(prometheus.Labels{"query": "download_file"}))
	defer t.ObserveDuration()

	// Open the file.
	r, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	// Read the file.
	file, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Close the file.
	err = r.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing file: %w", err)
	}

	return file, nil
}

func (s *localImpl) DeleteFile(_ context.Context, filePath string) error {
	// Start the prometheus timer.
	t := prometheus.NewTimer(StorageLatency.With(prometheus.Labels{"query": "delete_file"}))
	defer t.ObserveDuration()

	// Delete the file.
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("error deleting file: %w", err)
	}

	return nil
}

func (s *localImpl) Purge(_ context.Context, from time.Time) (int, error) {
	// Start the prometheus timer.
	t := prometheus.NewTimer(StorageLatency.With(prometheus.Labels{"query": "purge"}))
	defer t.ObserveDuration()

	// Get a list of all the files in the working directory and delete them.
	files, err := os.ReadDir(".")
	if err != nil {
		return 0, fmt.Errorf("error reading directory: %w", err)
	}

	count := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Ignore all non-SQL files.
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		fi, err := file.Info()
		if err != nil {
			return 0, fmt.Errorf("error getting file info: %w", err)
		}

		if fi.ModTime().Before(from) {
			err := os.Remove(file.Name())
			if err != nil {
				return 0, fmt.Errorf("error deleting file: %w", err)
			}
			count++
		}
	}

	return count, nil
}
