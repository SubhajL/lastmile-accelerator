use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::models::Ecosystem;

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, sqlx::FromRow)]
pub struct Dependency {
    pub id: Uuid,
    pub snapshot_id: Uuid,
    pub name: String,
    pub version: String,
    pub ecosystem: String,
    pub is_direct: bool,
    pub parent_id: Option<Uuid>,
    pub created_at: DateTime<Utc>,
}

impl Dependency {
    pub fn new(
        snapshot_id: Uuid,
        name: impl Into<String>,
        version: impl Into<String>,
        ecosystem: Ecosystem,
        is_direct: bool,
        parent_id: Option<Uuid>,
    ) -> Result<Self, String> {
        let name = name.into().trim().to_string();
        let version = version.into().trim().to_string();

        if name.is_empty() {
            return Err("Dependency name cannot be empty".into());
        }
        if version.is_empty() {
            return Err("Dependency version cannot be empty".into());
        }

        Ok(Self {
            id: Uuid::new_v4(),
            snapshot_id,
            name,
            version,
            ecosystem: ecosystem.to_string(),
            is_direct,
            parent_id,
            created_at: Utc::now(),
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_dependency_creation_validates_name() {
        let result = Dependency::new(Uuid::new_v4(), " ", "1.0.0", Ecosystem::Npm, true, None);
        assert!(result.is_err());
    }

    #[test]
    fn test_dependency_creation_validates_version() {
        let result = Dependency::new(Uuid::new_v4(), "lodash", " ", Ecosystem::Npm, true, None);
        assert!(result.is_err());
    }

    #[test]
    fn test_dependency_creation_ok() {
        let dep = Dependency::new(
            Uuid::new_v4(),
            "lodash",
            "4.17.21",
            Ecosystem::Npm,
            true,
            None,
        )
        .unwrap();

        assert_eq!(dep.name, "lodash");
        assert_eq!(dep.version, "4.17.21");
        assert_eq!(dep.ecosystem, "npm");
        assert!(dep.is_direct);
        assert!(dep.parent_id.is_none());
    }
}
