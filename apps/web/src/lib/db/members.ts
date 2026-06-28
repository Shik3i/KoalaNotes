import { db, uuid } from './database';
import { liveQuery } from 'dexie';
import type { CampaignMember, Role } from '$lib/types/models';

/** Observe all members of a campaign. */
export function observeMembers(campaignId: string) {
	return liveQuery(() =>
		db.campaign_members
			.where('campaign_id')
			.equals(campaignId)
			.toArray()
	);
}

/** Add a member to a campaign. Returns the member id. */
export async function addMember(
	campaignId: string,
	displayName: string,
	role: Role
): Promise<string> {
	const now = new Date().toISOString();
	const member: CampaignMember = {
		id: uuid(),
		campaign_id: campaignId,
		display_name: displayName.trim(),
		role,
		joined_at: now
	};
	await db.campaign_members.add(member);
	return member.id;
}

/** Update a member's role. */
export async function updateMemberRole(id: string, role: Role): Promise<void> {
	await db.campaign_members.update(id, { role });
}

/** Remove a member from a campaign. */
export async function removeMember(id: string): Promise<void> {
	await db.campaign_members.delete(id);
}
