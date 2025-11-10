use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SbomFormat {
    SpdxJson,
    CycloneDxJson,
}

impl core::fmt::Display for SbomFormat {
    fn fmt(&self, f: &mut core::fmt::Formatter<'_>) -> core::fmt::Result {
        match self {
            SbomFormat::SpdxJson => write!(f, "spdx_json"),
            SbomFormat::CycloneDxJson => write!(f, "cyclonedx_json"),
        }
    }
}

impl TryFrom<&str> for SbomFormat {
    type Error = String;
    fn try_from(value: &str) -> Result<Self, Self::Error> {
        match value.to_lowercase().as_str() {
            "spdx_json" | "spdx" => Ok(SbomFormat::SpdxJson),
            "cyclonedx_json" | "cyclonedx" => Ok(SbomFormat::CycloneDxJson),
            other => Err(format!("Unsupported SBOM format: {}", other)),
        }
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, sqlx::FromRow)]
pub struct Sbom {
    pub id: Uuid,
    pub snapshot_id: Uuid,
    pub format: String,
    pub storage_key: String,
    pub file_hash: String,
    pub generated_at: DateTime<Utc>,
}

impl Sbom {
    pub fn new(
        snapshot_id: Uuid,
        format: SbomFormat,
        storage_key: impl Into<String>,
        file_hash: impl Into<String>,
        generated_at: Option<DateTime<Utc>>,
    ) -> Result<Self, String> {
        let storage_key = storage_key.into();
        if !storage_key.starts_with("s3://") && !storage_key.starts_with("gs://") {
            return Err("storage_key must be an object URI (s3:// or gs://)".into());
        }
        let file_hash = file_hash.into();
        if file_hash.len() != 64 || !file_hash.chars().all(|c| c.is_ascii_hexdigit()) {
            return Err("file_hash must be a 64-char hex SHA256".into());
        }

        Ok(Self {
            id: Uuid::new_v4(),
            snapshot_id,
            format: format.to_string(),
            storage_key,
            file_hash,
            generated_at: generated_at.unwrap_or_else(Utc::now),
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_sbom_creation_ok() {
        let s = Sbom::new(
            Uuid::new_v4(),
            SbomFormat::SpdxJson,
            "s3://bucket/sboms/123.json",
            "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
            None,
        )
        .unwrap();
        assert_eq!(s.format, "spdx_json");
    }

    #[test]
    fn test_sbom_storage_key_validation() {
        let err = Sbom::new(
            Uuid::new_v4(),
            SbomFormat::CycloneDxJson,
            "/tmp/local/path.json",
            "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
            None,
        )
        .unwrap_err();
        assert!(err.contains("storage_key"));
    }

    #[test]
    fn test_sbom_hash_validation() {
        let err = Sbom::new(
            Uuid::new_v4(),
            SbomFormat::SpdxJson,
            "s3://bucket/path.json",
            "short",
            None,
        )
        .unwrap_err();
        assert!(err.contains("SHA256"));
    }
}
