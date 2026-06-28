import { db, uuid } from './database';
import type { Template, TemplateType } from '$lib/types/models';

/** Built-in TTRPG note templates. Seeded on first use. */
const BUILT_IN_TEMPLATES: Omit<Template, 'id'>[] = [
	{
		type: 'blank',
		name: 'Blank Page',
		description: 'A blank page for freeform notes.',
		sections: [
			{ heading: 'Notes', placeholder: 'Write anything...', default_visibility: 'shared', order: 0 }
		]
	},
	{
		type: 'npc',
		name: 'NPC',
		description: 'Track a non-player character.',
		sections: [
			{ heading: 'Description', placeholder: 'Physical appearance, demeanor, etc.', default_visibility: 'gm_only', order: 0 },
			{ heading: 'Personality', placeholder: 'Traits, ideals, bonds, flaws', default_visibility: 'gm_only', order: 1 },
			{ heading: 'Motivations', placeholder: 'What does this NPC want?', default_visibility: 'gm_only', order: 2 },
			{ heading: 'Relationships', placeholder: 'Links to other NPCs, factions, or PCs', default_visibility: 'gm_only', order: 3 },
			{ heading: 'Notes', placeholder: 'Session notes, rumors, secrets', default_visibility: 'shared', order: 4 }
		]
	},
	{
		type: 'location',
		name: 'Location',
		description: 'Describe a place in your world.',
		sections: [
			{ heading: 'Description', placeholder: 'What does this place look like?', default_visibility: 'shared', order: 0 },
			{ heading: 'Notable Features', placeholder: 'Landmarks, architecture, atmosphere', default_visibility: 'shared', order: 1 },
			{ heading: 'Inhabitants', placeholder: 'Who or what lives here?', default_visibility: 'shared', order: 2 },
			{ heading: 'Secrets', placeholder: 'Hidden knowledge (GM only)', default_visibility: 'gm_only', order: 3 },
			{ heading: 'Adventures', placeholder: 'Events that happened here', default_visibility: 'shared', order: 4 }
		]
	},
	{
		type: 'quest',
		name: 'Quest',
		description: 'Outline an adventure or mission.',
		sections: [
			{ heading: 'Summary', placeholder: 'Brief overview of the quest', default_visibility: 'shared', order: 0 },
			{ heading: 'Hook', placeholder: 'How do the players learn about this?', default_visibility: 'gm_only', order: 1 },
			{ heading: 'Objectives', placeholder: 'What must be done to complete it?', default_visibility: 'shared', order: 2 },
			{ heading: 'Rewards', placeholder: 'Treasure, XP, reputation, etc.', default_visibility: 'shared', order: 3 },
			{ heading: 'GM Notes', placeholder: 'Behind-the-scenes details', default_visibility: 'gm_only', order: 4 }
		]
	},
	{
		type: 'item',
		name: 'Item',
		description: 'Catalog a noteworthy item, artifact, or piece of equipment.',
		sections: [
			{ heading: 'Description', placeholder: 'Appearance, weight, materials', default_visibility: 'shared', order: 0 },
			{ heading: 'Properties', placeholder: 'Magical effects, special abilities', default_visibility: 'shared', order: 1 },
			{ heading: 'History', placeholder: 'Origin and known lore', default_visibility: 'gm_only', order: 2 },
			{ heading: 'Current Owner', placeholder: 'Who has it now?', default_visibility: 'shared', order: 3 }
		]
	},
	{
		type: 'faction',
		name: 'Faction',
		description: 'Document an organization, guild, or group.',
		sections: [
			{ heading: 'Overview', placeholder: 'Purpose, size, influence', default_visibility: 'shared', order: 0 },
			{ heading: 'Leadership', placeholder: 'Key members and hierarchy', default_visibility: 'shared', order: 1 },
			{ heading: 'Goals', placeholder: 'What does the faction want?', default_visibility: 'gm_only', order: 2 },
			{ heading: 'Resources', placeholder: 'Wealth, military, political power', default_visibility: 'gm_only', order: 3 },
			{ heading: 'Relations', placeholder: 'Allies, enemies, neutral parties', default_visibility: 'shared', order: 4 }
		]
	},
	{
		type: 'session',
		name: 'Session',
		description: 'Plan and track a specific play session.',
		sections: [
			{ heading: 'Summary', placeholder: 'What happened?', default_visibility: 'shared', order: 0 },
			{ heading: 'Prep Notes', placeholder: 'Scenes, encounters, NPCs to introduce', default_visibility: 'gm_only', order: 1 },
			{ heading: 'Player Actions', placeholder: 'Key decisions and outcomes', default_visibility: 'shared', order: 2 },
			{ heading: 'Loot & XP', placeholder: 'Treasure and experience awarded', default_visibility: 'shared', order: 3 },
			{ heading: 'Next Session', placeholder: 'Cliffhangers and hooks for next time', default_visibility: 'gm_only', order: 4 }
		]
	},
	{
		type: 'session_recap',
		name: 'Session Recap',
		description: 'A player-facing summary of the previous session.',
		sections: [
			{ heading: 'Last Time…', placeholder: 'One-paragraph recap of the session', default_visibility: 'shared', order: 0 },
			{ heading: 'Current Situation', placeholder: 'Where does the party stand now?', default_visibility: 'shared', order: 1 },
			{ heading: 'Open Threads', placeholder: 'Unresolved plot points', default_visibility: 'shared', order: 2 }
		]
	}
];

/** Seed built-in templates if the templates table is empty. Returns true if seeded. */
export async function seedTemplates(): Promise<boolean> {
	const count = await db.templates.count();
	if (count > 0) return false;

	const now = new Date().toISOString();
	const templates: Template[] = BUILT_IN_TEMPLATES.map((t) => ({
		...t,
		id: uuid()
	}));

	await db.templates.bulkAdd(templates);
	return true;
}

/** Return all templates. */
export async function getAllTemplates(): Promise<Template[]> {
	return db.templates.toArray();
}

/** Return templates by type. */
export async function getTemplatesByType(type: TemplateType): Promise<Template[]> {
	return db.templates.where('type').equals(type).toArray();
}

/** Return a single template by id. */
export async function getTemplate(id: string): Promise<Template | undefined> {
	return db.templates.get(id);
}
