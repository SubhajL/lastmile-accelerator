package docs

import (
    "bufio"
    "os"
    "path/filepath"
    "regexp"
    "runtime"
    "strings"
    "testing"
)

// resolve paths relative to this test file regardless of CWD
func mustPath(parts ...string) string {
    _, file, _, _ := runtime.Caller(0)
    base := filepath.Dir(file)
    p := filepath.Join(append([]string{base}, parts...)...)
    return filepath.Clean(p)
}

func readFile(t *testing.T, path string) string {
    t.Helper()
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read %s: %v", path, err) }
    return string(b)
}

// parse env var names from config.Load and map field->ENV_KEY
func parseEnvMapFromConfig(src string) map[string]string {
    envMap := map[string]string{}
    // match lines like: FieldName:     getEnv("ENV", "dev"),
    re := regexp.MustCompile(`(?m)\s*([A-Za-z0-9_]+)\s*:\s*getEnv\("([A-Z0-9_]+)"`)
    for _, m := range re.FindAllStringSubmatch(src, -1) {
        envMap[m[1]] = m[2]
    }
    return envMap
}

// parse required fields from Validate: if c.<Field> == "" {
func parseRequiredFieldsFromValidate(src string) map[string]struct{} {
    req := map[string]struct{}{}
    re := regexp.MustCompile(`(?m)if\s+c\.([A-Za-z0-9_]+)\s*==\s*""\s*{`)
    for _, m := range re.FindAllStringSubmatch(src, -1) {
        req[m[1]] = struct{}{}
    }
    return req
}

// parse Environment Matrix section of AGENTS.md and return set of env var names
func parseEnvMatrixFromAgents(md string) map[string]struct{} {
    lines := strings.Split(md, "\n")
    envs := map[string]struct{}{}
    in := false
    for _, ln := range lines {
        if strings.HasPrefix(strings.ToLower(strings.TrimSpace(ln)), "## environment matrix") { in = true; continue }
        if !in { continue }
        if strings.HasPrefix(strings.TrimSpace(ln), "## ") { break }
        if strings.HasPrefix(strings.TrimSpace(ln), "|") {
            cells := splitMarkdownRow(ln)
            if len(cells) == 0 || strings.EqualFold(cells[0], "variable") { continue }
            name := strings.TrimSpace(cells[0])
            if name != "" && name != ":-" && name != "-" { envs[name] = struct{}{} }
        }
    }
    return envs
}

func splitMarkdownRow(row string) []string {
    // naive split: | a | b | c |
    parts := []string{}
    cur := ""
    for i := 0; i < len(row); i++ {
        ch := row[i]
        if ch == '|' {
            if cur != "" { parts = append(parts, strings.TrimSpace(cur)) }
            cur = ""
            continue
        }
        cur += string(ch)
    }
    if cur != "" { parts = append(parts, strings.TrimSpace(cur)) }
    // drop trailing alignment lines like ---
    pruned := []string{}
    for _, p := range parts {
        if p == "" { continue }
        pruned = append(pruned, p)
    }
    return pruned
}

// parse Scopes Matrix: rows of RPC | scope
func parseScopesMatrixFromAgents(md string) map[string]string {
    lines := strings.Split(md, "\n")
    out := map[string]string{}
    in := false
    for _, ln := range lines {
        if strings.HasPrefix(strings.ToLower(strings.TrimSpace(ln)), "## scopes matrix") { in = true; continue }
        if !in { continue }
        if strings.HasPrefix(strings.TrimSpace(ln), "## ") { break }
        if strings.HasPrefix(strings.TrimSpace(ln), "|") {
            cells := splitMarkdownRow(ln)
            if len(cells) < 2 || strings.EqualFold(cells[0], "rpc") { continue }
            rpc := strings.TrimSpace(cells[0])
            scope := strings.TrimSpace(cells[1])
            if rpc != "" && scope != "" { out[rpc] = scope }
        }
    }
    return out
}

// parse proto to list RPC names (e.g., CreateSLO, GetSLO, ...)
func parseRPCsFromProto(src string) []string {
    rpcs := []string{}
    re := regexp.MustCompile(`(?m)^\s*rpc\s+([A-Za-z0-9_]+)\s*\(`)
    for _, m := range re.FindAllStringSubmatch(src, -1) {
        rpcs = append(rpcs, m[1])
    }
    return rpcs
}

