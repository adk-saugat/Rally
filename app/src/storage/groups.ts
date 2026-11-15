import AsyncStorage from '@react-native-async-storage/async-storage';
import type { Group } from '../services/groups';

const key = (userID: string) => `rally.groups.v1.${userID}`;
export async function loadGroups(userID: string): Promise<Group[]> {
  const stored = await AsyncStorage.getItem(key(userID));
  if (!stored) return [];
  const data: unknown = JSON.parse(stored);
  if (!Array.isArray(data)) return [];
  return data
    .filter(
      (value): value is Group =>
        value &&
        typeof value.id === 'string' &&
        typeof value.name === 'string' &&
        (value.inviteCode === undefined || typeof value.inviteCode === 'string'),
    )
    .slice(0, 30);
}
export async function saveGroups(userID: string, groups: Group[]): Promise<void> {
  await AsyncStorage.setItem(key(userID), JSON.stringify(groups.slice(0, 30)));
}
