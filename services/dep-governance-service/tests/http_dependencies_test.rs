use axum::{extract::{Path, Query, State}};
use dep_governance_service::handlers::api::deps::{list_dependencies_handler, ListDepsQuery};
use dep_governance_service::db::migrate::test_pool;
use uuid::Uuid;

#[tokio::test]
async fn test_get_dependencies_lists_all_for_snapshot() {
    let Some(pool) = test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };

    // Seed
    let snapshot_id = Uuid::new_v4();
    let dep1 = dep_governance_service::models::Dependency::new(snapshot_id, "alpha", "1.0.0", dep_governance_service::models::Ecosystem::Npm, true, None).unwrap();
    let dep2 = dep_governance_service::models::Dependency::new(snapshot_id, "beta", "1.0.0", dep_governance_service::models::Ecosystem::Npm, false, None).unwrap();
    let _ = dep_governance_service::db::dependencies::create_dependency(&pool, &dep1).await.unwrap();
    let _ = dep_governance_service::db::dependencies::create_dependency(&pool, &dep2).await.unwrap();

    let json = list_dependencies_handler(Path(snapshot_id), Query(ListDepsQuery { direct: None }), State(pool)).await.unwrap();
    let deps = json.0;
    assert_eq!(deps.len(), 2);
    assert!(deps[0].is_direct);
}

#[tokio::test]
async fn test_get_dependencies_filters_direct_only() {
    let Some(pool) = test_pool().await else { eprintln!("Skipping: TEST_DATABASE_URL not set"); return; };

    // Seed
    let snapshot_id = Uuid::new_v4();
    let dep1 = dep_governance_service::models::Dependency::new(snapshot_id, "alpha", "1.0.0", dep_governance_service::models::Ecosystem::Npm, true, None).unwrap();
    let dep2 = dep_governance_service::models::Dependency::new(snapshot_id, "beta", "1.0.0", dep_governance_service::models::Ecosystem::Npm, false, None).unwrap();
    let _ = dep_governance_service::db::dependencies::create_dependency(&pool, &dep1).await.unwrap();
    let _ = dep_governance_service::db::dependencies::create_dependency(&pool, &dep2).await.unwrap();

    let json = list_dependencies_handler(Path(snapshot_id), Query(ListDepsQuery { direct: Some(true) }), State(pool)).await.unwrap();
    let deps = json.0;
    assert_eq!(deps.len(), 1);
    assert!(deps[0].is_direct);
}
