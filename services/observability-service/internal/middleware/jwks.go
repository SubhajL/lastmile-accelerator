package middleware

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type JWKSVerifier struct {
	jwksURL   string
	issuer    string
	audience  string
	scopeClaim string
	hc        *http.Client
	mu        sync.RWMutex
	keys      map[string]any // kid -> public key
	lastFetch time.Time
	ttl       time.Duration
}

func NewJWKSVerifier(jwksURL, issuer, audience, scopeClaim string, hc *http.Client) *JWKSVerifier {
	if hc == nil { hc = &http.Client{Timeout: 5 * time.Second} }
	if scopeClaim == "" { scopeClaim = "scopes" }
	return &JWKSVerifier{jwksURL: jwksURL, issuer: issuer, audience: audience, scopeClaim: scopeClaim, hc: hc, keys: map[string]any{}, ttl: 5 * time.Minute}
}

func (v *JWKSVerifier) Verify(ctx context.Context, token string) (*Claims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256","RS384","RS512","ES256","ES384","ES512"}))
	claims := jwt.MapClaims{}
	_, err := parser.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" { return nil, errors.New("missing kid") }
		key, err := v.getKey(ctx, kid)
		if err != nil { return nil, err }
		return key, nil
	})
	if err != nil { return nil, fmt.Errorf("jwt parse: %w", err) }
	if iss, _ := claims["iss"].(string); v.issuer != "" && iss != v.issuer { return nil, fmt.Errorf("issuer mismatch") }
	// Audience optional
	if v.audience != "" {
		if !verifyAudience(claims, v.audience) { return nil, fmt.Errorf("audience mismatch") }
	}
	sub, _ := claims["sub"].(string)
	projectID, _ := claims["project_id"].(string)
	scopes := parseScopes(claims[v.scopeClaim])
	return &Claims{Subject: sub, ProjectID: projectID, Scopes: scopes, Raw: mapFromClaims(claims)}, nil
}

func (v *JWKSVerifier) getKey(ctx context.Context, kid string) (any, error) {
	v.mu.RLock(); key, ok := v.keys[kid]; exp := v.lastFetch.Add(v.ttl).Before(time.Now()); v.mu.RUnlock()
	if ok && !exp { return key, nil }
	if err := v.refresh(ctx); err != nil { return nil, err }
	v.mu.RLock(); key, ok = v.keys[kid]; v.mu.RUnlock()
	if !ok { return nil, fmt.Errorf("kid not found") }
	return key, nil
}

func (v *JWKSVerifier) refresh(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	resp, err := v.hc.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return fmt.Errorf("jwks http %d", resp.StatusCode) }
	var jwks struct{ Keys []map[string]any `json:"keys"` }
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil { return err }
	keys := map[string]any{}
	for _, k := range jwks.Keys {
		kid, _ := k["kid"].(string)
		kty, _ := k["kty"].(string)
		if kid == "" { continue }
		var pub any
		switch kty {
		case "RSA":
			ns, _ := k["n"].(string); es, _ := k["e"].(string)
			if ns != "" && es != "" { if pk, err := rsaFromModExp(ns, es); err == nil { pub = pk } }
			if pub == nil { // try x5c
				if x5c, ok := firstString(k["x5c"]); ok { if pk, err := rsaFromX5C(x5c); err == nil { pub = pk } }
			}
		case "EC":
			if x5c, ok := firstString(k["x5c"]); ok { if pk, err := ecFromX5C(x5c); err == nil { pub = pk } }
		}
		if pub != nil { keys[kid] = pub }
	}
	v.mu.Lock(); v.keys = keys; v.lastFetch = time.Now(); v.mu.Unlock()
	return nil
}

func firstString(v any) (string, bool) {
	arr, ok := v.([]any); if !ok || len(arr)==0 { return "", false }
	s, _ := arr[0].(string); return s, s != ""
}

func rsaFromModExp(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64); if err != nil { return nil, err }
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64); if err != nil { return nil, err }
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes).Int64()
	return &rsa.PublicKey{N: n, E: int(e)}, nil
}

func rsaFromX5C(x5c string) (*rsa.PublicKey, error) {
	der, err := base64.StdEncoding.DecodeString(x5c); if err != nil { return nil, err }
	cert, err := x509.ParseCertificate(der); if err != nil { return nil, err }
	pk, ok := cert.PublicKey.(*rsa.PublicKey); if !ok { return nil, errors.New("not rsa") }
	return pk, nil
}

func ecFromX5C(x5c string) (*ecdsa.PublicKey, error) {
	der, err := base64.StdEncoding.DecodeString(x5c); if err != nil { return nil, err }
	cert, err := x509.ParseCertificate(der); if err != nil { return nil, err }
	pk, ok := cert.PublicKey.(*ecdsa.PublicKey); if !ok { return nil, errors.New("not ec") }
	// ensure supported curve
	if pk.Curve != elliptic.P256() && pk.Curve != elliptic.P384() && pk.Curve != elliptic.P521() { return nil, errors.New("unsupported curve") }
	return pk, nil
}

func parseScopes(v any) []string {
	switch t := v.(type) {
	case []any:
		out := []string{}
		for _, x := range t { if s, ok := x.(string); ok { out = append(out, s) } }
		return out
	case []string:
		return t
	case string:
		sep := ","
		if strings.Contains(t, " ") { sep = " " }
		parts := strings.Split(t, sep)
		out := make([]string, 0, len(parts))
		for _, p := range parts { p = strings.TrimSpace(p); if p != "" { out = append(out, p) } }
		return out
	default:
		return nil
	}
}

func mapFromClaims(mc jwt.MapClaims) map[string]any {
	m := make(map[string]any, len(mc))
	for k, v := range mc { m[k] = v }
	return m
}

func verifyAudience(mc jwt.MapClaims, need string) bool {
	if need == "" { return true }
	if v, ok := mc["aud"]; ok {
		switch t := v.(type) {
		case string:
			return t == need
		case []any:
			for _, x := range t { if s, ok := x.(string); ok && s == need { return true } }
		case []string:
			for _, s := range t { if s == need { return true } }
		}
	}
	return false
}
