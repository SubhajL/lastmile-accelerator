use axum::http::HeaderMap;
use axum::response::IntoResponse;
use axum::{extract::Request, middleware::Next, response::Response};
use serde::{Deserialize, Serialize};

use crate::error::AppError;
use jsonwebtoken::{self, decode, decode_header, Algorithm, Validation};
// Scaffold auth context/config and helpers. Full JWT verification comes in later stacks.

#[derive(Debug, Clone)]
pub struct AuthConfig {
    pub jwks_url: Option<String>,
    pub issuer: Option<String>,
    pub audience: Option<String>,
}

pub trait JwksProvider: Send + Sync {
    fn get_key<'a>(
        &'a self,
        _kid: &'a str,
    ) -> futures::future::BoxFuture<'a, Option<std::sync::Arc<jsonwebtoken::DecodingKey>>> {
        Box::pin(async { None })
    }
}

#[derive(Debug, Default)]
pub struct NoopJwksProvider;

impl JwksProvider for NoopJwksProvider {}

#[derive(Clone)]
pub struct AuthContext {
    pub config: AuthConfig,
    // Boxed provider to allow future swap with real JWKS client
    provider: std::sync::Arc<dyn JwksProvider>,
}

impl AuthContext {
    pub fn new(config: AuthConfig, provider: std::sync::Arc<dyn JwksProvider>) -> Self {
        Self { config, provider }
    }

    pub fn for_tests() -> Self {
        Self {
            config: AuthConfig {
                jwks_url: None,
                issuer: None,
                audience: None,
            },
            provider: std::sync::Arc::new(NoopJwksProvider),
        }
    }

    pub fn scaffold(config: AuthConfig) -> Self {
        Self {
            config,
            provider: std::sync::Arc::new(NoopJwksProvider),
        }
    }

