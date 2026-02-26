package secrets

import "testing"

func TestExtractProjectDSN_PrefersDSNKey(t *testing.T) {
	dsn, err := extractProjectDSN(map[string]string{"dsn": "postgres://x", "url": "postgres://y"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if dsn != "postgres://x" {
		t.Fatalf("expected dsn, got %s", dsn)
	}
}

func TestExtractProjectDSN_FallbackToDatabaseURL(t *testing.T) {
	dsn, err := extractProjectDSN(map[string]string{"DATABASE_URL": "postgres://x"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if dsn != "postgres://x" {
		t.Fatalf("expected DATABASE_URL, got %s", dsn)
	}
}

func TestExtractProjectDSN_MissingKey(t *testing.T) {
	_, err := extractProjectDSN(map[string]string{"something": "else"})
	if err == nil {
		t.Fatal("expected error")
	}
}
