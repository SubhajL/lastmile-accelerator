pub mod config;
pub mod db;
pub mod error;
pub mod events;
pub mod handlers;
pub mod middleware;
pub mod models;
pub mod services;
pub mod errors;
pub mod metrics;
// openapi module is added in the openapi-spec branch upstack

pub use config::AppConfig;
pub use error::{AppError, Result};
