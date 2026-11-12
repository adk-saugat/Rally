# Local mock-data test result

PASS — location sharing worked against the real local services and freshly built gateway on 2026-11-12. Tested revision `89cc908` plus the working-tree changes in `gateway/internal/handler/location.go` and `gateway/internal/utils/bearer/token.go`.

Run: `go run ./scripts/location-smoke` with the Compose stack running and the current gateway on port 18080. The harness creates three synthetic users and two groups through real Auth and Group APIs; no real location data or mocked success responses are used.

- All 18 measured movement updates arrived with matching identity, group, coordinates, and timestamp across gRPC, WebSocket, and mixed clients.
- Latest positions persisted in Redis. Forged gRPC user IDs were replaced with authenticated identities.
- Reconnection, zero/boundary coordinates, and server-generated timestamps passed.
- No cross-group delivery or self-echo reached the isolated client during a one-second observation.
- Nonmembers, missing/invalid credentials, invalid latitude/longitude, and missing latitude were rejected. The denied nonmember had no stored position.
- `go test ./gateway/... ./services/location/...` passed compilation; these packages have no test files.

Not covered: device GPS/map UI, load or long-duration reliability, direct gRPC negative cases and group-switch rejection, malformed JSON, and dependency-failure recovery. A bounded absence check does not establish isolation for all future execution.

Final run identifier: `1794509592756357000`. Redis test positions were removed. Synthetic accounts and groups remain because the current APIs expose no deletion. An earlier successful core run also created three accounts and two groups (`1794509434401654000`).
