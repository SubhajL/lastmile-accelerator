use axum::{async_trait, extract::{FromRequest, Json, Request}};
use axum::extract::rejection::JsonRejection;

use crate::error::AppError;

pub struct StrictJson<T>(pub T);

#[async_trait]
impl<S, T> FromRequest<S> for StrictJson<T>
where
    Json<T>: FromRequest<S, Rejection = JsonRejection>,
    S: Send + Sync,
{
    type Rejection = AppError;

    async fn from_request(req: Request, state: &S) -> Result<Self, Self::Rejection> {
        match Json::<T>::from_request(req, state).await {
            Ok(Json(v)) => Ok(StrictJson(v)),
            Err(rej) => {
                let msg = rej.body_text();
                let message = if msg.is_empty() { "invalid JSON body".to_string() } else { msg };
                Err(AppError::BadRequest(message))
            }
        }
    }
}
