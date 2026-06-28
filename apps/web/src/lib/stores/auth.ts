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
	if (typeof localStorage === 'undefined') return;
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
	} catch (e) {
		console.warn('[auth] failed to persist auth state', e);
	}
}

function clearStored() {
	if (typeof localStorage === 'undefined') return;
	try {
		localStorage.removeItem(STORAGE_KEY);
	} catch { /* ignore */ }
}

/** Auth state: token, accountId, email. Persisted to localStorage. */
export const auth = writable<AuthState>(getStored());

// Sync auth state across tabs
if (typeof window !== 'undefined') {
	window.addEventListener('storage', (e) => {
		if (e.key === STORAGE_KEY) {
			auth.set(getStored());
		}
	});
}

let persistPending: AuthState | null = null;
let persistScheduled = false;
auth.subscribe((v) => {
	persistPending = v;
	if (!persistScheduled) {
		persistScheduled = true;
		queueMicrotask(() => {
			persistScheduled = false;
			const p = persistPending;
			persistPending = null;
			if (p?.token) persist(p);
			else clearStored();
		});
	}
});

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

interface AuthResponse {
	token: string;
	account_id: string;
	email: string;
}

/** Parse seconds-remaining from JWT token or -1 if unparseable. */
function tokenExpiresIn(token: string): number {
	try {
		const payload = JSON.parse(atob(token.split('.')[1]));
		if (payload.exp) {
			return Math.max(0, payload.exp * 1000 - Date.now());
		}
	} catch { /* ignore */ }
	return -1;
}

let expiryTimer: ReturnType<typeof setTimeout> | undefined;
auth.subscribe((v) => {
	if (expiryTimer) { clearTimeout(expiryTimer); expiryTimer = undefined; }
	if (v.token) {
		const ms = tokenExpiresIn(v.token);
		if (ms >= 0 && ms <= 86400000) {
			expiryTimer = setTimeout(() => {
				console.warn('[auth] token expired, logging out');
				auth.set({ token: null, accountId: null, email: null });
			}, ms);
		}
	}
});

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
