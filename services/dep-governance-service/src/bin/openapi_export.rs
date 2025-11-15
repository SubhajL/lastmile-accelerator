fn main() -> Result<(), Box<dyn std::error::Error>> {
    let doc = dep_governance_service::openapi::doc();
    let json_path = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("openapi.json");
    let yaml_path = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("openapi.yaml");

    let json = serde_json::to_vec_pretty(&doc)?;
    std::fs::write(&json_path, json)?;
    let yaml = serde_yaml::to_string(&serde_json::to_value(&doc)?)?;
    std::fs::write(&yaml_path, yaml)?;
    println!("Wrote {} and {}", json_path.display(), yaml_path.display());
    Ok(())
}
