# Rally — Learning Roadmap

### Go + gRPC + Protobuf Microservices

---

## The Core Idea

You're building **Rally** — a simplified "Google Maps location sharing for a trip group":

- Users sign up, create/join a group (like a trip or family circle)
- Each member's phone streams its GPS coordinates
- Everyone in the group sees everyone else moving live on a map

That's it. Everything else (geofencing, ETAs, chat) is a stretch goal.

---

## Why "simpler" matters here

The temptation is to design 6 microservices, Kubernetes, Kafka, and a mobile app on day one. Don't. You'll burn out before you see a dot move on a map. Instead:

1. Build services **one at a time**, each runnable and testable standalone before you wire it to the next.
2. Use **Docker Compose**, not Kubernetes, until everything works.
3. Use **React Native + `react-native-maps` (Google Maps provider)** for the app — one codebase, real native map component, no WebView/JS-bridge complexity to fight with on top of everything else.
4. Only split into microservices where it actually teaches you something (auth vs. group vs. location are naturally separate domains — that's your real boundary, not an arbitrary one).

---

## Final Architecture (what you're building toward)

```
                     ┌─────────────┐
                     │  Mobile App  │  (React Native, react-native-maps + WebSocket)
                     └──────┬──────┘
                            │ WS / HTTP
                     ┌──────▼──────┐
                     │   Gateway    │  (HTTP+WS ⇄ gRPC bridge)
                     └──┬───┬───┬──┘
              ┌─────────┘   │   └─────────┐
        ┌─────▼────┐  ┌─────▼─────┐ ┌─────▼──────┐
        │   Auth    │◄─┤   Group    │ │  Location   │
        │  Service  │  │  Service   │ │   Service   │
        └─────┬────┘  └──┬──────┬──┘ └─────┬──────┘
              │           │      ▲          │
              │           │      └──────────┘
         Postgres       Postgres      Postgres + Redis
        (users)        (groups)     (live loc + history)
```

All calls (Gateway→service and service→service) use gRPC + protobuf. Only the Gateway speaks HTTP/WebSocket to the mobile app, because React Native doesn't have a practical raw gRPC-streaming client (gRPC-Web exists but targets browsers, not React Native directly) — WebSocket is the pragmatic bridge either way. The two dashed-style arrows above are the interesting ones:

- **Group → Auth**: when listing members, Group Service asks Auth Service for each user's email/first name (Group Service only stores user _IDs_, not profile data)
- **Location → Group**: before accepting a user's location stream, Location Service asks Group Service "is this user actually in this group?" — authorization lives in the service that owns that knowledge, not duplicated everywhere

---

## Phase 0 — gRPC/Protobuf Fundamentals (2–4 days)

Don't skip this even though you know basic HTTP in Go — gRPC's mental model (services, streaming, code-gen) is different enough to deserve isolated practice.

**Do:**

- Install `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`.
- Write one trivial `.proto` (e.g. `Greeter` service, unary RPC), generate code, run client + server locally.
- Then extend it to a **server-streaming** RPC (server sends multiple responses) — this is the pattern you'll use for location broadcasts.
- Then try a **bidirectional-streaming** RPC — this is what a client will use to continuously push its own location while receiving others'.

**Learn:** unary vs server-streaming vs client-streaming vs bidi-streaming, gRPC status codes, `context.Context` deadlines, interceptors (middleware equivalent).

**Milestone:** You can explain, in your own words, why location updates are a good fit for streaming RPCs instead of polling.

---

## Phase 1 — Repo Structure & Tooling (1 day)

Set up a monorepo so all `.proto` files are shared and versioned together:

```
rally/
├── proto/
│   ├── auth/v1/auth.proto
│   ├── group/v1/group.proto
│   └── location/v1/location.proto
├── gen/                  # generated Go code lands here
├── services/
│   ├── auth/
│   ├── group/
│   └── location/
├── gateway/
├── app/                  # React Native app
├── docker-compose.yml
└── Makefile
```

Write a `Makefile` target (e.g. `make gen`) that runs `protoc` across all `.proto` files in one command, something like:

```makefile
gen:
	protoc \
	  --go_out=gen --go_opt=paths=source_relative \
	  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
	  proto/auth/v1/auth.proto \
	  proto/group/v1/group.proto \
	  proto/location/v1/location.proto
```

This keeps the actual `protoc` invocation visible and in your control — you'll understand exactly what each flag does, which pays off later when you're debugging import paths or package naming across services.

**Milestone:** `make gen` regenerates Go code for all three services from one command.

### Inside each service folder

Each of `services/auth/`, `services/group/`, `services/location/` should follow the same internal layout — a standard, idiomatic Go service shape. Doing this consistently across all three is itself the point: once you've built one, the others are the same skeleton with different logic inside.

```
services/auth/
├── cmd/
│   └── server/
│       └── main.go          # wires everything together, starts the gRPC server
├── internal/
│   ├── handler/              # gRPC handler layer — implements the generated
│   │   └── auth_handler.go   # <Service>Server interface, no business logic here
│   ├── service/               # business logic — validates input, orchestrates,
│   │   └── auth_service.go   # calls repository + clients, has no gRPC types in it
│   ├── repository/            # data access only — SQL queries, Redis calls
│   │   └── user_repository.go
│   ├── clients/                # gRPC clients for calling OTHER services
│   │   └── auth_client.go     # (empty for Auth itself; Group Service's clients/
│   │                          #  folder would hold its Auth Service client)
│   └── config/
│       └── config.go          # env var loading (DB URL, port, other service addrs)
├── migrations/                  # Goose SQL migrations (versioned schema changes)
│   └── 00001_create_users.sql
├── go.mod
└── Dockerfile
```

**Why split `handler` from `service` from `repository`** (this is the one idea worth internalizing, not just copying):

- **`handler`** only knows about protobuf types (`*authv1.LoginRequest`) and gRPC concerns (status codes, streaming). It converts those into plain Go values and calls `service`.
- **`service`** only knows plain Go types and business rules ("password must be hashed before storage", "reject login after 5 failed attempts"). It has zero knowledge of gRPC or SQL — you could unit-test it without a network or database.
- **`repository`** only knows how to read/write data. It doesn't know _why_ — no business rules live here, just queries.

This means a request flows one direction: `handler → service → repository`, and each layer is independently testable. It also directly explains where the Phase 3/4 inter-service calls live: Group Service's call to `AuthService.GetUsers` lives in its `internal/clients/auth_client.go`, and gets _called from_ `internal/service/group_service.go` — the service layer decides it needs that data, the clients layer knows how to fetch it.

The **Gateway** (`gateway/`) follows a lighter version of this same shape — it doesn't need a `repository` layer since it holds no data of its own, just `handler`(HTTP/WS)→`clients`(gRPC calls to the three services).

---

## Phase 2 — Auth Service (3–5 days)

The smallest full vertical slice: proto → server → storage → client test.

- `.proto`: `Signup(email, password, firstName) → UserID`, `Login(email, password) → JWT`
- Store users in Postgres (just one `users` table — id, email, password_hash, first_name)
- Manage that schema with **[Goose](https://github.com/pressly/goose)** (`github.com/pressly/goose/v3`) — versioned SQL migrations, not `CREATE TABLE` scattered in Go code
  - Put files in `services/auth/migrations/` (e.g. `00001_create_users.sql`) with `-- +goose Up` / `-- +goose Down`
  - Apply with Makefile targets (`make migrate-up`, `make migrate-down`, `make migrate-status`) against `DATABASE_URL`
  - The repository only runs queries; Goose owns schema changes
- `firstName` is required at signup — that's the official name shown on the map and member list
- Hash passwords with bcrypt
- Issue JWTs (use `github.com/golang-jwt/jwt`)
- Test with [`grpcurl`](https://github.com/fullstorydev/grpcurl) — no client code needed yet

**Milestone:** `grpcurl` can sign up a user and log in to get a JWT, talking only to this one service in isolation.

---

## Phase 3 — Group Service (3–5 days)

- `.proto`: `CreateGroup(name) → GroupID+InviteCode`, `JoinGroup(inviteCode) → OK`, `ListMembers(groupID) → []Member`
- Postgres tables: `groups`, `group_members` (just store `user_id`, not name/email — that data belongs to Auth Service) — add them as Goose migrations in `services/group/migrations/`, same pattern as Auth
- Add a **gRPC interceptor** that validates the JWT from Phase 2 on every call (your first taste of cross-service auth) — the interceptor just verifies the token signature locally; it doesn't need to call the Auth service over the network for this part.
- Add `.proto` to Auth Service: `GetUser(userID) → {email, firstName}` (a small unary RPC you probably don't have yet)
- **In `ListMembers`**, after fetching `user_id`s from Postgres, Group Service opens a **gRPC client connection to Auth Service** and calls `GetUser` for each member (or better: add a `GetUsers(userIDs) → []User` batch RPC to Auth Service and call it once — avoids an N+1 problem, and is a good early lesson in RPC design).
- This means Group Service now needs Auth Service's address at startup — pass it via an env var (`AUTH_SERVICE_ADDR=auth-service:50051`), which is also exactly how it'll work in Docker Compose later (services find each other by container name).

**Milestone:** `ListMembers` returns not just user IDs but real emails/first names — proof that Group Service successfully called Auth Service under the hood. Kill the Auth Service process and re-run `ListMembers` — watch it fail with a clear gRPC `Unavailable` error, and think about what a real system should do here (see "Handling Service-to-Service Failure" below).

---

## Phase 4 — Location Service (this is the heart of the project, 1–2 weeks)

- `.proto`: a **bidirectional streaming RPC** `StreamLocation(stream LocationUpdate) → stream LocationUpdate`
  - Client sends its own `{lat, lng, timestamp}` repeatedly
  - Server pushes back other group members' latest locations
- Architecture inside this service:
  - Write **every incoming update** to Redis (`SET group:<id>:user:<id> "lat,lng,ts"`) — this is your "current live position" store, fast and ephemeral
  - Periodically (or on every N updates) persist to Postgres (ideally with **PostGIS** extension) for history/playback later — schema via Goose in `services/location/migrations/`
  - Use Redis **Pub/Sub** (or just an in-memory fan-out map of channels per group) so that when User A's update arrives, it gets pushed to all other connected streams for that group

This is where you'll learn the most: managing many concurrent long-lived streams per group, goroutines per connection, and clean shutdown/cancellation via `context`.

**Add the second inter-service call here.** Add `IsMember(groupID, userID) → bool` to Group Service's `.proto`. When a client opens a `StreamLocation` connection to Location Service, the very first thing the handler does — before accepting a single update — is call `GroupService.IsMember(...)` over gRPC. If false, close the stream with a `PermissionDenied` status. This is real authorization logic living in the correct service, called at the correct moment (connection time, not per-message — no need to re-check on every location update).

**Milestone:** Two `grpcurl`-style test clients (or a tiny Go test client you write) in the same group see each other's coordinates update live within ~1 second. Then try opening a stream for a `groupID` the test user never joined — confirm Location Service rejects it by calling out to Group Service, not by checking its own local data.

---

## Interlude — Inter-Service Communication Mechanics

Once Group Service is calling Auth Service and Location Service is calling Group Service, you're doing "real" microservices. A few things worth learning deliberately here rather than by accident:

- **Each service is both a gRPC server and, now, a gRPC client of another service.** Keep a small `internal/clients/` package per service that owns the `grpc.NewClient(addr, ...)` connection setup — don't scatter dial calls around your handlers.
- **Service discovery, the simple way**: in Docker Compose, services resolve each other by container name (`auth-service:50051`) — no need for Consul/etcd yet. Pass addresses via env vars, never hardcode.
- **Propagate deadlines.** When Location Service calls Group Service inside a handler, derive the outgoing call's context from the incoming request's context (`ctx` you were already given), so if the original caller times out, the downstream call is cancelled too instead of running forever.
- **Handle downstream failure explicitly.** What should `ListMembers` do if Auth Service is down — fail the whole request, or return member IDs with a placeholder name? What should `StreamLocation` do if Group Service is briefly unreachable — reject the connection, or fail open? There's no universally right answer; deciding and justifying it is the actual skill.
- **Stretch, once the happy path works**: add a client-side timeout + one retry with backoff on the Group→Auth call (`google.golang.org/grpc` supports this via interceptors or the newer retry policy config). This is a small, contained way to practice resilience patterns without needing a service mesh.

---

## Phase 5 — Gateway (3–5 days)

React Native doesn't speak raw gRPC streams well, so bridge it:

- Simplest option: a plain Go HTTP server with a `/ws` WebSocket endpoint per group. It holds a gRPC bidi-stream connection to the Location Service on the backend, and relays messages between the app's WebSocket and that gRPC stream.
- Alternative (more "proper" but more setup): [`grpc-gateway`](https://github.com/grpc-ecosystem/grpc-gateway) for the unary Auth/Group calls (REST/JSON), plus your own WebSocket bridge just for the streaming Location part.
- This Gateway is also where you validate the JWT from the app before opening the stream.

**Milestone:** Using a WebSocket testing tool (e.g. `websocat`, or a quick throwaway script) or your browser's dev console as a stand-in client, open a WebSocket, send `{lat, lng}`, and receive other members' updates as JSON — before you've even built the app UI, you know the Gateway↔Location pipe works.

---

## Phase 6 — Native Mobile App with Google Maps (4–8 days)

**Prerequisite — Google Cloud setup (30–60 min, do this first):**

- Create a Google Cloud project, enable "Maps SDK for Android" and/or "Maps SDK for iOS", add a billing account (required even though map display is free at this scale — no charge unless you add Places/Geocoding/Directions later or go far beyond hobby-project volume).
- Generate an API key, then **restrict it** to your app's package name/bundle ID — don't ship an unrestricted key.
- Keep the key out of source control (`.env`, not committed) — this is also good practice to carry into every future project.

**Build the app:**

- `npx react-native init` (or use Expo if you'd rather avoid native build tooling initially — note Expo's managed workflow has some constraints around `react-native-maps` config, so check current Expo docs if you go that route)
- Add `react-native-maps`, configure the Google provider (`PROVIDER_GOOGLE`) and API key per platform
- On load: connect WebSocket to the Gateway, request location permission, use a geolocation library (e.g. `react-native-geolocation-service` or Expo's `Location.watchPositionAsync`) to stream your own coordinates
- On receiving updates from the WebSocket: update/add a `<Marker>` per user ID on the `<MapView>`
- Basic UI: group invite code input, member list, the map itself with colored markers

**Milestone:** Run the app on two devices (or one device + one emulator) in the same group, walk around (or mock GPS in the emulator), watch both markers update live on both.

---

## Phase 7 — Wire It All Together (2–3 days)

- `docker-compose.yml` with: postgres, redis, auth-service, group-service, location-service, gateway
- One `docker-compose up` should bring up the whole stack
- Write a short README with the run instructions — treat this like a real project

**Milestone:** Fresh clone → `docker-compose up` → working demo, no manual steps.

---

## Phase 8 — Stretch Goals (optional, pick what excites you)

Roughly in order of learning value vs. effort:

| Feature                                                                              | What it teaches                                                   |
| ------------------------------------------------------------------------------------ | ----------------------------------------------------------------- |
| Location history playback (route replay)                                             | PostGIS queries, time-series thinking                             |
| Geofencing ("notify when User X arrives")                                            | Background workers, event-driven design                           |
| **Async messaging (NATS or Kafka)** — see below                                      | Decoupling, pub/sub vs request/response                           |
| Add Prometheus + Grafana                                                             | Observability, gRPC interceptor metrics                           |
| Add OpenTelemetry tracing across services                                            | Distributed tracing, understanding latency across your 3 services |
| Move Compose → Kubernetes (k3d/minikube)                                             | Real orchestration, service discovery                             |
| Add a proper API gateway (Envoy/Kong) instead of hand-rolled one                     | How production gRPC-web setups actually work                      |
| Real push notifications (via Firebase Cloud Messaging) instead of in-app-only alerts | Background delivery, device tokens, platform-specific setup       |

Don't do all of these — pick 1–2 that interest you most after Phase 7 is working.

### Later: Introducing Async Messaging

Once the synchronous calls (Group→Auth, Location→Group) feel natural, this is the natural next step — and it should _replace_ something you already built, not just bolt on.

**Best candidate in this system:** `JoinGroup`. Right now Group Service handles a join entirely by itself. Once you add async, Group Service instead **publishes a `UserJoinedGroup` event** (via NATS or Kafka) when someone joins, and:

- A new **Notification Service** subscribes to it and, say, logs "Alice joined Weekend Trip" (stand-in for a push notification)
- Location Service could subscribe too, to pre-warm a Redis entry for that user

This is a good first async use case specifically because it's **fire-and-forget** — nothing needs to block on it, unlike your `IsMember` authorization check, which genuinely needs a synchronous answer before Location Service can decide whether to accept a stream. That contrast (this call must be sync vs. this one is fine as an event) is the actual lesson — don't convert everything to async just because you learned it.

Start with **NATS** over Kafka: it's a single small binary, trivial to run in Compose, and the client library is simple. Kafka is worth learning eventually but adds real operational weight (partitions, consumer groups, Zookeeper/KRaft) that isn't worth it for a project this size.

When you're ready to actually build this part, come back and we can design the event schema and wire it in.

---

## Suggested Pacing

- **Weeks 1**: Phase 0–1 (fundamentals + repo setup)
- **Weeks 2–3**: Phase 2–3 (Auth + Group services)
- **Weeks 4–5**: Phase 4 (Location service — give this the most time)
- **Week 6**: Phase 5 (Gateway)
- **Week 7**: Phase 6–7 (App + Compose)
- **Week 8+**: Stretch goals, at your own pace

~8 weeks part-time to a working end-to-end demo is realistic if you're learning gRPC as you go.

---

## Key Libraries/Tools Reference

- `google.golang.org/grpc` — core gRPC
- `google.golang.org/protobuf` — protobuf runtime
- `protoc` + `protoc-gen-go` + `protoc-gen-go-grpc` — proto code generation
- `github.com/fullstorydev/grpcurl` — CLI testing of gRPC services without writing a client
- `github.com/golang-jwt/jwt/v5` — JWT issue/verify
- `github.com/jackc/pgx/v5` — Postgres driver
- `github.com/pressly/goose/v3` — Postgres schema migrations (SQL files + `make migrate-up`)
- `github.com/redis/go-redis/v9` — Redis client
- `github.com/gorilla/websocket` — WebSocket for the Gateway
- PostGIS — Postgres extension for geo queries (later phases)
- `react-native-maps` — native map component (Google Maps provider)
- Google Cloud account + Maps SDK API key (Android/iOS) — free for map display; billing details required but no charge expected at hobby scale

---

## One Piece of Advice

Resist splitting into microservices _before_ Phase 2's single service is fully working end to end (proto → server → DB → grpcurl test passing). Get one vertical slice completely solid first — it teaches you the whole toolchain in a low-stakes way, and every service after that is just "do that again, with different fields."
