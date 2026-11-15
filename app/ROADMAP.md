# Rally mobile roadmap

## Product flow

Welcome → Signup/Login → personal map → optional active trip.

The map displays local GPS after foreground permission. Create/join actions,
trip members, invitations, and profile controls open as sheets over the map.
Permission alone does not transmit a position; Start sharing is explicit.

Groups are intended to be temporary trip rooms. Permanent group history and a
synced list of past groups are not product requirements.

## Implemented

- [x] Welcome, login, signup, and authenticated map navigation.
- [x] Secure token storage, session restoration, expiry, and logout.
- [x] Shared accessible controls, safe areas, and keyboard-aware forms/sheets.
- [x] Apple Maps in iPhone Expo Go; Google Maps in configured native builds.
- [x] Foreground GPS, native blue location dot, street-level zoom, and recentering.
- [x] Create/join APIs, invite sharing, member list, and group selection.
- [x] Explicit Start/Stop sharing, member markers, timestamps, and reconnect attempts.
- [x] Stop transmission on group changes, logout, and leaving the foreground.
- [x] TypeScript and formatting checks.

See [SETUP.md](SETUP.md) for setup commands, key restrictions, phone networking,
and the current backend limitations. Typechecking and formatting checks are not a substitute for
physical-device GPS and release-build validation.

## 1. Finish temporary trip lifecycle

- [ ] Replace the current device-local group history with active-trip recovery only.
- [ ] Record the trip creator and add an authorized End trip API.
- [ ] On End trip, invalidate the invite and prevent new joins or location updates.
- [ ] Close member streams, remove Redis positions and trip records, and notify clients.
- [ ] Clear active-trip state on all clients and return to the personal map.
- [ ] Test end-trip races with reconnecting members and in-flight updates.

The current backend does not support ending/deleting trips. Cached groups and
last-seen positions must not be described as already being deleted automatically.

## 2. Improve subscription and reconnect behavior

- [ ] Add an authorized view-only subscription that needs no GPS publication.
- [ ] Add a current-position snapshot for newly connected/reconnected members.
- [ ] Add stop-sharing/presence events so peers can distinguish stopped and stale users.
- [ ] Verify native WebSocket authentication and expired-session recovery on real devices.

Currently the first published position starts the subscription. Received positions
are last-seen data; the stream does not replay stored Redis positions.

## 3. Complete device acceptance

- [ ] Verify two phones in one trip and a third user outside it for isolation.
- [ ] Test walking, stationary GPS, approximate position, denial/revocation, and GPS off.
- [ ] Test offline/reconnect, token expiry, Stop sharing, suspension, and logout.
- [ ] Test smaller screens, enlarged text, keyboard behavior, and accessibility.
- [ ] Measure battery use, GPS sampling, and update latency; tune before background support.
- [ ] Verify Android and iOS release keys and signing restrictions in distributed builds.

The iOS simulator has rendered both map providers and the location dot using
simulated coordinates. Real physical-device coverage remains separate.

## 4. Release and later features

- [ ] Configure production HTTPS/WSS endpoints, app identity, icons, and versioning.
- [ ] Complete location-use explanations and store privacy disclosures.
- [ ] Validate signed builds through TestFlight and Android internal testing.
- [ ] Scope background sharing only after foreground sharing is reliable.

Defer chat, ETAs, directions, place search, geofencing, and location history.
Each requires its own product decision and any additional API/cost review.
