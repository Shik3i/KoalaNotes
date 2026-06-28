/**
 * Sync service — push/pull encrypted campaign data to/from the server.
 *
 * Flow:
 *   push: serialize all campaigns → encrypt each as blob → upload
 *   pull: download blobs → decrypt → merge into IndexedDB
 */

import { auth } from '$lib/stores/auth';
import { get as getStore } from 'svelte/store';
import { db } from '$lib/db/database';
import type { Campaign, Note, Session, TimelineEntry, CampaignMember } from '$lib/types/models';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const LAST_SYNC_KEY = 'koalanotes:last_sync';

interface BlobRecord {
	id: string;
	campaign_key_id: string;
	encrypted_payload: string;
	vector_clock: string;
}

interface PushResponse {
	accepted: number;
}

interface PullResponse {
	blobs: BlobRecord[];
}

// ---- Concurrency lock ----

let lock: Promise<void> = Promise.resolve();

async function withLock<T>(fn: () => Promise<T>): Promise<T> {
	const prev = lock;
	let release: () => void;
	lock = new Promise(resolve => { release = resolve; });
	await prev;
	try {
		return await fn();
	} finally {
		release!();
	}
}

// ---- Last-sync timestamp ----

function getLastSync(): string {
	try {
		return localStorage.getItem(LAST_SYNC_KEY) || '';
	} catch {
		return '';
	}
}

function setLastSync(ts: string) {
	try {
		localStorage.setItem(LAST_SYNC_KEY, ts);
	} catch { /* ignore */ }
}

// ---- Auto-sync registration ----

type AutoSyncFn = () => Promise<void>;
let registeredAutoSync: AutoSyncFn | null = null;

/** Register a global auto-sync function (called by settings page when keys are unlocked). */
export function registerAutoSync(fn: AutoSyncFn): void {
	registeredAutoSync = fn;
}

/** Unregister the auto-sync function (called on logout / key clear). */
export function unregisterAutoSync(): void {
	registeredAutoSync = null;
}

/** Trigger auto-sync if registered. Safe to call after any note/session mutation. */
export async function triggerAutoSync(): Promise<void> {
	if (registeredAutoSync) {
		try {
			await registeredAutoSync();
		} catch (err) {
			console.error('[sync] auto-sync failed', err);
		}
	}
}

// ---- Core sync operations ----

/** Full sync: pull then push. Acquires concurrency lock. */
export async function fullSync(
	encrypt: (plaintext: string) => Promise<{ iv: string; ciphertext: string }>,
	decrypt: (payload: { iv: string; ciphertext: string }) => Promise<string>
): Promise<{ pulled: number; pushed: number }> {
	return withLock(async () => {
		const pulled = await pull(decrypt);
		const pushed = await push(encrypt);
		return { pulled, pushed };
	});
}

