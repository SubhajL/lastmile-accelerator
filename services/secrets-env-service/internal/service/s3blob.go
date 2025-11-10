package service

import (
	"bufio"
	"context"
	"errors"
	"path"
	"path/filepath"
	"strings"
)

// S3Config holds minimal settings for S3-compatible storage wiring.
type S3Config struct {
	Endpoint       string
	Bucket         string
	Prefix         string
	UseTLS         bool
	IgnoreGlobs    []string
	SizeLimitBytes int64
}

// S3Blob implements blobStorage by listing and fetching files from a blob backend.
// For unit tests here we only expose helper filters; the actual network client is injected elsewhere.
type S3Client interface {
	ListObjects(ctx context.Context, bucket, prefix string) ([]string, error)
	GetObject(ctx context.Context, bucket, key string) ([]byte, error)
}

type FileBlobAlias = fileBlob

type S3Blob struct {
	cfg    S3Config
	client S3Client
	// fetcher/listing can be injected in real wiring
}

func NewS3Blob(cfg S3Config) *S3Blob { return &S3Blob{cfg: cfg} }
func NewS3BlobWithClient(cfg S3Config, client S3Client) *S3Blob { return &S3Blob{cfg: cfg, client: client} }

// ListFiles lists files under bucket/prefix/projectID/snapshotID and applies filters.
func (s *S3Blob) ListFiles(ctx context.Context, projectID, snapshotID string) ([]fileBlob, error) {
	if s.client == nil { return nil, errors.New("ListFiles not wired: provide concrete client in main") }
	base := strings.TrimSuffix(s.cfg.Prefix, "/")
	prefix := path.Join(base, projectID, snapshotID)
	keys, err := s.client.ListObjects(ctx, s.cfg.Bucket, prefix)
	if err != nil { return nil, err }
	var out []fileBlob
	for _, key := range keys {
		// derive relative path for reporting
		rel := strings.TrimPrefix(key, prefix)
		rel = strings.TrimPrefix(rel, "/")
		// ignore by glob against relative path
		if shouldIgnore(rel, s.cfg.IgnoreGlobs) { continue }
		// skip obvious binaries by extension prior to fetch
		if isBinary(rel, nil) { continue }
		b, err := s.client.GetObject(ctx, s.cfg.Bucket, key)
		if err != nil { continue }
		// post-fetch filters (size, null-bytes)
		if isLarge(s.cfg.SizeLimitBytes, b) { continue }
		if isBinary(rel, b) { continue }
		out = append(out, fileBlob{Path: rel, Content: b})
	}
	return out, nil
}

// ---- Helpers (unit tested) ----

func isBinary(path string, content []byte) bool {
	if hasNullByte(content) { return true }
	ext := strings.ToLower(filepath.Ext(path))
	switch ext := ext; ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".pdf", ".zip", ".tar", ".gz", ".mp4", ".mov", ".exe", ".dll":
		return true
	}
	return false
}

func hasNullByte(content []byte) bool {
	for _, b := range content { if b == 0 { return true } }
	return false
}

func isLarge(limit int64, content []byte) bool {
	if limit <= 0 { return false }
	return int64(len(content)) > limit
}

func shouldIgnore(path string, globs []string) bool {
	for _, g := range globs {
		// Support gitignore-like semantics:
		// - Patterns without a slash apply to the basename (e.g., "*.map")
		// - "**/" prefix means match suffix anywhere (e.g., "**/*.min.js")
		if strings.HasPrefix(g, "**/") {
			tail := strings.TrimPrefix(g, "**/")
			if strings.HasPrefix(tail, "*") {
				sfx := strings.TrimPrefix(tail, "*")
				if strings.HasSuffix(path, sfx) { return true }
				continue
			}
		}
		if !strings.Contains(g, string(filepath.Separator)) {
			if ok, _ := filepath.Match(g, filepath.Base(path)); ok { return true }
			continue
		}
		if ok, _ := filepath.Match(g, path); ok { return true }
	}
	return false
}

// filterBuffer returns the text lines of a buffer for scanning.
func filterBuffer(content []byte) []string {
	s := bufio.NewScanner(strings.NewReader(string(content)))
	var lines []string
	for s.Scan() { lines = append(lines, s.Text()) }
	return lines
}
