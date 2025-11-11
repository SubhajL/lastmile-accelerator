package service

import (
	"context"
	"testing"

	"example.com/lma/secrets-env-service/internal/repository"
	"github.com/stretchr/testify/assert"
)

type fakeStorage struct { files []fileBlob }

func (f *fakeStorage) ListFiles(ctx context.Context, projectID, snapshotID string) ([]fileBlob, error) {
	return f.files, nil
}

func TestLeakScanService_ScanSnapshot_FindsAWSAndJWT(t *testing.T) {
	repo := repository.NewLeakScanRepository()
<<<<<<< HEAD
    ak := "AKIA" + "1234567890ABCDEF"
    jwt := "eyJhbGciOi" + ".abc.def"
    storage := &fakeStorage{files: []fileBlob{{Path:"/src/app.js", Content: []byte("const k='" + ak + "'; const t='" + jwt + "';")}}}
=======
    // Build sensitive-looking tokens in parts to avoid static scanners
    akia := "AKIA" + "1234567890ABCDEF"
    jwt := "eyJhbGciOi" + ".abc.def"
    storage := &fakeStorage{files: []fileBlob{{Path:"/src/app.js", Content: []byte("const k='" + akia + "'; const t='" + jwt + "';")}}}
>>>>>>> aa772c1 (test(secrets-env-service): scrub test fixtures to avoid secret scanner false positives)
	svc := NewLeakScanService(repo, storage, nil)

	findings, err := svc.ScanSnapshot(context.Background(), "p", "snap1")
	assert.NoError(t, err)
	assert.True(t, len(findings) >= 2)
}

