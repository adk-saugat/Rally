import { forwardRef } from 'react';
import { Platform, StyleSheet, Text, View } from 'react-native';
import Constants, { ExecutionEnvironment } from 'expo-constants';
import MapView, { Marker, PROVIDER_GOOGLE, type LatLng } from 'react-native-maps';
import type { Position } from '../hooks/useDeviceLocation';
import type { MemberLocation } from '../utils/locations';
import type { Member } from '../services/groups';

const expoGo = Constants.executionEnvironment === ExecutionEnvironment.StoreClient;
const nativeConfigured = Constants.expoConfig?.extra?.mapsConfigured?.[Platform.OS] === true;
// Expo Go includes Apple Maps on iOS; Google Maps remains the native-build provider.
const useAppleMaps = Platform.OS === 'ios' && expoGo;
const googleMapAvailable =
  Platform.OS !== 'web' && (expoGo ? Platform.OS === 'android' : nativeConfigured);

export function cameraRegion(point: LatLng) {
  return {
    latitude: point.latitude,
    longitude: point.longitude,
    latitudeDelta: 0.003,
    longitudeDelta: 0.003,
  };
}

type RallyMapProps = {
  locationEnabled: boolean;
  position: Position | null;
  points: MemberLocation[];
  members: Member[];
  now: number;
  onReady: () => void;
  bottomPadding: number;
};

const RallyMap = forwardRef<MapView, RallyMapProps>(function RallyMap(
  { locationEnabled, position, points, members, now, onReady, bottomPadding },
  ref,
) {
  if (!useAppleMaps && !googleMapAvailable) {
    return (
      <View style={styles.setup}>
        <Text style={styles.symbol}>◎</Text>
        <Text style={styles.setupTitle}>Your map is almost here.</Text>
        <Text style={styles.setupText}>
          Configure the Google Maps key for this platform and rebuild Rally. You can still create or
          join a group here.
        </Text>
      </View>
    );
  }
  return (
    <MapView
      ref={ref}
      style={styles.fill}
      provider={useAppleMaps ? undefined : PROVIDER_GOOGLE}
      initialRegion={
        position
          ? cameraRegion(position)
          : { latitude: 39.5, longitude: -98.35, latitudeDelta: 35, longitudeDelta: 35 }
      }
      mapPadding={{ top: 135, right: 16, bottom: bottomPadding, left: 16 }}
      onMapReady={onReady}
      showsUserLocation={locationEnabled}
      showsMyLocationButton={false}
      showsCompass={false}
      toolbarEnabled={false}
    >
      {points.map((point) => {
        const name =
          members.find((member) => member.user_id === point.user_id)?.first_name ?? 'Group member';
        const seconds = Math.max(0, Math.floor((now - point.timestamp_unix_ms) / 1000));
        return (
          <Marker
            key={point.user_id}
            coordinate={point}
            title={name}
            pinColor={seconds > 30 ? '#999F93' : '#BC7654'}
            description={`Updated ${seconds}s ago${seconds > 30 ? ' · May no longer be sharing' : ''}`}
          />
        );
      })}
    </MapView>
  );
});
export default RallyMap;

const styles = StyleSheet.create({
  fill: { flex: 1 },
  setup: {
    flex: 1,
    backgroundColor: '#E3EADA',
    justifyContent: 'center',
    paddingHorizontal: 36,
    paddingBottom: 100,
    alignItems: 'center',
  },
  symbol: { color: '#8DA576', fontSize: 80 },
  setupTitle: {
    fontSize: 25,
    lineHeight: 32,
    color: '#233E33',
    fontWeight: '700',
    textAlign: 'center',
  },
  setupText: {
    maxWidth: 400,
    marginTop: 14,
    fontSize: 15,
    lineHeight: 24,
    color: '#5B7058',
    textAlign: 'center',
  },
});
