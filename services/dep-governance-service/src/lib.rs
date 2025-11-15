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
pub mod openapi;
// web & validation modules are added in api-validation-polish upstack

pub use config::AppConfig;
pub use error::{AppError, Result};
