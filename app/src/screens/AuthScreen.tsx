import { useEffect, useRef, useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import Button from '../components/Button';
import FormField from '../components/FormField';
import ScreenLayout from '../components/ScreenLayout';
import type { ScreenProps } from '../navigation/types';
import { authenticate, type AuthMode } from '../services/auth';
import { useSession } from '../session/SessionProvider';

type AuthScreenProps = ScreenProps<'Login'> | ScreenProps<'Signup'>;

export default function AuthScreen({ navigation, route }: AuthScreenProps) {
  const { signIn, notice } = useSession();
  const mode: AuthMode = route.name === 'Signup' ? 'signup' : 'login';
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState(
    route.name === 'Login' ? (route.params?.notice ?? notice) : '',
  );
  const pending = useRef<AbortController | null>(null);

  useEffect(() => () => pending.current?.abort(), []);

  async function submit() {
    if (pending.current) return;
    if (!email.trim() || !password || (mode === 'signup' && !name.trim())) {
      setMessage('Please fill in all fields.');
      return;
    }
    const controller = new AbortController();
    pending.current = controller;
    setBusy(true);
    setMessage('');
    const timeout = setTimeout(() => controller.abort(), 10000);
    try {
      const token = await authenticate(mode, { name, email, password }, controller.signal);
      if (controller.signal.aborted) return;
      setPassword('');
      if (mode === 'signup') {
        navigation.replace('Login', { notice: 'Account created. Log in to continue.' });
      } else {
        if (!token) throw new Error('The server did not return a login token.');
        await signIn(token);
      }
    } catch (error) {
      if (navigation.isFocused()) {
        setMessage(
          error instanceof Error ? error.message : 'Unable to continue. Please try again.',
        );
      }
    } finally {
      clearTimeout(timeout);
      pending.current = null;
      setBusy(false);
    }
  }

  return (
    <ScreenLayout
      eyebrow="YOUR NEXT ADVENTURE STARTS HERE"
      title={mode === 'signup' ? 'Join the crew.' : 'Welcome back.'}
      description={
        mode === 'signup' ? 'Make a little room for your people.' : 'Good to have you here again.'
      }
    >
      {mode === 'signup' && (
        <FormField
          label="First name"
          editable={!busy}
          autoComplete="given-name"
          value={name}
          onChangeText={setName}
        />
      )}
      <FormField
        label="Email"
        editable={!busy}
        autoComplete="email"
        keyboardType="email-address"
        autoCapitalize="none"
        autoCorrect={false}
        value={email}
        onChangeText={setEmail}
      />
      <FormField
        label="Password"
        editable={!busy}
        autoComplete={mode === 'signup' ? 'new-password' : 'current-password'}
        secureTextEntry
        autoCapitalize="none"
        value={password}
        onChangeText={setPassword}
        onSubmitEditing={submit}
        returnKeyType="go"
      />
      <View style={styles.submit}>
        <Button
          title={mode === 'signup' ? 'Create account' : 'Log in'}
          onPress={submit}
          loading={busy}
        />
      </View>
      <Text accessibilityLiveRegion="polite" style={styles.message}>
        {message}
      </Text>
      <Pressable
        disabled={busy}
        accessibilityRole="button"
        style={styles.switchMode}
        onPress={() =>
          mode === 'signup' ? navigation.replace('Login') : navigation.replace('Signup')
        }
      >
        <Text style={styles.link}>
          {mode === 'signup' ? 'Already have an account? Log in' : 'New to Rally? Sign up'}
        </Text>
      </Pressable>
    </ScreenLayout>
  );
}

const styles = StyleSheet.create({
  submit: { marginTop: 28 },
  message: { marginTop: 18, color: '#435B45', fontSize: 14, lineHeight: 22 },
  switchMode: { minHeight: 48, paddingVertical: 18, alignItems: 'center' },
  link: { color: '#294E3B', fontWeight: '600', fontSize: 14, textAlign: 'center' },
});
