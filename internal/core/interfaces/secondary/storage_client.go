package secondary

import (
	"context"
	"io"
)

type StorageClient interface {
	Upload(ctx context.Context, objectPath string, reader io.Reader, contentType string) (string, error)
	Delete(ctx context.Context, objectPath string) error
	SignURL(ctx context.Context, objectPath string, expiresInSeconds int) (string, error)
}
