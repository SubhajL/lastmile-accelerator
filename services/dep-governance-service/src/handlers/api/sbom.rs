use axum::{extract::{Path, State}, response::IntoResponse, Json};
use serde::Deserialize;
use utoipa::ToSchema;
use sqlx::PgPool;
use uuid::Uuid;

use crate::{db::sboms as sboms_repo, error::AppError, errors::db::map_db_error, models::{Sbom, SbomFormat}};

#[derive(Deserialize, ToSchema)]
#[serde(deny_unknown_fields)]
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
            crate::web::strict_json::StrictJson(req),
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

#[utoipa::path(
    post,
    path = "/v1/snapshots/{snapshot_id}/sbom",
    request_body = SbomCreateRequest,
    params(
        ("snapshot_id" = uuid::Uuid, Path, description = "Snapshot ID"),
    ),
    responses(
        (status = 201, description = "SBOM created", body = crate::models::Sbom),
        (status = 400, description = "Bad request", body = crate::openapi::ApiError),
        (status = 401, description = "Unauthorized"),
        (status = 409, description = "Conflict", body = crate::openapi::ApiError),
    ),
    security(("bearerAuth" = []))
)]
pub async fn create_sbom_handler(
    Path(snapshot_id): Path<Uuid>,
    State(pool): State<PgPool>,
    crate::web::strict_json::StrictJson(req): crate::web::strict_json::StrictJson<SbomCreateRequest>,
) -> Result<impl IntoResponse, AppError> {
    let format = SbomFormat::try_from(req.format.as_str()).map_err(AppError::BadRequest)?;

    let sbom = Sbom::new(snapshot_id, format, req.storage_key, req.file_hash, None)
        .map_err(AppError::BadRequest)?;

    let created = sboms_repo::create_sbom(&pool, &sbom).await
        .map_err(map_db_error)?;

    Ok((axum::http::StatusCode::CREATED, Json(created)))
}

#[utoipa::path(
    get,
    path = "/v1/snapshots/{snapshot_id}/sbom",
    params(
        ("snapshot_id" = uuid::Uuid, Path, description = "Snapshot ID"),
    ),
    responses(
        (status = 200, description = "Latest SBOM", body = crate::models::Sbom),
        (status = 401, description = "Unauthorized"),
        (status = 404, description = "Not found"),
    ),
    security(("bearerAuth" = []))
)]
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
