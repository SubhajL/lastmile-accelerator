use axum::{Router, routing::post};
use axum::http::{Request, StatusCode};
use tower::ServiceExt;

#[tokio::test]
async fn upsert_cve_unknown_field_returns_400() {
    let Some(pool) = dep_governance_service::db::migrate::test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };
    let app = Router::new()
        .route("/v1/cves", post(dep_governance_service::handlers::api::cves::upsert_cve_handler))
        .with_state(pool);

    let body = r#"{
        "cveId": "CVE-2025-0001",
        "severity": "low",
        "cvssScore": 4.5,
        "description": "desc",
        "publishedAt": null,
        "source": "OSV",
        "extra": "nope"
    }"#;
    let resp = app
        .oneshot(Request::builder().uri("/v1/cves").method("POST").header("content-type", "application/json").body(axum::body::Body::from(body)).unwrap())
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::BAD_REQUEST);
}

#[tokio::test]
async fn upsert_cve_invalid_cve_id_returns_400() {
    let Some(pool) = dep_governance_service::db::migrate::test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };
    let app = Router::new()
        .route("/v1/cves", post(dep_governance_service::handlers::api::cves::upsert_cve_handler))
        .with_state(pool);

    let body = r#"{
        "cveId": "CVE-20x5-1234",
        "severity": "low",
        "cvssScore": 4.5
    }"#;
    let resp = app
        .oneshot(Request::builder().uri("/v1/cves").method("POST").header("content-type", "application/json").body(axum::body::Body::from(body)).unwrap())
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::BAD_REQUEST);
}

#[tokio::test]
async fn upsert_cve_invalid_cvss_out_of_range_returns_400() {
    let Some(pool) = dep_governance_service::db::migrate::test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };
    let app = Router::new()
        .route("/v1/cves", post(dep_governance_service::handlers::api::cves::upsert_cve_handler))
        .with_state(pool);

    let body = r#"{
        "cveId": "CVE-2025-1234",
        "severity": "low",
        "cvssScore": 11.1
    }"#;
    let resp = app
        .oneshot(Request::builder().uri("/v1/cves").method("POST").header("content-type", "application/json").body(axum::body::Body::from(body)).unwrap())
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::BAD_REQUEST);
}

#[tokio::test]
async fn link_vuln_invalid_status_returns_400() {
    use axum::routing::post;
    use uuid::Uuid;
    let Some(pool) = dep_governance_service::db::migrate::test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };
    let app = Router::new()
        .route("/v1/dependencies/:id/vulns/link", post(dep_governance_service::handlers::api::cves::link_vuln_handler))
        .with_state(pool);

    let body = r#"{
        "cveId": "CVE-2025-0002",
        "status": "bogus"
    }"#;
    let dep_id = Uuid::new_v4();
    let uri = format!("/v1/dependencies/{}/vulns/link", dep_id);
    let resp = app
        .oneshot(Request::builder().uri(uri).method("POST").header("content-type", "application/json").body(axum::body::Body::from(body)).unwrap())
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::BAD_REQUEST);
}

#[tokio::test]
async fn sbom_unknown_field_returns_400() {
    let Some(pool) = dep_governance_service::db::migrate::test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };
    let app = Router::new()
        .route("/v1/snapshots/:snapshot_id/sbom", post(dep_governance_service::handlers::api::sbom::create_sbom_handler))
        .with_state(pool);

    let body = r#"{
        "format": "spdx_json",
        "storageKey": "s3://b/x.json",
        "fileHash": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
        "extra": true
    }"#;
    let resp = app
        .oneshot(Request::builder().uri("/v1/snapshots/00000000-0000-0000-0000-000000000000/sbom").method("POST").header("content-type", "application/json").body(axum::body::Body::from(body)).unwrap())
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::BAD_REQUEST);
}

#[tokio::test]
async fn extractor_json_syntax_error_returns_400() {
    let Some(pool) = dep_governance_service::db::migrate::test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };
    let app = Router::new()
        .route("/v1/cves", post(dep_governance_service::handlers::api::cves::upsert_cve_handler))
        .with_state(pool);

    let body = "{ not-json }";
    let resp = app
        .oneshot(Request::builder().uri("/v1/cves").method("POST").header("content-type", "application/json").body(axum::body::Body::from(body)).unwrap())
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::BAD_REQUEST);
}
