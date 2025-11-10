use sqlx::PgPool;

pub async fn run_migrations(pool: &PgPool) -> Result<(), sqlx::migrate::MigrateError> {
    // Embed migrations from ./migrations directory
    sqlx::migrate!("./migrations").run(pool).await
}

pub async fn test_pool() -> Option<PgPool> {
    use sqlx::postgres::PgPoolOptions;

    let url = match std::env::var("TEST_DATABASE_URL") {
        Ok(v) => v,
        Err(_) => return None,
    };
    let pool = PgPoolOptions::new()
        .max_connections(5)
        .connect(&url)
        .await
        .ok()?;
    run_migrations(&pool).await.ok()?;
    Some(pool)
}
