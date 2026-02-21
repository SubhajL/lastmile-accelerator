package service

import (
	"context"
	"testing"

	"example.com/lma/secrets-env-service/internal/repository"
)

type capPub2 struct{ topics []string }
func (c *capPub2) Publish(ctx context.Context, topic string, payload any) error { c.topics = append(c.topics, topic); return nil }

type fakeStorage2 struct{ files []fileBlob }
func (f *fakeStorage2) ListFiles(ctx context.Context, projectID, snapshotID string) ([]fileBlob, error) { return f.files, nil }

func TestLeakScanService_PublishesClientLeakFound_OnFindings(t *testing.T) {
	repo := repository.NewLeakScanRepository()
    storage := &fakeStorage2{files: []fileBlob{{Path:"src/app.js", Content: []byte("const t='eyJ" + "hbGciOi" + ".abc.def';")}}}
	cap := &capPub2{}
	svc := NewLeakScanService(repo, storage, cap)
	_, err := svc.ScanSnapshot(context.Background(), "p", "s1")
	if err != nil { t.Fatalf("scan err: %v", err) }
	found := false
	for _, tpc := range cap.topics { if tpc == "client.leak.found" { found = true; break } }
	if !found { t.Fatalf("expected client.leak.found publish, got %v", cap.topics) }
}
