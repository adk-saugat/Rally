import { Platform } from 'react-native';

// Expo replaces direct EXPO_PUBLIC_* references when bundling the app.
// Values are public configuration, never secrets.
const configured = process.env.EXPO_PUBLIC_API_URL?.trim();
const fallback = Platform.OS === 'android' ? 'http://10.0.2.2:8080' : 'http://localhost:8080';

export const API_URL = (configured || fallback).replace(/\/+$/, '');
