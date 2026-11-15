import { useCallback, useEffect, useRef, useState } from 'react';
import { useSession } from '../session/SessionProvider';
import { createGroup, joinGroup, listMembers, type Group, type Member } from '../services/groups';
import { ApiError } from '../services/http';
import { loadGroups, saveGroups } from '../storage/groups';

type GroupAction = 'create' | 'join';

export default function useGroups() {
  const { session, expire } = useSession();
  const [groups, setGroups] = useState<Group[]>([]);
  const [selected, setSelected] = useState<Group | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [ready, setReady] = useState(false);
  const [verified, setVerified] = useState(false);
  const [loadingMembers, setLoadingMembers] = useState(false);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState('');
  const [revision, setRevision] = useState(0);
  const pending = useRef<AbortController | null>(null);

  const handleError = useCallback(
    (error: unknown) => {
      if (error instanceof ApiError && error.status === 401) {
        expire();
        return;
      }
      setMessage(error instanceof Error ? error.message : 'Unable to load your group.');
    },
    [expire],
  );

  useEffect(() => {
    if (!session) return;
    let disposed = false;
    loadGroups(session.userID)
      .then((saved) => {
        if (disposed) return;
        setGroups(saved);
        setSelected(saved[0] ?? null);
      })
      .catch(() => {
        if (!disposed) setMessage('Could not restore groups saved on this device.');
      })
      .finally(() => {
        if (!disposed) setReady(true);
      });
    return () => {
      disposed = true;
      pending.current?.abort();
    };
  }, [session]);

  useEffect(() => {
    setMembers([]);
    setVerified(false);
    if (!session || !selected) {
      setLoadingMembers(false);
      return;
    }
    const controller = new AbortController();
    setLoadingMembers(true);
    listMembers(session.token, selected.id, controller.signal)
      .then((next) => {
        if (!controller.signal.aborted) {
          setMembers(next);
          setVerified(true);
        }
      })
      .catch((error) => {
        if (!controller.signal.aborted) handleError(error);
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoadingMembers(false);
      });
    return () => controller.abort();
  }, [session, selected, revision, handleError]);

  async function submit(action: GroupAction, value: string): Promise<boolean> {
    if (!session || pending.current || !ready) return false;
    if (!value.trim()) {
      setMessage(action === 'create' ? 'Give your group a name.' : 'Enter an invite code.');
      return false;
    }
    const controller = new AbortController();
    pending.current = controller;
    setBusy(true);
    setMessage('');
    try {
      const group =
        action === 'create'
          ? await createGroup(session.token, value, controller.signal)
          : await joinGroup(session.token, value, controller.signal);
      if (controller.signal.aborted) return false;
      const next = [group, ...groups.filter((item) => item.id !== group.id)].slice(0, 30);
      setGroups(next);
      setSelected(group);
      try {
        await saveGroups(session.userID, next);
      } catch {
        setMessage('Group opened, but could not save it on this device. Keep your invite code.');
      }
      return true;
    } catch (error) {
      if (!controller.signal.aborted) handleError(error);
      return false;
    } finally {
      pending.current = null;
      setBusy(false);
    }
  }

  function select(group: Group | null) {
    setMessage('');
    setVerified(false);
    setMembers([]);
    setSelected(group);
  }
  function refresh() {
    setMessage('');
    setRevision((value) => value + 1);
  }
  return {
    groups,
    selected,
    members,
    ready,
    verified,
    loadingMembers,
    busy,
    message,
    setMessage,
    submit,
    select,
    refresh,
  };
}
