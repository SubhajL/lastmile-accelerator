use sqlx::PgPool;
use uuid::Uuid;

use crate::models::Dependency;

pub async fn create_dependency(pool: &PgPool, dep: &Dependency) -> Result<Dependency, sqlx::Error> {
    let row = sqlx::query_as::<_, Dependency>(
        r#"
        INSERT INTO dependencies (id, snapshot_id, name, version, ecosystem, is_direct, parent_id, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, snapshot_id, name, version, ecosystem, is_direct, parent_id, created_at
        "#,
    )
    .bind(dep.id)
    .bind(dep.snapshot_id)
    .bind(&dep.name)
    .bind(&dep.version)
    .bind(&dep.ecosystem)
    .bind(dep.is_direct)
    .bind(dep.parent_id)
    .bind(dep.created_at)
    .fetch_one(pool)
    .await?;

    Ok(row)
}

pub async fn get_dependency_by_id(
    pool: &PgPool,
    id: Uuid,
) -> Result<Option<Dependency>, sqlx::Error> {
    let row = sqlx::query_as::<_, Dependency>(
        r#"
        SELECT id, snapshot_id, name, version, ecosystem, is_direct, parent_id, created_at
        FROM dependencies WHERE id = $1
        "#,
    )
    .bind(id)
    .fetch_optional(pool)
    .await?;

    Ok(row)
}

pub async fn list_dependencies_for_snapshot(
    pool: &PgPool,
    snapshot_id: Uuid,
) -> Result<Vec<Dependency>, sqlx::Error> {
    let rows = sqlx::query_as::<_, Dependency>(
        r#"
        SELECT id, snapshot_id, name, version, ecosystem, is_direct, parent_id, created_at
        FROM dependencies
        WHERE snapshot_id = $1
        ORDER BY is_direct DESC, name ASC
        "#,
    )
    .bind(snapshot_id)
    .fetch_all(pool)
    .await?;

    Ok(rows)
}

pub async fn list_dependencies_for_snapshot_filtered(
    pool: &PgPool,
    snapshot_id: Uuid,
    direct: Option<bool>,
) -> Result<Vec<Dependency>, sqlx::Error> {
    match direct {
        Some(d) => {
            let rows = sqlx::query_as::<_, Dependency>(
                r#"
                SELECT id, snapshot_id, name, version, ecosystem, is_direct, parent_id, created_at
                FROM dependencies
                WHERE snapshot_id = $1 AND is_direct = $2
                ORDER BY is_direct DESC, name ASC
                "#,
            )
            .bind(snapshot_id)
            .bind(d)
            .fetch_all(pool)
            .await?;
            Ok(rows)
        }
        None => list_dependencies_for_snapshot(pool, snapshot_id).await,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::db::migrate::test_pool;
    use crate::models::{Dependency, Ecosystem};

    #[tokio::test]
    async fn test_create_and_get_dependency() {
        let Some(pool) = test_pool().await else {
            eprintln!("Skipping: TEST_DATABASE_URL not set");
            return;
        };

        let dep = Dependency::new(
            Uuid::new_v4(),
            "serde",
            "1.0.0",
            Ecosystem::Cargo,
            true,
            None,
        )
        .unwrap();

        let created = create_dependency(&pool, &dep).await.unwrap();
        assert_eq!(created.name, dep.name);

        let fetched = get_dependency_by_id(&pool, created.id)
            .await
            .unwrap()
            .unwrap();
        assert_eq!(fetched.id, created.id);
    }

    #[tokio::test]
    async fn test_list_dependencies_for_snapshot() {
        let Some(pool) = test_pool().await else {
            eprintln!("Skipping: TEST_DATABASE_URL not set");
            return;
        };

        let snapshot_id = Uuid::new_v4();

        let dep1 = Dependency::new(snapshot_id, "A", "1.0.0", Ecosystem::Npm, true, None).unwrap();
        let dep2 = Dependency::new(snapshot_id, "B", "1.0.0", Ecosystem::Npm, false, None).unwrap();

        let _ = create_dependency(&pool, &dep1).await.unwrap();
        let _ = create_dependency(&pool, &dep2).await.unwrap();

        let list = list_dependencies_for_snapshot(&pool, snapshot_id)
            .await
            .unwrap();
        assert_eq!(list.len(), 2);
        assert!(list[0].is_direct);
    }
}
