export type Session = { token: string; userID: string; expiresAt: number };

// Decoding is only for identity/expiry in the UI. The backend verifies signatures.
export function sessionFromToken(token: string, now = Date.now()): Session {
  const payload = token.split('.')[1];
  if (!payload) throw new Error('The server returned an invalid login token.');
  try {
    const encoded = payload.replace(/-/g, '+').replace(/_/g, '/');
    const claims: unknown = JSON.parse(
      atob(encoded.padEnd(Math.ceil(encoded.length / 4) * 4, '=')),
    );
    if (
      !claims ||
      typeof claims !== 'object' ||
      !('sub' in claims) ||
      typeof claims.sub !== 'string' ||
      !claims.sub ||
      !('exp' in claims) ||
      typeof claims.exp !== 'number' ||
      !Number.isFinite(claims.exp) ||
      claims.exp * 1000 <= now
    ) {
      throw new Error('Invalid or expired token');
    }
    return { token, userID: claims.sub, expiresAt: claims.exp * 1000 };
  } catch {
    throw new Error('Your login has expired. Please log in again.');
  }
}
