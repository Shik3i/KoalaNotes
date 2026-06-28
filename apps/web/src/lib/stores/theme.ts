import { writable } from 'svelte/store';

const STORAGE_KEY = 'koalanotes:theme';
const MEDIA_QUERY = '(prefers-color-scheme: light)';

export type Theme = 'light' | 'dark';

function getStored(): Theme {
	if (typeof localStorage === 'undefined' || typeof window === 'undefined') return 'dark';
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (raw === 'light' || raw === 'dark') return raw;
	} catch { /* ignore */ }
	return window.matchMedia(MEDIA_QUERY).matches ? 'light' : 'dark';
}

function persist(theme: Theme) {
	if (typeof localStorage === 'undefined') return;
	try {
		localStorage.setItem(STORAGE_KEY, theme);
	} catch { /* ignore */ }
}

function apply(theme: Theme) {
	if (typeof document === 'undefined') return;
	document.documentElement.classList.toggle('light', theme === 'light');
	document.documentElement.classList.toggle('dark', theme === 'dark');
}

const initial = getStored();
apply(initial);

export const theme = writable<Theme>(initial);

theme.subscribe((val) => {
	apply(val);
	persist(val);
});

/** Toggle between light and dark theme. */
export function toggleTheme(): void {
	theme.update((t) => (t === 'light' ? 'dark' : 'light'));
}
