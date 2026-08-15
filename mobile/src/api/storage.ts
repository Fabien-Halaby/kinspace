import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

const TOKEN_KEY = 'kinspace.session';

/**
 * Persists the session token in the platform keychain on native and in
 * localStorage on web (where SecureStore is unavailable).
 */
export async function saveSessionToken(token: string | null): Promise<void> {
  if (Platform.OS === 'web') {
    if (token === null) {
      localStorage.removeItem(TOKEN_KEY);
    } else {
      localStorage.setItem(TOKEN_KEY, token);
    }
    return;
  }
  if (token === null) {
    await SecureStore.deleteItemAsync(TOKEN_KEY);
  } else {
    await SecureStore.setItemAsync(TOKEN_KEY, token);
  }
}

export async function loadSessionToken(): Promise<string | null> {
  if (Platform.OS === 'web') {
    return localStorage.getItem(TOKEN_KEY);
  }
  return SecureStore.getItemAsync(TOKEN_KEY);
}
