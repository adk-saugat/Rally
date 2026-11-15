# Rally

Rally is live location sharing for a trip group, in the same spirit as Google Maps’ “share my location,” but only for people you invited.

You sign up, create a group (a weekend trip, a family drive, a hike), and send the invite code. Everyone who joins streams their GPS. Each person sees the others moving on a map in near real time. That is the product: dots on a map, for your group, while you are out together.

No chat, ETAs, or geofences yet. Those can come later. The point is that when Alice walks down a street, Bob’s map updates without anyone refreshing.

The backend and HTTP/WebSocket gateway provide accounts, groups, and a live location stream. The mobile app has a Google Maps workspace after login, secure session persistence, foreground GPS, and group controls over the map. Native Maps keys and device setup are described in [app/SETUP.md](app/SETUP.md).

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

## Mobile app setup

The Expo TypeScript app lives in `app/`. Start the backend with `make up`, then
start its HTTP gateway in another terminal:

```bash
go run ./gateway/cmd/server
```

Install Node.js 22 LTS or newer, then start the app:

```bash
cd app
npm ci
npm start
```

Open it with a matching Expo Go version, or press `i` for an iOS simulator or
`a` for an Android emulator with the native tooling installed. Use **Sign up** or **Log in** to open the authentication forms.

The gateway URL defaults to `http://localhost:8080` on iOS and
`http://10.0.2.2:8080` on the Android emulator. For a physical phone, copy
`app/.env.example` to `app/.env` and set `EXPO_PUBLIC_API_URL` to your computer's
LAN address, for example `http://192.168.1.42:8080`. Keep both devices on the
same network and allow access to port 8080. Expo's tunnel does not tunnel the
gateway. Reload the app after changing the URL. Public Expo environment
variables are included in the app bundle; keep secrets out of them.

Run `npm run typecheck` inside `app/` to check TypeScript. See [app/README.md](app/README.md) for the source layout and current app capabilities.
See [app/SETUP.md](app/SETUP.md) to configure Google Maps and test location sharing.

## What belongs in Git

Real `.env` files, dependencies, Expo caches, and generated native iOS/Android
projects are ignored. Commit `.env.example`, app configuration, source files, and
package lockfiles so another checkout can reproduce the setup.

## Repo layout

```
proto/          service contracts
gen/            generated Go from those protos
services/       auth, group, location
shared/         JWT helpers and shared errors
infra/          Dockerfiles and Compose
docs/roadmap.md how the project is being built
```
