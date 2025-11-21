use serde::{Deserialize, Serialize};
use std::fmt;
use std::str::FromStr;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize, sqlx::Type)]
#[sqlx(type_name = "text")]
pub enum Ecosystem {
    #[sqlx(rename = "npm")]
    Npm,
    #[sqlx(rename = "cargo")]
    Cargo,
    #[sqlx(rename = "pypi")]
    PyPI,
    #[sqlx(rename = "maven")]
    Maven,
    #[sqlx(rename = "go")]
    Go,
    #[sqlx(rename = "nuget")]
    NuGet,
    #[sqlx(rename = "composer")]
    Composer,
    #[sqlx(rename = "rubygems")]
    RubyGems,
    #[sqlx(rename = "unknown")]
    Unknown,
}

impl fmt::Display for Ecosystem {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let s = match self {
            Self::Npm => "npm",
            Self::Cargo => "cargo",
            Self::PyPI => "pypi",
            Self::Maven => "maven",
            Self::Go => "go",
            Self::NuGet => "nuget",
            Self::Composer => "composer",
            Self::RubyGems => "rubygems",
            Self::Unknown => "unknown",
        };
        write!(f, "{}", s)
    }
}

impl FromStr for Ecosystem {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_lowercase().as_str() {
            "npm" => Ok(Self::Npm),
            "cargo" | "crates.io" => Ok(Self::Cargo),
            "pypi" | "pip" => Ok(Self::PyPI),
            "maven" => Ok(Self::Maven),
            "go" | "golang" => Ok(Self::Go),
            "nuget" => Ok(Self::NuGet),
            "composer" | "packagist" => Ok(Self::Composer),
            "rubygems" | "gem" => Ok(Self::RubyGems),
            _ => Ok(Self::Unknown),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize, sqlx::Type)]
#[sqlx(type_name = "text")]
pub enum Severity {
    #[sqlx(rename = "info")]
    Info,
    #[sqlx(rename = "low")]
    Low,
    #[sqlx(rename = "medium")]
    Medium,
    #[sqlx(rename = "high")]
    High,
    #[sqlx(rename = "critical")]
    Critical,
}

// Custom Ord implementation for severity (Critical > High > Medium > Low > Info)
impl PartialOrd for Severity {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        Some(self.cmp(other))
    }
}

impl Ord for Severity {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        let self_val = match self {
            Self::Critical => 5,
            Self::High => 4,
            Self::Medium => 3,
            Self::Low => 2,
            Self::Info => 1,
        };
        let other_val = match other {
            Self::Critical => 5,
            Self::High => 4,
            Self::Medium => 3,
            Self::Low => 2,
            Self::Info => 1,
        };
        self_val.cmp(&other_val)
    }
}

impl Severity {
    pub fn from_cvss_score(score: f32) -> Self {
        match score {
            s if s >= 9.0 => Self::Critical,
            s if s >= 7.0 => Self::High,
            s if s >= 4.0 => Self::Medium,
            s if s >= 0.1 => Self::Low,
            _ => Self::Info,
        }
    }

    pub fn to_cvss_range(&self) -> (f32, f32) {
        match self {
            Self::Critical => (9.0, 10.0),
            Self::High => (7.0, 8.9),
            Self::Medium => (4.0, 6.9),
            Self::Low => (0.1, 3.9),
            Self::Info => (0.0, 0.0),
        }
    }
}

impl fmt::Display for Severity {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let s = match self {
            Self::Critical => "critical",
            Self::High => "high",
            Self::Medium => "medium",
            Self::Low => "low",
            Self::Info => "info",
        };
        write!(f, "{}", s)
    }
}

impl FromStr for Severity {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_lowercase().as_str() {
            "critical" => Ok(Self::Critical),
            "high" => Ok(Self::High),
            "medium" => Ok(Self::Medium),
            "low" => Ok(Self::Low),
            "info" | "informational" => Ok(Self::Info),
            _ => Err(format!("Invalid severity: {}", s)),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize, sqlx::Type)]
#[sqlx(type_name = "text")]
pub enum LicenseType {
    #[sqlx(rename = "permissive")]
    Permissive,
    #[sqlx(rename = "copyleft")]
    Copyleft,
    #[sqlx(rename = "proprietary")]
    Proprietary,
    #[sqlx(rename = "unknown")]
    Unknown,
}

impl fmt::Display for LicenseType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let s = match self {
            Self::Permissive => "permissive",
            Self::Copyleft => "copyleft",
            Self::Proprietary => "proprietary",
            Self::Unknown => "unknown",
        };
        write!(f, "{}", s)
    }
}

impl FromStr for LicenseType {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_lowercase().as_str() {
            "permissive" => Ok(Self::Permissive),
            "copyleft" => Ok(Self::Copyleft),
            "proprietary" => Ok(Self::Proprietary),
            _ => Ok(Self::Unknown),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize, sqlx::Type)]
