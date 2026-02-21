package auth

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

    jwt "github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
    jwksURL string
    issuer  string
    audience string
    skew    time.Duration
    httpc   *http.Client
    mu      sync.RWMutex
    keys    map[string]*rsa.PublicKey // kid -> key
}

func NewJWTAuthenticator(jwksURL, issuer, audience string, clockSkew time.Duration) *JWTAuthenticator {
    return &JWTAuthenticator{
        jwksURL: jwksURL,
        issuer: issuer,
        audience: audience,
        skew: clockSkew,
        httpc: &http.Client{Timeout: 5 * time.Second},
        keys: make(map[string]*rsa.PublicKey),
    }
}

func (a *JWTAuthenticator) Verify(ctx context.Context, bearer string, required []string) (*Claims, error) {
    raw, err := parseBearer(bearer)
    if err != nil { return nil, err }

    // keyfunc loads key by kid
    keyFunc := func(t *jwt.Token) (interface{}, error) {
        if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
            return nil, fmt.Errorf("unexpected alg: %s", t.Method.Alg())
        }
        kid, _ := t.Header["kid"].(string)
        if kid == "" { return nil, errors.New("missing kid") }
        key, err := a.fetchKey(ctx, kid)
        if err != nil { return nil, err }
        return key, nil
    }

    var claims jwt.MapClaims
    tok, err := jwt.ParseWithClaims(raw, &claims, keyFunc)
    if err != nil || !tok.Valid { return nil, errors.New("invalid token") }

    now := time.Now()
    // Issuer
    if iss, _ := claims["iss"].(string); iss != a.issuer { return nil, errors.New("issuer mismatch") }
    // Audience: accept string or []string
    if audv, ok := claims["aud"]; ok {
        switch v := audv.(type) {
        case string:
            if v != a.audience { return nil, errors.New("audience mismatch") }
        case []interface{}:
            found := false
            for _, it := range v { if s, ok := it.(string); ok && s == a.audience { found = true; break } }
            if !found { return nil, errors.New("audience mismatch") }
        }
    } else { return nil, errors.New("audience missing") }
    // Times: exp, nbf, iat with skew
    if ev, ok := claims["exp"].(float64); ok {
        exp := time.Unix(int64(ev), 0)
        if now.After(exp.Add(a.skew)) { return nil, errors.New("token expired") }
    }
    if nbfv, ok := claims["nbf"].(float64); ok {
        nbf := time.Unix(int64(nbfv), 0)
        if now.Add(a.skew).Before(nbf) { return nil, errors.New("token not yet valid") }
    }
    if iatv, ok := claims["iat"].(float64); ok {
        iat := time.Unix(int64(iatv), 0)
        // disallow far-future iat
        if iat.After(now.Add(a.skew)) { return nil, errors.New("token issued in future") }
    }

    subject, _ := claims["sub"].(string)
    scopes := extractScopes(claims)
    if !hasAllScopes(scopes, required) { return nil, errors.New("missing scope") }
    return &Claims{Subject: subject, Scopes: scopes}, nil
}

func extractScopes(c jwt.MapClaims) []string {
    if s, ok := c["scope"].(string); ok && s != "" {
        parts := strings.Fields(s)
        return parts
    }
    if a, ok := c["scopes"].([]interface{}); ok {
        out := make([]string, 0, len(a))
        for _, v := range a { if s, ok := v.(string); ok { out = append(out, s) } }
        return out
    }
    return []string{}
}

func (a *JWTAuthenticator) fetchKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
    a.mu.RLock()
    if k, ok := a.keys[kid]; ok { a.mu.RUnlock(); return k, nil }
    a.mu.RUnlock()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.jwksURL, nil)
    if err != nil { return nil, err }
    resp, err := a.httpc.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 { return nil, fmt.Errorf("jwks status %d", resp.StatusCode) }
    var body struct{ Keys []struct{ Kty, Kid, N, E string } `json:"keys"` }
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil { return nil, err }
    for _, k := range body.Keys {
        if k.Kid != kid { continue }
        if strings.ToUpper(k.Kty) != "RSA" { continue }
        nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
        if err != nil { return nil, err }
        eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
        if err != nil { return nil, err }
        n := new(big.Int).SetBytes(nBytes)
        var eInt int
        for _, b := range eBytes { eInt = (eInt<<8) | int(b) }
        pub := &rsa.PublicKey{N: n, E: eInt}
        a.mu.Lock(); a.keys[kid] = pub; a.mu.Unlock()
        return pub, nil
    }
    return nil, errors.New("kid not found")
}

func hasAllScopes(have, required []string) bool {
    if len(required) == 0 { return true }
    set := make(map[string]struct{}, len(have))
    for _, s := range have { set[s] = struct{}{} }
    for _, r := range required { if _, ok := set[r]; !ok { return false } }
    return true
}

func parseBearer(h string) (string, error) {
    if !strings.HasPrefix(h, "Bearer ") { return "", errors.New("invalid authorization header") }
    tok := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
    if tok == "" { return "", errors.New("empty token") }
    return tok, nil
}
