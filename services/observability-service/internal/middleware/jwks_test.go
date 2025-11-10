package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func TestJWKSVerifier_Verify_Success(t *testing.T) {
	// Generate RSA key pair for testing
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	// Mock JWKS server
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Build JWKS response with RSA key
		nBytes := privKey.PublicKey.N.Bytes()
		eBytes := make([]byte, 4)
		eBytes[3] = byte(privKey.PublicKey.E)
		eBytes[2] = byte(privKey.PublicKey.E >> 8)
		eBytes[1] = byte(privKey.PublicKey.E >> 16)
		eBytes[0] = byte(privKey.PublicKey.E >> 24)
		
		// Trim leading zeros from e
		for len(eBytes) > 1 && eBytes[0] == 0 {
			eBytes = eBytes[1:]
		}

		jwks := map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": "test-key",
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
				"use": "sig",
				"alg": "RS256",
			}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}))
	defer jwksServer.Close()

	verifier := NewJWKSVerifier(jwksServer.URL, "test-issuer", "test-audience", "scopes", nil)

	// Create test JWT
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":        "test-issuer",
		"sub":        "test-user",
		"aud":        "test-audience",
		"exp":        time.Now().Add(time.Hour).Unix(),
		"iat":        time.Now().Unix(),
		"project_id": "test-project",
		"scopes":     []string{"observability:read", "observability:write"},
	})
	token.Header["kid"] = "test-key"

	tokenString, err := token.SignedString(privKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// Verify token
	claims, err := verifier.Verify(context.Background(), tokenString)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if claims.Subject != "test-user" {
		t.Errorf("expected subject test-user, got %s", claims.Subject)
	}
	if claims.ProjectID != "test-project" {
		t.Errorf("expected project_id test-project, got %s", claims.ProjectID)
	}
	if len(claims.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(claims.Scopes))
	}
	if claims.Scopes[0] != "observability:read" {
		t.Errorf("expected first scope observability:read, got %s", claims.Scopes[0])
	}
}

func TestJWKSVerifier_Verify_IssuerMismatch(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nBytes := privKey.PublicKey.N.Bytes()
		eBytes := []byte{1, 0, 1} // 65537
		jwks := map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": "test-key",
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			}},
		}
		json.NewEncoder(w).Encode(jwks)
	}))
	defer jwksServer.Close()

	verifier := NewJWKSVerifier(jwksServer.URL, "expected-issuer", "", "scopes", nil)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "wrong-issuer",
		"sub": "test-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	token.Header["kid"] = "test-key"

	tokenString, _ := token.SignedString(privKey)

	_, err := verifier.Verify(context.Background(), tokenString)
	if err == nil {
		t.Fatal("expected issuer mismatch error")
	}
	if err.Error() != "issuer mismatch" {
		t.Errorf("expected 'issuer mismatch', got %s", err.Error())
	}
}

func TestJWKSVerifier_Verify_AudienceMismatch(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nBytes := privKey.PublicKey.N.Bytes()
		eBytes := []byte{1, 0, 1}
		jwks := map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": "test-key",
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			}},
		}
		json.NewEncoder(w).Encode(jwks)
	}))
	defer jwksServer.Close()

	verifier := NewJWKSVerifier(jwksServer.URL, "test-issuer", "expected-audience", "scopes", nil)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "test-issuer",
		"sub": "test-user",
		"aud": "wrong-audience",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	token.Header["kid"] = "test-key"

	tokenString, _ := token.SignedString(privKey)

	_, err := verifier.Verify(context.Background(), tokenString)
	if err == nil {
		t.Fatal("expected audience mismatch error")
	}
	if err.Error() != "audience mismatch" {
		t.Errorf("expected 'audience mismatch', got %s", err.Error())
	}
}

func TestJWKSVerifier_Verify_MissingKid(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"keys": []any{}})
	}))
	defer jwksServer.Close()

	verifier := NewJWKSVerifier(jwksServer.URL, "test-issuer", "", "scopes", nil)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "test-issuer",
		"sub": "test-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	// No kid header

	tokenString, _ := token.SignedString(privKey)

	_, err := verifier.Verify(context.Background(), tokenString)
	if err == nil {
		t.Fatal("expected missing kid error")
	}
}

func TestJWKSVerifier_Verify_KeyNotFound(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"keys": []any{}})
	}))
	defer jwksServer.Close()

	verifier := NewJWKSVerifier(jwksServer.URL, "test-issuer", "", "scopes", nil)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "test-issuer",
		"sub": "test-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	token.Header["kid"] = "nonexistent-key"

	tokenString, _ := token.SignedString(privKey)

	_, err := verifier.Verify(context.Background(), tokenString)
	if err == nil {
		t.Fatal("expected key not found error")
	}
}

func TestParseScopes_Array(t *testing.T) {
	scopes := parseScopes([]any{"read", "write", "admin"})
	expected := []string{"read", "write", "admin"}
	
	if len(scopes) != len(expected) {
		t.Fatalf("expected %d scopes, got %d", len(expected), len(scopes))
	}
	for i, s := range scopes {
		if s != expected[i] {
			t.Errorf("expected scope %s, got %s", expected[i], s)
		}
	}
}

func TestParseScopes_StringSpace(t *testing.T) {
	scopes := parseScopes("read write admin")
	expected := []string{"read", "write", "admin"}
	
	if len(scopes) != len(expected) {
		t.Fatalf("expected %d scopes, got %d", len(expected), len(scopes))
	}
	for i, s := range scopes {
		if s != expected[i] {
			t.Errorf("expected scope %s, got %s", expected[i], s)
		}
	}
}

func TestParseScopes_StringComma(t *testing.T) {
	scopes := parseScopes("read,write,admin")
	expected := []string{"read", "write", "admin"}
	
	if len(scopes) != len(expected) {
		t.Fatalf("expected %d scopes, got %d", len(expected), len(scopes))
	}
	for i, s := range scopes {
		if s != expected[i] {
			t.Errorf("expected scope %s, got %s", expected[i], s)
		}
	}
}

func TestParseScopes_Empty(t *testing.T) {
	scopes := parseScopes("")
	if len(scopes) != 0 {
		t.Errorf("expected empty scopes, got %v", scopes)
	}

	scopes = parseScopes(nil)
	if scopes != nil {
		t.Errorf("expected nil scopes, got %v", scopes)
	}
}

func TestVerifyAudience_String(t *testing.T) {
	claims := jwt.MapClaims{"aud": "test-audience"}
	if !verifyAudience(claims, "test-audience") {
		t.Error("expected audience verification to pass")
	}
	if verifyAudience(claims, "wrong-audience") {
		t.Error("expected audience verification to fail")
	}
}

func TestVerifyAudience_Array(t *testing.T) {
	claims := jwt.MapClaims{"aud": []any{"aud1", "test-audience", "aud3"}}
	if !verifyAudience(claims, "test-audience") {
		t.Error("expected audience verification to pass")
	}
	if verifyAudience(claims, "wrong-audience") {
		t.Error("expected audience verification to fail")
	}
}

func TestVerifyAudience_EmptyNeed(t *testing.T) {
	claims := jwt.MapClaims{}
	if !verifyAudience(claims, "") {
		t.Error("expected empty audience need to always pass")
	}
}