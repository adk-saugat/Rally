---
name: test-rally-location
description: Verify Rally's newly implemented location sharing using mock users, groups, and coordinate updates against the real local implementation. Test gRPC and WebSocket delivery and report whether sharing works, fails, or is blocked.
---

# Test Rally location sharing

Determine whether live location sharing works using observable delivery and persistence. Use synthetic coordinates and disposable local test accounts. This verifies the backend; it does not prove device GPS acquisition or mobile map rendering works.

## Mock data and implementation under test

Interpret mock data as fabricated input, not fabricated success responses. Exercise the actual location service, broker, persistence, and gateway locally. Do not replace the location implementation with mocks for the end-to-end verdict. Isolated unit tests with fake dependencies may supplement that verdict, but must be reported separately.

Inspect the working-tree diff and recent relevant changes to identify the newly implemented behavior and ensure those sources are built and exercised. Do not rely on containers built before the changes. Record the tested revision and whether uncommitted changes were included.

Create a unique run identifier and use these synthetic fixtures:

| Fixture | Purpose |
| --- | --- |
| Mock Alice and Mock Bob | Two authenticated members of mock trip G |
| Mock Charlie | Authenticated member of mock trip H, excluded from G |
| Alice points | `(41.8810, -87.6230)`, `(41.8815, -87.6225)`, `(41.8820, -87.6220)` |
| Bob points | `(41.8800, -87.6240)`, `(41.8805, -87.6235)`, `(41.8810, -87.6230)` |

Use distinct current millisecond timestamps for each update. Send each user's route in order after stream readiness, and assert the other member receives every measured point with its exact sender, group, coordinates, and timestamp. Verify the final point is persisted. These points simulate movement without requesting device location or using real people's location history. Obtain fixture IDs from real API responses rather than assuming literal names are valid IDs.

The primary acceptance check is: Alice's movement reaches Bob and Bob's movement reaches Alice through the newly implemented sharing path. Include a mixed WebSocket/gRPC pair when testing gateway-to-service interoperability. Failure of this check means sharing is not working for that path, even if connections open or builds succeed.

## Establish the target

- Read repository instructions and inspect the current location proto, handlers, service, repositories, gateway routes/config, and Compose file. Source code takes precedence over stale README examples.
- Default to a local development stack. Confirm the target before writing test data to a shared or remote environment. Never print passwords or bearer tokens in reports.
- Check available tools and running services before starting anything. The current Compose stack runs Auth, Group, Location, Redis, and Postgres; it does not start the gateway. Discover gateway startup and its port from the current code.
- Use existing disposable fixtures when available; otherwise create uniquely named test users and groups through the normal Auth and Group APIs. Obtain tokens through Login rather than minting bypass tokens. Prepare two distinct members A and B in group G, plus a user C who is not a member of G and belongs to a separate group H.
- Reuse existing test tooling when practical. If a temporary client is needed, use an isolated temporary directory and the generated Go clients / existing WebSocket dependency. Keep application changes and permanent test harness additions outside this diagnostic task unless requested.

## Verify the stream

Current interfaces:

- gRPC: `location.v1.LocationService/StreamLocation`, normally on `localhost:50053`, with `authorization: Bearer <token>` metadata. Read `proto/location/v1/location.proto` for fields.
- WebSocket: `/groups/:group_id/ws`, using an `Authorization: Bearer <token>` header. Send JSON containing `latitude`, `longitude`, and optionally `timestamp_unix_ms`.
- Redis latest position: `group:<group_id>:user:<user_id>`, containing JSON coordinates and timestamp. Read only the exact fixture keys.

Each stream must send its first coordinate update before membership checking and subscription occur. A successful connection alone does not prove authorization or delivery. Updates are sent to other users; the sender receives no self-echo. The service currently offers no initial snapshot to a new subscriber.

Use bounded connection/read deadlines (for example, 5 seconds locally) and unique timestamps or coordinates to correlate messages. Avoid startup races: initialize B, initialize A, establish readiness with a bounded exchange, then start measured assertions. Do not treat an update sent before subscription as a delivery failure. Drain initialization messages before checking negative cases.

For both gRPC and the gateway where available:

1. **Member delivery:** Open A and B streams in G, initialize both, then send distinguishable updates A→B and B→A. Verify group, authenticated sender identity, coordinates, and explicit timestamp in each received update.
2. **Latest position:** Send a second point and inspect the exact Redis key until the deadline. Verify the latest coordinates and timestamp replaced the previous point. Check that omitting or zeroing the timestamp produces a plausible server timestamp.
3. **Self-echo and isolation:** After initialization is drained, send a marked A update. B must receive it; A must not receive that marked update. Initialize C in H and confirm no G update reaches C during the bounded observation period. Report absence checks with their observation duration.
4. **Authentication and membership:** Test missing/invalid credentials, then C attempting G with valid credentials. Send the first update when needed to trigger authorization. Verify rejection and no position key creation/update for the denied fixture. Distinguish HTTP handshake failures, WebSocket close codes, and gRPC status codes.
5. **Coordinates:** Use fresh streams for invalid cases. Check latitude beyond ±90, longitude beyond ±180, and valid zero/boundary values. Through the gateway, also test missing coordinate fields and malformed JSON. Invalid gateway payloads currently close with code 1007. Do not assert missing proto scalar coordinates are rejected: omitted doubles decode as zero in proto3.
6. **Identity and stream scope:** For gRPC, send a forged `user_id` and verify the server uses the authenticated identity. Attempt changing `group_id` after stream initialization and verify rejection. Through the gateway, verify received identity and group come from authentication and the route rather than extra client JSON fields.
7. **Lifecycle:** Close a client cleanly, reconnect, reinitialize, and verify new updates flow without stale delivery or a hung connection. Close every test stream and client when finished.

Where useful, run existing focused tests for the gateway and location packages. Report packages with no test files accurately; compilation or unit-test success alone is not evidence of end-to-end delivery.

## Diagnose and report

Report **PASS**, **FAIL**, or **BLOCKED** per check, with the target, protocol, observed evidence, and any skipped coverage. A timeout is a failure only when prerequisites and subscription readiness were established; otherwise identify the blocker. Treat unavailable dependencies separately from a demonstrated application defect.

On failure, inspect bounded relevant service logs and trace gateway → Location → Group/Auth and Redis. Record a reproducible case and likely failing layer. Do not change production code or weaken authentication to make a check pass. If asked to implement a fix, preserve the failing reproduction and verify the fix against it.

Finish with a clear verdict: working for the tested scope, failing with the observed reason, or inconclusive due to blockers. Distinguish gRPC coverage from WebSocket coverage and backend behavior from device GPS behavior. Clean up only fixtures/resources created by this run using supported APIs; list leftovers when deletion is unsupported. Never flush Redis, delete volumes, or stop pre-existing services as cleanup.
