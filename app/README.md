# Rally mobile app

Expo + React Native app with a welcome screen, signup/login, persistent sessions,
and a map main screen (Apple Maps in iPhone Expo Go; Google Maps in configured native builds). Group controls open as sheets over the map.
Location permission centers your private map; sharing requires an explicit action.

Start with [SETUP.md](SETUP.md) for Google Cloud keys, environment settings, native
build commands, and the two-device test flow. [ROADMAP.md](ROADMAP.md) tracks the
remaining work and acceptance checks.

## Development

```sh
npm ci
npm start -- --clear
npm run typecheck
npm run format:check
```

Use Node.js 22 or newer. Native Google Maps on iOS needs your own build;
iPhone Expo Go uses Apple Maps with the same location dot and group controls. Native build instructions and local
backend addresses are in [SETUP.md](SETUP.md).

## Source layout

- `src/navigation/`: typed navigation; signed-in sessions open Map.
- `src/session/`: SecureStore token persistence, session restore, expiry, and logout.
- `src/screens/`: welcome, authentication, and the map workspace.
- `src/components/`: map, group sheets, fields, and shared controls.
- `src/hooks/`: foreground GPS, group actions, and location stream lifecycle.
- `src/services/`: HTTP APIs and authenticated native WebSocket connection.
- `src/storage/`: account-scoped groups remembered on this device.
- `src/utils/`: session decoding and location-message validation.
- `src/styles/`: shared styles.
- `app.config.js`: native SDK keys, identifiers, and foreground-only permissions.

The gateway currently lacks end-trip deletion, independent view-only subscriptions,
location snapshots, and stop-sharing events. These limits and their visible
behavior are documented in [SETUP.md](SETUP.md#current-boundaries).
