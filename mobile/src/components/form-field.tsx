import { StyleSheet, TextInput, View, type TextInputProps } from 'react-native';

import { ThemedText } from '@/components/themed-text';
import { Colors } from '@/constants/theme';
import { useColorScheme } from '@/hooks/use-color-scheme';

interface FormFieldProps extends TextInputProps {
  label: string;
}

/** Labeled text input following the app's light/dark theme. */
export function FormField({ label, style, ...inputProps }: FormFieldProps) {
  const colorScheme = useColorScheme();
  const colors = Colors[colorScheme ?? 'light'];

  return (
    <View style={styles.wrapper}>
      <ThemedText type="defaultSemiBold">{label}</ThemedText>
      <TextInput
        style={[styles.input, { color: colors.text, borderColor: colors.icon }, style]}
        placeholderTextColor={colors.icon}
        autoCapitalize="none"
        autoCorrect={false}
        {...inputProps}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: {
    gap: 6,
    width: '100%',
  },
  input: {
    borderWidth: 1,
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 12,
    fontSize: 16,
  },
});
