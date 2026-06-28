import { db, uuid } from './database';
import { resolveWikiLinks, retargetWikiLinks } from './wiki';
import type { Note, TemplateType } from '$lib/types/models';

/** Create a new note. If no sections provided, starts with a single blank section. */
export async function createNote(
	campaign_id: string,
	title: string,
	content = '',
	template_type?: TemplateType,
	tags: string[] = []
): Promise<string> {
	const trimmed = title.trim() || 'Untitled';
	const now = new Date().toISOString();
	const note: Note = {
		id: uuid(),
		campaign_id,
		title: trimmed,
		title_lower: trimmed.toLowerCase(),
		content,
		template_type,
		tags,
		sections: [],
		created_at: now,
		updated_at: now,
		pinned: false
	};
	await db.notes.add(note);
	// Resolve wiki links for the new content
	await resolveWikiLinks(note.id, campaign_id, content, note.title);
	return note.id;
}

/** Update note fields and re-resolve wiki links if content or title changed. */
export async function updateNote(
	id: string,
	changes: Partial<Pick<Note, 'title' | 'content' | 'template_type' | 'tags' | 'pinned' | 'sections'>>
): Promise<void> {
	const old = await db.notes.get(id);
	if (!old) return;

	await db.transaction('rw', [db.notes, db.wiki_links], async () => {
		const payload: Record<string, unknown> = { ...changes, updated_at: new Date().toISOString() };
		if (changes.title !== undefined) {
			payload.title_lower = changes.title.trim().toLowerCase();
		}
		await db.notes.update(id, payload as any);
	});

	// Re-resolve wiki links if content changed
	if (changes.content !== undefined) {
		const newTitle = changes.title ?? old.title;
		await resolveWikiLinks(id, old.campaign_id, changes.content, newTitle);
	}

	// Retarget existing wiki links if title changed
	if (changes.title !== undefined && changes.title !== old.title) {
		await retargetWikiLinks(id, old.campaign_id, old.title, changes.title);
	}
}

/** Get a single note by id. */
export async function getNote(id: string): Promise<Note | undefined> {
	return db.notes.get(id);
}

/** Get all notes for a campaign, sorted by title using compound index. */
export async function getNotesByCampaign(campaign_id: string): Promise<Note[]> {
	return db.notes
		.where('[campaign_id+title]')
		.between([campaign_id, ''], [campaign_id, '\uffff'])
		.toArray();
}

/** Delete a note, its wiki links, and orphaned timeline entries. */
export async function deleteNote(id: string): Promise<void> {
	await db.transaction('rw', [db.notes, db.wiki_links, db.timeline_entries], async () => {
		await db.notes.delete(id);
		// Use indexed queries for wiki_links (source_note_id and target_note_id are indexed)
		const sourceKeys = await db.wiki_links.where('source_note_id').equals(id).primaryKeys();
		const targetKeys = await db.wiki_links.where('target_note_id').equals(id).primaryKeys();
		const seen = new Set<string>();
		const allKeys: [string, string][] = [];
		for (const k of [...sourceKeys, ...targetKeys]) {
			const key = `${k[0]}\0${k[1]}`;
			if (!seen.has(key)) { seen.add(key); allKeys.push(k); }
		}
		if (allKeys.length > 0) await db.wiki_links.bulkDelete(allKeys);
		await db.timeline_entries.where('note_id').equals(id).delete();
	});
}