/** Pull blobs from the server, decrypt, and merge into local DB. Acquires concurrency lock. */
export async function pull(
	decrypt: (payload: { iv: string; ciphertext: string }) => Promise<string>
): Promise<number> {
	return withLock(async () => {
		const state = getStore(auth);
		if (!state.token) throw new Error('Not authenticated');

		const controller = new AbortController();
		const timeout = setTimeout(() => controller.abort(), 30000);

		const since = getLastSync();
		const url = since
			? `${API_BASE}/api/sync/pull?since=${encodeURIComponent(since)}`
			: `${API_BASE}/api/sync/pull`;

		let res: Response;
		try {
			res = await fetch(url, {
				headers: { 'Authorization': `Bearer ${state.token}` },
				signal: controller.signal
			});
		} finally {
			clearTimeout(timeout);
		}

		if (!res.ok) {
			if (res.status === 401) auth.set({ token: null, accountId: null, email: null });
			const err = await res.json().catch(() => ({ error: 'Pull failed' }));
			throw new Error(err.error || 'Pull failed');
		}

		const data = await res.json();
		if (!data || !Array.isArray(data.blobs) || data.blobs.length === 0) return 0;

		// Track latest vector_clock for incremental sync
		let latestVectorClock = since;

		// Process each blob and collect all entities, then batch-write in a single transaction
		const campaigns: Campaign[] = [];
		const notes: Note[] = [];
		const sessions: Session[] = [];
		const timeline_entries: TimelineEntry[] = [];
		const members: CampaignMember[] = [];

		for (const blob of data.blobs) {
			try {
				const payload = JSON.parse(blob.encrypted_payload);
				const plaintext = await decrypt({ iv: payload.iv, ciphertext: payload.ciphertext });
				const campaignData = JSON.parse(plaintext);

				if (campaignData.campaign) campaigns.push(campaignData.campaign);
				if (campaignData.notes) notes.push(...campaignData.notes);
				if (campaignData.sessions) sessions.push(...campaignData.sessions);
				if (campaignData.timeline_entries) timeline_entries.push(...campaignData.timeline_entries);
				if (campaignData.members) members.push(...campaignData.members);

				// Track newest timestamp
				if (blob.vector_clock > latestVectorClock) {
					latestVectorClock = blob.vector_clock;
				}
			} catch (err) {
				console.error('[sync] failed to process blob', blob.id, err);
			}
		}

		// Batch-write all entities in a single readwrite transaction
		if (campaigns.length > 0 || notes.length > 0 || sessions.length > 0 || timeline_entries.length > 0 || members.length > 0) {
			const tables = [db.campaigns, db.notes, db.sessions, db.timeline_entries, db.campaign_members] as const;
			await db.transaction('rw', tables, async () => {
				if (campaigns.length > 0) await db.campaigns.bulkPut(campaigns);
				if (notes.length > 0) await db.notes.bulkPut(notes);
				if (sessions.length > 0) await db.sessions.bulkPut(sessions);
				if (timeline_entries.length > 0) await db.timeline_entries.bulkPut(timeline_entries);
				if (members.length > 0) await db.campaign_members.bulkPut(members);
			});
		}

		// Persist latest vector_clock for next incremental pull
		if (latestVectorClock) {
			setLastSync(latestVectorClock);
		}

		return campaigns.length;
	});
}

/** Push all local campaign data to the server as encrypted blobs. Acquires concurrency lock. */
export async function push(
	encrypt: (plaintext: string) => Promise<{ iv: string; ciphertext: string }>
): Promise<number> {
	return withLock(async () => {
		const state = getStore(auth);
		if (!state.token) throw new Error('Not authenticated');

		// Gather all campaigns
		const campaigns = await db.campaigns.toArray();
		const blobs: BlobRecord[] = [];

		for (const campaign of campaigns) {
			const [notes, sessions, timeline_entries, members] = await Promise.all([
				db.notes.where('campaign_id').equals(campaign.id).toArray(),
				db.sessions.where('campaign_id').equals(campaign.id).toArray(),
				db.timeline_entries.where('campaign_id').equals(campaign.id).toArray(),
				db.campaign_members.where('campaign_id').equals(campaign.id).toArray(),
			]);

			const campaignData = {
				campaign,
				notes,
				sessions,
				timeline_entries,
				members
			};

			const plaintext = JSON.stringify(campaignData);
			const encrypted = await encrypt(plaintext);

			blobs.push({
				id: campaign.id,
				campaign_key_id: campaign.id,
				encrypted_payload: JSON.stringify(encrypted),
				vector_clock: new Date().toISOString()
			});
		}

		if (blobs.length === 0) return 0;

		const controller = new AbortController();
		const timeout = setTimeout(() => controller.abort(), 30000);

		let res: Response;
		try {
			res = await fetch(`${API_BASE}/api/sync/push`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${state.token}`
				},
				body: JSON.stringify({ blobs }),
				signal: controller.signal
			});
		} finally {
			clearTimeout(timeout);
		}

		if (!res.ok) {
			if (res.status === 401) auth.set({ token: null, accountId: null, email: null });
			const err = await res.json().catch(() => ({ error: 'Push failed' }));
			throw new Error(err.error || 'Push failed');
		}

		const data = await res.json();

		// Update last sync timestamp on successful push
		setLastSync(new Date().toISOString());

		return data?.accepted ?? 0;
	});
}
