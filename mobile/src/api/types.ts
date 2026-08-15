export type RelationType = 'parent' | 'child' | 'spouse' | 'sibling';

export interface User {
  id: number;
  name: string;
  email: string;
  family_id: number | null;
  created_at: string;
}

export interface Family {
  id: number;
  name: string;
  invite_code: string;
  created_at: string;
}

export interface Relation {
  id: number;
  family_id: number;
  user_id: number;
  related_user_id: number;
  type: RelationType;
  created_at: string;
}

export interface AuthResponse {
  user: User;
  token: string;
}
