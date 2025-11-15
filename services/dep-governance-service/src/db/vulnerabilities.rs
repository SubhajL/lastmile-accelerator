use sqlx::PgPool;
use uuid::Uuid;

use crate::models::{Cve, DependencyVulnerability};

pub async fn get_cve_by_cve_id(pool: &PgPool, cve_id: &str) -> Result<Option<Cve>, sqlx::Error> {
    let row = sqlx::query_as::<_, Cve>(
        r#"
        SELECT id, cve_id, severity, description, published_at, source, cvss_score, updated_at
        FROM cves WHERE cve_id = $1
        "#,
    )
    .bind(cve_id)
    .fetch_optional(pool)
    .await?;
    Ok(row)
}

pub async fn upsert_cve(pool: &PgPool, cve: &Cve) -> Result<Cve, sqlx::Error> {
    let start = std::time::Instant::now();
    let row = sqlx::query_as::<_, Cve>(
        r#"
        INSERT INTO cves (id, cve_id, severity, description, published_at, source, cvss_score, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (cve_id) DO UPDATE SET
            severity = EXCLUDED.severity,
            description = EXCLUDED.description,
            published_at = EXCLUDED.published_at,
            source = EXCLUDED.source,
            cvss_score = EXCLUDED.cvss_score,
            updated_at = NOW()
        RETURNING id, cve_id, severity, description, published_at, source, cvss_score, updated_at
        "#,
    )
    .bind(cve.id)
    .bind(&cve.cve_id)
    .bind(&cve.severity)
    .bind(&cve.description)
.bind(cve.published_at)
    .bind(&cve.source)
    .bind(cve.cvss_score)
    .bind(cve.updated_at)
    .fetch_one(pool)
    .await?;

    crate::metrics::db()
        .query_duration_seconds
        .with_label_values(&["upsert_cve"]) 
        .observe(start.elapsed().as_secs_f64());

    Ok(row)
}

pub async fn link_vulnerability_to_dependency(
    pool: &PgPool,
    dep_id: Uuid,
    cve_id: Uuid,
    status: &str,
    affected_version_range: Option<&str>,
    fixed_version: Option<&str>,
) -> Result<DependencyVulnerability, sqlx::Error> {
    let start = std::time::Instant::now();
    let row = sqlx::query_as::<_, DependencyVulnerability>(
        r#"
        INSERT INTO dependency_vulnerabilities (id, dependency_id, cve_id, affected_version_range, fixed_version, discovered_at, status)
        VALUES ($1, $2, $3, $4, $5, NOW(), $6)
        RETURNING id, dependency_id, cve_id, affected_version_range, fixed_version, discovered_at, status
        "#,
    )
    .bind(Uuid::new_v4())
    .bind(dep_id)
    .bind(cve_id)
    .bind(affected_version_range)
    .bind(fixed_version)
    .bind(status)
    .fetch_one(pool)
    .await?;

    crate::metrics::db()
        .query_duration_seconds
        .with_label_values(&["link_vulnerability_to_dependency"]) 
        .observe(start.elapsed().as_secs_f64());

    Ok(row)
}

pub async fn get_vulnerabilities_for_dependency(
    pool: &PgPool,
    dep_id: Uuid,
) -> Result<Vec<(Cve, DependencyVulnerability)>, sqlx::Error> {
    let start = std::time::Instant::now();
    use sqlx::Row;
    let rows = sqlx::query(
        r#"
        SELECT
            c.id as c_id, c.cve_id as c_cve_id, c.severity as c_severity, c.description as c_description,
            c.published_at as c_published_at, c.source as c_source, c.cvss_score as c_cvss_score, c.updated_at as c_updated_at,
            dv.id as dv_id, dv.dependency_id as dv_dependency_id, dv.cve_id as dv_cve_id,
            dv.affected_version_range as dv_affected_version_range, dv.fixed_version as dv_fixed_version,
            dv.discovered_at as dv_discovered_at, dv.status as dv_status
        FROM dependency_vulnerabilities dv
        JOIN cves c ON dv.cve_id = c.id
        WHERE dv.dependency_id = $1
        ORDER BY c.severity DESC
        "#,
    )
    .bind(dep_id)
    .fetch_all(pool)
    .await?;

    crate::metrics::db()
        .query_duration_seconds
        .with_label_values(&["get_vulnerabilities_for_dependency"]) 
        .observe(start.elapsed().as_secs_f64());

    let mut result = Vec::with_capacity(rows.len());
    for r in rows {
        let cve = Cve {
            id: r.get("c_id"),
            cve_id: r.get::<String, _>("c_cve_id"),
            severity: r.get::<String, _>("c_severity"),
            description: r.try_get("c_description").ok(),
            published_at: r.try_get::<chrono::DateTime<chrono::Utc>, _>("c_published_at").ok(),
            source: r.try_get("c_source").ok(),
            cvss_score: r.try_get::<f32, _>("c_cvss_score").ok(),
            updated_at: r.get::<chrono::DateTime<chrono::Utc>, _>("c_updated_at"),
        };
        let dv = DependencyVulnerability {
            id: r.get("dv_id"),
            dependency_id: r.get("dv_dependency_id"),
            cve_id: r.get("dv_cve_id"),
            affected_version_range: r.try_get("dv_affected_version_range").ok(),
            fixed_version: r.try_get("dv_fixed_version").ok(),
            discovered_at: r.get::<chrono::DateTime<chrono::Utc>, _>("dv_discovered_at"),
            status: r.get::<String, _>("dv_status"),
        };
        result.push((cve, dv));
    }
    Ok(result)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::db::{dependencies::create_dependency, migrate::test_pool};
    use crate::models::{Cve, Dependency, Ecosystem, Severity, VulnerabilityStatus};

    #[tokio::test]
    async fn test_upsert_and_link_vulnerability() {
        let Some(pool) = test_pool().await else {
            eprintln!("Skipping: TEST_DATABASE_URL not set");
            return;
        };

        // Create dependency
        let dep = Dependency::new(
            Uuid::new_v4(),
            "openssl",
            "1.1.1",
            Ecosystem::Cargo,
            true,
            None,
        )
        .unwrap();
        let dep = create_dependency(&pool, &dep).await.unwrap();

        // Upsert CVE
        let cve = Cve::new(
            "CVE-2024-1234",
            Severity::High,
            Some(7.5),
            Some("Example vulnerability".into()),
            None,
            Some("OSV".into()),
        );
        let cve = upsert_cve(&pool, &cve).await.unwrap();

        // Link
        let dv = link_vulnerability_to_dependency(
            &pool,
            dep.id,
            cve.id,
            &VulnerabilityStatus::Open.to_string(),
            Some(">=1.1.0 <1.1.2"),
            Some("1.1.2"),
        )
        .await
        .unwrap();
        assert_eq!(dv.status, "open");

        // Query
        let list = get_vulnerabilities_for_dependency(&pool, dep.id)
            .await
            .unwrap();
        assert_eq!(list.len(), 1);
        assert_eq!((list[0].0).cve_id, "CVE-2024-1234");
    }
}
