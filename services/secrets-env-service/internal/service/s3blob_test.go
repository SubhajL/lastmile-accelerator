package service

import (
	"context"
	"testing"
)

func Test_isBinary_ByExtAndNull(t *testing.T) {
	if !isBinary("/img/logo.png", []byte("abc")) { t.Fatalf("png should be binary") }
	if !isBinary("/any.txt", []byte{'a', 0x00, 'b'}) { t.Fatalf("null-byte content should be binary") }
	if isBinary("/notes.txt", []byte("hello")) { t.Fatalf("txt plain should not be binary") }
}

func Test_isLarge_SizeLimit(t *testing.T) {
	if !isLarge(3, []byte("abcd")) { t.Fatalf("over limit should be large") }
	if isLarge(4, []byte("abcd")) { t.Fatalf("equal to limit not large") }
	if isLarge(0, []byte("abcd")) { t.Fatalf("no limit should not be large") }
}

func Test_shouldIgnore_Globs(t *testing.T) {
	globs := []string{"*.map", "dist/*", "**/*.min.js"}
	cases := []struct{ path string; want bool }{
		{"app.min.js", true},
		{"dist/app.js", true},
		{"src/app.js.map", true},
		{"src/app.js", false},
	}
	for _, c := range cases {
		if got := shouldIgnore(c.path, globs); got != c.want { t.Fatalf("%s expected %v got %v", c.path, c.want, got) }
	}
}

func Test_filterBuffer_SplitsLines(t *testing.T) {
	lines := filterBuffer([]byte("a\nb\n"))
	if len(lines) != 2 || lines[0] != "a" || lines[1] != "b" { t.Fatalf("unexpected lines: %#v", lines) }
}

// --- ListFiles integration with fake client ---

type fakeS3Client struct{
	keys []string
	objects map[string][]byte
}

func (f *fakeS3Client) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	// return keys under prefix
	var out []string
	for _, k := range f.keys {
		if len(prefix) == 0 || (len(k) >= len(prefix) && k[:len(prefix)] == prefix) {
			out = append(out, k)
		}
	}
	return out, nil
}

func (f *fakeS3Client) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	return f.objects[key], nil
}

func TestS3Blob_ListFiles_Filters(t *testing.T) {
	cfg := S3Config{Bucket: "b", Prefix: "snapshots", IgnoreGlobs: []string{"*.map", "dist/*", "**/*.min.js"}, SizeLimitBytes: 1024}
	client := &fakeS3Client{
		keys: []string{
			"snapshots/p1/s1/src/app.js",
			"snapshots/p1/s1/src/app.js.map",
			"snapshots/p1/s1/dist/bundle.min.js",
			"snapshots/p1/s1/assets/logo.png",
		},
		objects: map[string][]byte{
			"snapshots/p1/s1/src/app.js":       []byte("const k='AKIA1234567890ABCDEF';"),
			"snapshots/p1/s1/src/app.js.map":   []byte("{\"version\":3}"),
			"snapshots/p1/s1/dist/bundle.min.js": []byte("minified"),
			"snapshots/p1/s1/assets/logo.png":  []byte{0x89, 'P', 'N', 'G'},
		},
	}
	s3 := NewS3BlobWithClient(cfg, client)
	files, err := s3.ListFiles(context.Background(), "p1", "s1")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(files) != 1 { t.Fatalf("expected 1 file after filters, got %d", len(files)) }
	if files[0].Path != "src/app.js" { t.Fatalf("unexpected path: %s", files[0].Path) }
}
