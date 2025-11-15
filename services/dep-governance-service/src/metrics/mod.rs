use once_cell::sync::Lazy;
use prometheus::{Encoder, HistogramOpts, HistogramVec, IntCounterVec, IntGaugeVec, Opts, Registry, TextEncoder};

pub struct HttpMetrics {
    pub requests_total: IntCounterVec,
    pub request_duration_seconds: HistogramVec,
    pub inflight: IntGaugeVec,
}

pub struct JwksMetrics {
    pub fetch_total: IntCounterVec,
}

pub struct DbMetrics {
    pub query_duration_seconds: HistogramVec,
}

pub struct Metrics {
    pub registry: Registry,
    pub http: HttpMetrics,
    pub jwks: JwksMetrics,
    pub db: DbMetrics,
}

pub static METRICS: Lazy<Metrics> = Lazy::new(|| {
    let registry = Registry::new();

    let requests_total = IntCounterVec::new(
        Opts::new("dep_governance_http_requests_total", "HTTP requests total"),
        &["method", "route", "status", "outcome"],
    )
    .unwrap();
    let request_duration_seconds = HistogramVec::new(
        HistogramOpts::new(
            "dep_governance_http_request_duration_seconds",
            "HTTP request duration in seconds",
        )
        .buckets(vec![0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 5.0]),
        &["method", "route", "status"],
    )
    .unwrap();
    let inflight = IntGaugeVec::new(
        Opts::new("dep_governance_inflight_requests", "In-flight requests"),
        &["route"],
    )
    .unwrap();

    let fetch_total = IntCounterVec::new(
        Opts::new("dep_governance_jwks_fetch_total", "JWKS fetch outcomes"),
        &["outcome"],
    )
    .unwrap();

    let query_duration_seconds = HistogramVec::new(
        HistogramOpts::new(
            "dep_governance_db_query_duration_seconds",
            "DB query duration in seconds",
        )
        .buckets(vec![0.002, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0]),
        &["op"],
    )
    .unwrap();

    registry.register(Box::new(requests_total.clone())).unwrap();
    registry
        .register(Box::new(request_duration_seconds.clone()))
        .unwrap();
    registry.register(Box::new(inflight.clone())).unwrap();
    registry.register(Box::new(fetch_total.clone())).unwrap();
    registry.register(Box::new(query_duration_seconds.clone())).unwrap();

    Metrics {
        registry,
        http: HttpMetrics { requests_total, request_duration_seconds, inflight },
        jwks: JwksMetrics { fetch_total },
        db: DbMetrics { query_duration_seconds },
    }
});

pub fn registry() -> &'static Registry { &METRICS.registry }
pub fn http() -> &'static HttpMetrics { &METRICS.http }
pub fn jwks() -> &'static JwksMetrics { &METRICS.jwks }
pub fn db() -> &'static DbMetrics { &METRICS.db }

pub fn encode() -> String {
    let metric_families = METRICS.registry.gather();
    let mut buf = Vec::new();
    TextEncoder::new().encode(&metric_families, &mut buf).unwrap();
    String::from_utf8(buf).unwrap()
}
