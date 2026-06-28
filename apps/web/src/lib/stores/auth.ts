import { writable } from 'svelte/store';

export interface AuthState {
	token: string | null;
	accountId: string | null;
	email: string | null;
	locked: boolean;
}

const STORAGE_KEY = 'koalanotes:auth';
const BROADCAST_CHANNEL = 'koalanotes:auth';
const IDLE_TIMEOUT_MS = 15 * 60 * 1000; // 15 minutes

function getStored(): AuthState {
	if (typeof localStorage === 'undefined') return { token: null, accountId: null, email: null, locked: true };
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (raw) {
			const parsed = JSON.parse(raw);
			return {
				token: parsed.token ?? null,
				accountId: parsed.accountId ?? null,
				email: parsed.email ?? null,
				locked: parsed.locked ?? true
			};
		}
	} catch { /* ignore */ }
	return { token: null, accountId: null, email: null, locked: true };
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

/** Auth state: token, accountId, email, locked. Persisted to localStorage. */
export const auth = writable<AuthState>(getStored());

// ---- BroadcastChannel cross-tab sync ----

const bc = typeof BroadcastChannel !== 'undefined' ? new BroadcastChannel(BROADCAST_CHANNEL) : null;

if (bc) {
	bc.onmessage = (event) => {
		const data = event.data;
		if (data?.type === 'auth') {
			auth.set({
				token: data.token ?? null,
				accountId: data.accountId ?? null,
				email: data.email ?? null,
				locked: data.locked ?? true
			});
		}
	};
}

// Fallback: storage event for older browsers
if (typeof window !== 'undefined') {
	window.addEventListener('storage', (e) => {
		if (e.key === STORAGE_KEY) {
			auth.set(getStored());
		}
	});
}

function broadcast(state: AuthState) {
	if (bc) {
		bc.postMessage({ type: 'auth', ...state });
	}
}

// ---- Persist on subscribe (batched via queueMicrotask) ----

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
				auth.set({ token: null, accountId: null, email: null, locked: true });
			}, ms);
		}
	}
});

// ---- Idle timeout (auto-lock) ----

let idleTimer: ReturnType<typeof setTimeout> | undefined;
let activityHandler: (() => void) | undefined;

function resetIdleTimer() {
	if (idleTimer) { clearTimeout(idleTimer); idleTimer = undefined; }
	const current = getStored();
	if (!current.token || current.locked) return;
	idleTimer = setTimeout(() => {
		console.warn('[auth] idle timeout — locking key');
		auth.update((prev) => ({ ...prev, locked: true }));
	}, IDLE_TIMEOUT_MS);
}

function handleActivity() {
	resetIdleTimer();
}

function startIdleTracking() {
	if (typeof document === 'undefined' || activityHandler) return;
	activityHandler = handleActivity;
	document.addEventListener('mousedown', handleActivity, { passive: true });
	document.addEventListener('keydown', handleActivity, { passive: true });
	document.addEventListener('touchstart', handleActivity, { passive: true });
	document.addEventListener('scroll', handleActivity, { passive: true });
	resetIdleTimer();
}

function stopIdleTracking() {
	if (idleTimer) { clearTimeout(idleTimer); idleTimer = undefined; }
	if (typeof document === 'undefined' || !activityHandler) return;
	document.removeEventListener('mousedown', activityHandler);
	document.removeEventListener('keydown', activityHandler);
	document.removeEventListener('touchstart', activityHandler);
	document.removeEventListener('scroll', activityHandler);
	activityHandler = undefined;
}

// Start/stop idle tracking based on lock state
auth.subscribe((v) => {
	if (v.token && !v.locked) {
		startIdleTracking();
	} else {
		stopIdleTracking();
	}
});

/** Update lock state and broadcast to other tabs. */
export function setLocked(locked: boolean): void {
	auth.update((prev) => {
		const next = { ...prev, locked };
		broadcast(next);
		return next;
	});
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
	const state: AuthState = { token: data.token, accountId: data.account_id, email: data.email, locked: true };
	auth.set(state);
	broadcast(state);
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
	const state: AuthState = { token: data.token, accountId: data.account_id, email: data.email, locked: true };
	auth.set(state);
	broadcast(state);
}

/** Logout: clear auth state and lock key. */
export function logout(): void {
	const state: AuthState = { token: null, accountId: null, email: null, locked: true };
	auth.set(state);
	broadcast(state);
}
