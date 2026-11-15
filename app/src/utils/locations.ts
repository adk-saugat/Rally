export type MemberLocation = {
  group_id: string;
  user_id: string;
  latitude: number;
  longitude: number;
  timestamp_unix_ms: number;
};

export function parseMemberLocation(
  value: unknown,
  groupID: string,
  now = Date.now(),
): MemberLocation | null {
  if (!value || typeof value !== 'object') return null;
  const point = value as Record<string, unknown>;
  if (
    point.group_id !== groupID ||
    typeof point.user_id !== 'string' ||
    !point.user_id ||
    typeof point.latitude !== 'number' ||
    !Number.isFinite(point.latitude) ||
    Math.abs(point.latitude) > 90 ||
    typeof point.longitude !== 'number' ||
    !Number.isFinite(point.longitude) ||
    Math.abs(point.longitude) > 180 ||
    typeof point.timestamp_unix_ms !== 'number' ||
    !Number.isFinite(point.timestamp_unix_ms) ||
    point.timestamp_unix_ms <= 0 ||
    point.timestamp_unix_ms > now + 60000
  )
    return null;
  return point as MemberLocation;
}

export function websocketURL(apiURL: string, groupID: string): string {
  return `${apiURL.replace(/^http/, 'ws')}/groups/${encodeURIComponent(groupID)}/ws`;
}
