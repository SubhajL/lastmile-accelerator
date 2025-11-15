use axum::http::StatusCode;
use tower::ServiceExt; // for oneshot

#[tokio::test]
async fn openapi_json_exposes_v1_paths_and_security() {
    // Build a minimal app with only OpenAPI/Swagger routes
    let res = axum::http::Request::builder()
        .uri("/openapi.json")
        .method("GET")
        .body(axum::body::Body::empty())
        .unwrap();

    let resp = axum::Router::new()
        .merge(dep_governance_service::openapi::routes())
        .into_service()
        .oneshot(res)
        .await
        .unwrap();

    assert_eq!(resp.status(), StatusCode::OK);

    let body_bytes = axum::body::to_bytes(resp.into_body(), 1024 * 1024).await.unwrap();
    let doc: serde_json::Value = serde_json::from_slice(&body_bytes).unwrap();

    let paths = doc.get("paths").and_then(|v| v.as_object()).expect("paths object");
    assert!(paths.contains_key("/v1/snapshots/{snapshot_id}/sbom"));
    assert!(paths.contains_key("/v1/snapshots/{snapshot_id}/dependencies"));
    assert!(paths.contains_key("/v1/dependencies/{dependency_id}/vulns"));
    assert!(paths.contains_key("/v1/dependencies/{dependency_id}/vulns/link"));
    assert!(paths.contains_key("/v1/cves"));

    // Security: bearerAuth should be present globally or per-path
    let has_global = doc.get("security").is_some();
    let mut has_per_path = false;
    for p in paths.values() {
        if p.get("get").and_then(|m| m.get("security")).is_some()
            || p.get("post").and_then(|m| m.get("security")).is_some()
        {
            has_per_path = true;
            break;
        }
    }
    assert!(has_global || has_per_path, "expected security requirements");
}

#[tokio::test]
async fn swagger_ui_served_at_docs_path() {
    let res = axum::http::Request::builder()
        .uri("/docs/")
        .method("GET")
        .body(axum::body::Body::empty())
        .unwrap();
    let resp = axum::Router::new().merge(dep_governance_service::openapi::routes()).into_service().oneshot(res).await.unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
}