#[sqlx(type_name = "text")]
pub enum ScanStatus {
    #[sqlx(rename = "pending")]
    Pending,
    #[sqlx(rename = "running")]
    Running,
    #[sqlx(rename = "completed")]
    Completed,
    #[sqlx(rename = "failed")]
    Failed,
}

impl fmt::Display for ScanStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let s = match self {
            Self::Pending => "pending",
            Self::Running => "running",
            Self::Completed => "completed",
            Self::Failed => "failed",
        };
        write!(f, "{}", s)
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize, sqlx::Type)]
#[sqlx(type_name = "text")]
pub enum PolicyScope {
    #[sqlx(rename = "organization")]
    Organization,
    #[sqlx(rename = "project")]
    Project,
}

impl fmt::Display for PolicyScope {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let s = match self {
            Self::Organization => "organization",
            Self::Project => "project",
        };
        write!(f, "{}", s)
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize, sqlx::Type)]
#[sqlx(type_name = "text")]
pub enum VulnerabilityStatus {
    #[sqlx(rename = "open")]
    Open,
    #[sqlx(rename = "acknowledged")]
    Acknowledged,
    #[sqlx(rename = "fixed")]
    Fixed,
    #[sqlx(rename = "ignored")]
    Ignored,
    #[sqlx(rename = "false_positive")]
    FalsePositive,
}

impl fmt::Display for VulnerabilityStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let s = match self {
            Self::Open => "open",
            Self::Acknowledged => "acknowledged",
            Self::Fixed => "fixed",
            Self::Ignored => "ignored",
            Self::FalsePositive => "false_positive",
        };
        write!(f, "{}", s)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_ecosystem_parsing() {
        assert_eq!("npm".parse::<Ecosystem>().unwrap(), Ecosystem::Npm);
        assert_eq!("CARGO".parse::<Ecosystem>().unwrap(), Ecosystem::Cargo);
        assert_eq!("crates.io".parse::<Ecosystem>().unwrap(), Ecosystem::Cargo);
        assert_eq!(
            "unknown-eco".parse::<Ecosystem>().unwrap(),
            Ecosystem::Unknown
        );
    }

    #[test]
    fn test_ecosystem_display() {
        assert_eq!(Ecosystem::Npm.to_string(), "npm");
        assert_eq!(Ecosystem::PyPI.to_string(), "pypi");
    }

    #[test]
    fn test_severity_ordering() {
        assert!(Severity::Critical > Severity::High);
        assert!(Severity::High > Severity::Medium);
        assert!(Severity::Medium > Severity::Low);
        assert!(Severity::Low > Severity::Info);
    }

    #[test]
    fn test_severity_from_cvss_score() {
        assert_eq!(Severity::from_cvss_score(9.8), Severity::Critical);
        assert_eq!(Severity::from_cvss_score(7.5), Severity::High);
        assert_eq!(Severity::from_cvss_score(5.0), Severity::Medium);
        assert_eq!(Severity::from_cvss_score(2.0), Severity::Low);
        assert_eq!(Severity::from_cvss_score(0.0), Severity::Info);
    }

    #[test]
    fn test_severity_to_cvss_range() {
        let (min, max) = Severity::Critical.to_cvss_range();
        assert_eq!(min, 9.0);
        assert_eq!(max, 10.0);
    }

    #[test]
    fn test_severity_parsing() {
        assert_eq!("critical".parse::<Severity>().unwrap(), Severity::Critical);
        assert_eq!("HIGH".parse::<Severity>().unwrap(), Severity::High);
        assert!("invalid".parse::<Severity>().is_err());
    }

    #[test]
    fn test_license_type_parsing() {
        assert_eq!(
            "permissive".parse::<LicenseType>().unwrap(),
            LicenseType::Permissive
        );
        assert_eq!(
            "copyleft".parse::<LicenseType>().unwrap(),
            LicenseType::Copyleft
        );
        assert_eq!(
            "unknown-license".parse::<LicenseType>().unwrap(),
            LicenseType::Unknown
        );
    }

    #[test]
    fn test_scan_status_display() {
        assert_eq!(ScanStatus::Pending.to_string(), "pending");
        assert_eq!(ScanStatus::Running.to_string(), "running");
        assert_eq!(ScanStatus::Completed.to_string(), "completed");
    }

    #[test]
    fn test_policy_scope_display() {
        assert_eq!(PolicyScope::Organization.to_string(), "organization");
        assert_eq!(PolicyScope::Project.to_string(), "project");
    }

    #[test]
    fn test_vulnerability_status_display() {
        assert_eq!(VulnerabilityStatus::Open.to_string(), "open");
        assert_eq!(VulnerabilityStatus::Fixed.to_string(), "fixed");
        assert_eq!(
            VulnerabilityStatus::FalsePositive.to_string(),
            "false_positive"
        );
    }
}
