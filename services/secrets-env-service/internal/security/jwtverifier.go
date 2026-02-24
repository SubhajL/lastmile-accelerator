package security

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"example.com/lma/secrets-env-service/internal/handlers"
	"github.com/golang-jwt/jwt/v5"
)

// JWKSVerifier validates JWTs using a remote JWKS endpoint and basic iss/aud checks.
type JWKSVerifier struct {
	jwksURL string
	aud     string
	iss     string
	ttl     time.Duration

	mu        sync.RWMutex
	keys      map[string]interface{} // kid -> public key (*rsa.PublicKey)
	lastFetch time.Time
	client    *http.Client
}

type jwkSet struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"` // base64url
	E   string `json:"e"` // base64url
}

// NewJWKSVerifier constructs a verifier with caching TTL.
func NewJWKSVerifier(jwksURL, audience, issuer string, ttl time.Duration) *JWKSVerifier {
	return &JWKSVerifier{
		jwksURL: jwksURL,
		aud:     audience,
		iss:     issuer,
		ttl:     ttl,
		keys:    make(map[string]interface{}),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// Verify parses and validates a JWT using JWKS; returns app Claims.
func (v *JWKSVerifier) Verify(ctx context.Context, tokenStr string) (*handlers.Claims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))
	// Use MapClaims for flexibility; validate manually below.
	claims := jwt.MapClaims{}

	keyfunc := func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("missing kid")
		}
		if k := v.getKey(kid); k != nil {
			return k, nil
		}
		if err := v.refresh(ctx); err != nil {
			return nil, err
		}
		if k := v.getKey(kid); k != nil {
			return k, nil
		}
		return nil, fmt.Errorf("kid %s not found in JWKS", kid)
	}

	_, err := parser.ParseWithClaims(tokenStr, claims, keyfunc)
	if err != nil {
		return nil, err
	}
	// Validate iss
	if v.iss != "" {
		if iss, _ := claims["iss"].(string); iss != v.iss {
			return nil, errors.New("issuer mismatch")
		}
	}
	// Validate aud (string or array)
	if v.aud != "" {
		if !audContains(claims["aud"], v.aud) {
			return nil, errors.New("audience mismatch")
		}
	}
	// Exp and NBF
	if err := checkExpNbf(claims); err != nil {
		return nil, err
	}

	// Extract custom claims
	tenantID, _ := stringField(claims, "tenant_id")
	projectID, _ := stringField(claims, "project_id")
	subject, _ := stringField(claims, "sub")
	scopes := scopesFromClaims(claims)
	roles := rolesFromClaims(claims)
	return &handlers.Claims{Subject: subject, TenantID: tenantID, ProjectID: projectID, Scopes: scopes, Roles: roles}, nil
}

func (v *JWKSVerifier) getKey(kid string) interface{} {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.ttl > 0 && time.Since(v.lastFetch) > v.ttl {
		return nil
	}
	return v.keys[kid]
}

func (v *JWKSVerifier) refresh(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	resp, err := v.client.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	var set jwkSet
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil { return err }

	keys := make(map[string]interface{})
	for _, k := range set.Keys {
		if strings.ToUpper(k.Kty) == "RSA" && k.N != "" && k.E != "" && k.Kid != "" {
			pub, err := rsaFromJWK(k.N, k.E)
			if err == nil { keys[k.Kid] = pub }
		}
	}
	v.mu.Lock()
	v.keys = keys
	v.lastFetch = time.Now()
	v.mu.Unlock()
	return nil
}

func rsaFromJWK(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil { return nil, err }
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil { return nil, err }
	var n big.Int
	n.SetBytes(nBytes)
	e := 0
	for _, b := range eBytes { e = e*256 + int(b) }
	return &rsa.PublicKey{N: &n, E: e}, nil
}

func audContains(audField any, expected string) bool {
	switch v := audField.(type) {
	case string:
		return v == expected
	case []any:
		for _, x := range v { if s, ok := x.(string); ok && s == expected { return true } }
	case []string:
		for _, s := range v { if s == expected { return true } }
	}
	return false
}

func stringField(m jwt.MapClaims, key string) (string, bool) {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok { return s, true }
	}
	return "", false
}

func scopesFromClaims(m jwt.MapClaims) []string {
	v, ok := m["scopes"]
	if !ok || v == nil { return nil }
	switch s := v.(type) {
	case []any:
		out := make([]string, 0, len(s))
		for _, x := range s { if str, ok := x.(string); ok { out = append(out, str) } }
		return out
	case []string:
		return s
	case string:
		// allow space or comma separated
		fields := strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == ',' })
		return fields
	default:
		return nil
	}
}

func rolesFromClaims(m jwt.MapClaims) []string {
	v, ok := m["roles"]
	if !ok || v == nil { return nil }
	switch s := v.(type) {
	case []any:
		out := make([]string, 0, len(s))
		for _, x := range s { if str, ok := x.(string); ok { out = append(out, str) } }
		return out
	case []string:
		return s
	case string:
		// allow space or comma separated
		fields := strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == ',' })
		return fields
	default:
		return nil
	}
}

func checkExpNbf(m jwt.MapClaims) error {
	now := time.Now().Unix()
	if expV, ok := m["exp"]; ok {
		exp, ok := toInt64(expV)
		if !ok || now >= exp { return errors.New("token expired") }
	}
	if nbfV, ok := m["nbf"]; ok {
		nbf, ok := toInt64(nbfV)
		if ok && now < nbf { return errors.New("token not yet valid") }
	}
	return nil
}

func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case int64:
		return x, true
	case json.Number:
		i, err := x.Int64(); if err == nil { return i, true }
	}
	return 0, false
}
