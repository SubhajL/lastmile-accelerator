use axum::{extract::{Path, State}, response::IntoResponse, Json};
use serde::Deserialize;
use sqlx::PgPool;
use uuid::Uuid;

use crate::{db::sboms as sboms_repo, error::AppError, models::{Sbom, SbomFormat}};

#[derive(Deserialize)]
pub struct SbomCreateRequest {
    pub format: String,
    #[serde(rename = "storageKey")]
    pub storage_key: String,
    #[serde(rename = "fileHash")]
    pub file_hash: String,
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::response::IntoResponse;

    #[tokio::test]
    async fn test_post_and_get_sbom() {
        let Some(pool) = crate::db::migrate::test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };
        let snapshot_id = Uuid::new_v4();
        let req = super::SbomCreateRequest {
            format: "spdx_json".into(),
            storage_key: "s3://bucket/sboms/123.json".into(),
            file_hash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef".into(),
        };
        let created = super::create_sbom_handler(
            axum::extract::Path(snapshot_id),
            axum::extract::State(pool.clone()),
            axum::Json(req),
        )
        .await
        .unwrap();
        assert_eq!(created.into_response().status(), axum::http::StatusCode::CREATED);

        let latest = super::get_latest_sbom_handler(
            axum::extract::Path(snapshot_id),
            axum::extract::State(pool),
        )
        .await
        .unwrap();
        assert_eq!(latest.into_response().status(), axum::http::StatusCode::OK);
    }
}

pub async fn create_sbom_handler(
    Path(snapshot_id): Path<Uuid>,
    State(pool): State<PgPool>,
    Json(req): Json<SbomCreateRequest>,
) -> Result<impl IntoResponse, AppError> {
    let format = SbomFormat::try_from(req.format.as_str()).map_err(AppError::BadRequest)?;

    let sbom = Sbom::new(snapshot_id, format, req.storage_key, req.file_hash, None)
        .map_err(AppError::BadRequest)?;

    let created = sboms_repo::create_sbom(&pool, &sbom).await
        .map_err(AppError::Database)?;

    Ok((axum::http::StatusCode::CREATED, Json(created)))
}

pub async fn get_latest_sbom_handler(
    Path(snapshot_id): Path<Uuid>,
    State(pool): State<PgPool>,
) -> Result<impl IntoResponse, AppError> {
    let maybe = sboms_repo::get_latest_sbom_by_snapshot(&pool, snapshot_id).await
        .map_err(AppError::Database)?;

    match maybe {
        Some(sbom) => Ok((axum::http::StatusCode::OK, Json(sbom)).into_response()),
        None => Ok((axum::http::StatusCode::NOT_FOUND).into_response()),
    }
}
