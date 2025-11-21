use axum::extract::{Path, State};
use dep_governance_service::db::migrate::test_pool;
use dep_governance_service::handlers::api::vulns::get_dependency_vulns_handler;
use uuid::Uuid;

#[tokio::test]
async fn test_get_dependency_vulns_returns_joined_rows() {
    let Some(pool) = test_pool().await else {
        eprintln!("Skipping: TEST_DATABASE_URL not set");
        return;
    };

    // Seed dependency
    let dep = dep_governance_service::models::Dependency::new(
        Uuid::new_v4(),
        "openssl",
        "1.1.1",
        dep_governance_service::models::Ecosystem::Cargo,
        true,
        None,
    )
    .unwrap();
    let dep = dep_governance_service::db::dependencies::create_dependency(&pool, &dep)
        .await
        .unwrap();

    // Seed CVE and link
    let cve = dep_governance_service::models::Cve::new(
        "CVE-2024-9999",
        dep_governance_service::models::Severity::High,
        Some(8.0),
        Some("Example".into()),
        None,
        Some("OSV".into()),
    );
    let cve = dep_governance_service::db::vulnerabilities::upsert_cve(&pool, &cve)
        .await
        .unwrap();
    let _link = dep_governance_service::db::vulnerabilities::link_vulnerability_to_dependency(
        &pool,
        dep.id,
        cve.id,
        "open",
        Some(">=1.1.0 <1.1.2"),
        Some("1.1.2"),
    )
    .await
    .unwrap();

    let json = get_dependency_vulns_handler(Path(dep.id), State(pool))
        .await
        .unwrap();
    let list = json.0;

    assert!(!list.is_empty());
    assert_eq!(list[0].cve.cve_id, "CVE-2024-9999");
}
