/** Core data model types for KoalaNotes. See docs/DATA_MODEL.md for details. */

export type Visibility = 'gm_only' | 'shared' | 'observer' | 'private';

export type TemplateType =
	| 'blank'
	| 'npc'
	| 'location'
	| 'quest'
	| 'item'
	| 'faction'
	| 'session'
	| 'session_recap';

export type Role = 'gm' | 'player' | 'observer';

export type SessionStatus = 'planned' | 'active' | 'completed';

export interface Campaign {
	id: string;
	name: string;
	description: string;
	system?: string;
	created_at: string;
	updated_at: string;
	archived: boolean;
}

export interface Note {
	id: string;
	campaign_id: string;
	title: string;
	content: string;
	template_type?: TemplateType;
	tags: string[];
	sections: NoteSection[];
	created_at: string;
	updated_at: string;
	pinned: boolean;
}

export interface NoteSection {
	id: string;
	heading: string;
	content: string;
	visibility: Visibility;
	order: number;
}

export interface Session {
	id: string;
	campaign_id: string;
	name: string;
	session_number: number;
	status: SessionStatus;
	started_at?: string;
	ended_at?: string;
	planned_date?: string;
	recap_note_id?: string;
	created_at: string;
	updated_at: string;
}

export interface TimelineEntry {
	id: string;
	campaign_id: string;
	session_id: string;
	note_id?: string;
	content: string;
	clock_time: string;
	session_elapsed: number; // seconds
	tags: string[];
	pinned: boolean;
	created_at: string;
}

export interface Template {
	id: string;
	type: TemplateType;
	name: string;
	description: string;
	sections: TemplateSection[];
}

export interface TemplateSection {
	heading: string;
	placeholder: string;
	default_visibility: Visibility;
	order: number;
}

export interface Tag {
	id: string;
	name: string;
	color?: string;
	campaign_id?: string;
	usage_count: number;
}

export interface WikiLink {
	id: string;
	source_note_id: string;
	target_note_id: string;
	context: string;
	created_at: string;
}

export interface CampaignMember {
	id: string;
	campaign_id: string;
	user_id?: string;
	display_name: string;
	role: Role;
	joined_at: string;
}
