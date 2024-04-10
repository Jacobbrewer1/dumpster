package dataaccess

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/api/iterator"
)

const (
	EnvGCSCredentials = "GCS_CREDENTIALS"
)

type gcsImpl struct {
	// gcs is the Google Cloud Storage client.
	gcs *storage.Client

	// bucket is the name of the bucket to use.
	bucket string
}

func NewGCS(gcs *storage.Client, bucket string) Storage {
	return &gcsImpl{
		gcs:    gcs,
		bucket: bucket,
	}
}

func (s *gcsImpl) SaveFile(ctx context.Context, filePath string, file []byte) error {
	// Start the prometheus timer.
	t := prometheus.NewTimer(StorageLatency.With(prometheus.Labels{"query": "save_file"}))
	defer t.ObserveDuration()

	// Connect to the bucket.
	bkt := s.gcs.Bucket(s.bucket)

	// Create a new file in the bucket.
	w := bkt.Object(filePath).NewWriter(ctx)

	// Write the file to the bucket.
	_, err := w.Write(file)
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

func (s *gcsImpl) DownloadFile(ctx context.Context, filePath string) ([]byte, error) {
	// Start the prometheus timer.
	t := prometheus.NewTimer(StorageLatency.With(prometheus.Labels{"query": "download_file"}))
	defer t.ObserveDuration()

	// Connect to the bucket.
	bkt := s.gcs.Bucket(s.bucket)

	// Open the file.
	r, err := bkt.Object(filePath).NewReader(ctx)
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

func (s *gcsImpl) DeleteFile(ctx context.Context, filePath string) error {
	// Start the prometheus timer.
	t := prometheus.NewTimer(StorageLatency.With(prometheus.Labels{"query": "delete_file"}))
	defer t.ObserveDuration()

	// Connect to the bucket.
	bkt := s.gcs.Bucket(s.bucket)

	// Delete the file.
	err := bkt.Object(filePath).Delete(ctx)
	if err != nil {
		return fmt.Errorf("error deleting file: %w", err)
	}

	return nil
}

func (s *gcsImpl) Purge(ctx context.Context, from time.Time) (int, error) {
	// Start the prometheus timer.
	t := prometheus.NewTimer(StorageLatency.With(prometheus.Labels{"query": "purge"}))
	defer t.ObserveDuration()

	// Connect to the bucket.
	bkt := s.gcs.Bucket(s.bucket)

	// Get a list of all the files in the bucket.
	it := bkt.Objects(ctx, nil)

	count := 0

	// Iterate through the files.
	for {
		// Get the next file.
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			// There are no more files, so break out of the loop.
			break
		} else if err != nil {
			return 0, fmt.Errorf("error getting file next: %w", err)
		}

		// Get the file name.
		fileName := attrs.Name

		// Ignore all non-SQL files.
		if !strings.HasSuffix(fileName, ".sql") {
			continue
		}

		// Remove the path (report/environment/fqdn) from the file name.
		fileName = fileName[strings.LastIndex(fileName, "/")+1:]

		// Remove the file extension.
		fileName = fileName[:len(fileName)-len(".sql")]

		// Parse the file date from the file name.
		fileDate, err := time.Parse(time.RFC3339, fileName)
		if err != nil {
			slog.Warn(fmt.Sprintf("Error parsing file date from file name: %s", fileName))
			continue
		}

		// Check if the file date is before the purge date.
		if fileDate.After(from) {
			continue
		}

		// Delete the file.
		err = bkt.Object(attrs.Name).Delete(ctx)
		if err != nil {
			return 0, fmt.Errorf("error deleting file: %w", err)
		}

		count++
	}

	return count, nil
}
