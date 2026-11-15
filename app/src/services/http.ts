import { API_URL } from '../config';

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message);
  }
}

type RequestOptions = {
  method?: 'GET' | 'POST';
  body?: unknown;
  token?: string;
  signal?: AbortSignal;
};

export async function request(path: string, options: RequestOptions = {}): Promise<unknown> {
  const controller = new AbortController();
  const abort = () => controller.abort();
  options.signal?.addEventListener('abort', abort, { once: true });
  if (options.signal?.aborted) controller.abort();
  const timeout = setTimeout(abort, 10000);
  try {
    const response = await fetch(`${API_URL}${path}`, {
      method: options.method ?? 'GET',
      headers: {
        'Content-Type': 'application/json',
        ...(options.token ? { Authorization: `Bearer ${options.token}` } : {}),
      },
      ...(options.body !== undefined ? { body: JSON.stringify(options.body) } : {}),
      signal: controller.signal,
    });
    const data: unknown = await response.json().catch(() => null);
    if (!response.ok) {
      const message =
        data && typeof data === 'object' && 'error' in data && typeof data.error === 'string'
          ? data.error
          : 'Unable to continue. Please try again.';
      throw new ApiError(message, response.status);
    }
    return data;
  } catch (error) {
    if (error instanceof ApiError) throw error;
    throw new Error('Could not reach Rally. Check your connection and try again.');
  } finally {
    clearTimeout(timeout);
    options.signal?.removeEventListener('abort', abort);
  }
}
