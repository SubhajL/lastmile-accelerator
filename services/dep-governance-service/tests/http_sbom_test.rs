use axum::{extract::Path, http::StatusCode};
use axum::response::IntoResponse;
use dep_governance_service::handlers::api::sbom::{create_sbom_handler, get_latest_sbom_handler, SbomCreateRequest};
use dep_governance_service::db::migrate::test_pool;
use axum::extract::State;
use axum::Json;
use uuid::Uuid;

#[tokio::test]
async fn test_post_and_get_sbom_happy_path() {
    let Some(pool) = test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };

    let snapshot_id = Uuid::new_v4();
    let req = SbomCreateRequest {
        format: "spdx_json".into(),
        storage_key: "s3://bucket/sboms/123.json".into(),
        file_hash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef".into(),
    };

    let created = create_sbom_handler(Path(snapshot_id), State(pool.clone()), Json(req)).await.unwrap();
    let created_resp = created.into_response();
    assert_eq!(created_resp.status(), StatusCode::CREATED);

    let latest = get_latest_sbom_handler(Path(snapshot_id), State(pool)).await.unwrap();
    let latest_resp = latest.into_response();
    assert_eq!(latest_resp.status(), StatusCode::OK);
}

#[tokio::test]
async fn test_post_sbom_validates_input() {
    let Some(pool) = test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };

    let snapshot_id = Uuid::new_v4();
    let bad_req = SbomCreateRequest {
        format: "spdx_json".into(),
        storage_key: "/not/object/uri.json".into(),
        file_hash: "not-a-sha".into(),
    };

    let res = create_sbom_handler(Path(snapshot_id), State(pool), Json(bad_req)).await;
    assert!(res.is_err());
}
