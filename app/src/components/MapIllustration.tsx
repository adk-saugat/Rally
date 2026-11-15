import { StyleSheet, Text, View } from 'react-native';

export default function MapIllustration() {
  return (
    <View
      style={styles.map}
      accessibilityRole="image"
      accessibilityLabel="Illustrated map showing three friends together along a route. Preview only, not live locations."
    >
      <View style={styles.neighborhood} />
      <View style={styles.parkOne} />
      <View style={styles.parkTwo} />
      {[0, 1, 2, 3, 4].map((i) => (
        <View key={`v${i}`} style={[styles.streetVertical, { left: `${12 + i * 21}%` }]} />
      ))}
      {[0, 1, 2, 3].map((i) => (
        <View key={`h${i}`} style={[styles.streetHorizontal, { top: `${16 + i * 23}%` }]} />
      ))}
      <View style={styles.riverBank} />
      <View style={styles.river} />
      <View style={styles.bridge} />
      <Text style={styles.streetLabel}>WILLOW AVENUE</Text>
      <Text style={styles.waterLabel}>the riverside</Text>
      <View style={styles.youHalo} />
      <View style={styles.routeHorizontal} />
      <View style={styles.routeVertical} />
      <Text style={styles.parkLabel}>RIVERSIDE PARK</Text>
      <View style={[styles.pin, styles.pinOne]}>
        <Text style={styles.pinText}>J</Text>
      </View>
      <View style={[styles.pin, styles.pinTwo]}>
        <Text style={styles.pinText}>A</Text>
      </View>
      <View style={[styles.pin, styles.pinThree]}>
        <Text style={styles.pinText}>You</Text>
      </View>
      <View style={styles.friendLabel}>
        <View style={styles.liveDot} />
        <Text style={styles.friendLabelText}>Your people, nearby</Text>
      </View>
      <View style={styles.mapBadge}>
        <View style={styles.miniAvatars}>
          <Text style={[styles.miniAvatar, styles.avatarPeach]}>J</Text>
          <Text style={[styles.miniAvatar, styles.avatarSage]}>A</Text>
          <Text style={[styles.miniAvatar, styles.avatarGreen]}>Y</Text>
        </View>
        <View>
          <Text style={styles.mapBadgeText}>A little closer, together.</Text>
          <Text style={styles.badgeSubtext}>Wherever the day takes you</Text>
        </View>
      </View>
      <Text style={styles.preview}>ILLUSTRATED PREVIEW</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  map: {
    position: 'absolute',
    top: 0,
    right: 0,
    bottom: 0,
    left: 0,
    backgroundColor: '#EAE9DF',
    overflow: 'hidden',
  },
  parkOne: {
    position: 'absolute',
    width: '46%',
    height: '60%',
    backgroundColor: '#CAD8B4',
    top: -20,
    left: -20,
    borderRadius: 50,
    transform: [{ rotate: '-16deg' }],
  },
  parkTwo: {
    position: 'absolute',
    width: '34%',
    height: '35%',
    backgroundColor: '#D1DDBD',
    bottom: -30,
    right: 50,
    borderRadius: 25,
  },
  streetVertical: {
    position: 'absolute',
    height: '150%',
    width: 9,
    backgroundColor: '#FAF9F1',
    top: -40,
    transform: [{ rotate: '22deg' }],
  },
  streetHorizontal: {
    position: 'absolute',
    width: '120%',
    height: 10,
    left: -20,
    backgroundColor: '#FAF9F1',
    transform: [{ rotate: '-12deg' }],
  },
  river: {
    position: 'absolute',
    width: 68,
    height: '150%',
    backgroundColor: '#B9D6D7',
    right: 10,
    top: -70,
    transform: [{ rotate: '22deg' }],
  },
  routeHorizontal: {
    position: 'absolute',
    left: '23%',
    top: '55%',
    width: '39%',
    height: 5,
    borderRadius: 3,
    backgroundColor: '#68865B',
  },
  routeVertical: {
    position: 'absolute',
    left: '61%',
    top: '35%',
    width: 5,
    height: '21%',
    backgroundColor: '#68865B',
  },
  parkLabel: {
    position: 'absolute',
    left: 28,
    top: '35%',
    width: 90,
    fontSize: 9,
    lineHeight: 14,
    letterSpacing: 1.5,
    color: '#738461',
    fontWeight: '600',
  },
  pin: {
    position: 'absolute',
    width: 43,
    height: 43,
    borderRadius: 22,
    borderWidth: 4,
    borderColor: '#fff',
    alignItems: 'center',
    justifyContent: 'center',
    shadowColor: '#233E33',
    shadowOpacity: 0.16,
    shadowRadius: 6,
    shadowOffset: { width: 0, height: 3 },
    elevation: 3,
  },
  pinOne: { top: '50%', left: '18%', backgroundColor: '#BC7654' },
  pinTwo: { top: '30%', left: '56%', backgroundColor: '#7C8B62' },
  pinThree: { top: '51%', left: '52%', backgroundColor: '#294E3B' },
  pinText: { fontWeight: '700', color: '#fff', fontSize: 12 },
  mapBadge: {
    position: 'absolute',
    bottom: 48,
    left: 26,
    right: 26,
    maxWidth: 350,
    paddingHorizontal: 15,
    paddingVertical: 13,
    borderRadius: 19,
    shadowColor: '#233E33',
    shadowOpacity: 0.08,
    shadowRadius: 18,
    shadowOffset: { width: 0, height: 5 },
    elevation: 3,
    backgroundColor: '#FFFFFF',
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  liveDot: { width: 7, height: 7, borderRadius: 4, backgroundColor: '#7E9A5C' },
  mapBadgeText: { fontSize: 11, fontWeight: '600', color: '#3F5141' },
  preview: {
    position: 'absolute',
    bottom: 30,
    right: 18,
    fontSize: 7,
    letterSpacing: 1,
    color: '#56645C',
  },
  neighborhood: {
    position: 'absolute',
    top: '24%',
    left: '27%',
    width: '27%',
    height: '31%',
    borderRadius: 22,
    backgroundColor: '#DFDFD0',
    transform: [{ rotate: '-12deg' }],
  },
  riverBank: {
    position: 'absolute',
    width: 88,
    height: '150%',
    backgroundColor: '#D5DFD0',
    right: 0,
    top: -70,
    transform: [{ rotate: '22deg' }],
  },
  bridge: {
    position: 'absolute',
    top: '55%',
    right: 0,
    width: 115,
    height: 14,
    borderWidth: 2,
    borderColor: '#E0DDCB',
    backgroundColor: '#FAF9F1',
    transform: [{ rotate: '-12deg' }],
  },
  streetLabel: {
    position: 'absolute',
    left: '28%',
    top: '72%',
    fontSize: 7,
    letterSpacing: 2,
    color: '#8A8D7D',
    transform: [{ rotate: '-12deg' }],
  },
  waterLabel: {
    position: 'absolute',
    right: 5,
    top: '39%',
    fontSize: 10,
    fontStyle: 'italic',
    color: '#688F92',
    transform: [{ rotate: '-68deg' }],
  },
  youHalo: {
    position: 'absolute',
    top: '51%',
    left: '52%',
    width: 79,
    height: 79,
    marginLeft: -18,
    marginTop: -18,
    borderRadius: 40,
    borderWidth: 1,
    borderColor: '#7C9E6C55',
    backgroundColor: '#7C9E6C22',
  },
  friendLabel: {
    position: 'absolute',
    top: '23%',
    right: 24,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    borderRadius: 12,
    paddingVertical: 7,
    paddingHorizontal: 10,
    backgroundColor: '#FAF9F1',
  },
  friendLabelText: { fontSize: 9, fontWeight: '600', color: '#53674C' },
  miniAvatars: { flexDirection: 'row', paddingRight: 5 },
  miniAvatar: {
    width: 29,
    height: 29,
    borderRadius: 15,
    borderWidth: 2,
    borderColor: '#fff',
    overflow: 'hidden',
    textAlign: 'center',
    lineHeight: 25,
    fontSize: 10,
    fontWeight: '700',
    color: '#fff',
    marginRight: -7,
  },
  avatarPeach: { backgroundColor: '#BC7654' },
  avatarSage: { backgroundColor: '#7C8B62' },
  avatarGreen: { backgroundColor: '#294E3B' },
  badgeSubtext: { fontSize: 9, color: '#7B8476', marginTop: 3 },
});
