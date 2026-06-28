import { writable } from 'svelte/store';
import type { Role } from '$lib/types/models';

const STORAGE_KEY = 'koalanotes:viewingRole';

function getStored(): Role {
	if (typeof localStorage === 'undefined') return 'gm';
	const v = localStorage.getItem(STORAGE_KEY);
	if (v === 'gm' || v === 'player' || v === 'observer') return v;
	return 'gm';
}

/** The role the current user is viewing the campaign as. Persisted in localStorage. */
export const viewingRole = writable<Role>(getStored());

// Persist changes to localStorage
viewingRole.subscribe((v) => {
	if (typeof localStorage !== 'undefined') {
		localStorage.setItem(STORAGE_KEY, v);
	}
});
