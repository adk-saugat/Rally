import { ScrollView, StyleSheet, Text, useWindowDimensions, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import Button from '../components/Button';
import type { ScreenProps } from '../navigation/types';
import MapIllustration from '../components/MapIllustration';
import { sharedStyles } from '../styles/shared';

export default function HomeScreen({ navigation }: ScreenProps<'Welcome'>) {
  const { height } = useWindowDimensions();
  const insets = useSafeAreaInsets();

  return (
    <View style={styles.screen}>
      <ScrollView
        automaticallyAdjustContentInsets={false}
        contentInsetAdjustmentBehavior="never"
        showsVerticalScrollIndicator={false}
        contentContainerStyle={[styles.content, { minHeight: height }]}
      >
        <View style={[styles.mapHalf, { minHeight: height / 2 + 12 }]}>
          <MapIllustration />
          <View style={[styles.header, { top: insets.top + 12, left: Math.max(insets.left, 28) }]}>
            <Text style={styles.brand}>
              rally<Text style={styles.brandDot}>.</Text>
            </Text>
          </View>
        </View>
        <View
          style={[
            styles.bottomHalf,
            {
              minHeight: height / 2 + 12,
              paddingBottom: Math.max(insets.bottom, 24) + 10,
              paddingLeft: Math.max(insets.left, 30),
              paddingRight: Math.max(insets.right, 30),
            },
          ]}
        >
          <View style={styles.welcome}>
            <View style={styles.kicker}>
              <View style={styles.kickerLine} />
              <Text style={styles.eyebrow}>MADE FOR YOUR PEOPLE</Text>
            </View>
            <Text accessibilityRole="header" style={styles.title}>
              Go together.{'\n'}
              <Text style={styles.titleAccent}>Stay close.</Text>
            </Text>
            <Text style={styles.homeDescription}>Less “where are you?” More being there.</Text>
            <View style={styles.actions}>
              <Button title="Sign up" onPress={() => navigation.navigate('Signup')} />
              <Button title="Log in" secondary onPress={() => navigation.navigate('Login')} />
            </View>
            <Text style={styles.footnote}>Invite-only groups. Shared adventures.</Text>
          </View>
        </View>
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  ...sharedStyles,
  content: { flexGrow: 1 },
  mapHalf: { flex: 1, overflow: 'hidden', backgroundColor: '#E9E8DD' },
  bottomHalf: {
    flex: 1,
    justifyContent: 'center',
    paddingHorizontal: 30,
    paddingTop: 32,
    paddingBottom: 34,
    borderTopLeftRadius: 32,
    borderTopRightRadius: 32,
    marginTop: -24,
    backgroundColor: '#F8F7F1',
  },
  welcome: { width: '100%', maxWidth: 480, alignSelf: 'center' },
  header: { position: 'absolute' },
  brand: { fontSize: 40, fontWeight: '800', letterSpacing: -2, color: '#233E33' },
  brandDot: { color: '#86A760' },
  title: {
    marginTop: 15,
    fontSize: 44,
    lineHeight: 49,
    fontWeight: '700',
    letterSpacing: -2,
    color: '#233E33',
  },
  kicker: { flexDirection: 'row', alignItems: 'center', gap: 10 },
  kickerLine: { width: 22, height: 2, backgroundColor: '#91A773' },
  titleAccent: { color: '#7B9165' },
  homeDescription: { marginTop: 12, fontSize: 14, lineHeight: 22, color: '#737B6E' },
  footnote: {
    marginTop: 20,
    textAlign: 'center',
    fontSize: 10,
    letterSpacing: 0.3,
    color: '#818978',
  },
  actions: { marginTop: 26, gap: 12 },
});
