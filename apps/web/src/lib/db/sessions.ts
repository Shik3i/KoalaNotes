import { db, uuid } from './database';
import { liveQuery } from 'dexie';
import type { Note, Session, TimelineEntry } from '$lib/types/models';

/** Observe the single active session (status = 'active'), if any. */
export function observeActiveSession() {
	return liveQuery(() =>
		db.sessions.where('status').equals('active').first()
	);
}

/** Observe all sessions for a campaign, newest first. */
export function observeSessionsByCampaign(campaignId: string) {
	return liveQuery(() =>
		db.sessions
			.where('campaign_id')
			.equals(campaignId)
			.reverse()
			.sortBy('started_at')
	);
}

/** Get a session by id. */
export async function getSession(id: string): Promise<Session | undefined> {
	return db.sessions.get(id);
}

/** Start a new session for a campaign. Returns the session id. */
export async function startSession(campaignId: string, name?: string): Promise<string> {
	const now = new Date().toISOString();

	// End any currently active session first
	const active = await db.sessions.where('status').equals('active').first();
	if (active) {
		await db.sessions.update(active.id, { status: 'completed', ended_at: now });
	}

	// Determine next session number
	const existing = await db.sessions
		.where('campaign_id')
		.equals(campaignId)
		.count();

	const session: Session = {
		id: uuid(),
		campaign_id: campaignId,
		name: name || `Session ${existing + 1}`,
		session_number: existing + 1,
		status: 'active',
		started_at: now,
		created_at: now,
		updated_at: now
	};

	await db.sessions.add(session);
	return session.id;
}

/** Stop an active session. */
export async function stopSession(id: string): Promise<void> {
	await db.sessions.update(id, {
		status: 'completed',
		ended_at: new Date().toISOString(),
		updated_at: new Date().toISOString()
	});
}

/** Delete a session and its timeline entries. */
export async function deleteSession(id: string): Promise<void> {
	await db.transaction('rw', [db.sessions, db.timeline_entries], async () => {
		await db.sessions.delete(id);
		await db.timeline_entries.where('session_id').equals(id).delete();
	});
}

/** Format elapsed seconds as HH:MM:SS for recap text. */
function formatElapsed(seconds: number): string {
	const h = Math.floor(seconds / 3600);
	const m = Math.floor((seconds % 3600) / 60);
	const s = seconds % 60;
	return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
}

/**
 * Create a recap note from a completed session's timeline entries.
 * The note is pre-filled with compiled entries and linked via recap_note_id.
 * Returns the new note's id.
 */
export async function createSessionRecap(sessionId: string): Promise<string> {
	const session = await db.sessions.get(sessionId);
	if (!session) throw new Error('Session not found');

	const entries: TimelineEntry[] = await db.timeline_entries
		.where('session_id')
		.equals(sessionId)
		.sortBy('clock_time');

	// Resolve note titles for wiki link formatting
	const noteIds = [...new Set(entries.map(e => e.note_id).filter(Boolean) as string[])];
	const notes: Note[] = noteIds.length > 0
		? await db.notes.filter(n => noteIds.includes(n.id)).toArray()
		: [];
	const titleMap = new Map(notes.map(n => [n.id, n.title]));

	// Build the "Last Time…" section as a bullet list
	const lines: string[] = [];
	lines.push(`## Last Time…`);
	lines.push('');

	for (const entry of entries) {
		const elapsed = formatElapsed(entry.session_elapsed);
		let line = `- \`[${elapsed}]\` ${entry.content}`;
		if (entry.note_id && titleMap.has(entry.note_id)) {
			line += `  *(via [[${titleMap.get(entry.note_id)}]])*`;
		}
		lines.push(line);
	}

	lines.push('');
	lines.push(`## Current Situation`);
	lines.push('');
	lines.push('*Write a brief summary of where things stand...*');
	lines.push('');
	lines.push(`## Open Threads`);
	lines.push('');
	lines.push('*List any unresolved plot points...*');

	const content = lines.join('\n');
	const now = new Date().toISOString();

	const recapNote: Note = {
		id: uuid(),
		campaign_id: session.campaign_id,
		title: `Recap: ${session.name}`,
		content,
		template_type: 'session_recap',
		tags: ['recap'],
		sections: [],
		created_at: now,
		updated_at: now,
		pinned: false
	};

	await db.transaction('rw', [db.notes, db.sessions, db.wiki_links], async () => {
		await db.notes.add(recapNote);
		await db.sessions.update(sessionId, { recap_note_id: recapNote.id, updated_at: now });
		// Resolve wiki links within the recap content
		const { resolveWikiLinks } = await import('./wiki');
		await resolveWikiLinks(recapNote.id, session.campaign_id, content, recapNote.title);
	});

	return recapNote.id;
}
