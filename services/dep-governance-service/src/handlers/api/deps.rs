use axum::{extract::{Path, Query, State}, Json};
use serde::Deserialize;
use sqlx::PgPool;
use uuid::Uuid;

use crate::{db::dependencies as deps_repo, error::AppError, models::Dependency};
use utoipa::ToSchema;

#[derive(Deserialize)]
pub struct ListDepsQuery {
    pub direct: Option<bool>,
}

#[cfg(test)]
mod tests {
    use super::*;
    #[tokio::test]
    async fn test_list_dependencies() {
        let Some(pool) = crate::db::migrate::test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };

        let snapshot_id = Uuid::new_v4();
        // seed
        let d1 = crate::models::Dependency::new(snapshot_id, "a", "1.0.0", crate::models::Ecosystem::Npm, true, None).unwrap();
        let d2 = crate::models::Dependency::new(snapshot_id, "b", "1.0.0", crate::models::Ecosystem::Npm, false, None).unwrap();
        let _ = crate::db::dependencies::create_dependency(&pool, &d1).await.unwrap();
        let _ = crate::db::dependencies::create_dependency(&pool, &d2).await.unwrap();

        let json = super::list_dependencies_handler(
            axum::extract::Path(snapshot_id),
            axum::extract::Query(super::ListDepsQuery { direct: Some(true) }),
            axum::extract::State(pool),
        )
        .await
        .unwrap();
        assert_eq!(json.0.len(), 1);
    }
}

#[derive(serde::Serialize, ToSchema)]
#[serde(rename_all = "camelCase")]
pub struct DependencyResponse {
    pub id: Uuid,
    pub name: String,
    pub version: String,
    pub ecosystem: String,
    pub is_direct: bool,
    pub parent_id: Option<Uuid>,
}

impl From<Dependency> for DependencyResponse {
    fn from(d: Dependency) -> Self {
        Self {
            id: d.id,
            name: d.name,
            version: d.version,
            ecosystem: d.ecosystem,
            is_direct: d.is_direct,
            parent_id: d.parent_id,
        }
    }
}

#[utoipa::path(
    get,
    path = "/v1/snapshots/{snapshot_id}/dependencies",
    params(
        ("snapshot_id" = uuid::Uuid, Path, description = "Snapshot ID"),
        ("direct" = Option<bool>, Query, description = "Filter direct dependencies only"),
    ),
    responses(
        (status = 200, description = "Dependencies", body = [DependencyResponse]),
        (status = 401, description = "Unauthorized"),
    ),
    security(("bearerAuth" = []))
)]
pub async fn list_dependencies_handler(
    Path(snapshot_id): Path<Uuid>,
    Query(q): Query<ListDepsQuery>,
    State(pool): State<PgPool>,
) -> Result<Json<Vec<DependencyResponse>>, AppError> {
    let rows = deps_repo::list_dependencies_for_snapshot_filtered(&pool, snapshot_id, q.direct)
        .await
        .map_err(AppError::Database)?;

    let out: Vec<DependencyResponse> = rows.into_iter().map(DependencyResponse::from).collect();
    Ok(Json(out))
}
