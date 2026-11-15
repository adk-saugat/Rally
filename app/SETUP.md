# Run Rally with Google Maps

The app now opens its main map after login. These steps configure your Google
project and build the native app. No API keys have been created or added for you.

## 1. Set up Google Cloud

1. Create/select a project in [Google Cloud Console](https://console.cloud.google.com/).
2. Enable billing and review the current Maps SDK pricing. Configure billing alerts;
   alerts do not impose a hard spending cap.
3. Enable **Maps SDK for Android** and **Maps SDK for iOS**.
4. Choose the stable app identifier you want to use. The local default is
   `com.rally.app` for both platforms; change `RALLY_APP_ID` before distribution.
5. Create separate Android and iOS keys:
   - Android: restrict to your package identifier and signing certificate SHA-1;
     restrict the API to Maps SDK for Android. Use the debug certificate for local
     builds and the actual release/Play app-signing certificate for distribution.
   - iOS: restrict to your bundle identifier and Maps SDK for iOS.

Follow Google's [Android key setup](https://developers.google.com/maps/documentation/android-sdk/get-api-key),
[iOS key setup](https://developers.google.com/maps/documentation/ios-sdk/get-api-key),
and [API security guidance](https://developers.google.com/maps/api-security-best-practices).
For local Android debug builds, obtain the fingerprint from the generated Android
project with `./gradlew signingReport` inside `app/android/` after prebuild. The
fingerprint must match the certificate that actually signed the installed app.

## 2. Configure the app

From `app/`, copy `.env.example` to `.env` **if you do not already have a `.env`**.
Keep any existing values you need, and fill in:

```dotenv
EXPO_PUBLIC_API_URL=http://localhost:8080
GOOGLE_MAPS_ANDROID_API_KEY=your_android_sdk_key
GOOGLE_MAPS_IOS_API_KEY=your_ios_sdk_key
RALLY_APP_ID=com.rally.app
```

- iOS simulator: the gateway can use `http://localhost:8080`.
- Android emulator: use `http://10.0.2.2:8080`.
- Physical phone: use your computer's LAN IP (for example `http://192.168.1.42:8080`),
  keep both devices on the same network, and allow port 8080 through the firewall.
- Production: use HTTPS; the app derives WSS from an HTTPS API URL.

`app.config.js` injects the Maps keys into native configuration. They are embedded
in the native binary, so restrict them even though `.env` is ignored by Git.
Keep JWT signing secrets and database credentials out of the mobile app.
Rebuild whenever a Maps key, identifier, permission description, or native plugin
changes. A JavaScript reload is insufficient for native configuration changes.

## 3. Start the backend and gateway

From the repository root:

```sh
make up
```

In another terminal at the repository root:

```sh
go run ./gateway/cmd/server
```

## 4. Build and run the native app

Use Node.js 22 or newer. Install Xcode for
iOS and Android Studio/SDK tooling for Android. From `app/`:

```sh
npm ci
npx expo run:ios
# Or, with an Android emulator/device available:
npx expo run:android
```

These commands generate the local native projects and build/install the app.
Do not run them simultaneously while generating native files. After the first
build, `npm start -- --clear` restarts Metro with a clean dependency cache.
If using a development client built with `expo-dev-client`, use
`npx expo start --dev-client --clear` instead.

**iPhone Expo Go works without a native build:** it uses Apple Maps, including the
location dot, zoom/recenter, group controls, and live member markers. Scan the
project QR code in Expo Go and keep your phone and Mac on the same Wi-Fi.
**Google Maps on iOS requires your own native build.** Configured native Rally
builds use Google Maps; iPhone Expo Go selects Apple Maps automatically.
Android Expo Go can exercise its bundled Google Maps integration, but your own
build is still necessary to validate your restricted key and signing configuration.
See [Expo's Google Maps integration](https://docs.expo.dev/versions/latest/sdk/map-view/).

If a map is blank, check the enabled SDK, billing, key restrictions, app identifier,
signing fingerprint, internet access, and whether you rebuilt after configuring
keys. Test the actual distributed app's signing identity before release.

## 5. Try the app

1. Sign up, then log in. Successful login opens the map and saves the JWT securely.
2. Tap **Use my location**, read the OS prompt, and grant foreground permission.
   The map centers on your GPS fix. Denying permission leaves the map usable.
3. In the iOS simulator, set a location under **Features → Location** if no GPS fix
   arrives. On a phone, enable device location services.
4. Tap **Create group** over the map and enter a name.
5. Open **Crew**, then **Share invite**. On another device, log in with a different
   account and use **Join group** with that code.
6. On each device, tap **Start sharing**. Each receives updates from the other
   group members while sharing is connected and Rally is in the foreground.
7. Try **Stop sharing**, switching groups, app suspension, logging out, and restarting.
   Sharing never automatically resumes after suspension or restart.

## Current boundaries

- Local GPS is used to display your position after permission is granted. Group
  transmission starts only after explicit Start sharing. Stop sharing closes the
  stream immediately; local GPS may still update your private map while it is open.
- Groups are intended to last only for a trip. The current implementation still
  caches groups per account on this device; it does not yet have an End trip API
  to delete trip data, invalidate invites, and disconnect everyone. Active-trip
  recovery and end-trip cleanup are the next lifecycle work, not permanent group history.
  Reinstalling or using another device does not restore the local cache. Joining
  an already-joined group can currently return “already a member”.
- The current backend requires a first published position to start a subscription.
  This app therefore receives member positions while sharing is on. Independent
  view-only subscriptions and initial stored-position snapshots need backend work.
- Stop sharing stops new transmissions. Peers may retain the last point, labelled
  with its age and marked stale after 30 seconds; no immediate stop event exists.
- Positions received before reconnect are last-seen data. There is no initial
  Redis snapshot replay; fresh updates arrive when members publish again.
- Sharing is foreground-only, with capped reconnect attempts and no offline backlog.
- JWT expiry returns to login. There is no refresh-token endpoint.

## Development checks

From `app/`:

```sh
npm run typecheck
npm run format:check
```

Before using live sharing, manually verify two physical devices, permission
denial/revocation, suspended apps, group isolation, logout, reconnect, and a
distributed native build. Typechecking does not validate Maps keys or real GPS.
