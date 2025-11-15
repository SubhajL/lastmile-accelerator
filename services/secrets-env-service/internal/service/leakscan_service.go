package service

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"math"
	"regexp"
	"strings"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"example.com/lma/secrets-env-service/internal/events"
	"example.com/lma/secrets-env-service/internal/metrics"
	"example.com/lma/secrets-env-service/internal/repository"
	"github.com/google/uuid"
)

type fileBlob struct { Path string; Content []byte }

type blobStorage interface {
	ListFiles(ctx context.Context, projectID, snapshotID string) ([]fileBlob, error)
}

type LeakScanService struct {
	repo    *repository.LeakScanRepository
	storage blobStorage
	pub     events.Publisher
}

func NewLeakScanService(repo *repository.LeakScanRepository, storage blobStorage, pub events.Publisher) *LeakScanService {
	return &LeakScanService{repo: repo, storage: storage, pub: pub}
}

var (
	reAWS       = regexp.MustCompile(`AKIA[0-9A-Z]{16}`)
	reJWT       = regexp.MustCompile(`(?i)eyJ[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+`)
	reStripe    = regexp.MustCompile(`sk_(?:live|test)_[0-9a-zA-Z]{24,}`)
	reGitHub    = regexp.MustCompile(`(?:ghp|gho|ghs)_[A-Za-z0-9]{36,}|github_pat_[A-Za-z0-9_]{60,}`)
	reSlack     = regexp.MustCompile(`xox[baprs]-[A-Za-z0-9\-]{10,}`)
	reGoogleAPI = regexp.MustCompile(`AIza[0-9A-Za-z\-_]{35}`)
	rePrivKey   = regexp.MustCompile(`-----BEGIN (?:RSA|EC|DSA|OPENSSH) PRIVATE KEY-----`)
)

func (s *LeakScanService) ScanSnapshot(ctx context.Context, projectID, snapshotID string) ([]*domain.ClientLeakScan, error) {
    files, err := s.storage.ListFiles(ctx, projectID, snapshotID)
    if err != nil {
        if metrics.Default != nil { metrics.Default.ObserveLeaks("scan", "error") }
        return nil, err
    }
	var findings []*domain.ClientLeakScan
	for _, f := range files {
		ff, _ := s.scanFile(f.Path, f.Content)
		findings = append(findings, ff...)
	}
	_ = s.repo.CreateBatch(ctx, findings)
	if s.pub != nil {
		_ = s.pub.Publish(ctx, "leak.scan.completed", map[string]any{"project_id": projectID, "snapshot_id": snapshotID, "count": len(findings)})
		if len(findings) > 0 {
			_ = s.pub.Publish(ctx, "client.leak.found", map[string]any{"project_id": projectID, "snapshot_id": snapshotID, "count": len(findings)})
		}
	}
    if metrics.Default != nil { metrics.Default.ObserveLeaks("scan", "success") }
    return findings, nil
}

func (s *LeakScanService) GetScanResults(ctx context.Context, projectID, snapshotID, severity string) ([]*domain.ClientLeakScan, error) {
    res, err := s.repo.GetBySnapshotID(ctx, snapshotID, severity)
    if metrics.Default != nil {
        if err != nil { metrics.Default.ObserveLeaks("results", "error") } else { metrics.Default.ObserveLeaks("results", "success") }
    }
    return res, err
}

func (s *LeakScanService) MarkAsFixed(ctx context.Context, scanID string) error {
    err := s.repo.MarkAsFixed(ctx, scanID)
    if metrics.Default != nil {
        if err != nil { metrics.Default.ObserveLeaks("fix", "error") } else { metrics.Default.ObserveLeaks("fix", "success") }
    }
    return err
}

func (s *LeakScanService) scanFile(path string, content []byte) ([]*domain.ClientLeakScan, error) {
	var out []*domain.ClientLeakScan
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	line := 0
for scanner.Scan() {
		line++
		text := scanner.Text()
		lower := strings.ToLower(text)
		// AWS
		if m := reAWS.FindString(text); m != "" {
			out = append(out, s.newFinding(path, line, "aws_access_key", "critical"))
		}
		// JWT
		if m := reJWT.FindString(text); m != "" && !strings.Contains(lower, "example") {
			out = append(out, s.newFinding(path, line, "jwt_token", s.classify("jwt", text)))
		}
		// Stripe
		if m := reStripe.FindString(text); m != "" {
			sev := "high"
			if strings.HasPrefix(m, "sk_live_") { sev = "critical" }
			out = append(out, s.newFinding(path, line, "stripe_key", sev))
		}
		// GitHub
		if m := reGitHub.FindString(text); m != "" {
			out = append(out, s.newFinding(path, line, "github_token", "high"))
		}
		// Slack
		if m := reSlack.FindString(text); m != "" {
			out = append(out, s.newFinding(path, line, "slack_token", "high"))
		}
		// Google API Key
		if m := reGoogleAPI.FindString(text); m != "" {
			out = append(out, s.newFinding(path, line, "google_api_key", s.classify("google_api_key", text)))
		}
		// Private key markers
		if m := rePrivKey.FindString(text); m != "" {
			out = append(out, s.newFinding(path, line, "private_key", "critical"))
		}
	}
	return out, nil
}

func (s *LeakScanService) classify(kind, text string) string {
	entropy := calcEntropy(text)
	switch kind {
	case "jwt":
		if entropy >= 3.5 { return "high" }
		return "low"
	case "google_api_key":
		if entropy >= 3.5 { return "high" }
		return "medium"
	default:
		if entropy >= 3.5 { return "high" }
		return "low"
	}
}

func (s *LeakScanService) newFinding(path string, line int, pattern, severity string) *domain.ClientLeakScan {
	id := uuid.New().String()
	return &domain.ClientLeakScan{ID: id, ProjectID: "", SnapshotID: hashPath(path), FilePath: path, LineNumber: line, Pattern: pattern, Severity: severity, Fixed: false, CreatedAt: time.Now()}
}

func calcEntropy(s string) float64 {
	if s == "" { return 0 }
	freq := map[rune]int{}
	for _, r := range s { freq[r]++ }
	total := 0
	for _, c := range freq { total += c }
	h := 0.0
	for _, c := range freq {
		p := float64(c) / float64(total)
		if p > 0 { h += -p * math.Log2(p) }
	}
	return h
}

func hashPath(p string) string {
	h := sha1.Sum([]byte(p))
	return hex.EncodeToString(h[:])
}
