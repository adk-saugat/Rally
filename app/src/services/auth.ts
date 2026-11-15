import { request } from './http';

export type AuthMode = 'login' | 'signup';
type Credentials = { name: string; email: string; password: string };

export async function authenticate(
  mode: AuthMode,
  credentials: Credentials,
  signal: AbortSignal,
): Promise<string | null> {
  const data = await request(`/auth/${mode}`, {
    method: 'POST',
    signal,
    body: {
      email: credentials.email.trim(),
      password: credentials.password,
      ...(mode === 'signup' ? { first_name: credentials.name.trim() } : {}),
    },
  });
  if (mode === 'signup') return null;
  if (!data || typeof data !== 'object' || !('token' in data) || typeof data.token !== 'string') {
    throw new Error('The server returned an invalid login response.');
  }
  return data.token;
}
