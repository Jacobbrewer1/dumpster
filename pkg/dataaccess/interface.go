package dataaccess

import (
	"context"
	"time"
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
