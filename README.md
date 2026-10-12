# Rally

Rally is live location sharing for a trip group, in the same spirit as Google Maps’ “share my location,” but only for people you invited.

You sign up, create a group (a weekend trip, a family drive, a hike), and send the invite code. Everyone who joins streams their GPS. Each person sees the others moving on a map in near real time. That is the product: dots on a map, for your group, while you are out together.

No chat, ETAs, or geofences yet. Those can come later. The point is that when Alice walks down a street, Bob’s map updates without anyone refreshing.

The mobile app and the HTTP gateway in front of it are not built yet. The backend that would feed that map is: accounts, groups, and a live location stream.

---

## How it is built

Three Go services talk to each other over gRPC. A client (today [grpcurl](https://github.com/fullstorydev/grpcurl) or [grpcui](https://github.com/fullstorydev/grpcui); later a phone) calls them with a JWT from login.

- **Auth** — sign up, log in, issue a JWT, look up users by id
- **Group** — create a group, join with an invite code, list members (names come from Auth)
- **Location** — you send your coordinates on a bidirectional stream; other members of that group get those updates. Latest point is stored in Redis. Redis pub/sub fans each update out to everyone else who is connected.

Authorization stays with the service that owns it. Location asks Group “is this user in this group?” before it accepts a stream. Group asks Auth for names when it lists members. Group only stores user ids, not emails.

```
  client (grpcurl / grpcui today, app later)
        │
   ┌────┼────────────┐
   ▼    ▼            ▼
 Auth  Group     Location
   │     │            │
   │     └──► Auth    └──► Group
   ▼     ▼            ▼
Postgres Postgres   Redis
```

Each service is `handler` → `service` → `repository`. Calls to another service go through `internal/clients`.

The longer build plan, including the gateway and React Native map, is in [docs/roadmap.md](docs/roadmap.md).

## Run it

You need Docker.

```bash
make up
```

That builds the images, runs Auth and Group migrations, and starts Postgres, Redis, and the three services.

```bash
make logs          # all services
make down          # stop everything
```

| Service  | Port  |
|----------|-------|
| Auth     | 50051 |
| Group    | 50052 |
| Location | 50053 |

gRPC reflection is on, so you do not need extra proto flags:

```bash
grpcurl -plaintext localhost:50051 list
grpcui -plaintext localhost:50051
```

Group and Location expect `authorization: Bearer <token>` from Auth `Login`. Open two Location streams for the same group (two users) and send coordinates from each — the other stream should print the update.

## Repo layout

```
proto/          service contracts
gen/            generated Go from those protos
services/       auth, group, location
shared/         JWT helpers and shared errors
infra/          Dockerfiles and Compose
docs/roadmap.md how the project is being built
```
