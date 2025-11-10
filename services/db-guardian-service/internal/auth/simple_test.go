package auth

import (
	"context"
	"testing"
)

func TestSimpleAuthenticator_Verify_OK(t *testing.T) {
	a := NewSimpleAuthenticator()
	c, err := a.Verify(context.Background(), "Bearer token", nil)
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if c == nil || c.Subject == "" { t.Fatalf("want claims, got %#v", c) }
}

func TestSimpleAuthenticator_Verify_RejectsMissingBearer(t *testing.T) {
	a := NewSimpleAuthenticator()
	if _, err := a.Verify(context.Background(), "token", nil); err == nil { t.Fatal("expected error") }
	if _, err := a.Verify(context.Background(), "Bearer  ", nil); err == nil { t.Fatal("expected error for empty token") }
}