use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use utoipa::ToSchema;
use uuid::Uuid;

#[derive(Deserialize, ToSchema)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct UpsertCveRequest {
    pub cve_id: String,
    pub severity: String,
    pub cvss_score: Option<f32>,
    pub description: Option<String>,
    pub published_at: Option<String>, // RFC3339
    pub source: Option<String>,
}

#[derive(Debug, Serialize, ToSchema)]
#[serde(rename_all = "camelCase")]
pub struct CveResponse {
    pub id: Uuid,
    pub cve_id: String,
    pub severity: String,
    pub cvss_score: Option<f32>,
    pub description: Option<String>,
    pub published_at: Option<DateTime<Utc>>,
    pub source: Option<String>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Deserialize, ToSchema)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct LinkVulnRequest {
    pub cve_id: String,
    pub status: String,
    pub affected_version_range: Option<String>,
    pub fixed_version: Option<String>,
}

#[derive(Debug, Serialize, ToSchema)]
#[serde(rename_all = "camelCase")]
pub struct LinkResponse {
    pub id: Uuid,
    pub dependency_id: Uuid,
    pub cve_id: Uuid,
    pub status: String,
    pub affected_version_range: Option<String>,
    pub fixed_version: Option<String>,
}
