package doclint

import (
    "os"
    "path/filepath"
    "runtime"
    "testing"
)

func readREADME(t *testing.T) []byte {
    t.Helper()
    _, thisFile, _, _ := runtime.Caller(0)
    // internal/doclint/ -> ../../README.md
    p := filepath.Join(filepath.Dir(thisFile), "..", "..", "README.md")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("readme read err: %v", err) }
    return b
}

func TestREADME_RequiredSections_Present(t *testing.T) {
    data := readREADME(t)
    if err := ValidateSections(data); err != nil { t.Fatalf("sections: %v", err) }
}

func TestREADME_EnvKeys_Covered(t *testing.T) {
    data := readREADME(t)
    if err := ValidateEnvKeys(data); err != nil { t.Fatalf("env keys: %v", err) }
}

func TestREADME_Endpoints_Mentioned(t *testing.T) {
    data := readREADME(t)
    if err := ValidateEndpoints(data); err != nil { t.Fatalf("endpoints: %v", err) }
}

func TestREADME_Observability_Documented(t *testing.T) {
    data := readREADME(t)
    if err := ValidateObservability(data); err != nil { t.Fatalf("observability: %v", err) }
}

func TestREADME_GRPC_Toggles_Documented(t *testing.T) {
    data := readREADME(t)
    if err := ValidateGRPCToggles(data); err != nil { t.Fatalf("grpc toggles: %v", err) }
}

func TestREADME_RBAC_RolesToScopes_Documented(t *testing.T) {
    data := readREADME(t)
    if err := ValidateRBAC(data); err != nil { t.Fatalf("rbac: %v", err) }
}

func TestREADME_Quickstart_Exists(t *testing.T) {
    data := readREADME(t)
    if err := ValidateQuickstart(data); err != nil { t.Fatalf("quickstart: %v", err) }
}

func TestREADME_Troubleshooting_Minimum(t *testing.T) {
    data := readREADME(t)
    if err := ValidateTroubleshooting(data); err != nil { t.Fatalf("troubleshooting: %v", err) }
}
