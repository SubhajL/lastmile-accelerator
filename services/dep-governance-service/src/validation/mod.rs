use once_cell::sync::Lazy;
use regex::Regex;

use crate::error::AppError;

static CVE_RE: Lazy<Regex> = Lazy::new(|| Regex::new(r"^CVE-\d{4}-\d{4,7}$").unwrap());

pub fn validate_cve_id(s: &str) -> Result<(), AppError> {
    if CVE_RE.is_match(s) { Ok(()) } else { Err(AppError::BadRequest("invalid cveId format".into())) }
}

pub fn validate_cvss_score(score: Option<f32>) -> Result<(), AppError> {
    if let Some(v) = score {
        if (0.0..=10.0).contains(&v) { Ok(()) } else { Err(AppError::BadRequest("cvssScore must be between 0.0 and 10.0".into())) }
    } else { Ok(()) }
}
