import { db, uuid } from './database';
import { liveQuery } from 'dexie';
import type { TimelineEntry } from '$lib/types/models';

/** Observe all timeline entries for a session, oldest first. */
export function observeSessionEntries(sessionId: string) {
	return liveQuery(() =>
		db.timeline_entries
			.where('[session_id+clock_time]')
			.between([sessionId, ''], [sessionId, '\uffff'])
			.toArray()
	);
}

/** Observe all timeline entries for a campaign, newest first. */
export function observeCampaignEntries(campaignId: string) {
	return liveQuery(() =>
		db.timeline_entries
			.where('[campaign_id+clock_time]')
			.between([campaignId, ''], [campaignId, '\uffff'])
			.reverse()
			.toArray()
	);
}

/** Create a new timeline entry. Returns the entry id. */
export async function createEntry(
	sessionId: string,
	campaignId: string,
	content: string,
	noteId?: string,
	sessionElapsed?: number
): Promise<string> {
	const now = new Date().toISOString();
	const entry: TimelineEntry = {
		id: uuid(),
		campaign_id: campaignId,
		session_id: sessionId,
		note_id: noteId,
		content: content.trim(),
		clock_time: now,
		session_elapsed: sessionElapsed ?? 0,
		tags: [],
		pinned: false,
		created_at: now
	};
	await db.timeline_entries.add(entry);
	return entry.id;
}

/** Delete a single timeline entry. */
export async function deleteEntry(id: string): Promise<void> {
	await db.timeline_entries.delete(id);
}
