export interface Tenant {
  id: string;
  name: string;
  slug: string;
  created_at: string;
  updated_at: string;
  deleted_at: string | null;
}

export interface Project {
  id: string;
  tenant_id: string;
  name: string;
  description?: string | null;
  created_at: string;
  updated_at: string;
  deleted_at: string | null;
}

export interface Member {
  id: string;
  tenant_id: string;
  user_id: string;
  email: string;
  role: 'owner' | 'admin' | 'developer' | 'viewer';
  created_at: string;
  updated_at: string;
}

export interface Environment {
  id: string;
  project_id: string;
  name: string;
  config?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface IngestionMode {
  id: string;
  project_id: string;
  mode: 'A' | 'B' | 'C';
  is_default: boolean;
  created_at: string;
}