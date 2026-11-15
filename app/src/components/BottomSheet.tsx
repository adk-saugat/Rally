import type { ReactNode } from 'react';
import {
  KeyboardAvoidingView,
  Modal,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

type BottomSheetProps = {
  visible: boolean;
  title: string;
  onClose: () => void;
  busy?: boolean;
  children: ReactNode;
};

export default function BottomSheet({
  visible,
  title,
  onClose,
  busy = false,
  children,
}: BottomSheetProps) {
  const insets = useSafeAreaInsets();
  return (
    <Modal
      visible={visible}
      transparent
      animationType="slide"
      onRequestClose={() => {
        if (!busy) onClose();
      }}
    >
      <KeyboardAvoidingView
        style={styles.overlay}
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      >
        <Pressable
          style={styles.backdrop}
          accessibilityRole="button"
          accessibilityLabel="Close panel"
          disabled={busy}
          onPress={onClose}
        />
        <View
          style={[
            styles.panel,
            {
              paddingBottom: Math.max(insets.bottom, 20),
              paddingLeft: Math.max(insets.left, 24),
              paddingRight: Math.max(insets.right, 24),
            },
          ]}
          accessibilityViewIsModal
        >
          <View style={styles.grabber} />
          <View style={styles.header}>
            <Text accessibilityRole="header" style={styles.title}>
              {title}
            </Text>
            <Pressable
              accessibilityRole="button"
              accessibilityLabel="Close panel"
              disabled={busy}
              onPress={onClose}
              style={styles.close}
            >
              <Text style={styles.closeText}>×</Text>
            </Pressable>
          </View>
          <ScrollView keyboardShouldPersistTaps="handled" contentContainerStyle={styles.content}>
            {children}
          </ScrollView>
        </View>
      </KeyboardAvoidingView>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: { flex: 1, justifyContent: 'flex-end' },
  backdrop: {
    position: 'absolute',
    top: 0,
    left: 0,
    bottom: 0,
    right: 0,
    backgroundColor: '#132A2455',
  },
  panel: {
    maxHeight: '85%',
    backgroundColor: '#F8F7F1',
    borderTopLeftRadius: 30,
    borderTopRightRadius: 30,
    paddingTop: 10,
  },
  grabber: {
    width: 40,
    height: 4,
    borderRadius: 2,
    backgroundColor: '#D1D7CA',
    alignSelf: 'center',
  },
  header: { flexDirection: 'row', alignItems: 'center', gap: 12, marginTop: 6 },
  title: { flex: 1, fontSize: 26, fontWeight: '700', letterSpacing: -0.8, color: '#233E33' },
  close: { minWidth: 48, minHeight: 48, justifyContent: 'center', alignItems: 'center' },
  closeText: { fontSize: 30, color: '#294E3B' },
  content: { paddingTop: 8, paddingBottom: 16, gap: 16 },
});
