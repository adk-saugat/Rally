import { StyleSheet } from 'react-native';

export const sharedStyles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: '#F8F7F1' },
  eyebrow: { fontSize: 10, fontWeight: '700', letterSpacing: 2, color: '#718066' },
  primary: {
    minHeight: 56,
    paddingHorizontal: 22,
    paddingVertical: 15,
    borderRadius: 18,
    backgroundColor: '#294E3B',
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  primaryText: { color: '#fff', fontSize: 16, fontWeight: '600' },
  pressed: { opacity: 0.8, transform: [{ scale: 0.985 }] },
});
