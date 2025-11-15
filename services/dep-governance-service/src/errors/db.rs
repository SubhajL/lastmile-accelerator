use crate::error::AppError;

// Maps common Postgres SQLSTATE codes to HTTP-friendly AppError variants.
// No backward-compat behavior: deterministic, minimal mapping.
pub fn map_db_error(e: sqlx::Error) -> AppError {
    match &e {
        sqlx::Error::Database(db_err) => {
            match db_err.code().as_deref() {
                Some("23505") => AppError::Conflict("unique violation".into()),
                Some("23503") => AppError::BadRequest("foreign key violation".into()),
                _ => AppError::Database(e),
            }
        }
        _ => AppError::Database(e),
    }
}
