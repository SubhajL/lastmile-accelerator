use axum::{
    http::{Request, StatusCode},
    middleware,
    Extension, Router,
};
use tower::ServiceExt; // for oneshot
use uuid::Uuid;

#[tokio::test]
async fn metrics_endpoint_exposes_counters_and_histograms_text_plain() {
    let app = test_app();

    // Trigger a request
    let _ = app
        .clone()
        .oneshot(Request::builder().uri("/healthz").body(axum::body::Body::empty()).unwrap())
        .await
        .unwrap();

    // Now query metrics
    let resp = app
        .oneshot(Request::builder().uri("/metrics").body(axum::body::Body::empty()).unwrap())
        .await
        .unwrap();

    assert_eq!(resp.status(), StatusCode::OK);
    let ct = resp.headers().get("content-type").unwrap().to_str().unwrap();
    assert!(ct.contains("text/plain"));
    let body = axum::body::to_bytes(resp.into_body(), 1024 * 1024).await.unwrap();
    let s = String::from_utf8(body.to_vec()).unwrap();
    assert!(s.contains("dep_governance_http_requests_total"));
    assert!(s.contains("dep_governance_http_request_duration_seconds_bucket"));
}

#[tokio::test]
async fn request_counter_increments_on_healthz_success() {
    let app = test_app();
    let _ = app.clone().oneshot(Request::builder().uri("/healthz").body(axum::body::Body::empty()).unwrap()).await.unwrap();

    let resp = app
        .oneshot(Request::builder().uri("/metrics").body(axum::body::Body::empty()).unwrap())
        .await
        .unwrap();
    let body = axum::body::to_bytes(resp.into_body(), 1024 * 1024).await.unwrap();
    let s = String::from_utf8(body.to_vec()).unwrap();
    assert!(s.contains("route=\"/healthz\""));
    assert!(s.contains("status=\"200\""));
}

#[tokio::test]
async fn request_counter_increments_on_unauthorized_v1() {
    let app = test_app();
    // No Authorization header, expect 401
    let resp = app
        .clone()
        .oneshot(Request::builder().uri("/v1/ping").body(axum::body::Body::empty()).unwrap())
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::UNAUTHORIZED);

    let resp2 = app
        .oneshot(Request::builder().uri("/metrics").body(axum::body::Body::empty()).unwrap())
        .await
        .unwrap();
    let body = axum::body::to_bytes(resp2.into_body(), 1024 * 1024).await.unwrap();
    let s = String::from_utf8(body.to_vec()).unwrap();
    assert!(s.contains("route=\"/v1/ping\""));
    assert!(s.contains("status=\"401\""));
}

#[tokio::test]
async fn metrics_db_histogram_records_create_sbom_op() {
    let Some(pool) = dep_governance_service::db::migrate::test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };
    // Perform a DB write directly to trigger DB metrics
    let sbom = dep_governance_service::models::Sbom::new(
        Uuid::new_v4(),
        dep_governance_service::models::SbomFormat::SpdxJson,
        "s3://bucket/metrics-test.json",
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
        None,
    ).unwrap();
    let _ = dep_governance_service::db::sboms::create_sbom(&pool, &sbom).await;

    let app = test_app();
    let resp = app
        .oneshot(Request::builder().uri("/metrics").body(axum::body::Body::empty()).unwrap())
        .await
        .unwrap();
    let body = axum::body::to_bytes(resp.into_body(), 1024 * 1024).await.unwrap();
    let s = String::from_utf8(body.to_vec()).unwrap();
    assert!(s.contains("dep_governance_db_query_duration_seconds_bucket"));
    assert!(s.contains("op=\"create_sbom\""));
}

fn test_app() -> Router {
    use dep_governance_service::middleware::{jwt_auth_middleware, AuthConfig, AuthContext};
    use dep_governance_service::middleware::auth::{JwksProvider, NoopJwksProvider};
    use std::sync::Arc;

    let provider: Arc<dyn JwksProvider> = Arc::new(NoopJwksProvider);
    let auth_ctx = AuthContext::new(AuthConfig { jwks_url: None, issuer: None, audience: None }, provider);

    axum::Router::new()
        .route("/healthz", axum::routing::get(dep_governance_service::handlers::healthz))
        .route("/metrics", axum::routing::get(dep_governance_service::handlers::metrics))
        .nest(
            "/v1",
            axum::Router::new()
                .route("/ping", axum::routing::get(|| async { StatusCode::OK }))
                .layer(middleware::from_fn(jwt_auth_middleware))
                .layer(Extension(auth_ctx)),
        )
        .layer(dep_governance_service::middleware::metrics_layer())
}
