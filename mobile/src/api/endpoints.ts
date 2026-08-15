import { api } from './client';
import type { AuthResponse, Family, Relation, RelationType, User } from './types';

export const authApi = {
  register(name: string, email: string, password: string): Promise<AuthResponse> {
    return api.request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: { name, email, password },
    });
  },

  login(email: string, password: string): Promise<AuthResponse> {
    return api.request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: { email, password },
    });
  },

  me(token: string): Promise<User> {
    return api.request<{ user: User }>('/auth/me', { token }).then((r) => r.user);
  },
};

export const familyApi = {
  me(token: string): Promise<Family> {
    return api.request<{ family: Family }>('/families/me', { token }).then((r) => r.family);
  },

  create(token: string, name: string): Promise<Family> {
    return api.request<{ family: Family }>('/families', {
      method: 'POST',
      body: { name },
      token,
    }).then((r) => r.family);
  },

  join(token: string, inviteCode: string): Promise<Family> {
    return api.request<{ family: Family }>('/families/join', {
      method: 'POST',
      body: { invite_code: inviteCode },
      token,
    }).then((r) => r.family);
  },
};

export const relationsApi = {
  list(token: string): Promise<Relation[]> {
    return api.request<{ relations: Relation[] }>('/relations', { token }).then((r) => r.relations);
  },

  create(token: string, relatedUserId: number, type: RelationType): Promise<Relation> {
    return api.request<{ relation: Relation }>('/relations', {
      method: 'POST',
      body: { related_user_id: relatedUserId, type },
      token,
    }).then((r) => r.relation);
  },
};
