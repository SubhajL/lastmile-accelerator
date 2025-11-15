use sqlx::PgPool;
use uuid::Uuid;

use crate::models::Sbom;

pub async fn create_sbom(pool: &PgPool, sbom: &Sbom) -> Result<Sbom, sqlx::Error> {
    let start = std::time::Instant::now();
    let row = sqlx::query_as::<_, Sbom>(
        r#"
        INSERT INTO sboms (id, snapshot_id, format, storage_key, file_hash, generated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, snapshot_id, format, storage_key, file_hash, generated_at
        "#,
    )
    .bind(sbom.id)
    .bind(sbom.snapshot_id)
    .bind(&sbom.format)
    .bind(&sbom.storage_key)
    .bind(&sbom.file_hash)
    .bind(sbom.generated_at)
    .fetch_one(pool)
    .await?;

    crate::metrics::db()
        .query_duration_seconds
        .with_label_values(&["create_sbom"]) 
        .observe(start.elapsed().as_secs_f64());

    Ok(row)
}

pub async fn get_latest_sbom_by_snapshot(
    pool: &PgPool,
    snapshot_id: Uuid,
) -> Result<Option<Sbom>, sqlx::Error> {
    let start = std::time::Instant::now();
    let row = sqlx::query_as::<_, Sbom>(
        r#"
        SELECT id, snapshot_id, format, storage_key, file_hash, generated_at
        FROM sboms
        WHERE snapshot_id = $1
        ORDER BY generated_at DESC
        LIMIT 1
        "#,
    )
    .bind(snapshot_id)
    .fetch_optional(pool)
    .await?;

    crate::metrics::db()
        .query_duration_seconds
        .with_label_values(&["get_latest_sbom_by_snapshot"]) 
        .observe(start.elapsed().as_secs_f64());

    Ok(row)
}

pub async fn update_sbom_storage_key(
    pool: &PgPool,
    id: Uuid,
    storage_key: &str,
) -> Result<(), sqlx::Error> {
    let start = std::time::Instant::now();
    sqlx::query(
        r#"
        UPDATE sboms SET storage_key = $2 WHERE id = $1
        "#,
    )
    .bind(id)
    .bind(storage_key)
    .execute(pool)
    .await?;

    crate::metrics::db()
        .query_duration_seconds
        .with_label_values(&["update_sbom_storage_key"]) 
        .observe(start.elapsed().as_secs_f64());

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::db::migrate::test_pool;
    use crate::models::{Sbom, SbomFormat};

    #[tokio::test]
    async fn test_create_and_get_latest_sbom() {
        let Some(pool) = test_pool().await else {
            eprintln!("Skipping: TEST_DATABASE_URL not set");
            return;
        };

        let snapshot_id = Uuid::new_v4();
        let sbom1 = Sbom::new(
            snapshot_id,
            SbomFormat::SpdxJson,
            "s3://bucket/a.json",
            "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
            None,
        )
        .unwrap();
        let sbom2 = Sbom::new(
            snapshot_id,
            SbomFormat::CycloneDxJson,
            "s3://bucket/b.json",
            "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
            None,
        )
        .unwrap();

        let _ = create_sbom(&pool, &sbom1).await.unwrap();
        let _ = create_sbom(&pool, &sbom2).await.unwrap();

        let latest = get_latest_sbom_by_snapshot(&pool, snapshot_id)
            .await
            .unwrap()
            .unwrap();
        assert_eq!(latest.storage_key, sbom2.storage_key);
    }

    #[tokio::test]
    async fn test_update_sbom_storage_key() {
        let Some(pool) = test_pool().await else {
            eprintln!("Skipping: TEST_DATABASE_URL not set");
            return;
        };

        let snapshot_id = Uuid::new_v4();
        let sbom = Sbom::new(
            snapshot_id,
            SbomFormat::SpdxJson,
            "s3://bucket/a.json",
            "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
            None,
        )
        .unwrap();
        let created = create_sbom(&pool, &sbom).await.unwrap();
        update_sbom_storage_key(&pool, created.id, "s3://bucket/new.json")
            .await
            .unwrap();

        let latest = get_latest_sbom_by_snapshot(&pool, snapshot_id)
            .await
            .unwrap()
            .unwrap();
        assert_eq!(latest.storage_key, "s3://bucket/new.json");
    }
}
