use axum::{routing::post, Router};
use sqlx::PgPool;

pub mod cves;
pub mod deps;
pub mod sbom;
pub mod types;
pub mod vulns;

pub fn router(pool: PgPool) -> Router<PgPool> {
    Router::new()
        .route(
            "/snapshots/:snapshot_id/sbom",
            post(sbom::create_sbom_handler).get(sbom::get_latest_sbom_handler),
        )
        .route(
            "/snapshots/:snapshot_id/dependencies",
            axum::routing::get(deps::list_dependencies_handler),
        )
        .route(
            "/dependencies/:dependency_id/vulns",
            axum::routing::get(vulns::get_dependency_vulns_handler),
        )
        .route("/cves", axum::routing::post(cves::upsert_cve_handler))
        .route(
            "/dependencies/:dependency_id/vulns/link",
            axum::routing::post(cves::link_vuln_handler),
        )
        .with_state(pool)
}
