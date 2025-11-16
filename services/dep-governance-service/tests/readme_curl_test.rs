use std::fs;
use std::path::PathBuf;

fn read_readme() -> String {
    let p = PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("README.md");
    fs::read_to_string(p).expect("read README.md")
}

fn extract_curl_blocks(md: &str) -> Vec<String> {
    let mut out = Vec::new();
    let mut in_block = false;
    let mut current = String::new();
    for line in md.lines() {
        if line.trim_start().starts_with("```") {
            if in_block {
                // end block
                if current.trim_start().starts_with("curl ") {
                    out.push(current.clone());
                }
                current.clear();
                in_block = false;
            } else {
                in_block = true;
            }
            continue;
        }
        if in_block {
            current.push_str(line);
            current.push('\n');
        }
    }
    out
}

#[test]
fn readme_contains_required_curl_examples_for_paths() {
    let md = read_readme();
    let blocks = extract_curl_blocks(&md);
    let text = md.as_str();

    // public endpoints
    assert!(text.contains("/healthz"));
    assert!(text.contains("/readyz"));
    assert!(text.contains("/docs") || text.contains("/docs/"));
    assert!(text.contains("/openapi.json"));

    // v1 endpoints
    let required = [
        "/v1/snapshots/", // sbom post/get
        "/v1/cves",
        "/v1/dependencies/", // link + list vulns
    ];
    for r in required.iter() {
        assert!(blocks.iter().any(|b| b.contains(r)), "missing curl for {r}");
    }
}

#[test]
fn v1_examples_include_bearer_authorization_header() {
    let md = read_readme();
    let blocks = extract_curl_blocks(&md);
    for b in blocks {
        if b.contains("/v1/") {
            assert!(b.contains("Authorization: Bearer $TOKEN"), "v1 example missing Authorization header: {b}");
        }
    }
}

#[test]
fn example_bodies_exist_and_are_valid_json() {
    let root = PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("examples");
    let files = ["sbom.json", "upsert-cve.json", "link-vuln.json"];
    for f in files {
        let p = root.join(f);
        let s = fs::read_to_string(&p).expect("read example json");
        let _: serde_json::Value = serde_json::from_str(&s).expect("parse example json");
    }
}
