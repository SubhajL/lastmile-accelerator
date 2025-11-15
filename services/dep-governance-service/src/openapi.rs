use utoipa::OpenApi;
use utoipa::Modify;
use utoipa::openapi::security::{Http, HttpAuthScheme, SecurityRequirement, SecurityScheme};

#[derive(serde::Serialize, utoipa::ToSchema)]
struct ApiError { error: String }

#[derive(OpenApi)]
#[openapi(
    paths(
        crate::handlers::api::sbom::create_sbom_handler,
        crate::handlers::api::sbom::get_latest_sbom_handler,
        crate::handlers::api::deps::list_dependencies_handler,
        crate::handlers::api::vulns::get_dependency_vulns_handler,
        crate::handlers::api::cves::upsert_cve_handler,
        crate::handlers::api::cves::link_vuln_handler,
    ),
    components(
        schemas(
            crate::handlers::api::sbom::SbomCreateRequest,
            crate::models::Sbom,
            crate::handlers::api::deps::DependencyResponse,
            crate::handlers::api::vulns::CveSummary,
            crate::handlers::api::vulns::DependencyVulnResponse,
            crate::handlers::api::types::UpsertCveRequest,
            crate::handlers::api::types::CveResponse,
            crate::handlers::api::types::LinkVulnRequest,
            crate::handlers::api::types::LinkResponse,
            ApiError
        )
    ),
    info(title = "dep-governance-service API", version = "1.0.0"),
    modifiers(&SecurityAddon)
)]
struct ApiDoc;

struct SecurityAddon;

impl Modify for SecurityAddon {
    fn modify(&self, openapi: &mut utoipa::openapi::OpenApi) {
        let components = openapi
            .components
            .get_or_insert_with(utoipa::openapi::Components::new);
        components.add_security_scheme(
            "bearerAuth",
            SecurityScheme::Http(Http::new(HttpAuthScheme::Bearer)),
        );
        openapi.security = Some(vec![SecurityRequirement::new("bearerAuth", Vec::<String>::new())]);
    }
}

pub fn routes() -> utoipa_swagger_ui::SwaggerUi {
    let doc = ApiDoc::openapi();
    utoipa_swagger_ui::SwaggerUi::new("/docs").url("/openapi.json", doc)
}

pub fn doc() -> utoipa::openapi::OpenApi { ApiDoc::openapi() }
