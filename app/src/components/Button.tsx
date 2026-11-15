import { ActivityIndicator, Pressable, StyleSheet, Text } from 'react-native';
import { sharedStyles } from '../styles/shared';

type ButtonProps = {
  title: string;
  onPress: () => void;
  secondary?: boolean;
  disabled?: boolean;
  loading?: boolean;
};

export default function Button({
  title,
  onPress,
  secondary = false,
  disabled = false,
  loading = false,
}: ButtonProps) {
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ disabled: disabled || loading, busy: loading }}
      disabled={disabled || loading}
      onPress={onPress}
      style={({ pressed }) => [
        styles.button,
        secondary && styles.secondary,
        (disabled || loading) && styles.disabled,
        pressed && sharedStyles.pressed,
      ]}
    >
      {loading ? (
        <ActivityIndicator color="#fff" accessibilityLabel="Please wait" />
      ) : (
        <Text style={[sharedStyles.primaryText, secondary && styles.secondaryText]}>{title}</Text>
      )}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  button: { ...sharedStyles.primary, justifyContent: 'center', width: '100%' },
  secondary: { borderWidth: 1, borderColor: '#D7DDCF', backgroundColor: '#FDFCF8' },
  secondaryText: { color: '#294E3B' },
  disabled: { opacity: 0.55 },
});
