pub mod config;
pub mod db;
pub mod error;
pub mod events;
pub mod handlers;
pub mod middleware;
pub mod models;
pub mod services;
pub mod errors;
pub mod openapi;
pub mod metrics;
pub mod web;
pub mod validation;
// modules above are added across stacked branches

pub use config::AppConfig;
pub use error::{AppError, Result};
