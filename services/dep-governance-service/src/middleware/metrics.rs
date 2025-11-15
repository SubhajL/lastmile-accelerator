use std::{future::Future, pin::Pin, task::{Context, Poll}, time::Instant};

use axum::{extract::MatchedPath, http::{Request, StatusCode}};
use tower::{Layer, Service};

pub fn metrics_layer<S>() -> MetricsLayer<S> { MetricsLayer { _inner: std::marker::PhantomData } }

#[derive(Clone)]
pub struct MetricsLayer<S> { _inner: std::marker::PhantomData<S> }

impl<S> Layer<S> for MetricsLayer<S> {
    type Service = MetricsService<S>;
    fn layer(&self, inner: S) -> Self::Service { MetricsService { inner } }
}

#[derive(Clone)]
pub struct MetricsService<S> { inner: S }

impl<S, B> Service<Request<B>> for MetricsService<S>
where
    S: Service<Request<B>, Response = axum::response::Response> + Clone + Send + 'static,
    S::Response: Send + 'static,
    S::Future: Send + 'static,
    S::Error: Send + 'static,
    B: Send + 'static,
{
    type Response = S::Response;
    type Error = S::Error;
    type Future = Pin<Box<dyn Future<Output = Result<Self::Response, Self::Error>> + Send>>;

    fn poll_ready(&mut self, cx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
        self.inner.poll_ready(cx)
    }

    fn call(&mut self, req: Request<B>) -> Self::Future {
        let method = req.method().to_string();
        let route = req
            .extensions()
            .get::<MatchedPath>()
            .map(|m| m.as_str().to_string())
            .unwrap_or_else(|| "UNKNOWN".to_string());

        let inflight = crate::metrics::http().inflight.with_label_values(&[&route]);
        inflight.inc();
        let start = Instant::now();

        let mut inner = self.inner.clone();
        Box::pin(async move {
            let res = inner.call(req).await;
            let code_str = match &res {
                Ok(ref r) => format!("{}", r.status().as_u16()),
                Err(_) => format!("{}", StatusCode::INTERNAL_SERVER_ERROR.as_u16()),
            };
            let outcome = if code_str.starts_with('2') || code_str.starts_with('3') { "success" } else { "error" };

            let http = crate::metrics::http();
            http.requests_total
                .with_label_values(&[&method, &route, &code_str, outcome])
                .inc();
            let dur = start.elapsed().as_secs_f64();
            http.request_duration_seconds
                .with_label_values(&[&method, &route, &code_str])
                .observe(dur);
            inflight.dec();
            res
        })
    }
}
