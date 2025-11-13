use axum::{middleware, Router, Extension};
use dep_governance_service::{
    config::AppConfig,
    db::{create_db_pool, migrate::run_migrations},
    events::NatsPublisher,
    handlers::{self, healthz, metrics, readyz},
    middleware::{init_telemetry, jwt_auth_middleware, shutdown_telemetry, trace_layer, AuthContext, AuthConfig},
};
use dep_governance_service::middleware::auth::{CachingJwksProvider, JwksProvider, NoopJwksProvider};
use std::sync::Arc;
use std::net::SocketAddr;
use tokio::signal;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Load configuration
    let config = AppConfig::from_env()?;

    // Initialize telemetry
    init_telemetry(&config.service_name, config.otel_endpoint.as_deref())?;

    tracing::info!(
        "Starting {} in {} environment",
        config.service_name,
        config.env
    );

    // Create database pool
    let pool = create_db_pool(&config.database_url).await?;

    // Run migrations at startup
    run_migrations(&pool).await?;

    // Connect to NATS (optional)
    let nats = if let Some(ref nats_url) = config.nats_url {
        match NatsPublisher::connect(nats_url).await {
            Ok(publisher) => Some(publisher),
            Err(e) => {
                tracing::warn!("Failed to connect to NATS: {}. Continuing without NATS", e);
                None
            }
        }
    } else {
        tracing::info!("NATS URL not configured, skipping NATS connection");
        None
    };

    // Build router
    let provider: Arc<dyn JwksProvider> = match &config.jwt_public_key_url {
        Some(url) => Arc::new(CachingJwksProvider::new(url.clone(), std::time::Duration::from_secs(300))),
        None => Arc::new(NoopJwksProvider),
    };
    let auth_ctx = AuthContext::new(
        AuthConfig {
            jwks_url: config.jwt_public_key_url.clone(),
            issuer: config.jwt_issuer.clone(),
            audience: config.jwt_audience.clone(),
        },
        provider,
    );

    let app = build_router(pool, nats, auth_ctx);

    // Bind to address
    let addr = SocketAddr::from(([0, 0, 0, 0], config.service_port));
    tracing::info!("Listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await?;

    // Start server with graceful shutdown
    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await?;

    // Cleanup
    shutdown_telemetry();
    tracing::info!("Service shut down gracefully");

    Ok(())
}

fn build_router(pool: sqlx::PgPool, _nats: Option<NatsPublisher>, auth_ctx: AuthContext) -> Router {
    Router::new()
        // Health endpoints (no auth)
        .route("/healthz", axum::routing::get(healthz))
        .route("/readyz", axum::routing::get(readyz))
        .route("/metrics", axum::routing::get(metrics))
        // v1 API
        .nest("/v1", handlers::api::router(pool.clone()))
        .layer(Extension(auth_ctx))
        .layer(middleware::from_fn(jwt_auth_middleware))
        .layer(trace_layer())
        .with_state(pool)
}

async fn shutdown_signal() {
    let ctrl_c = async {
        signal::ctrl_c()
            .await
            .expect("failed to install Ctrl+C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("failed to install signal handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {
            tracing::info!("Received Ctrl+C, starting graceful shutdown");
        },
        _ = terminate => {
            tracing::info!("Received terminate signal, starting graceful shutdown");
        },
    }
}
