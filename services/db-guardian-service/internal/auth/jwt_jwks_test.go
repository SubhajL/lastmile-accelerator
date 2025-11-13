package auth

import (
    "context"
    "crypto/rand"
    "crypto/rsa"
    "crypto/sha1"
    "crypto/x509"
    "encoding/base64"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    jwt "github.com/golang-jwt/jwt/v5"
)

type jwk struct{
    Kty string `json:"kty"`
    Kid string `json:"kid"`
    N   string `json:"n"`
    E   string `json:"e"`
}
type jwks struct{ Keys []jwk `json:"keys"` }

func makeRSAJWK(t *testing.T) (*rsa.PrivateKey, jwks, string) {
    t.Helper()
    priv, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil { t.Fatalf("rsa key: %v", err) }
    pub := &priv.PublicKey
    // kid: sha1 of public key DER
    der, _ := x509.MarshalPKIXPublicKey(pub)
    sum := sha1.Sum(der)
    kid := base64.RawURLEncoding.EncodeToString(sum[:])
    n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
    e := base64.RawURLEncoding.EncodeToString(bigEndianBytes(pub.E))
    return priv, jwks{Keys: []jwk{{Kty: "RSA", Kid: kid, N: n, E: e}}}, kid
}

func bigEndianBytes(e int) []byte {
    if e == 0 { return []byte{0} }
    out := []byte{}
    for e > 0 { out = append([]byte{byte(e & 0xff)}, out...); e >>= 8 }
    return out
}

func newJWKSHTTPServer(t *testing.T, set jwks) *httptest.Server {
    t.Helper()
    h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        _ = json.NewEncoder(w).Encode(set)
    })
    return httptest.NewServer(h)
}

func signJWT(t *testing.T, priv *rsa.PrivateKey, kid, iss, aud, sub string, scopes []string, exp time.Time) string {
    t.Helper()
    claims := jwt.MapClaims{
        "iss": iss,
        "aud": aud,
        "sub": sub,
        "exp": exp.Unix(),
        "iat": time.Now().Add(-time.Minute).Unix(),
        "nbf": time.Now().Add(-time.Minute).Unix(),
        "scope": joinScopes(scopes),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    token.Header["kid"] = kid
    s, err := token.SignedString(priv)
    if err != nil { t.Fatalf("sign: %v", err) }
    return s
}

func joinScopes(sc []string) string {
    if len(sc) == 0 { return "" }
    out := sc[0]
    for i:=1;i<len(sc);i++ { out += " " + sc[i] }
    return out
}

func TestJWTAuthenticator_Verify_ValidTokenOK(t *testing.T) {
    priv, set, kid := makeRSAJWK(t)
    srv := newJWKSHTTPServer(t, set)
    defer srv.Close()

    authn := NewJWTAuthenticator(srv.URL, "https://issuer/", "aud", 30*time.Second)
    tok := signJWT(t, priv, kid, "https://issuer/", "aud", "user-1", []string{"db.read","db.write"}, time.Now().Add(5*time.Minute))
    claims, err := authn.Verify(context.Background(), "Bearer "+tok, []string{"db.read"})
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if claims == nil || claims.Subject != "user-1" { t.Fatalf("bad claims: %+v", claims) }
}

func TestJWTAuthenticator_Verify_MissingScope(t *testing.T) {
    priv, set, kid := makeRSAJWK(t)
    srv := newJWKSHTTPServer(t, set)
    defer srv.Close()
    authn := NewJWTAuthenticator(srv.URL, "https://issuer/", "aud", 0)
    tok := signJWT(t, priv, kid, "https://issuer/", "aud", "user-1", []string{"db.read"}, time.Now().Add(5*time.Minute))
    if _, err := authn.Verify(context.Background(), "Bearer "+tok, []string{"db.write"}); err == nil {
        t.Fatalf("expected scope error")
    }
}

func TestJWTAuthenticator_Verify_BadIssuerAudienceSignature(t *testing.T) {
    priv, set, kid := makeRSAJWK(t)
    srv := newJWKSHTTPServer(t, set)
    defer srv.Close()
    authn := NewJWTAuthenticator(srv.URL, "https://issuer/", "aud", 0)

    // wrong audience
    tok := signJWT(t, priv, kid, "https://issuer/", "wrong-aud", "u", []string{}, time.Now().Add(5*time.Minute))
    if _, err := authn.Verify(context.Background(), "Bearer "+tok, nil); err == nil {
        t.Fatalf("expected audience error")
    }

    // wrong issuer
    tok = signJWT(t, priv, kid, "https://wrong/", "aud", "u", []string{}, time.Now().Add(5*time.Minute))
    if _, err := authn.Verify(context.Background(), "Bearer "+tok, nil); err == nil {
        t.Fatalf("expected issuer error")
    }

    // wrong signature (different key/kid)
    _, set2, _ := makeRSAJWK(t)
    srv2 := newJWKSHTTPServer(t, set2)
    defer srv2.Close()
    authn2 := NewJWTAuthenticator(srv2.URL, "https://issuer/", "aud", 0)
    tok = signJWT(t, priv, kid, "https://issuer/", "aud", "u", []string{}, time.Now().Add(5*time.Minute))
    if _, err := authn2.Verify(context.Background(), "Bearer "+tok, nil); err == nil {
        t.Fatalf("expected signature error")
    }
}
