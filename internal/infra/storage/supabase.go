package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"go.uber.org/zap"
)

type Client struct {
	baseURL        string
	serviceRoleKey string
	bucket         string
	httpClient     *http.Client
}

func NewClient(baseURL, serviceRoleKey, bucket string) *Client {
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "https://" + baseURL
	}
	return &Client{
		baseURL:        strings.TrimRight(baseURL, "/"),
		serviceRoleKey: serviceRoleKey,
		bucket:         bucket,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

func NewClientFromEnv() *Client {
	return NewClient(
		os.Getenv("SUPABASE_URL"),
		os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		os.Getenv("SUPABASE_BUCKET"),
	)
}

func (c *Client) Upload(ctx context.Context, objectPath string, reader io.Reader, contentType string) (string, error) {
	encodedPath := strings.ReplaceAll(url.PathEscape(objectPath), "%2F", "/")
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.baseURL, c.bucket, encodedPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, reader)
	if err != nil {
		logger.Log.Error("supabase upload: new request failed", zap.Error(err), zap.String("path", objectPath))
		return "", domain_errors.WrapUnexpectedMsg(err, "failed to create upload request")
	}

	req.Header.Set("Authorization", "Bearer "+c.serviceRoleKey)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("x-upsert", "false")

	res, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("supabase upload: request failed", zap.Error(err), zap.String("path", objectPath))
		return "", domain_errors.WrapUnexpectedMsg(err, "failed to upload file to storage")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logger.Log.Error("failed to close response body", zap.Error(err))
		}
	}()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		err := fmt.Errorf("storage upload returned status %d: %s", res.StatusCode, string(body))
		logger.Log.Error("supabase upload: error response", zap.Error(err), zap.Int("status", res.StatusCode), zap.String("path", objectPath))
		return "", domain_errors.WrapUnexpectedMsg(err, "storage upload error")
	}

	signedURL, err := c.SignURL(ctx, objectPath, 3600)
	if err != nil {
		return "", err
	}
	return signedURL, nil
}

func (c *Client) Delete(ctx context.Context, objectPath string) error {
	encodedPath := strings.ReplaceAll(url.PathEscape(objectPath), "%2F", "/")
	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.baseURL, c.bucket, encodedPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
	if err != nil {
		logger.Log.Error("supabase delete: new request failed", zap.Error(err), zap.String("path", objectPath))
		return domain_errors.WrapUnexpectedMsg(err, "failed to create delete request")
	}
	req.Header.Set("Authorization", "Bearer "+c.serviceRoleKey)

	res, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("supabase delete: request failed", zap.Error(err), zap.String("path", objectPath))
		return domain_errors.WrapUnexpectedMsg(err, "failed to delete object from storage")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logger.Log.Error("failed to close response body", zap.Error(err))
		}
	}()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		err := fmt.Errorf("storage delete returned status %d: %s", res.StatusCode, string(body))
		logger.Log.Error("supabase delete: error response", zap.Error(err), zap.Int("status", res.StatusCode), zap.String("path", objectPath))
		return domain_errors.WrapUnexpectedMsg(err, "storage delete error")
	}

	return nil
}

func (c *Client) SignURL(ctx context.Context, objectPath string, expiresInSeconds int) (string, error) {
	encodedPath := strings.ReplaceAll(url.PathEscape(objectPath), "%2F", "/")
	signURL := fmt.Sprintf("%s/storage/v1/object/sign/%s/%s", c.baseURL, c.bucket, encodedPath)

	payload := map[string]int{"expiresIn": expiresInSeconds}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, signURL, bytes.NewReader(b))
	if err != nil {
		logger.Log.Error("supabase sign: new request failed", zap.Error(err), zap.String("path", objectPath))
		return "", domain_errors.WrapUnexpectedMsg(err, "failed to create sign request")
	}

	req.Header.Set("Authorization", "Bearer "+c.serviceRoleKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("supabase sign: request failed", zap.Error(err), zap.String("path", objectPath))
		return "", domain_errors.WrapUnexpectedMsg(err, "failed to request signed url")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logger.Log.Error("failed to close response body", zap.Error(err))
		}
	}()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		err := fmt.Errorf("storage sign returned status %d: %s", res.StatusCode, string(body))
		logger.Log.Error("supabase sign: error response", zap.Error(err), zap.Int("status", res.StatusCode), zap.String("path", objectPath))
		return "", domain_errors.WrapUnexpectedMsg(err, "storage sign url error")
	}

	var m map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		logger.Log.Error("supabase sign: decode response failed", zap.Error(err), zap.String("path", objectPath))
		return "", domain_errors.WrapUnexpectedMsg(err, "failed to decode sign url response")
	}

	if v, ok := m["signedURL"].(string); ok && v != "" {
		return c.makeFullURL(v), nil
	}
	if v, ok := m["signedUrl"].(string); ok && v != "" {
		return c.makeFullURL(v), nil
	}

	return "", domain_errors.NewUnexpectedMsg("signed url not found in response")
}

func (c *Client) makeFullURL(signed string) string {
	s := strings.TrimSpace(signed)
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	if strings.HasPrefix(s, "/object/") {
		s = "/storage/v1" + s
	}
	return c.baseURL + s
}
