use crate::error::{AppError, Result};
use std::env;

#[derive(Debug, Clone)]
pub struct AppConfig {
    pub env: String,
    pub service_name: String,
    pub service_port: u16,
    pub database_url: String,
    pub redis_url: Option<String>,
    pub nats_url: Option<String>,
    pub otel_endpoint: Option<String>,
    pub jwt_public_key_url: Option<String>,
    pub vault_addr: Option<String>,
    pub vault_role_id: Option<String>,
    pub vault_secret_id: Option<String>,
    pub log_level: String,
}

impl AppConfig {
    pub fn from_env() -> Result<Self> {
        let env = env::var("ENV").unwrap_or_else(|_| "dev".to_string());
        let service_name =
            env::var("SERVICE_NAME").unwrap_or_else(|_| "dep-governance-service".to_string());
        let service_port = env::var("SERVICE_PORT")
            .unwrap_or_else(|_| "7106".to_string())
            .parse::<u16>()
            .map_err(|e| AppError::Config(format!("Invalid SERVICE_PORT: {}", e)))?;

        let database_url = env::var("DATABASE_URL")
            .map_err(|_| AppError::Config("DATABASE_URL is required".to_string()))?;

        let redis_url = env::var("REDIS_URL").ok();
        let nats_url = env::var("NATS_URL").ok();
        let otel_endpoint = env::var("OTEL_EXPORTER_OTLP_ENDPOINT").ok();
        let jwt_public_key_url = env::var("JWT_PUBLIC_KEY_URL").ok();
        let vault_addr = env::var("VAULT_ADDR").ok();
        let vault_role_id = env::var("VAULT_ROLE_ID").ok();
        let vault_secret_id = env::var("VAULT_SECRET_ID").ok();
        let log_level = env::var("LOG_LEVEL").unwrap_or_else(|_| "info".to_string());

        let config = Self {
            env,
            service_name,
            service_port,
            database_url,
            redis_url,
            nats_url,
            otel_endpoint,
            jwt_public_key_url,
            vault_addr,
            vault_role_id,
            vault_secret_id,
            log_level,
        };

        config.validate()?;
        Ok(config)
    }

    pub fn validate(&self) -> Result<()> {
        if self.service_port == 0 {
            return Err(AppError::Config(
                "Invalid port: 0. Must be between 1 and 65535".to_string(),
            ));
        }

        if !self.database_url.starts_with("postgres://")
            && !self.database_url.starts_with("postgresql://")
        {
            return Err(AppError::Config(
                "DATABASE_URL must start with postgres:// or postgresql://".to_string(),
            ));
        }

        if let Some(ref redis_url) = self.redis_url {
            if !redis_url.starts_with("redis://") && !redis_url.starts_with("rediss://") {
                return Err(AppError::Config(
                    "REDIS_URL must start with redis:// or rediss://".to_string(),
                ));
            }
        }

        if let Some(ref nats_url) = self.nats_url {
            if !nats_url.starts_with("nats://") {
                return Err(AppError::Config(
                    "NATS_URL must start with nats://".to_string(),
                ));
            }
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::env;

    fn setup_minimal_env() {
        env::set_var("DATABASE_URL", "postgres://localhost/test");
        env::set_var("SERVICE_PORT", "7106");
    }

    fn cleanup_env() {
        env::remove_var("ENV");
        env::remove_var("SERVICE_NAME");
        env::remove_var("SERVICE_PORT");
        env::remove_var("DATABASE_URL");
        env::remove_var("REDIS_URL");
        env::remove_var("NATS_URL");
        env::remove_var("OTEL_EXPORTER_OTLP_ENDPOINT");
        env::remove_var("JWT_PUBLIC_KEY_URL");
        env::remove_var("LOG_LEVEL");
    }

    #[test]
    fn test_config_loads_from_env() {
        cleanup_env();
        setup_minimal_env();
        env::set_var("ENV", "staging");
        env::set_var("SERVICE_NAME", "test-service");

        let config = AppConfig::from_env().unwrap();

        assert_eq!(config.env, "staging");
        assert_eq!(config.service_name, "test-service");
        assert_eq!(config.service_port, 7106);
        assert_eq!(config.database_url, "postgres://localhost/test");

        cleanup_env();
    }

    #[test]
    fn test_config_fails_without_required_vars() {
        cleanup_env();
        env::set_var("SERVICE_PORT", "7106");

        let result = AppConfig::from_env();
        assert!(result.is_err());
        assert!(result.unwrap_err().to_string().contains("DATABASE_URL"));

        cleanup_env();
    }

    #[test]
    fn test_config_validates_port_range() {
        cleanup_env();
        env::set_var("DATABASE_URL", "postgres://localhost/test");
        env::set_var("SERVICE_PORT", "70000"); // This will fail parsing to u16

        let result = AppConfig::from_env();
        assert!(result.is_err());
        assert!(result.unwrap_err().to_string().contains("Invalid SERVICE_PORT"));

        cleanup_env();
    }

    #[test]
    fn test_config_validates_database_url_format() {
        cleanup_env();
        env::set_var("DATABASE_URL", "http://invalid");
        env::set_var("SERVICE_PORT", "7106");

        let result = AppConfig::from_env();
        assert!(result.is_err());
        assert!(result
            .unwrap_err()
            .to_string()
            .contains("must start with postgres"));

        cleanup_env();
    }

    #[test]
    fn test_config_defaults_applied_correctly() {
        cleanup_env();
        setup_minimal_env();

        let config = AppConfig::from_env().unwrap();

        assert_eq!(config.log_level, "info");
        assert_eq!(config.service_name, "dep-governance-service");
        assert_eq!(config.env, "dev");

        cleanup_env();
    }

    #[test]
    fn test_config_validates_redis_url_format() {
        cleanup_env();
        setup_minimal_env();
        env::set_var("REDIS_URL", "http://invalid");

        let result = AppConfig::from_env();
        assert!(result.is_err());
        assert!(result
            .unwrap_err()
            .to_string()
            .contains("REDIS_URL must start with redis"));

        cleanup_env();
    }

    #[test]
    fn test_config_validates_nats_url_format() {
        cleanup_env();
        setup_minimal_env();
        env::set_var("NATS_URL", "http://invalid");

        let result = AppConfig::from_env();
        assert!(result.is_err());
        assert!(result
            .unwrap_err()
            .to_string()
            .contains("NATS_URL must start with nats"));

        cleanup_env();
    }
}
