import { db, uuid } from './database';
import { liveQuery } from 'dexie';
import type { WikiLink } from '$lib/types/models';

const WIKI_LINK_RE = /\[\[([^\]]+)\]\]/g;

/**
 * Parse note content, extract [[Wiki Links]], resolve them to note IDs,
 * and update the wiki_links table to reflect the current set of links.
 *
 * Links to note titles that don't exist yet are stored with empty
 * target_note_id for later resolution.
 */
export async function resolveWikiLinks(
	noteId: string,
	campaignId: string,
	content: string,
	title: string
): Promise<void> {
	const targets = new Set<string>();
	let match: RegExpExecArray | null;
	const re = new RegExp(WIKI_LINK_RE.source, 'g');

	while ((match = re.exec(content)) !== null) {
		targets.add(match[1].trim());
	}

	// Also consider the note's own title for self-links
	targets.delete(title);

	if (targets.size === 0) {
		// No wiki links — remove any existing links from this source
		await db.wiki_links.where('source_note_id').equals(noteId).delete();
		return;
	}

	// Resolve each target title to a note ID
	const linkData: Array<{ title: string; noteId: string | null }> = [];
	for (const targetTitle of targets) {
		const targetNote = await db.notes
			.where('campaign_id')
			.equals(campaignId)
			.filter((n) => n.title.toLowerCase() === targetTitle.toLowerCase())
			.first();
		linkData.push({ title: targetTitle, noteId: targetNote?.id ?? null });
	}

	// Rewrite all wiki_links for this source note
	await db.transaction('rw', db.wiki_links, async () => {
		// Remove existing links from this source
		await db.wiki_links.where('source_note_id').equals(noteId).delete();

		// Insert new links
		const now = new Date().toISOString();
		const links: WikiLink[] = linkData
			.map(({ title, noteId: targetId }) => ({
				id: uuid(),
				source_note_id: noteId,
				target_note_id: targetId ?? '',
				context: title,
				created_at: now
			}))
			.filter((l) => l.source_note_id !== l.target_note_id);

		if (links.length > 0) {
			await db.wiki_links.bulkAdd(links);
		}
	});
}

/**
 * Return an observable of backlinks for a given note
 * (wiki_links where target_note_id matches this note's id).
 */
export function observeBacklinks(noteId: string) {
	return liveQuery(() =>
		db.wiki_links
			.where('target_note_id')
			.equals(noteId)
			.toArray()
	);
}

/**
 * Return an observable of outgoing wiki links for a given note.
 */
export function observeOutgoingLinks(noteId: string) {
	return liveQuery(() =>
		db.wiki_links
			.where('source_note_id')
			.equals(noteId)
			.toArray()
	);
}

/**
 * Build a map of note IDs to titles for resolving backlink display.
 */
export async function getNoteTitleMap(noteIds: string[]): Promise<Map<string, string>> {
	const notes = await db.notes
		.filter(n => noteIds.includes(n.id))
		.toArray();
	const map = new Map<string, string>();
	for (const n of notes) {
		map.set(n.id, n.title);
	}
	return map;
}

/**
 * Update wiki link references when a note's title changes.
 * Finds all wiki_links targeting this note and re-resolves by context name.
 */
export async function retargetWikiLinks(
	noteId: string,
	campaignId: string,
	oldTitle: string,
	newTitle: string
): Promise<void> {
	if (oldTitle === newTitle) return;

	const linksToUpdate = await db.wiki_links
		.where('target_note_id')
		.equals(noteId)
		.toArray();

	await db.transaction('rw', db.wiki_links, async () => {
		for (const link of linksToUpdate) {
			const key: [string, string] = [link.source_note_id, link.target_note_id];
			if (link.context === oldTitle) {
				await db.wiki_links.update(key, { context: newTitle });
			}
		}
	});
}
