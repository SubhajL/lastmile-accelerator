 # dep-governance-service

Runs on port 7106 with a `/healthz` endpoint.

## Curl quickstart

Set environment and token:

```
export BASE=http://localhost:8080
export TOKEN="<your_jwt_here>"
```

Public endpoints:

```
curl -sS $BASE/healthz
curl -sS $BASE/readyz
curl -sS $BASE/docs/
curl -sS $BASE/openapi.json
```

SBOM endpoints:

```
SNAPSHOT_ID=$(uuidgen)
curl -sS -X POST \
  -H 'Authorization: Bearer $TOKEN' \
  -H 'Content-Type: application/json' \
  -d @examples/sbom.json \
  "$BASE/v1/snapshots/$SNAPSHOT_ID/sbom"

curl -sS -H 'Authorization: Bearer $TOKEN' \
  "$BASE/v1/snapshots/$SNAPSHOT_ID/sbom"
```

Dependencies for a snapshot:

```
curl -sS -H 'Authorization: Bearer $TOKEN' \
  "$BASE/v1/snapshots/$SNAPSHOT_ID/dependencies?direct=true"
```

CVE upsert and link:

```
curl -sS -X POST \
  -H 'Authorization: Bearer $TOKEN' \
  -H 'Content-Type: application/json' \
  -d @examples/upsert-cve.json \
  "$BASE/v1/cves"

DEPENDENCY_ID="<uuid>"
curl -sS -X POST \
  -H 'Authorization: Bearer $TOKEN' \
  -H 'Content-Type: application/json' \
  -d @examples/link-vuln.json \
  "$BASE/v1/dependencies/$DEPENDENCY_ID/vulns/link"

curl -sS -H 'Authorization: Bearer $TOKEN' \
  "$BASE/v1/dependencies/$DEPENDENCY_ID/vulns"
```

 ## Dev
 ```bash
 make build && make run
 ```
