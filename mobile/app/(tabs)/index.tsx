import { useCallback, useEffect, useState } from 'react';
import { ScrollView, StyleSheet } from 'react-native';

import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';
import { authApi, familyApi } from '@/src/api/endpoints';
import type { Family, User } from '@/src/api/types';
import { useAuth } from '@/src/auth/auth-context';
import { PrimaryButton } from '@/src/components/primary-button';

export default function HomeScreen() {
  const { token, signOut } = useAuth();
  const [user, setUser] = useState<User | null>(null);
  const [family, setFamily] = useState<Family | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!token) {
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const [me, myFamily] = await Promise.all([
        authApi.me(token),
        familyApi.me(token).catch(() => null),
      ]);
      setUser(me);
      setFamily(myFamily);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Could not load your profile');
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleSignOut() {
    await signOut();
  }

  if (loading) {
    return (
      <ThemedView style={styles.centered}>
        <ThemedText>Loading…</ThemedText>
      </ThemedView>
    );
  }

  return (
    <ThemedView style={styles.screen}>
      <ScrollView contentContainerStyle={styles.content}>
        <ThemedText type="title">Home</ThemedText>

        {error ? <ThemedText style={styles.error}>{error}</ThemedText> : null}

        {user ? (
          <ThemedView style={styles.card}>
            <ThemedText type="subtitle">{user.name}</ThemedText>
            <ThemedText>{user.email}</ThemedText>
            {family ? (
              <ThemedView style={styles.row}>
                <ThemedText type="defaultSemiBold">Family:</ThemedText>
                <ThemedText>{family.name}</ThemedText>
              </ThemedView>
            ) : (
              <ThemedText>You are not part of a family yet.</ThemedText>
            )}
          </ThemedView>
        ) : null}

        <ThemedView style={styles.card}>
          <ThemedText type="subtitle">Feed</ThemedText>
          <ThemedText>Family posts will appear here once the feed is live.</ThemedText>
        </ThemedView>

        <PrimaryButton title="Sign out" onPress={handleSignOut} variant="secondary" />
      </ScrollView>
    </ThemedView>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
  },
  centered: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
  },
  content: {
    padding: 24,
    gap: 16,
  },
  card: {
    gap: 8,
  },
  row: {
    flexDirection: 'row',
    gap: 8,
  },
  error: {
    color: '#e5484d',
  },
});
