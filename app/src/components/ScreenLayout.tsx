import type { ReactNode } from 'react';
import { KeyboardAvoidingView, Platform, ScrollView, StyleSheet, Text } from 'react-native';
import { useHeaderHeight } from '@react-navigation/elements';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { sharedStyles } from '../styles/shared';

type ScreenLayoutProps = {
  title: string;
  description?: string;
  eyebrow?: string;
  children: ReactNode;
};

export default function ScreenLayout({ title, description, eyebrow, children }: ScreenLayoutProps) {
  const insets = useSafeAreaInsets();
  const headerHeight = useHeaderHeight();

  return (
    <KeyboardAvoidingView
      style={sharedStyles.screen}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      keyboardVerticalOffset={headerHeight}
    >
      <ScrollView
        keyboardShouldPersistTaps="handled"
        keyboardDismissMode="on-drag"
        contentInsetAdjustmentBehavior="never"
        contentContainerStyle={[
          styles.content,
          {
            paddingBottom: Math.max(insets.bottom, 24) + 16,
            paddingLeft: Math.max(insets.left, 28),
            paddingRight: Math.max(insets.right, 28),
          },
        ]}
      >
        {eyebrow && <Text style={sharedStyles.eyebrow}>{eyebrow}</Text>}
        <Text accessibilityRole="header" style={styles.title}>
          {title}
        </Text>
        {description && <Text style={styles.description}>{description}</Text>}
        {children}
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  content: { flexGrow: 1, paddingTop: 28, maxWidth: 560, width: '100%', alignSelf: 'center' },
  title: { fontSize: 36, fontWeight: '700', color: '#233E33', marginTop: 14, letterSpacing: -1 },
  description: { marginTop: 16, fontSize: 16, lineHeight: 25, color: '#6A7369' },
});
