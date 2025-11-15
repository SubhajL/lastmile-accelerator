use axum::{extract::{Path, State}, Json};
use serde::Serialize;
use utoipa::ToSchema;
use sqlx::PgPool;
use uuid::Uuid;

use crate::{db::vulnerabilities as vulns_repo, error::AppError};

#[derive(Serialize, ToSchema)]
#[serde(rename_all = "camelCase")]
pub struct CveSummary {
    pub cve_id: String,
    pub severity: String,
    pub cvss_score: Option<f32>,
}

#[cfg(test)]
mod tests {
    use super::*;
    #[tokio::test]
    async fn test_get_vulns() {
        let Some(pool) = crate::db::migrate::test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };

        // seed
        let dep = crate::models::Dependency::new(Uuid::new_v4(), "x", "1.0.0", crate::models::Ecosystem::Cargo, true, None).unwrap();
        let dep = crate::db::dependencies::create_dependency(&pool, &dep).await.unwrap();
        let cve = crate::models::Cve::new("CVE-2025-0001", crate::models::Severity::High, Some(7.5), None, None, Some("OSV".into()));
        let cve = crate::db::vulnerabilities::upsert_cve(&pool, &cve).await.unwrap();
        let _ = crate::db::vulnerabilities::link_vulnerability_to_dependency(&pool, dep.id, cve.id, "open", None, None).await.unwrap();

        let json = super::get_dependency_vulns_handler(
            axum::extract::Path(dep.id),
            axum::extract::State(pool),
        )
        .await
        .unwrap();
        assert!(!json.0.is_empty());
    }
}

#[derive(Serialize, ToSchema)]
#[serde(rename_all = "camelCase")]
pub struct DependencyVulnResponse {
    pub cve: CveSummary,
    pub status: String,
    pub affected_version_range: Option<String>,
    pub fixed_version: Option<String>,
}

#[utoipa::path(
    get,
    path = "/v1/dependencies/{dependency_id}/vulns",
    params(("dependency_id" = uuid::Uuid, Path, description = "Dependency ID")),
    responses(
        (status = 200, description = "Dependency vulnerabilities", body = [DependencyVulnResponse]),
        (status = 401, description = "Unauthorized"),
    ),
    security(("bearerAuth" = []))
)]
pub async fn get_dependency_vulns_handler(
    Path(dependency_id): Path<Uuid>,
    State(pool): State<PgPool>,
) -> Result<Json<Vec<DependencyVulnResponse>>, AppError> {
    let rows = vulns_repo::get_vulnerabilities_for_dependency(&pool, dependency_id)
        .await
        .map_err(AppError::Database)?;

    let out = rows
        .into_iter()
        .map(|(c, dv)| DependencyVulnResponse {
            cve: CveSummary {
                cve_id: c.cve_id,
                severity: c.severity,
                cvss_score: c.cvss_score,
            },
            status: dv.status,
            affected_version_range: dv.affected_version_range,
            fixed_version: dv.fixed_version,
        })
        .collect();

    Ok(Json(out))
}
