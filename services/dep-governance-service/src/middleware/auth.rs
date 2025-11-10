use axum::{extract::Request, middleware::Next, response::Response};
use serde::{Deserialize, Serialize};

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
    // Extract Authorization header
    let auth_header = request
        .headers()
        .get("authorization")
        .and_then(|h| h.to_str().ok());

    if let Some(auth) = auth_header {
        if auth.starts_with("Bearer ") {
            // In Phase 1, we'll accept any Bearer token for basic structure
            // Full JWT validation will be implemented in later phases
            tracing::debug!("Auth header present");
            return next.run(request).await;
        }
    }

    // For Phase 1, allow requests without auth for health endpoints
    // This will be enforced properly in later phases
    tracing::debug!("No auth header, proceeding anyway (Phase 1)");
    next.run(request).await
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
            .layer(middleware::from_fn(jwt_auth_middleware));

        let response = app
            .oneshot(Request::builder().uri("/test").body(Body::empty()).unwrap())
            .await
            .unwrap();

        assert_eq!(response.status(), StatusCode::OK);
    }
}
