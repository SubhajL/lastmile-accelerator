use axum::extract::{State, Path};
use dep_governance_service::web::strict_json::StrictJson;
use uuid::Uuid;

use dep_governance_service::db::migrate::test_pool;
use dep_governance_service::handlers::api::types::{UpsertCveRequest, LinkVulnRequest};
use dep_governance_service::handlers::api::cves::{upsert_cve_handler, link_vuln_handler};

#[tokio::test]
async fn test_upsert_cve_create_and_update() {
    let Some(pool) = test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };

    // Create
    let req = UpsertCveRequest {
        cve_id: "CVE-2025-1111".into(),
        severity: "high".into(),
        cvss_score: Some(7.5),
        description: Some("desc".into()),
        published_at: None,
        source: Some("OSV".into()),
    }; 
    let res = upsert_cve_handler(State(pool.clone()), StrictJson(req)).await.unwrap();
    assert_eq!(res.0.cve_id, "CVE-2025-1111");
    assert_eq!(res.0.severity, "high");

    // Update severity to critical
    let req2 = UpsertCveRequest {
        cve_id: "CVE-2025-1111".into(),
        severity: "critical".into(),
        cvss_score: Some(9.8),
        description: Some("changed".into()),
        published_at: None,
        source: Some("OSV".into()),
    };
    let res2 = upsert_cve_handler(State(pool.clone()), StrictJson(req2)).await.unwrap();
    assert_eq!(res2.0.severity, "critical");
}

#[tokio::test]
async fn test_link_vuln_happy_and_conflict() {
    let Some(pool) = test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };

    // Seed dependency
    let dep = dep_governance_service::models::Dependency::new(
        Uuid::new_v4(), "pkg", "1.0.0", dep_governance_service::models::Ecosystem::Cargo, true, None
    ).unwrap();
    let dep = dep_governance_service::db::dependencies::create_dependency(&pool, &dep).await.unwrap();

    // Upsert CVE
    let req = UpsertCveRequest {
        cve_id: "CVE-2025-2222".into(),
        severity: "low".into(),
        cvss_score: Some(2.0),
        description: None,
        published_at: None,
        source: Some("OSV".into()),
    };
    let _ = upsert_cve_handler(State(pool.clone()), StrictJson(req)).await.unwrap();

    // Link
    let link_req = LinkVulnRequest { cve_id: "CVE-2025-2222".into(), status: "open".into(), affected_version_range: None, fixed_version: None };
    let (status, _) = link_vuln_handler(Path(dep.id), State(pool.clone()), StrictJson(link_req)).await.unwrap();
    assert_eq!(status, axum::http::StatusCode::CREATED);

    // Conflict
    let link_req2 = LinkVulnRequest { cve_id: "CVE-2025-2222".into(), status: "open".into(), affected_version_range: None, fixed_version: None };
    let err = link_vuln_handler(Path(dep.id), State(pool.clone()), StrictJson(link_req2)).await.unwrap_err();
    assert!(format!("{}", err).to_lowercase().contains("conflict"));
}

#[tokio::test]
async fn test_link_vuln_404_when_cve_missing() {
    let Some(pool) = test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };

    let dep = dep_governance_service::models::Dependency::new(
        Uuid::new_v4(), "pkg2", "1.0.0", dep_governance_service::models::Ecosystem::Cargo, true, None
    ).unwrap();
    let dep = dep_governance_service::db::dependencies::create_dependency(&pool, &dep).await.unwrap();

    let link_req = LinkVulnRequest { cve_id: "NOT-EXIST".into(), status: "open".into(), affected_version_range: None, fixed_version: None };
    let err = link_vuln_handler(Path(dep.id), State(pool.clone()), StrictJson(link_req)).await.unwrap_err();
    assert!(format!("{}", err).to_lowercase().contains("not found"));
}

#[tokio::test]
async fn test_link_vuln_duplicate_returns_409_status() {
    let Some(pool) = test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };

    let dep = dep_governance_service::models::Dependency::new(
        Uuid::new_v4(), "pkg3", "1.0.0", dep_governance_service::models::Ecosystem::Cargo, true, None
    ).unwrap();
    let dep = dep_governance_service::db::dependencies::create_dependency(&pool, &dep).await.unwrap();

    let req = UpsertCveRequest {
        cve_id: "CVE-2025-3333".into(),
        severity: "low".into(),
        cvss_score: Some(2.0),
        description: None,
        published_at: None,
        source: Some("OSV".into()),
    };
    let _ = upsert_cve_handler(State(pool.clone()), StrictJson(req)).await.unwrap();

    let link_req = LinkVulnRequest { cve_id: "CVE-2025-3333".into(), status: "open".into(), affected_version_range: None, fixed_version: None };
    let _ = link_vuln_handler(Path(dep.id), State(pool.clone()), StrictJson(link_req)).await.unwrap();

    let link_req2 = LinkVulnRequest { cve_id: "CVE-2025-3333".into(), status: "open".into(), affected_version_range: None, fixed_version: None };
    let err = link_vuln_handler(Path(dep.id), State(pool.clone()), StrictJson(link_req2)).await.unwrap_err();
    let resp = axum::response::IntoResponse::into_response(err);
    assert_eq!(resp.status(), axum::http::StatusCode::CONFLICT);
}

#[tokio::test]
async fn test_link_vuln_missing_dependency_returns_400() {
    let Some(pool) = test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };

    // Create CVE so only dependency FK fails
    let req = UpsertCveRequest {
        cve_id: "CVE-2025-4444".into(),
        severity: "medium".into(),
        cvss_score: Some(5.0),
        description: None,
        published_at: None,
        source: Some("OSV".into()),
    };
    let _ = upsert_cve_handler(State(pool.clone()), StrictJson(req)).await.unwrap();

    let missing_dep_id = Uuid::new_v4();
    let link_req = LinkVulnRequest { cve_id: "CVE-2025-4444".into(), status: "open".into(), affected_version_range: None, fixed_version: None };
    let err = link_vuln_handler(Path(missing_dep_id), State(pool.clone()), StrictJson(link_req)).await.unwrap_err();
    let resp = axum::response::IntoResponse::into_response(err);
    assert_eq!(resp.status(), axum::http::StatusCode::BAD_REQUEST);
}