    #[allow(dead_code)]
    pub fn provider(&self) -> &dyn JwksProvider {
        &*self.provider
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Claims {
    pub sub: String,
    pub tenant_id: String,
    pub user_id: String,
    pub scopes: Vec<String>,
    pub exp: usize,
    pub iss: String,
}

pub async fn jwt_auth_middleware(request: Request, next: Next) -> Response {
    let mut req = request;

    let ctx = match req.extensions().get::<AuthContext>() {
        Some(c) => c,
        None => return AppError::Auth("Auth context not configured".into()).into_response(),
    };

    let token = match extract_bearer_token(req.headers()) {
        Some(t) => t,
        None => return AppError::Auth("Missing bearer token".into()).into_response(),
    };

    match verify_and_extract_claims(&token, ctx).await {
        Ok(claims) => {
            attach_claims_extension(&mut req, claims);
            next.run(req).await
        }
        Err(e) => e.into_response(),
    }
}

// Helpers
pub fn extract_bearer_token(headers: &HeaderMap) -> Option<String> {
    let raw = headers.get("authorization")?.to_str().ok()?;
    let trimmed = raw.trim();
    let prefix = "Bearer ";
    if trimmed.len() >= prefix.len() && trimmed[..prefix.len()].eq_ignore_ascii_case(prefix) {
        Some(trimmed[prefix.len()..].trim().to_string())
    } else {
        None
    }
}

pub fn attach_claims_extension(req: &mut Request, claims: Claims) {
    req.extensions_mut().insert::<Claims>(claims);
}

pub fn parse_and_build_placeholder_claims(token: &str) -> Claims {
    // Non-validating placeholder; derive deterministic sub from token
    use sha2::{Digest, Sha256};
    let mut hasher = Sha256::new();
    hasher.update(token.as_bytes());
    let hash = format!("{:x}", hasher.finalize());
    Claims {
        sub: hash,
        tenant_id: "unknown".into(),
        user_id: "unknown".into(),
        scopes: vec![],
        exp: 0,
        iss: "unknown".into(),
    }
}

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct RawClaims {
    #[serde(default)]
    sub: Option<String>,
    #[serde(default)]
    tenant_id: Option<String>,
    #[serde(default)]
    user_id: Option<String>,
    #[serde(default)]
    scopes: Option<Vec<String>>,
    #[serde(default)]
    exp: Option<usize>,
    #[serde(default)]
    iss: Option<String>,
    #[serde(default)]
    aud: Option<serde_json::Value>,
}

async fn verify_and_extract_claims(token: &str, ctx: &AuthContext) -> Result<Claims, AppError> {
    let header =
        decode_header(token).map_err(|e| AppError::Auth(format!("Invalid JWT header: {}", e)))?;
    if header.alg != Algorithm::RS256 {
        return Err(AppError::Auth("Unsupported JWT alg".into()));
    }
    let kid = header
        .kid
        .ok_or_else(|| AppError::Auth("JWT missing kid".into()))?;
    let key = ctx
        .provider()
        .get_key(&kid)
        .await
        .ok_or_else(|| AppError::Auth("Unknown key id".into()))?;

    let mut validation = Validation::new(Algorithm::RS256);
    if let Some(ref iss) = ctx.config.issuer {
        validation.set_issuer(std::slice::from_ref(iss));
    }
    if let Some(ref aud) = ctx.config.audience {
        validation.set_audience(std::slice::from_ref(aud));
    }

    let data = decode::<RawClaims>(token, &key, &validation)
        .map_err(|e| AppError::Auth(format!("JWT validation failed: {}", e)))?;
    let rc = data.claims;
    Ok(Claims {
        sub: rc.sub.unwrap_or_else(|| "unknown".into()),
        tenant_id: rc.tenant_id.unwrap_or_else(|| "unknown".into()),
        user_id: rc.user_id.unwrap_or_else(|| "unknown".into()),
        scopes: rc.scopes.unwrap_or_default(),
        exp: rc.exp.unwrap_or(0),
        iss: rc.iss.unwrap_or_else(|| "unknown".into()),
    })
}

// --- JWKS caching provider ---

#[derive(Debug, Clone, Deserialize)]
struct Jwks {
    keys: Vec<Jwk>,
}

#[derive(Debug, Clone, Deserialize)]
struct Jwk {
    #[serde(default)]
    kid: String,
    #[serde(default)]
    alg: Option<String>,
    kty: String,
    #[serde(default)]
    n: Option<String>,
    #[serde(default)]
    e: Option<String>,
}

struct CacheState {
    map: std::collections::HashMap<String, std::sync::Arc<jsonwebtoken::DecodingKey>>,
    expires_at: std::time::Instant,
}

pub struct CachingJwksProvider {
    url: String,
    ttl: std::time::Duration,
    client: reqwest::Client,
    state: std::sync::Arc<tokio::sync::RwLock<CacheState>>,
}

impl CachingJwksProvider {
    pub fn new(url: String, ttl: std::time::Duration) -> Self {
        let client = reqwest::Client::builder()
            .user_agent("dep-governance-service-jwks/1.0")
            .build()
            .expect("reqwest client");
        Self {
            url,
            ttl,
            client,
            state: std::sync::Arc::new(tokio::sync::RwLock::new(CacheState {
                map: std::collections::HashMap::new(),
                expires_at: std::time::Instant::now(),
            })),
        }
    }

    fn next_expiry(ttl: std::time::Duration) -> std::time::Instant {
        let now = std::time::Instant::now();
        // jitter up to 10%
        let jitter_ns = (ttl.as_nanos() as f64 * 0.1) as u128;
        let jitter = if jitter_ns == 0 {
            0
        } else {
            rand::random::<u64>() as u128 % jitter_ns
        };
        let ttl_ns = ttl.as_nanos().saturating_sub(jitter);
        now + std::time::Duration::from_nanos(ttl_ns as u64)
    }

    async fn refresh_if_needed(&self) {
        let expired = {
            let guard = self.state.read().await;
            std::time::Instant::now() >= guard.expires_at
        };
        if !expired {
            return;
        }

        let mut guard = self.state.write().await;
        if std::time::Instant::now() < guard.expires_at {
            return;
        }

        match self
            .client
            .get(&self.url)
            .send()
            .await
            .and_then(|r| r.error_for_status())
        {
            Ok(resp) => {
                if let Ok(jwks) = resp.json::<Jwks>().await {
                    let mut newmap = std::collections::HashMap::new();
                    for k in jwks.keys {
                        if let (Some(n), Some(e)) = (k.n.as_ref(), k.e.as_ref()) {
                            if k.alg.as_deref() == Some("RS256")
                                || (k.alg.is_none() && k.kty == "RSA")
                            {
                                if let Ok(dk) = jsonwebtoken::DecodingKey::from_rsa_components(n, e)
                                {
                                    newmap.insert(k.kid.clone(), std::sync::Arc::new(dk));
                                }
                            }
                        }
                        // ES256 support can be added in later stacks
                    }
                    guard.map = newmap;
                    guard.expires_at = Self::next_expiry(self.ttl);
                }
            }
            Err(err) => {
                tracing::warn!("JWKS fetch failed: {}", err);
                guard.expires_at = Self::next_expiry(self.ttl);
            }
        }
    }
}

impl JwksProvider for CachingJwksProvider {
    fn get_key<'a>(
        &'a self,
        kid: &'a str,
    ) -> futures::future::BoxFuture<'a, Option<std::sync::Arc<jsonwebtoken::DecodingKey>>> {
        Box::pin(async move {
            self.refresh_if_needed().await;
            let guard = self.state.read().await;
            guard.map.get(kid).cloned()
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::{
        body::Body,
        http::{Request, StatusCode},
        middleware, Extension, Router,
    };
    use base64::engine::general_purpose::URL_SAFE_NO_PAD;
    use base64::Engine;
    use chrono::Utc;
    use jsonwebtoken::{encode, Algorithm, EncodingKey, Header};
    use rand::{rngs::StdRng, RngCore, SeedableRng};
    use rsa::{pkcs1::EncodeRsaPrivateKey, traits::PublicKeyParts, RsaPrivateKey};
    use tower::ServiceExt;
    use wiremock::matchers::{method, path};
    use wiremock::{Mock, MockServer, ResponseTemplate};

    async fn handler() -> &'static str {
        "ok"
    }

    // Phase 3 enforces auth; legacy permissive tests removed.

    #[test]
    fn extract_bearer_token_parses_bearer_value() {
        let mut headers = HeaderMap::new();
        headers.insert("authorization", "Bearer abc.def.ghi".parse().unwrap());
        let token = extract_bearer_token(&headers);
        assert_eq!(token.as_deref(), Some("abc.def.ghi"));
    }

    #[test]
    fn extract_bearer_token_ignores_invalid_schemes() {
        let mut headers = HeaderMap::new();
        headers.insert("authorization", "Basic something".parse().unwrap());
        let token = extract_bearer_token(&headers);
        assert!(token.is_none());
    }

    #[tokio::test]
    async fn caching_jwks_provider_fetches_and_caches_rs256() {
        let server = MockServer::start().await;

        // Build a fake RSA modulus (base64url) and exponent AQAB (65537)
        let mut rng = StdRng::seed_from_u64(42);
        let mut modulus = vec![0u8; 256];
        rng.fill_bytes(&mut modulus);
        let n_b64 = URL_SAFE_NO_PAD.encode(&modulus);
        let jwks = serde_json::json!({
            "keys": [
                {"kty": "RSA", "kid": "kid1", "use": "sig", "alg": "RS256", "n": n_b64, "e": "AQAB"}
            ]
        });

        Mock::given(method("GET"))
            .and(path("/jwks.json"))
            .respond_with(ResponseTemplate::new(200).set_body_json(&jwks))
            .expect(1)
            .mount(&server)
            .await;

        let url = format!("{}/jwks.json", &server.uri());
        let prov = CachingJwksProvider::new(url, std::time::Duration::from_secs(300));

        // First fetch triggers network
        let k1 = prov.get_key("kid1").await;
        assert!(k1.is_some());

        // Second fetch served from cache
        let k2 = prov.get_key("kid1").await;
        assert!(k2.is_some());
    }

    #[tokio::test]
    async fn caching_jwks_provider_refreshes_on_ttl_expiry() {
        let server = MockServer::start().await;

        let n1 = URL_SAFE_NO_PAD.encode(vec![1u8; 256]);
        let jwks1 = serde_json::json!({"keys":[{"kty":"RSA","kid":"kid2","use":"sig","alg":"RS256","n":n1,"e":"AQAB"}]});

        let n2 = URL_SAFE_NO_PAD.encode(vec![2u8; 256]);
        let jwks2 = serde_json::json!({"keys":[{"kty":"RSA","kid":"kid2","use":"sig","alg":"RS256","n":n2,"e":"AQAB"}]});

        Mock::given(method("GET"))
            .and(path("/jwks.json"))
            .respond_with(ResponseTemplate::new(200).set_body_json(&jwks1))
            .mount(&server)
            .await;

        let url = format!("{}/jwks.json", &server.uri());
        let prov = CachingJwksProvider::new(url, std::time::Duration::from_millis(50));

        // First call populates cache
        assert!(prov.get_key("kid2").await.is_some());

        // After TTL, next request should refetch
        // Mount updated response for next fetch
        Mock::given(method("GET"))
            .and(path("/jwks.json"))
            .respond_with(ResponseTemplate::new(200).set_body_json(&jwks2))
            .mount(&server)
            .await;

        tokio::time::sleep(std::time::Duration::from_millis(60)).await;
        assert!(prov.get_key("kid2").await.is_some());
    }

    #[tokio::test]
    async fn caching_jwks_provider_ignores_unsupported_alg() {
        let server = MockServer::start().await;
        let jwks = serde_json::json!({
            "keys": [ {"kty": "oct", "kid": "bad", "alg":"HS256", "k":"abcd"} ]
        });
        Mock::given(method("GET"))
            .and(path("/jwks.json"))
            .respond_with(ResponseTemplate::new(200).set_body_json(&jwks))
            .mount(&server)
            .await;

        let url = format!("{}/jwks.json", &server.uri());
        let prov = CachingJwksProvider::new(url, std::time::Duration::from_secs(300));
        assert!(prov.get_key("bad").await.is_none());
    }

    fn make_rsa_jwks_and_pem(kid: &str) -> (String, serde_json::Value) {
        let mut rng = rand::thread_rng();
        let priv_key = RsaPrivateKey::new(&mut rng, 2048).expect("gen key");
        let pub_key = rsa::RsaPublicKey::from(&priv_key);
        let n_b64 = URL_SAFE_NO_PAD.encode(pub_key.n().to_bytes_be());
        let e_b64 = URL_SAFE_NO_PAD.encode(pub_key.e().to_bytes_be());
        let pem = priv_key
            .to_pkcs1_pem(rsa::pkcs1::LineEnding::LF)
            .unwrap()
            .to_string();
        let jwks = serde_json::json!({
            "keys": [{"kty":"RSA","kid":kid,"alg":"RS256","use":"sig","n":n_b64,"e":e_b64}]
        });
        (pem, jwks)
    }

    #[tokio::test]
    async fn valid_jwt_allows_request() {
        let server = MockServer::start().await;
        let kid = "kid-valid";
        let (pem, jwks) = make_rsa_jwks_and_pem(kid);
        Mock::given(method("GET"))
            .and(path("/jwks.json"))
            .respond_with(ResponseTemplate::new(200).set_body_json(&jwks))
            .mount(&server)
            .await;

        let iss = "test-iss".to_string();
        let aud = "test-aud".to_string();
        let url = format!("{}/jwks.json", server.uri());
        let prov = CachingJwksProvider::new(url.clone(), std::time::Duration::from_secs(300));
        // warm cache
        assert!(prov.get_key(kid).await.is_some());
        let auth_ctx = AuthContext::new(
            AuthConfig {
                jwks_url: Some(url),
                issuer: Some(iss.clone()),
                audience: Some(aud.clone()),
            },
            std::sync::Arc::new(prov),
        );

        #[derive(serde::Serialize)]
        struct TestClaims {
            sub: String,
            tenant_id: String,
            user_id: String,
            scopes: Vec<String>,
            exp: usize,
            iss: String,
            aud: String,
        }
        let claims = TestClaims {
            sub: "user-1".into(),
            tenant_id: "t-1".into(),
            user_id: "u-1".into(),
            scopes: vec![],
            exp: (Utc::now().timestamp() as usize) + 3600,
            iss: iss.clone(),
            aud: aud.clone(),
        };
        let mut header = Header::new(Algorithm::RS256);
        header.kid = Some(kid.into());
        let token = encode(
            &header,
            &claims,
            &EncodingKey::from_rsa_pem(pem.as_bytes()).unwrap(),
        )
        .unwrap();

        // Sanity: token decodes with same validation
        let mut v = jsonwebtoken::Validation::new(jsonwebtoken::Algorithm::RS256);
        v.set_issuer(std::slice::from_ref(&iss));
        v.set_audience(std::slice::from_ref(&aud));
        let n = jwks["keys"][0]["n"].as_str().unwrap();
        let e = jwks["keys"][0]["e"].as_str().unwrap();
        let _ = jsonwebtoken::decode::<serde_json::Value>(
            &token,
            &jsonwebtoken::DecodingKey::from_rsa_components(n, e).unwrap(),
            &v,
        )
        .unwrap();

        // Also ensure our verifier path succeeds
        verify_and_extract_claims(&token, &auth_ctx).await.unwrap();

        let app = Router::new()
            .route("/v1/protected", axum::routing::get(handler))
            .layer(middleware::from_fn(jwt_auth_middleware))
            .layer(Extension(auth_ctx));

        let res = app
            .oneshot(
                Request::builder()
                    .uri("/v1/protected")
                    .header("authorization", format!("Bearer {}", token))
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(res.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn missing_token_returns_401() {
        let auth_ctx = AuthContext::for_tests();
        let app = Router::new()
            .route("/v1/protected", axum::routing::get(handler))
            .layer(middleware::from_fn(jwt_auth_middleware))
            .layer(Extension(auth_ctx));

        let res = app
            .oneshot(
                Request::builder()
                    .uri("/v1/protected")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(res.status(), StatusCode::UNAUTHORIZED);
    }

    #[tokio::test]
    async fn expired_token_returns_401() {
        let server = MockServer::start().await;
        let kid = "kid-exp";
        let (pem, jwks) = make_rsa_jwks_and_pem(kid);
        Mock::given(method("GET"))
            .and(path("/jwks.json"))
            .respond_with(ResponseTemplate::new(200).set_body_json(&jwks))
            .mount(&server)
            .await;
        let iss = "test-iss".to_string();
        let aud = "test-aud".to_string();
        let auth_ctx = AuthContext::new(
            AuthConfig {
                jwks_url: Some(format!("{}/jwks.json", server.uri())),
                issuer: Some(iss.clone()),
                audience: Some(aud.clone()),
            },
            std::sync::Arc::new(CachingJwksProvider::new(
                format!("{}/jwks.json", server.uri()),
                std::time::Duration::from_secs(1),
            )),
        );

        #[derive(serde::Serialize)]
        struct TestClaims {
            sub: String,
            tenant_id: String,
            user_id: String,
            scopes: Vec<String>,
            exp: usize,
            iss: String,
            aud: String,
        }
        let claims = TestClaims {
            sub: "user-1".into(),
            tenant_id: "t-1".into(),
            user_id: "u-1".into(),
            scopes: vec![],
            exp: (Utc::now().timestamp() as usize) - 10,
            iss: iss.clone(),
            aud: aud.clone(),
        };
        let mut header = Header::new(Algorithm::RS256);
        header.kid = Some(kid.into());
        let token = encode(
            &header,
            &claims,
            &EncodingKey::from_rsa_pem(pem.as_bytes()).unwrap(),
        )
        .unwrap();

        let app = Router::new()
            .route("/v1/protected", axum::routing::get(handler))
            .layer(Extension(auth_ctx))
            .layer(middleware::from_fn(jwt_auth_middleware));

        let res = app
            .oneshot(
                Request::builder()
                    .uri("/v1/protected")
                    .header("authorization", format!("Bearer {}", token))
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(res.status(), StatusCode::UNAUTHORIZED);
    }
}
