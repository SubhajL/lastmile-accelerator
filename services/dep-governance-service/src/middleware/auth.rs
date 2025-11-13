use axum::{extract::Request, middleware::Next, response::Response};
use serde::{Deserialize, Serialize};
use axum::http::HeaderMap;
use axum::Extension;

// Scaffold auth context/config and helpers. Full JWT verification comes in later stacks.

#[derive(Debug, Clone)]
pub struct AuthConfig {
    pub jwks_url: Option<String>,
    pub issuer: Option<String>,
    pub audience: Option<String>,
}

pub trait JwksProvider: Send + Sync {
    fn get_key(&self, _kid: &str) -> Option<jsonwebtoken::DecodingKey> { None }
}

#[derive(Debug, Default)]
struct NoopJwksProvider;

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
            config: AuthConfig { jwks_url: None, issuer: None, audience: None },
            provider: std::sync::Arc::new(NoopJwksProvider),
        }
    }

    pub fn scaffold(config: AuthConfig) -> Self {
        Self { config, provider: std::sync::Arc::new(NoopJwksProvider) }
    }

    #[allow(dead_code)]
    pub fn provider(&self) -> &dyn JwksProvider { &*self.provider }
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

    // Extract Extension<AuthContext> if present (added by router layer)
    let _ctx: Option<Extension<AuthContext>> = req.extensions().get::<AuthContext>().cloned().map(Extension);

    if let Some(token) = extract_bearer_token(req.headers()) {
        let claims = parse_and_build_placeholder_claims(&token);
        attach_claims_extension(&mut req, claims);
        tracing::debug!("Auth header present; placeholder claims attached");
        return next.run(req).await;
    }

    tracing::debug!("No/invalid auth header; proceeding without claims (scaffold)");
    next.run(req).await
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

#[cfg(test)]
mod tests {
    use super::*;
    use axum::{body::Body, http::{Request, StatusCode}, middleware, Router};
    use tower::ServiceExt;

    async fn handler() -> &'static str {
        "ok"
    }

    #[tokio::test]
    async fn test_jwt_middleware_accepts_bearer_token() {
        let app = Router::new()
            .route("/test", axum::routing::get(handler))
            .layer(Extension(AuthContext::for_tests()))
            .layer(middleware::from_fn(jwt_auth_middleware));

        let response = app
            .oneshot(
                Request::builder()
                    .uri("/test")
                    .header("authorization", "Bearer test-token")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();

        assert_eq!(response.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_jwt_middleware_accepts_no_token_phase1() {
        let app = Router::new()
            .route("/test", axum::routing::get(handler))
            .layer(Extension(AuthContext::for_tests()))
            .layer(middleware::from_fn(jwt_auth_middleware));

        let response = app
            .oneshot(Request::builder().uri("/test").body(Body::empty()).unwrap())
            .await
            .unwrap();

        assert_eq!(response.status(), StatusCode::OK);
    }

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
}
