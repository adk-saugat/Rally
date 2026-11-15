import { request } from './http';

export type Group = { id: string; name: string; inviteCode?: string };
export type Member = { user_id: string; first_name: string; email: string };

function record(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object') throw new Error('Unexpected group response.');
  return value as Record<string, unknown>;
}

export async function createGroup(
  token: string,
  name: string,
  signal?: AbortSignal,
): Promise<Group> {
  const data = record(
    await request('/groups', { method: 'POST', body: { name: name.trim() }, token, signal }),
  );
  if (typeof data.group_id !== 'string' || typeof data.invite_code !== 'string')
    throw new Error('Unexpected group response.');
  return { id: data.group_id, name: name.trim(), inviteCode: data.invite_code };
}

export async function joinGroup(token: string, code: string, signal?: AbortSignal): Promise<Group> {
  const data = record(
    await request('/groups/join', {
      method: 'POST',
      body: { invite_code: code.trim() },
      token,
      signal,
    }),
  );
  if (typeof data.group_id !== 'string' || typeof data.name !== 'string')
    throw new Error('Unexpected group response.');
  return { id: data.group_id, name: data.name, inviteCode: code.trim() };
}

export async function listMembers(
  token: string,
  id: string,
  signal?: AbortSignal,
): Promise<Member[]> {
  const data = record(
    await request(`/groups/${encodeURIComponent(id)}/members`, { token, signal }),
  );
  if (!Array.isArray(data.members)) throw new Error('Unexpected members response.');
  return data.members.map((value) => {
    const member = record(value);
    if (
      typeof member.user_id !== 'string' ||
      typeof member.first_name !== 'string' ||
      typeof member.email !== 'string'
    )
      throw new Error('Unexpected member response.');
    return { user_id: member.user_id, first_name: member.first_name, email: member.email };
  });
}
