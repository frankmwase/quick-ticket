const API_HEADERS = (tenantId: string, apiKey: string) => ({
  'Content-Type': 'application/json',
  'X-Tenant-ID': tenantId,
  'X-API-Key': apiKey,
  'Authorization': `Bearer ${apiKey}`,
});

export interface Credentials {
  baseUrl: string;
  tenantId: string;
  apiKey: string;
}

export interface UserProfile {
  id: string;
  tenant_id: string;
  email: string;
  full_name: string;
  avatar_url: string;
  bio: string;
  created_at: string;
}

export interface Member {
  id: string;
  tenant_id: string;
  name: string;
  role: string;
  is_active: boolean;
  created_at: string;
}

export interface Ticket {
  ID: string;
  TenantID: string;
  BatchID: string;
  OwnerID: string;
  ManagedBy: string;
  SecureToken: string;
  Status: string;
  CreatedAt: string;
}

export interface GenerateResponse {
  count: number;
  tickets: Ticket[];
}

export interface VerifyResult {
  valid: boolean;
  ticket_id?: string;
  status?: string;
  error?: string;
}

async function apiRequest<T>(
  creds: Credentials,
  method: string,
  path: string,
  body?: unknown
): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${creds.baseUrl}${path}`, {
      method,
      headers: API_HEADERS(creds.tenantId, creds.apiKey),
      body: body ? JSON.stringify(body) : undefined,
    });
  } catch (err) {
    // Network error — backend unreachable
    throw new Error(
      `Network error: unable to reach ${creds.baseUrl}. Is the backend running?`
    );
  }

  // Handle empty responses (204 No Content)
  if (res.status === 204) {
    return {} as T;
  }

  let data: any;
  try {
    data = await res.json();
  } catch {
    throw new Error(`Invalid JSON response (${res.status})`);
  }

  if (!res.ok) {
    throw new Error(data.error || `API Error (${res.status})`);
  }

  return data as T;
}


export const api = {
  // Health (unauthenticated)
  health: (creds: Credentials) =>
    apiRequest<{ status: string }>(creds, 'GET', '/health'),

  // Profile
  getProfile: (creds: Credentials) =>
    apiRequest<UserProfile>(creds, 'GET', '/api/v1/profile'),

  updateProfile: (creds: Credentials, profile: Partial<UserProfile>) =>
    apiRequest<UserProfile>(creds, 'PUT', '/api/v1/profile', profile),

  // Members
  getMembers: (creds: Credentials) =>
    apiRequest<Member[]>(creds, 'GET', '/api/v1/members'),

  createMember: (creds: Credentials, name: string, role: string) =>
    apiRequest<Member>(creds, 'POST', '/api/v1/members', { name, role }),

  // Tickets
  generateTickets: (
    creds: Credentials,
    count: number,
    ownerId: string,
    managedBy?: string
  ) =>
    apiRequest<GenerateResponse>(creds, 'POST', '/api/v1/tickets/generate', {
      count,
      owner_id: ownerId,
      managed_by: managedBy || '',
    }),

  verifyTicket: (creds: Credentials, token: string) =>
    apiRequest<VerifyResult>(creds, 'POST', '/api/v1/tickets/verify', {
      token,
    }),

  revokeTicket: (creds: Credentials, ticketId: string, actorId: string) =>
    apiRequest<{ status: string }>(creds, 'POST', '/api/v1/tickets/revoke', {
      ticket_id: ticketId,
      actor_id: actorId,
      reason: 'Revoked via WebUI',
    }),
};
