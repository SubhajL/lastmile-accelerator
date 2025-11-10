use sqlx::postgres::{PgPool, PgPoolOptions};
use std::time::Duration;

pub async fn create_db_pool(database_url: &str) -> Result<PgPool, sqlx::Error> {
    tracing::info!("Creating database connection pool");

    let pool = PgPoolOptions::new()
        .min_connections(5)
        .max_connections(20)
        .acquire_timeout(Duration::from_secs(30))
        .idle_timeout(Duration::from_secs(600))
        .max_lifetime(Duration::from_secs(1800))
        .connect(database_url)
        .await?;

    // Test connectivity
    health_check(&pool).await?;

    tracing::info!("Database connection pool created successfully");
    Ok(pool)
}

pub async fn health_check(pool: &PgPool) -> Result<(), sqlx::Error> {
    sqlx::query("SELECT 1").execute(pool).await?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_db_pool_connects_successfully() {
        // Skip if no test database available
        if std::env::var("TEST_DATABASE_URL").is_err() {
            eprintln!("Skipping test: TEST_DATABASE_URL not set");
            return;
        }

        let database_url = std::env::var("TEST_DATABASE_URL").unwrap();
        let result = create_db_pool(&database_url).await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn test_db_pool_health_check_succeeds() {
        if std::env::var("TEST_DATABASE_URL").is_err() {
            eprintln!("Skipping test: TEST_DATABASE_URL not set");
            return;
        }

        let database_url = std::env::var("TEST_DATABASE_URL").unwrap();
        let pool = create_db_pool(&database_url).await.unwrap();
        let result = health_check(&pool).await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn test_db_pool_handles_connection_errors() {
        let invalid_url = "postgres://invalid:invalid@nonexistent:5432/test";
        let result = create_db_pool(invalid_url).await;
        assert!(result.is_err());
    }
}
