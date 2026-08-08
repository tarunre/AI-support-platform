// Package attachments provides a storage abstraction for email attachments.
// The local implementation writes files to disk; swap the interface for S3, GCS, etc.
package attachments

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Storage is the provider-agnostic interface for persisting attachment data.
type Storage interface {
	// Save writes data and returns the storage path (opaque string for later retrieval).
	Save(ticketID, filename string, data []byte) (string, error)

	// BasePath returns the root directory (for health checks).
	BasePath() string
}

// LocalStorage saves attachments to the local filesystem under basePath.
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a LocalStorage that writes under basePath.
// The directory is created on first use.
func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

// Save writes data to <basePath>/<ticketID>/<timestamp>_<filename> and returns
// the relative path.  All path components are sanitised.
func (s *LocalStorage) Save(ticketID, filename string, data []byte) (string, error) {
	// Sanitise inputs — never allow path traversal
	safeTicket := filepath.Base(ticketID)
	safeFile := filepath.Base(filename)
	if safeTicket == "" || safeTicket == "." {
		safeTicket = "unknown"
	}
	if safeFile == "" || safeFile == "." {
		safeFile = "attachment"
	}

	dir := filepath.Join(s.basePath, safeTicket)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", fmt.Errorf("attachments: create directory: %w", err)
	}

	// Timestamp prefix prevents filename collisions
	name := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeFile)
	fullPath := filepath.Join(dir, name)

	if err := os.WriteFile(fullPath, data, 0640); err != nil {
		return "", fmt.Errorf("attachments: write file: %w", err)
	}

	return fullPath, nil
}

// BasePath returns the configured root directory.
func (s *LocalStorage) BasePath() string {
	return s.basePath
}
