import { useCallback, useEffect, useRef, useState } from 'react';
import {
  Pressable,
  ScrollView,
  Share,
  StyleSheet,
  Text,
  useWindowDimensions,
  View,
} from 'react-native';
import type MapView from 'react-native-maps';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import BottomSheet from '../components/BottomSheet';
import Button from '../components/Button';
import FormField from '../components/FormField';
import RallyMap, { cameraRegion } from '../components/RallyMap';
import useDeviceLocation from '../hooks/useDeviceLocation';
import useGroups from '../hooks/useGroups';
import useLiveLocations from '../hooks/useLiveLocations';
import { useSession } from '../session/SessionProvider';
import type { Group } from '../services/groups';

type Panel = 'groups' | 'create' | 'join' | 'profile' | 'members' | null;
const panelTitles = {
  groups: 'Your groups',
  create: 'Create a group',
  join: 'Join your people',
  profile: 'Your Rally',
  members: 'Your crew',
};

export default function MapScreen() {
  const insets = useSafeAreaInsets();
  const { height } = useWindowDimensions();
  const { signOut } = useSession();
  const location = useDeviceLocation();
  const group = useGroups();
  const live = useLiveLocations(
    group.selected?.id,
    location.position,
    location.access === 'granted' && location.active && group.verified,
  );
  const map = useRef<MapView>(null);
  const centered = useRef(false);
  const [mapReady, setMapReady] = useState(false);
  const [panel, setPanel] = useState<Panel>(null);
  const [value, setValue] = useState('');
  const [now, setNow] = useState(Date.now());
  const [cardHeight, setCardHeight] = useState(240);
  const [loggingOut, setLoggingOut] = useState(false);
  const [error, setError] = useState('');
  const onMapReady = useCallback(() => setMapReady(true), []);

  useEffect(() => {
    if (!location.active) return;
    const timer = setInterval(() => setNow(Date.now()), 5000);
    return () => clearInterval(timer);
  }, [location.active]);

  useEffect(() => {
    if (mapReady && location.position && !centered.current) {
      map.current?.animateToRegion(cameraRegion(location.position), 700);
      centered.current = true;
    }
  }, [mapReady, location.position]);

  function open(next: Panel) {
    setError('');
    group.setMessage('');
    setValue('');
    setPanel(next);
  }
  function select(next: Group | null) {
    live.stop();
    group.select(next);
    setPanel(null);
  }

  async function submit() {
    if (panel !== 'create' && panel !== 'join') return;
    live.stop();
    if (await group.submit(panel, value)) setPanel(null);
  }

  async function logout() {
    if (loggingOut) return;
    live.stop();
    setLoggingOut(true);
    setError('');
    try {
      await signOut();
    } catch {
      setError('Could not remove the saved login. Please try logging out again.');
    } finally {
      setLoggingOut(false);
    }
  }

  async function shareInvite() {
    if (!group.selected?.inviteCode) return;
    try {
      await Share.share({
        message: `Join ${group.selected.name} on Rally. Invite code: ${group.selected.inviteCode}`,
      });
    } catch {
      setError('Could not open the share menu. Try again.');
    }
  }

  function recenter() {
    if (location.position && location.access === 'granted')
      map.current?.animateToRegion(cameraRegion(location.position), 500);
    else void location.requestAccess();
  }

  function fitGroup() {
    const coordinates = [
      ...(location.position ? [location.position] : []),
      ...Object.values(live.points).filter((point) => now - point.timestamp_unix_ms < 120000),
    ];
    if (coordinates.length === 1) map.current?.animateToRegion(cameraRegion(coordinates[0]), 500);
    else if (coordinates.length > 1)
      map.current?.fitToCoordinates(coordinates, {
        animated: true,
        edgePadding: {
          top: insets.top + 150,
          right: 45,
          bottom: cardHeight + insets.bottom + 30,
          left: 45,
        },
      });
  }

  const requestLocation = location.access !== 'granted';
  const points = Object.values(live.points);
  const status = live.sharing
    ? live.status
    : live.status === 'Not sharing'
      ? 'Your location stays on your device.'
      : live.status;

  return (
    <View style={styles.screen}>
      <RallyMap
        ref={map}
        locationEnabled={location.access === 'granted' && location.active}
        position={location.access === 'granted' ? location.position : null}
        points={points}
        members={group.members}
        now={now}
        onReady={onMapReady}
        bottomPadding={cardHeight + insets.bottom + 20}
      />
      <View
        style={[
          styles.top,
          {
            top: insets.top + 12,
            left: Math.max(insets.left, 20),
            right: Math.max(insets.right, 20),
          },
        ]}
      >
        <View style={styles.header}>
          <Text style={styles.brand}>
            rally<Text style={styles.dot}>.</Text>
          </Text>
          <Pressable
            accessibilityRole="button"
            accessibilityLabel="Open profile and settings"
            onPress={() => open('profile')}
            style={styles.round}
          >
            <Text style={styles.roundText}>☰</Text>
          </Pressable>
        </View>
        <View style={styles.toolbar}>
          <Pressable accessibilityRole="button" onPress={() => open('create')} style={styles.tool}>
            <Text style={styles.toolText}>＋ Create group</Text>
          </Pressable>
          <Pressable accessibilityRole="button" onPress={() => open('join')} style={styles.tool}>
            <Text style={styles.toolText}>↗ Join group</Text>
          </Pressable>
        </View>
      </View>
      <View
        style={[
          styles.bottom,
          {
            bottom: Math.max(insets.bottom, 16),
            left: Math.max(insets.left, 20),
            right: Math.max(insets.right, 20),
          },
        ]}
      >
        <View style={styles.mapActions}>
          <Pressable
            accessibilityRole="button"
            accessibilityLabel="Fit group locations on map"
            onPress={fitGroup}
            style={styles.round}
          >
            <Text style={styles.roundText}>⊙</Text>
          </Pressable>
          <Pressable
            accessibilityRole="button"
            accessibilityLabel="Center map on my location"
            onPress={recenter}
            style={styles.round}
          >
            <Text style={styles.roundText}>⌖</Text>
          </Pressable>
        </View>
        <View
          style={[styles.card, { maxHeight: Math.max(180, height * 0.5) }]}
          onLayout={(event) => setCardHeight(event.nativeEvent.layout.height + 60)}
        >
          <ScrollView
            contentContainerStyle={styles.cardContent}
            showsVerticalScrollIndicator={false}
            keyboardShouldPersistTaps="handled"
          >
            <Pressable
              accessibilityRole="button"
              accessibilityLabel="Choose a group"
              onPress={() => open('groups')}
              style={styles.groupTitle}
            >
              <View style={styles.groupIcon}>
                <Text style={styles.groupIconText}>◎</Text>
              </View>
              <View style={styles.grow}>
                <Text style={styles.eyebrow}>YOUR MAP</Text>
                <Text style={styles.cardTitle}>{group.selected?.name ?? 'Find your people'}</Text>
              </View>
              <Text style={styles.chevron}>⌄</Text>
            </Pressable>
            {requestLocation ? (
              <>
                <Text style={styles.copy}>
                  Let Rally show where you are. Sharing with a group stays off until you turn it on.
                </Text>
                <Button
                  title={
                    location.access === 'blocked' ? 'Open location settings' : 'Use my location'
                  }
                  onPress={() => {
                    void location.requestAccess();
                  }}
                  loading={location.requesting}
                  disabled={location.access === 'loading'}
                />
                {location.access === 'denied' && (
                  <Text style={styles.hint}>
                    Location was declined. You can keep browsing the map.
                  </Text>
                )}
              </>
            ) : (
              <>
                <Text accessibilityLiveRegion="polite" style={styles.copy}>
                  {status}
                </Text>
                {group.selected ? (
                  <View style={styles.buttonRow}>
                    <View style={styles.grow}>
                      <Button
                        title={live.sharing ? 'Stop sharing' : 'Start sharing'}
                        onPress={live.sharing ? live.stop : live.start}
                        disabled={
                          !live.sharing &&
                          (!location.position || !group.verified || !location.active)
                        }
                      />
                    </View>
                    <Pressable
                      accessibilityRole="button"
                      accessibilityLabel="View group members"
                      onPress={() => open('members')}
                      style={styles.membersButton}
                    >
                      <Text style={styles.toolText}>Crew</Text>
                    </Pressable>
                  </View>
                ) : (
                  <Text style={styles.hint}>Create or join a group using the buttons above.</Text>
                )}
              </>
            )}
            {location.message ? <Text style={styles.hint}>{location.message}</Text> : null}
            {group.loadingMembers && group.selected ? (
              <Text style={styles.hint}>Loading your crew…</Text>
            ) : null}
            {group.message || error ? (
              <Text accessibilityLiveRegion="polite" style={styles.error}>
                {error || group.message}
              </Text>
            ) : null}
          </ScrollView>
        </View>
      </View>
      <BottomSheet
        visible={panel !== null}
        title={panel ? panelTitles[panel] : ''}
        onClose={() => setPanel(null)}
        busy={group.busy || loggingOut}
      >
        {(panel === 'create' || panel === 'join') && (
          <>
            <Text style={styles.copy}>
              {panel === 'create'
                ? 'Give your adventure a name. We’ll make an invite for your friends.'
                : 'Enter the invite code shared by someone in your group.'}
            </Text>
            <FormField
              label={panel === 'create' ? 'Group name' : 'Invite code'}
              value={value}
              onChangeText={setValue}
              editable={!group.busy}
              autoCapitalize={panel === 'join' ? 'characters' : 'sentences'}
              autoCorrect={panel !== 'join'}
              returnKeyType="go"
              onSubmitEditing={submit}
              maxLength={panel === 'create' ? 100 : 64}
            />
            <Button
              title={panel === 'create' ? 'Create group' : 'Join group'}
              onPress={submit}
              loading={group.busy}
              disabled={!group.ready || !value.trim()}
            />
          </>
        )}
        {panel === 'groups' && (
          <>
            <Text style={styles.copy}>
              Groups you’ve opened on this device. Creating, joining, or switching a group never
              starts location sharing.
            </Text>
            {!group.ready && <Text style={styles.hint}>Restoring your groups…</Text>}
            {group.ready && group.groups.length === 0 && (
              <Text style={styles.hint}>No groups yet. Start your next adventure below.</Text>
            )}
            {group.groups.map((item) => (
              <Pressable
                key={item.id}
                accessibilityRole="button"
                accessibilityState={{ selected: item.id === group.selected?.id }}
                style={styles.groupRow}
                onPress={() => select(item)}
              >
                <Text style={styles.groupRowText}>{item.name}</Text>
                <Text style={styles.chevron}>{item.id === group.selected?.id ? '✓' : '→'}</Text>
              </Pressable>
            ))}
            {group.selected && (
              <Button title="Use my personal map" secondary onPress={() => select(null)} />
            )}
            <Button title="Create group" onPress={() => open('create')} />
            <Button title="Join with an invite" secondary onPress={() => open('join')} />
          </>
        )}
        {panel === 'members' && (
          <>
            <Text style={styles.copy}>{group.selected?.name}</Text>
            {group.selected?.inviteCode && (
              <View style={styles.invite}>
                <Text style={styles.eyebrow}>INVITE CODE</Text>
                <Text selectable style={styles.inviteCode}>
                  {group.selected.inviteCode}
                </Text>
                <Button
                  title="Share invite"
                  secondary
                  onPress={() => {
                    void shareInvite();
                  }}
                />
              </View>
            )}
            {group.members.map((member) => {
              const point = live.points[member.user_id];
              const seconds = point
                ? Math.max(0, Math.floor((now - point.timestamp_unix_ms) / 1000))
                : null;
              return (
                <View key={member.user_id} style={styles.memberRow}>
                  <Text style={styles.groupRowText}>{member.first_name}</Text>
                  <Text style={styles.hint}>
                    {seconds === null
                      ? 'No shared position received'
                      : `Updated ${seconds}s ago${seconds > 30 ? ' · Last seen' : ''}`}
                  </Text>
                </View>
              );
            })}
            {!live.sharing && (
              <Text style={styles.hint}>
                Live updates connect when you start sharing. Stopped members’ last positions may
                remain visible to peers until marked stale.
              </Text>
            )}
            <Button
              title="Refresh members"
              secondary
              onPress={group.refresh}
              loading={group.loadingMembers}
            />
          </>
        )}
        {panel === 'profile' && (
          <>
            <Text style={styles.copy}>
              You’re signed in on this device. Your location is private until you choose a group and
              start sharing.
            </Text>
            <Text style={styles.hint}>
              Sharing pauses when Rally leaves the foreground. Tap Start sharing when you return.
            </Text>
            <Button
              title={
                location.access === 'blocked' ? 'Open location settings' : 'Location permission'
              }
              secondary
              onPress={() => {
                void location.requestAccess();
              }}
              loading={location.requesting}
            />
            <Button
              title="Log out"
              onPress={() => {
                void logout();
              }}
              loading={loggingOut}
            />
          </>
        )}
        {(group.message || error) && (
          <Text accessibilityLiveRegion="polite" style={styles.error}>
            {error || group.message}
          </Text>
        )}
      </BottomSheet>
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: '#E3EADA' },
  top: { position: 'absolute', gap: 14 },
  header: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  brand: { fontSize: 40, fontWeight: '800', letterSpacing: -2, color: '#233E33' },
  dot: { color: '#86A760' },
  toolbar: { flexDirection: 'row', gap: 10, flexWrap: 'wrap' },
  tool: {
    backgroundColor: '#FDFCF8',
    minHeight: 48,
    borderRadius: 18,
    paddingHorizontal: 16,
    paddingVertical: 14,
    shadowColor: '#233E33',
    shadowOpacity: 0.12,
    shadowRadius: 10,
    shadowOffset: { width: 0, height: 3 },
    elevation: 3,
  },
  toolText: { color: '#294E3B', fontSize: 14, fontWeight: '600' },
  round: {
    width: 48,
    height: 48,
    borderRadius: 24,
    backgroundColor: '#FDFCF8',
    alignItems: 'center',
    justifyContent: 'center',
    shadowColor: '#233E33',
    shadowOpacity: 0.1,
    shadowRadius: 8,
    shadowOffset: { width: 0, height: 2 },
    elevation: 2,
  },
  roundText: { fontSize: 24, color: '#294E3B' },
  bottom: { position: 'absolute', gap: 12 },
  mapActions: { flexDirection: 'row', justifyContent: 'flex-end', gap: 10 },
  cardContent: { padding: 20, gap: 14 },
  card: {
    borderRadius: 26,
    backgroundColor: '#FDFCF8',
    gap: 14,
    shadowColor: '#233E33',
    shadowOpacity: 0.1,
    shadowRadius: 18,
    shadowOffset: { width: 0, height: 5 },
    elevation: 4,
  },
  groupTitle: { minHeight: 48, flexDirection: 'row', gap: 12, alignItems: 'center' },
  groupIcon: {
    width: 44,
    height: 44,
    backgroundColor: '#E8EEDB',
    borderRadius: 15,
    alignItems: 'center',
    justifyContent: 'center',
  },
  groupIconText: { fontSize: 27, color: '#7B9165' },
  grow: { flex: 1 },
  eyebrow: { fontSize: 9, fontWeight: '700', letterSpacing: 1.6, color: '#7B9165' },
  cardTitle: {
    marginTop: 4,
    fontSize: 20,
    fontWeight: '700',
    letterSpacing: -0.5,
    color: '#233E33',
  },
  chevron: { fontSize: 23, color: '#7B9165' },
  copy: { fontSize: 14, lineHeight: 22, color: '#5E6E59' },
  hint: { fontSize: 12, lineHeight: 19, color: '#75806D' },
  error: { fontSize: 13, lineHeight: 20, color: '#A1392A' },
  buttonRow: { flexDirection: 'row', gap: 12 },
  membersButton: {
    minWidth: 58,
    minHeight: 56,
    padding: 12,
    borderRadius: 18,
    borderWidth: 1,
    borderColor: '#D7DDCF',
    alignItems: 'center',
    justifyContent: 'center',
  },
  groupRow: {
    paddingVertical: 18,
    paddingHorizontal: 14,
    borderRadius: 16,
    backgroundColor: '#EAEEDD',
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
  },
  groupRowText: { flex: 1, fontSize: 16, fontWeight: '600', color: '#233E33' },
  memberRow: { gap: 6, paddingVertical: 12, borderBottomWidth: 1, borderBottomColor: '#DFE4D6' },
  invite: { gap: 12, padding: 20, backgroundColor: '#EAEEDD', borderRadius: 18 },
  inviteCode: { fontSize: 25, fontWeight: '700', letterSpacing: 2, color: '#233E33' },
});
