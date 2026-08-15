import { Link, router } from 'expo-router';
import { useState } from 'react';
import { KeyboardAvoidingView, Platform, ScrollView, StyleSheet } from 'react-native';

import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';
import { authApi } from '@/src/api/endpoints';
import { useAuth } from '@/src/auth/auth-context';
import { FormField } from '@/src/components/form-field';
import { PrimaryButton } from '@/src/components/primary-button';

export default function RegisterScreen() {
  const { signIn } = useAuth();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const canSubmit = name.length >= 2 && email.length > 0 && password.length >= 8;

  async function handleSubmit() {
    setSubmitting(true);
    setError(null);
    try {
      const { token } = await authApi.register(name, email, password);
      await signIn(token);
      router.replace('/');
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Something went wrong');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <ThemedView style={styles.screen}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.flex}>
        <ScrollView contentContainerStyle={styles.content}>
          <ThemedText type="title">KinSpace</ThemedText>
          <ThemedText type="subtitle">Create your account</ThemedText>

          <FormField
            label="Name"
            value={name}
            onChangeText={setName}
            placeholder="Fabien"
            autoComplete="name"
            textContentType="name"
          />
          <FormField
            label="Email"
            value={email}
            onChangeText={setEmail}
            placeholder="you@example.com"
            keyboardType="email-address"
            autoComplete="email"
            textContentType="emailAddress"
          />
          <FormField
            label="Password"
            value={password}
            onChangeText={setPassword}
            placeholder="At least 8 characters"
            secureTextEntry
            autoComplete="new-password"
            textContentType="newPassword"
          />

          {error ? <ThemedText style={styles.error}>{error}</ThemedText> : null}

          <PrimaryButton
            title="Create account"
            onPress={handleSubmit}
            loading={submitting}
            disabled={!canSubmit}
          />

          <ThemedText style={styles.switch}>
            Already have an account? <Link href="/(auth)/login"><ThemedText type="link">Sign in</ThemedText></Link>
          </ThemedText>
        </ScrollView>
      </KeyboardAvoidingView>
    </ThemedView>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
  },
  flex: {
    flex: 1,
  },
  content: {
    flexGrow: 1,
    justifyContent: 'center',
    padding: 24,
    gap: 16,
  },
  error: {
    color: '#e5484d',
  },
  switch: {
    textAlign: 'center',
    marginTop: 8,
  },
});
