package dataaccess

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	EnvGCSCredentials = "GCS_CREDENTIALS"
)

type Storage interface {
	// SaveFile uploads a file to the storage bucket. This will replace any existing file with the same name.
	SaveFile(ctx context.Context, filePath string, file []byte) error

	// DownloadFile downloads a file from the storage bucket.
	DownloadFile(ctx context.Context, filePath string) ([]byte, error)

	// DeleteFile deletes a file from the storage bucket.
	DeleteFile(ctx context.Context, filePath string) error

	// Purge purges the data from the storage bucket out of the given range.
	Purge(ctx context.Context, from time.Time) (int, error)
}

type storageImpl struct {
	// gcs is the Google Cloud Storage client.
	gcs *storage.Client

	// bucket is the name of the bucket to use.
	bucket string
}

func newStorage(gcs *storage.Client, bucket string) Storage {
	return &storageImpl{
		gcs:    gcs,
		bucket: bucket,
	}
}

func (s *storageImpl) SaveFile(ctx context.Context, filePath string, file []byte) error {
	// Start the prometheus timer.
	t := prometheus.NewTimer(GCSLatency.With(prometheus.Labels{"query": "save_file"}))
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

func (s *storageImpl) DownloadFile(ctx context.Context, filePath string) ([]byte, error) {
	// Start the prometheus timer.
	t := prometheus.NewTimer(GCSLatency.With(prometheus.Labels{"query": "download_file"}))
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

func (s *storageImpl) DeleteFile(ctx context.Context, filePath string) error {
	// Start the prometheus timer.
	t := prometheus.NewTimer(GCSLatency.With(prometheus.Labels{"query": "delete_file"}))
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

func (s *storageImpl) Purge(ctx context.Context, from time.Time) (int, error) {
	// Start the prometheus timer.
	t := prometheus.NewTimer(GCSLatency.With(prometheus.Labels{"query": "purge"}))
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

func ConnectGCS(ctx context.Context, gcsBucket string) (Storage, error) {
	// Get the service account credentials from the environment variable.
	gcsCredentials := os.Getenv(EnvGCSCredentials)
	if gcsCredentials == "" {
		return nil, errors.New("no GCS credentials provided")
	}

	client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(gcsCredentials)))
	if err != nil {
		return nil, fmt.Errorf("error connecting to GCS: %w", err)
	}
	cs := client

	// Get the bucket name from the environment variable and validate that it exists.
	if gcsBucket == "" {
		return nil, errors.New("no GCS bucket provided")
	}

	_, err = cs.Bucket(gcsBucket).Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("error validating GCS bucket: %w", err)
	}

	sc := newStorage(cs, gcsBucket)
	slog.Debug("Connected to GCS")
	return sc, nil
}
