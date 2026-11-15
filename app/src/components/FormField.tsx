import { StyleSheet, Text, TextInput, View, type TextInputProps } from 'react-native';
import { useId } from 'react';

type FormFieldProps = TextInputProps & { label: string };

export default function FormField({ label, ...props }: FormFieldProps) {
  const labelID = useId();
  return (
    <View style={styles.field}>
      <Text nativeID={labelID} style={styles.label}>
        {label}
      </Text>
      <TextInput
        accessibilityLabel={label}
        accessibilityLabelledBy={labelID}
        placeholderTextColor="#7B8476"
        {...props}
        style={[styles.input, props.style]}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  field: { marginTop: 24 },
  label: { marginBottom: 9, fontSize: 14, fontWeight: '600', color: '#233E33' },
  input: {
    minHeight: 54,
    borderWidth: 1,
    borderColor: '#CCD3C5',
    borderRadius: 12,
    backgroundColor: '#fff',
    paddingHorizontal: 16,
    paddingVertical: 14,
    color: '#233E33',
    fontSize: 16,
  },
});
