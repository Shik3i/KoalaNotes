import { db, uuid } from './database';
import { liveQuery } from 'dexie';
import type { Campaign } from '$lib/types/models';

/** Return an observable of all non-archived campaigns sorted by name. */
export function observeCampaigns() {
	return liveQuery(() =>
		db.campaigns
			.where('archived')
			.equals(0)
			.sortBy('name')
	);
}

/** Return an observable of a single campaign by id. */
export function observeCampaign(id: string) {
	return liveQuery(() => db.campaigns.get(id));
}

/** Create a new campaign and return its id. */
export async function createCampaign(
	name: string,
	description = '',
	system?: string
): Promise<string> {
	const now = new Date().toISOString();
	const campaign: Campaign = {
		id: uuid(),
		name: name.trim(),
		description,
		system,
		created_at: now,
		updated_at: now,
		archived: 0
	};
	await db.campaigns.add(campaign);
	return campaign.id;
}

/** Update campaign fields. Only provided fields are changed. */
export async function updateCampaign(
	id: string,
	changes: Partial<Pick<Campaign, 'name' | 'description' | 'system' | 'archived'>>
): Promise<void> {
	const update: Partial<Campaign> = { ...changes, updated_at: new Date().toISOString() };
	await db.campaigns.update(id, update);
}

/** Rename a campaign. */
export async function renameCampaign(id: string, newName: string): Promise<void> {
	await updateCampaign(id, { name: newName.trim() });
}

/** Archive a campaign (hide from default list). */
export async function archiveCampaign(id: string): Promise<void> {
	await updateCampaign(id, { archived: 1 });
}

/** Permanently delete a campaign and all its related data. */
export async function deleteCampaign(id: string): Promise<void> {
	const tables = [db.campaigns, db.notes, db.sessions, db.timeline_entries, db.tags, db.wiki_links, db.campaign_members] as const;
	await db.transaction('rw', tables, async () => {
		// Read note IDs BEFORE deleting notes, then delete wiki_links via indexed queries
		const noteIds = await db.notes.where('campaign_id').equals(id).primaryKeys();
		if (noteIds.length > 0) {
			const sourceKeys = await db.wiki_links.where('source_note_id').anyOf(noteIds).primaryKeys();
			const targetKeys = await db.wiki_links.where('target_note_id').anyOf(noteIds).primaryKeys();
			const seen = new Set<string>();
			const allKeys: [string, string][] = [];
			for (const k of [...sourceKeys, ...targetKeys]) {
				const key = `${k[0]}\0${k[1]}`;
				if (!seen.has(key)) { seen.add(key); allKeys.push(k); }
			}
			if (allKeys.length > 0) await db.wiki_links.bulkDelete(allKeys);
		}
		await db.campaigns.delete(id);
		await db.notes.where('campaign_id').equals(id).delete();
		await db.sessions.where('campaign_id').equals(id).delete();
		await db.timeline_entries.where('campaign_id').equals(id).delete();
		await db.tags.where('campaign_id').equals(id).delete();
		await db.campaign_members.where('campaign_id').equals(id).delete();
		await db.crypto_keys.where('campaign_id').equals(id).delete();
	});
}