// compute method->scope from auth_interceptor.go requiredScopesFor
func parseRequiredScopesFromInterceptor(src string) map[string]string {
    scopes := map[string]string{}
    // handle case groupings by suffix, with explicit listings per scope class
    // read-only
    add := func(names []string, scope string) { for _, n := range names { scopes[n] = scope } }
    namesFor := func(block string) []string {
        out := []string{}
        // match strings.HasSuffix(fullMethod, "/Name") occurrences in a contiguous block
        re := regexp.MustCompile(`HasSuffix\(fullMethod,\s*"/([A-Za-z0-9_]+)"\)`) // capture Name
        for _, m := range re.FindAllStringSubmatch(block, -1) { out = append(out, m[1]) }
        return out
    }
    // coarse split by case groups
    reader := bufio.NewScanner(strings.NewReader(src))
    buf := ""
    for reader.Scan() {
        line := reader.Text()
        buf += line + "\n"
    }
    // identify three groups by keywords
    readGroup := between(buf, "// read-only", "// write")
    writeGroup := between(buf, "// write", "// ingest")
    ingestGroup := between(buf, "// ingest", "default:")
    add(namesFor(readGroup), "observability:read")
    add(namesFor(writeGroup), "observability:write")
    add(namesFor(ingestGroup), "observability:ingest")
    return scopes
}

func between(s, start, end string) string {
    i := strings.Index(s, start)
    if i == -1 { return "" }
    s = s[i:]
    j := strings.Index(s, end)
    if j == -1 { return s }
    return s[:j]
}

func TestEnvDocSync_AllRequiredVarsDocumented(t *testing.T) {
    svcRoot := mustPath("..", "..")
    cfgSrc := readFile(t, filepath.Join(svcRoot, "internal", "config", "config.go"))
    agents := readFile(t, filepath.Join(svcRoot, "AGENTS.md"))

    envMap := parseEnvMapFromConfig(cfgSrc)
    reqFields := parseRequiredFieldsFromValidate(cfgSrc)
    requiredEnvs := map[string]struct{}{}
    for field, env := range envMap {
        if _, ok := reqFields[field]; ok { requiredEnvs[env] = struct{}{} }
    }
    documented := parseEnvMatrixFromAgents(agents)
    missing := []string{}
    for env := range requiredEnvs { if _, ok := documented[env]; !ok { missing = append(missing, env) } }
    if len(missing) > 0 { t.Fatalf("missing required envs in Environment Matrix: %v", missing) }
}

func TestScopesDocSync_AllRPCsHaveRows_AndMatchEnforcement(t *testing.T) {
    svcRoot := mustPath("..", "..")
    agents := readFile(t, filepath.Join(svcRoot, "AGENTS.md"))
    proto := readFile(t, filepath.Join(svcRoot, "..", "..", "packages", "contracts", "observability", "observability.proto"))
    interceptor := readFile(t, filepath.Join(svcRoot, "internal", "grpcserver", "auth_interceptor.go"))

    rpcs := parseRPCsFromProto(proto)
    doc := parseScopesMatrixFromAgents(agents)
    req := parseRequiredScopesFromInterceptor(interceptor)

    // coverage
    missing := []string{}
    for _, rpc := range rpcs { if _, ok := doc[rpc]; !ok { missing = append(missing, rpc) } }
    if len(missing) > 0 { t.Fatalf("missing RPCs in Scopes Matrix: %v", missing) }

    // match scope (default to read if interceptor not explicit)
    for _, rpc := range rpcs {
        want := req[rpc]
        if want == "" { want = "observability:read" }
        got := doc[rpc]
        if got != want { t.Fatalf("scope mismatch for %s: want %s got %s", rpc, want, got) }
    }
}

func TestProtoDocs_HaveCommentsForServiceAndRPCs(t *testing.T) {
    svcRoot := mustPath("..", "..")
    protoPath := filepath.Join(svcRoot, "..", "..", "packages", "contracts", "observability", "observability.proto")
    src := readFile(t, protoPath)
    lines := strings.Split(src, "\n")

    // service comment present directly above line starting with 'service ObservabilityService'
    svcIdx := -1
    for i, ln := range lines { if strings.HasPrefix(strings.TrimSpace(ln), "service ObservabilityService") { svcIdx = i; break } }
    if svcIdx == -1 { t.Fatalf("service ObservabilityService not found in proto") }
    if !hasLeadingComment(lines, svcIdx) { t.Fatalf("service ObservabilityService should have leading // comment") }

    // each rpc line must have a leading comment
    for i, ln := range lines {
        ts := strings.TrimSpace(ln)
        if strings.HasPrefix(ts, "rpc ") { if !hasLeadingComment(lines, i) { t.Fatalf("rpc at line %d missing leading // comment", i+1) } }
    }
}

func hasLeadingComment(lines []string, idx int) bool {
    // scan upward to first non-empty line and require it starts with //
    for j := idx-1; j >= 0; j-- {
        s := strings.TrimSpace(lines[j])
        if s == "" { continue }
        return strings.HasPrefix(s, "//")
    }
    return false
}
