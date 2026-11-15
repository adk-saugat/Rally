// Native Maps keys are embedded in binaries. Restrict them by app identity and SDK.
module.exports = ({ config }) => {
  const androidKey = process.env.GOOGLE_MAPS_ANDROID_API_KEY?.trim();
  const iosKey = process.env.GOOGLE_MAPS_IOS_API_KEY?.trim();
  const appID = process.env.RALLY_APP_ID?.trim() || 'com.rally.app';
  return {
    ...config,
    ios: { ...config.ios, bundleIdentifier: appID },
    android: { ...config.android, package: appID },
    plugins: [
      ...(config.plugins || []),
      'expo-secure-store',
      [
        'expo-location',
        {
          locationWhenInUsePermission:
            'Rally uses your location to center your map. Sharing with your group starts only when you choose Start sharing.',
          locationAlwaysAndWhenInUsePermission: false,
          locationAlwaysPermission: false,
          motionUsagePermission: false,
          locationFullAccuracyPermission: false,
          isIosBackgroundLocationEnabled: false,
          isAndroidBackgroundLocationEnabled: false,
          isAndroidForegroundServiceEnabled: false,
        },
      ],
      [
        'react-native-maps',
        {
          ...(androidKey ? { androidGoogleMapsApiKey: androidKey } : {}),
          ...(iosKey ? { iosGoogleMapsApiKey: iosKey } : {}),
        },
      ],
    ],
    extra: {
      ...config.extra,
      mapsConfigured: { android: Boolean(androidKey), ios: Boolean(iosKey) },
    },
  };
};
