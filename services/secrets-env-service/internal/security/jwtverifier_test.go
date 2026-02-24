package security

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

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jwksDoc struct{ Keys []map[string]string `json:"keys"` }

func rsaJWK(pub *rsa.PublicKey, kid string) map[string]string {
	return map[string]string{
		"kty": "RSA",
		"kid": kid,
		"alg": "RS256",
		"use": "sig",
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(bigEndian(pub.E)),
	}
}

func bigEndian(e int) []byte {
	// encode small int exponent (usually 65537)
	if e == 0 { return []byte{0} }
	var b []byte
	for e > 0 { b = append([]byte{byte(e & 0xff)}, b...); e >>= 8 }
	return b
}

func makeJWT(t *testing.T, priv *rsa.PrivateKey, kid, iss, aud, sub, tenant, project string, scopes []string, roles []string, expired bool) string {
	now := time.Now()
	exp := now.Add(1 * time.Hour)
	if expired { exp = now.Add(-1 * time.Hour) }
	claims := jwt.MapClaims{
		"iss": iss,
		"aud": aud,
		"sub": sub,
		"tenant_id": tenant,
		"project_id": project,
		"scopes": scopes,
		"roles": roles,
		"exp": exp.Unix(),
		"nbf": now.Unix()-10,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = kid
	s, err := tok.SignedString(priv)
	require.NoError(t, err)
	return s
}

func TestJWKSVerifier_ValidToken(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	kid := "kid1"
	iss := "https://issuer"
	aud := "myaud"
	jwks := jwksDoc{Keys: []map[string]string{rsaJWK(&priv.PublicKey, kid)}}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer ts.Close()

	ver := NewJWKSVerifier(ts.URL, aud, iss, 5*time.Minute)
	token := makeJWT(t, priv, kid, iss, aud, "subj", "tenant-1", "proj-1", []string{"secrets:read"}, []string{"admin"}, false)
	claims, err := ver.Verify(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, "tenant-1", claims.TenantID)
	assert.Contains(t, claims.Scopes, "secrets:read")
	assert.Equal(t, []string{"admin"}, claims.Roles)
}

func TestJWKSVerifier_AudienceMismatch(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	jwks := jwksDoc{Keys: []map[string]string{rsaJWK(&priv.PublicKey, "kid1")}}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(jwks) }))
	defer ts.Close()
	ver := NewJWKSVerifier(ts.URL, "otheraud", "https://issuer", time.Minute)
	token := makeJWT(t, priv, "kid1", "https://issuer", "badaud", "s", "t", "p", nil, nil, false)
	_, err := ver.Verify(context.Background(), token)
	assert.Error(t, err)
}

func TestJWKSVerifier_ExpiredToken(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	jwks := jwksDoc{Keys: []map[string]string{rsaJWK(&priv.PublicKey, "kid1")}}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(jwks) }))
	defer ts.Close()
	ver := NewJWKSVerifier(ts.URL, "aud", "iss", time.Minute)
	token := makeJWT(t, priv, "kid1", "iss", "aud", "s", "t", "p", nil, nil, true)
	_, err := ver.Verify(context.Background(), token)
	assert.Error(t, err)
}
