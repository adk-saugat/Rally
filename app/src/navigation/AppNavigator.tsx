import { DefaultTheme, NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { ActivityIndicator, Pressable, StyleSheet, Text, View } from 'react-native';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { StatusBar } from 'expo-status-bar';
import HomeScreen from '../screens/HomeScreen';
import AuthScreen from '../screens/AuthScreen';
import MapScreen from '../screens/MapScreen';
import { SessionProvider, useSession } from '../session/SessionProvider';
import type { RootStackParamList } from './types';

const Stack = createNativeStackNavigator<RootStackParamList>();
const theme = {
  ...DefaultTheme,
  colors: {
    ...DefaultTheme.colors,
    primary: '#294E3B',
    background: '#F8F7F1',
    card: '#F8F7F1',
    text: '#233E33',
  },
};

function SessionNavigator() {
  const { session, restoring, notice } = useSession();
  if (restoring)
    return (
      <View style={styles.loading}>
        <ActivityIndicator color="#294E3B" accessibilityLabel="Restoring your login" />
      </View>
    );
  return (
    <NavigationContainer theme={theme}>
      <Stack.Navigator
        screenOptions={({ navigation }) => ({
          headerTitle: '',
          headerShadowVisible: false,
          headerStyle: { backgroundColor: '#F8F7F1' },
          contentStyle: { backgroundColor: '#F8F7F1' },
          headerBackVisible: false,
          headerLeft: ({ canGoBack }) =>
            canGoBack ? (
              <Pressable
                accessibilityRole="button"
                accessibilityLabel="Go back"
                onPress={() => navigation.goBack()}
                style={styles.back}
              >
                <Text style={styles.arrow}>←</Text>
              </Pressable>
            ) : null,
        })}
      >
        {session ? (
          <Stack.Screen
            name="Map"
            component={MapScreen}
            options={{ headerShown: false }}
            navigationKey={session.userID}
          />
        ) : (
          <>
            <Stack.Screen name="Welcome" component={HomeScreen} options={{ headerShown: false }} />
            <Stack.Screen
              name="Login"
              component={AuthScreen}
              initialParams={notice ? { notice } : undefined}
            />
            <Stack.Screen name="Signup" component={AuthScreen} />
          </>
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}

export default function AppNavigator() {
  return (
    <SafeAreaProvider>
      <StatusBar style="dark" />
      <SessionProvider>
        <SessionNavigator />
      </SessionProvider>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  loading: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: '#F8F7F1' },
  back: { minWidth: 48, minHeight: 48, alignItems: 'center', justifyContent: 'center' },
  arrow: { color: '#294E3B', fontSize: 30 },
});
