import { writable } from 'svelte/store';

export interface AuthState {
	token: string | null;
	accountId: string | null;
	email: string | null;
}

const STORAGE_KEY = 'koalanotes:auth';

function getStored(): AuthState {
	if (typeof localStorage === 'undefined') return { token: null, accountId: null, email: null };
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (raw) return JSON.parse(raw);
	} catch { /* ignore */ }
	return { token: null, accountId: null, email: null };
}

function persist(state: AuthState) {
	if (typeof localStorage !== 'undefined') {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
	}
}

function clearStored() {
	if (typeof localStorage !== 'undefined') {
		localStorage.removeItem(STORAGE_KEY);
	}
}

/** Auth state: token, accountId, email. Persisted to localStorage. */
export const auth = writable<AuthState>(getStored());

auth.subscribe((v) => {
	if (v.token) {
		persist(v);
	} else {
		clearStored();
	}
});

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

interface AuthResponse {
	token: string;
	account_id: string;
	email: string;
}

/** Register a new account and update the auth store. */
export async function register(email: string, password: string): Promise<void> {
	const res = await fetch(`${API_BASE}/api/auth/register`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, password })
	});

	if (!res.ok) {
		const err = await res.json().catch(() => ({ error: 'Registration failed' }));
		throw new Error(err.error || 'Registration failed');
	}

	const data: AuthResponse = await res.json();
	auth.set({ token: data.token, accountId: data.account_id, email: data.email });
}

/** Login with email and password, update the auth store. */
export async function login(email: string, password: string): Promise<void> {
	const res = await fetch(`${API_BASE}/api/auth/login`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, password })
	});

	if (!res.ok) {
		const err = await res.json().catch(() => ({ error: 'Login failed' }));
		throw new Error(err.error || 'Login failed');
	}

	const data: AuthResponse = await res.json();
	auth.set({ token: data.token, accountId: data.account_id, email: data.email });
}

/** Logout: clear auth state. */
export function logout(): void {
	auth.set({ token: null, accountId: null, email: null });
}