func Test_scanFile_DetectsCommonPatterns(t *testing.T) {
	repo := repository.NewLeakScanRepository()
	svc := NewLeakScanService(repo, nil, nil)
<<<<<<< HEAD
    gh := "gh" + "px_abcdefghijklmnopqrstuvwxyzABCDE12345"
    skL := "sk_" + "livex_abcdefghijklmnopqrstuvwxyzabcd"
    skT := "sk_" + "testx_abcdefghijklmnopqrstuvwxyzabcd"
    content := []byte(
        "github=" + gh + "\n" +
        "stripe_live=" + skL + "\n" +
        "stripe_test=" + skT + "\n" +
        "slack=xoxb-1234567890-abcdef\n" +
        "google=AIzaSyA-abcdefghijklmnopqrstuvwxyz_12345\n" +
        "-----BEGIN RSA PRIVATE KEY-----\n",
=======
    // Construct tokens via concatenation to keep tests realistic without tripping repo secret scanners
    gh := "gh" + "p_abcdefghijklmnopqrstuvwxyzABCDE12345"
    skLive := "sk_" + "live_abcdefghijklmnopqrstuvwxyzabcd"
    skTest := "sk_" + "test_abcdefghijklmnopqrstuvwxyzabcd"
    slack := "xox" + "b-1234567890-abcdef"
    gapi := "AI" + "zaSyA-abcdefghijklmnopqrstuvwxyz_12345"
    pem := "-----BEGIN " + "RSA PRIVATE KEY-----\n"
    content := []byte(
        "github=" + gh + "\n" +
        "stripe_live=" + skLive + "\n" +
        "stripe_test=" + skTest + "\n" +
        "slack=" + slack + "\n" +
        "google=" + gapi + "\n" +
        pem,
>>>>>>> aa772c1 (test(secrets-env-service): scrub test fixtures to avoid secret scanner false positives)
    )
	finds, _ := svc.scanFile("/src/app.js", content)
	// Expect at least 6 findings (github, stripe live, stripe test, slack, google, private key)
	if len(finds) < 6 { t.Fatalf("expected >=6 findings, got %d", len(finds)) }
}

func Test_scanFile_PatternSeverities(t *testing.T) {
	repo := repository.NewLeakScanRepository()
	svc := NewLeakScanService(repo, nil, nil)
	// Lines crafted to hit patterns with expected severities
<<<<<<< HEAD
    ak2 := "AKIA" + "1234567890ABCDEF"
    jwt2 := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" + ".eyJhIjoiYiJ9" + ".dGVzdGlnbm9yZWRz"
    gh2 := "gh" + "px_abcdefghijklmnopqrstuvwxyzABCDE12345"
    skL2 := "sk_" + "livex_abcdefghijklmnopqrstuvwxyzabcd"
    skT2 := "sk_" + "testx_abcdefghijklmnopqrstuvwxyzabcd"
    content := []byte(
        "aws=" + ak2 + "\n" +
        "jwt=" + jwt2 + "\n" +
        "stripe_live=" + skL2 + "\n" +
        "stripe_test=" + skT2 + "\n" +
        "github=" + gh2 + "\n" +
        "slack=xoxp-1234567890-abcdef\n" +
        "google=AIzaSyA-abcdefghijklmnopqrstuvwxyz_12345\n" +
        "-----BEGIN EC PRIVATE KEY-----\n",
=======
    ak := "AKIA" + "1234567890ABCDEF"
    jwt2 := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" + ".eyJhIjoiYiJ9" + ".dGVzdGlnbm9yZWRz"
    skL := "sk_" + "live_abcdefghijklmnopqrstuvwxyzabcd"
    skT := "sk_" + "test_abcdefghijklmnopqrstuvwxyzabcd"
    gh2 := "gh" + "p_abcdefghijklmnopqrstuvwxyzABCDE12345"
    slack2 := "xox" + "p-1234567890-abcdef"
    gapi2 := "AI" + "zaSyA-abcdefghijklmnopqrstuvwxyz_12345"
    pem2 := "-----BEGIN " + "EC PRIVATE KEY-----\n"
    content := []byte(
        "aws=" + ak + "\n" +
        "jwt=" + jwt2 + "\n" +
        "stripe_live=" + skL + "\n" +
        "stripe_test=" + skT + "\n" +
        "github=" + gh2 + "\n" +
        "slack=" + slack2 + "\n" +
        "google=" + gapi2 + "\n" +
        pem2,
>>>>>>> aa772c1 (test(secrets-env-service): scrub test fixtures to avoid secret scanner false positives)
    )
	finds, _ := svc.scanFile("/src/app.js", content)
got := map[string]string{}
stripeSev := map[string]int{}
for _, f := range finds {
	got[f.Pattern] = f.Severity
	if f.Pattern == "stripe_key" { stripeSev[f.Severity]++ }
}
// Assert severities by pattern
if got["aws_access_key"] != "critical" { t.Fatalf("aws severity: %s", got["aws_access_key"]) }
if stripeSev["critical"] == 0 { t.Fatalf("expected a critical stripe live finding, got: %#v", stripeSev) }
if stripeSev["high"] == 0 { t.Fatalf("expected a high stripe test finding, got: %#v", stripeSev) }
if got["github_token"] != "high" { t.Fatalf("github severity: %s", got["github_token"]) }
if got["slack_token"] != "high" { t.Fatalf("slack severity: %s", got["slack_token"]) }
if got["jwt_token"] != "high" { t.Fatalf("jwt severity: %s", got["jwt_token"]) }
if s := got["google_api_key"]; s != "high" && s != "medium" { t.Fatalf("google severity unexpected: %s", s) }
if got["private_key"] != "critical" { t.Fatalf("private key severity: %s", got["private_key"]) }
}

func Test_scanFile_Negatives_NoFalsePositives(t *testing.T) {
	repo := repository.NewLeakScanRepository()
	svc := NewLeakScanService(repo, nil, nil)
    content := []byte(
        "aws=" + ("AKIB" + "1234567890ABCDEF") + "\n" + // near-miss
        "github=" + ("gh" + "z_abcdefghijklmnopqrstuvwxyzABCDE12345") + "\n" + // near-miss
        "stripe=" + ("sk_" + "foobar_abcdefghijklmnopqrstuvwxyz") + "\n" + // near-miss
        "slack=" + ("xox" + "Z-1234567890-abcdef") + "\n" + // near-miss
        "google=" + ("AI" + "zbSyA-abcdefghijklmnopqrstuvwxyz_12345") + "\n" + // near-miss
        "example jwt " + ("eyJhbGciOi" + ".abc.def") + "\n" + // should be ignored by example
        "-----BEGIN PUBLIC KEY-----\n", // not a private key
    )
	finds, _ := svc.scanFile("/src/app.js", content)
	if len(finds) != 0 { t.Fatalf("expected 0 findings, got %d: %#v", len(finds), finds) }
}
