use axum::{extract::{Path, State}, http::StatusCode, Json};
use chrono::{DateTime, Utc};
use sqlx::PgPool;
use uuid::Uuid;

use crate::{db::vulnerabilities as repo, error::AppError, models::{Cve, Severity, VulnerabilityStatus}};
use super::types::{UpsertCveRequest, CveResponse, LinkVulnRequest, LinkResponse};

fn parse_severity(s: &str) -> Result<Severity, AppError> {
    s.parse::<Severity>().map_err(|_| AppError::BadRequest("invalid severity".into()))
}

fn parse_status(s: &str) -> Result<VulnerabilityStatus, AppError> {
    match s.to_lowercase().as_str() {
        "open" => Ok(VulnerabilityStatus::Open),
        "acknowledged" => Ok(VulnerabilityStatus::Acknowledged),
        "fixed" => Ok(VulnerabilityStatus::Fixed),
        "ignored" => Ok(VulnerabilityStatus::Ignored),
        "false_positive" => Ok(VulnerabilityStatus::FalsePositive),
        _ => Err(AppError::BadRequest("invalid status".into())),
    }
}

fn parse_published_at(s: &Option<String>) -> Result<Option<DateTime<Utc>>, AppError> {
    if let Some(v) = s {
        let dt = DateTime::parse_from_rfc3339(v)
            .map_err(|_| AppError::BadRequest("invalid publishedAt".into()))?;
        Ok(Some(dt.with_timezone(&Utc)))
    } else {
        Ok(None)
    }
}

pub async fn upsert_cve_handler(
    State(pool): State<PgPool>,
    Json(req): Json<UpsertCveRequest>,
) -> Result<Json<CveResponse>, AppError> {
    let severity = parse_severity(&req.severity)?;
    let published_at = parse_published_at(&req.published_at)?;

    let cve = Cve::new(
        req.cve_id,
        severity,
        req.cvss_score,
        req.description,
        published_at,
        req.source,
    );

    let saved = repo::upsert_cve(&pool, &cve).await.map_err(AppError::Database)?;
    let resp = CveResponse {
        id: saved.id,
        cve_id: saved.cve_id,
        severity: saved.severity,
        cvss_score: saved.cvss_score,
        description: saved.description,
        published_at: saved.published_at,
        source: saved.source,
        updated_at: saved.updated_at,
    };
    Ok(Json(resp))
}

pub async fn link_vuln_handler(
    Path(dependency_id): Path<Uuid>,
    State(pool): State<PgPool>,
    Json(req): Json<LinkVulnRequest>,
) -> Result<(StatusCode, Json<LinkResponse>), AppError> {
    let status = parse_status(&req.status)?;
    let maybe = repo::get_cve_by_cve_id(&pool, &req.cve_id).await.map_err(AppError::Database)?;
    let cve = match maybe { Some(c) => c, None => return Err(AppError::NotFound("CVE not found".into())) };

    match repo::link_vulnerability_to_dependency(
        &pool,
        dependency_id,
        cve.id,
        &status.to_string(),
        req.affected_version_range.as_deref(),
        req.fixed_version.as_deref(),
    ).await {
        Ok(link) => Ok((
            StatusCode::CREATED,
            Json(LinkResponse {
                id: link.id,
                dependency_id: link.dependency_id,
                cve_id: link.cve_id,
                status: link.status,
                affected_version_range: link.affected_version_range,
                fixed_version: link.fixed_version,
            })
        )),
        Err(sqlx::Error::Database(db_err)) if db_err.code().as_deref() == Some("23505") => {
            Err(AppError::Conflict("dependency vulnerability link already exists".into()))
        }
        Err(e) => Err(AppError::Database(e)),
    }
}
