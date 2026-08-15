import { useCallback, useEffect, useState } from 'react';
import { ScrollView, StyleSheet } from 'react-native';

import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';
import { familyApi, relationsApi } from '@/src/api/endpoints';
import type { Family, Relation, RelationType } from '@/src/api/types';
import { useAuth } from '@/src/auth/auth-context';
import { FormField } from '@/src/components/form-field';
import { PrimaryButton } from '@/src/components/primary-button';

const RELATION_TYPES: RelationType[] = ['parent', 'child', 'spouse', 'sibling'];

export default function FamilyScreen() {
  const { token } = useAuth();

  const [family, setFamily] = useState<Family | null>(null);
  const [relations, setRelations] = useState<Relation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  // Create / join form
  const [createName, setCreateName] = useState('');
  const [inviteCode, setInviteCode] = useState('');

  // Add-relation form
  const [relatedUserId, setRelatedUserId] = useState('');
  const [relationType, setRelationType] = useState<RelationType>('parent');

  const load = useCallback(async () => {
    if (!token) {
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const myFamily = await familyApi.me(token).catch(() => null);
      setFamily(myFamily);
      setRelations(myFamily ? await relationsApi.list(token) : []);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Could not load your family');
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleCreate() {
    if (!token) {
      return;
    }
    setError(null);
    try {
      const created = await familyApi.create(token, createName);
      setFamily(created);
      setCreateName('');
      setNotice(`Family created. Share the invite code below.`);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Could not create the family');
    }
  }

  async function handleJoin() {
    if (!token) {
      return;
    }
    setError(null);
    try {
      const joined = await familyApi.join(token, inviteCode);
      setFamily(joined);
      setInviteCode('');
      setNotice(`Welcome to ${joined.name}!`);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Could not join the family');
    }
  }

  async function handleAddRelation() {
    if (!token) {
      return;
    }
    const relatedUserIdValue = Number(relatedUserId);
    if (!Number.isInteger(relatedUserIdValue) || relatedUserIdValue <= 0) {
      setError('Enter a valid member id');
      return;
    }
    setError(null);
    setNotice(null);
    try {
      await relationsApi.create(token, relatedUserIdValue, relationType);
      setRelatedUserId('');
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Could not add the relation');
    }
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
        <ThemedText type="title">Family</ThemedText>

        {error ? <ThemedText style={styles.error}>{error}</ThemedText> : null}
        {notice ? <ThemedText style={styles.notice}>{notice}</ThemedText> : null}

        {!family ? (
          <>
            <ThemedText type="subtitle">Create a family</ThemedText>
            <FormField
              label="Family name"
              value={createName}
              onChangeText={setCreateName}
              placeholder="The Smiths"
            />
            <PrimaryButton
              title="Create family"
              onPress={handleCreate}
              disabled={createName.length < 2}
            />

            <ThemedText type="subtitle">…or join one</ThemedText>
            <FormField
              label="Invite code"
              value={inviteCode}
              onChangeText={setInviteCode}
              placeholder="ABC123"
              autoCapitalize="characters"
            />
            <PrimaryButton
              title="Join family"
              onPress={handleJoin}
              variant="secondary"
              disabled={inviteCode.length === 0}
            />
          </>
        ) : (
          <>
            <ThemedView style={styles.card}>
              <ThemedText type="subtitle">{family.name}</ThemedText>
              <ThemedView style={styles.row}>
                <ThemedText type="defaultSemiBold">Invite code:</ThemedText>
                <ThemedText>{family.invite_code}</ThemedText>
              </ThemedView>
            </ThemedView>

            <ThemedText type="subtitle">Family tree</ThemedText>
            {relations.length === 0 ? (
              <ThemedText>No relations yet. Add your first family member below.</ThemedText>
            ) : (
              relations.map((relation) => (
                <ThemedView key={relation.id} style={styles.relationRow}>
                  <ThemedText>
                    #{relation.user_id} is <ThemedText type="defaultSemiBold">{relation.type}</ThemedText>{' '}
                    of #{relation.related_user_id}
                  </ThemedText>
                </ThemedView>
              ))
            )}

            <ThemedText type="subtitle">Add a relation</ThemedText>
            <FormField
              label="Member id"
              value={relatedUserId}
              onChangeText={setRelatedUserId}
              placeholder="2"
              keyboardType="number-pad"
            />
            <ThemedView style={styles.chipRow}>
              {RELATION_TYPES.map((type) => (
                <PrimaryButton
                  key={type}
                  title={type}
                  onPress={() => setRelationType(type)}
                  variant={relationType === type ? 'primary' : 'secondary'}
                />
              ))}
            </ThemedView>
            <PrimaryButton
              title="Add relation"
              onPress={handleAddRelation}
              disabled={relatedUserId.length === 0}
            />
          </>
        )}
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
  relationRow: {
    paddingVertical: 8,
    borderBottomWidth: StyleSheet.hairlineWidth,
  },
  chipRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  error: {
    color: '#e5484d',
  },
  notice: {
    color: '#30a46c',
  },
});
