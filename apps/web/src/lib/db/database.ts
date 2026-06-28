import Dexie, { type Table } from 'dexie';
import type {
	Campaign,
	Note,
	Session,
	TimelineEntry,
	Template,
	Tag,
	WikiLink,
	CampaignMember,
	CryptoKeyRecord
} from '$lib/types/models';

export class KoalaDB extends Dexie {
	campaigns!: Table<Campaign, string>;
	notes!: Table<Note, string>;
	sessions!: Table<Session, string>;
	timeline_entries!: Table<TimelineEntry, string>;
	templates!: Table<Template, string>;
	tags!: Table<Tag, string>;
	wiki_links!: Table<WikiLink, [string, string]>;
	campaign_members!: Table<CampaignMember, string>;
	crypto_keys!: Table<CryptoKeyRecord, string>;

	constructor() {
		super('koalanotes');

		this.version(1).stores({
			campaigns: 'id, name, created_at, updated_at, archived',
			notes: 'id, campaign_id, title, template_type, created_at, updated_at, *tags, pinned',
			sessions: 'id, campaign_id, status, started_at, session_number',
			timeline_entries: 'id, campaign_id, session_id, note_id, clock_time, session_elapsed',
			templates: 'id, type',
			tags: 'id, name, campaign_id',
			wiki_links: '[source_note_id+target_note_id], source_note_id, target_note_id, created_at',
			campaign_members: 'id, campaign_id, user_id, role'
		});

		this.version(2).stores({
			campaigns: 'id, name, created_at, updated_at, archived',
			notes: 'id, campaign_id, title, template_type, created_at, updated_at, *tags, pinned',
			sessions: 'id, campaign_id, status, started_at, session_number',
			timeline_entries: 'id, campaign_id, session_id, note_id, clock_time, session_elapsed',
			templates: 'id, type',
			tags: 'id, name, campaign_id',
			wiki_links: '[source_note_id+target_note_id], source_note_id, target_note_id, created_at',
			campaign_members: 'id, campaign_id, user_id, role',
			crypto_keys: 'id, campaign_id'
		});
	}
}

export const db = new KoalaDB();

/** Generate a UUID v4 string. */
export function uuid(): string {
	return crypto.randomUUID();
}
