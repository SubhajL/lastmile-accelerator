use axum::{extract::Path, http::StatusCode};
use axum::response::IntoResponse;
use dep_governance_service::handlers::api::sbom::{create_sbom_handler, get_latest_sbom_handler, SbomCreateRequest};
use dep_governance_service::db::migrate::test_pool;
use axum::extract::State;
use dep_governance_service::web::strict_json::StrictJson;
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

    let created = create_sbom_handler(Path(snapshot_id), State(pool.clone()), dep_governance_service::web::strict_json::StrictJson(req)).await.unwrap();
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

    let res = create_sbom_handler(Path(snapshot_id), State(pool), dep_governance_service::web::strict_json::StrictJson(bad_req)).await;
    assert!(res.is_err());
}

#[tokio::test]
async fn test_post_sbom_duplicate_storage_key_returns_409() {
    let Some(pool) = test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };

    let snapshot_id = Uuid::new_v4();
    let req = SbomCreateRequest {
        format: "spdx_json".into(),
        storage_key: "s3://bucket/sboms/dupe.json".into(),
        file_hash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa".into(),
    };

    let created = create_sbom_handler(Path(snapshot_id), State(pool.clone()), StrictJson(req)).await.unwrap();
    assert_eq!(created.into_response().status(), StatusCode::CREATED);

    // duplicate storage_key should violate unique constraint -> 409
    let req2 = SbomCreateRequest {
        format: "spdx_json".into(),
        storage_key: "s3://bucket/sboms/dupe.json".into(),
        file_hash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb".into(),
    };
    let res = create_sbom_handler(Path(Uuid::new_v4()), State(pool.clone()), StrictJson(req2)).await;
    let err = match res {
        Ok(ok) => panic!("expected error, got {}", ok.into_response().status()),
        Err(e) => e,
    };
    let resp = axum::response::IntoResponse::into_response(err);
    assert_eq!(resp.status(), StatusCode::CONFLICT);
}
