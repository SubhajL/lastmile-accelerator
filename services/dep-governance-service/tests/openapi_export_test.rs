use tower::ServiceExt;

#[tokio::test]
async fn openapi_yaml_includes_bearer_and_paths() {
    let doc = dep_governance_service::openapi::doc();
    let yaml = serde_yaml::to_string(&serde_json::to_value(&doc).unwrap()).unwrap();
    assert!(yaml.contains("bearerAuth"));
    assert!(yaml.contains("/v1/snapshots/{snapshot_id}/sbom"));
    assert!(yaml.contains("/v1/cves"));
}

#[tokio::test]
async fn openapi_json_matches_runtime_route() {
    let expected = serde_json::to_value(dep_governance_service::openapi::doc()).unwrap();
    let res = axum::Router::new()
        .merge(dep_governance_service::openapi::routes())
        .into_service()
        .oneshot(axum::http::Request::builder().uri("/openapi.json").body(axum::body::Body::empty()).unwrap())
        .await
        .unwrap();
    assert_eq!(res.status(), axum::http::StatusCode::OK);
    let body = axum::body::to_bytes(res.into_body(), 1024 * 1024).await.unwrap();
    let got: serde_json::Value = serde_json::from_slice(&body).unwrap();
    assert_eq!(got, expected);
}
