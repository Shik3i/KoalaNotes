<script lang="ts">
	import { liveQuery } from 'dexie';
	import { db } from '$lib/db/database';
	import { page } from '$app/stores';
	import { createCampaign, observeCampaigns } from '$lib/db/campaigns';
	import { autofocus } from '$lib/utils/actions';
	import type { Campaign, Note } from '$lib/types/models';

	interface Props {
		collapsed?: boolean;
	}

	let { collapsed = false }: Props = $props();

	// Reactive campaign list
	let campaigns = $state<Campaign[]>([]);
	$effect(() => {
		const observable = liveQuery(() =>
			db.campaigns.where('archived').equals(0).sortBy('name')
		);
		const sub = observable.subscribe({
			next: (result) => { campaigns = result; },
			error: (err) => console.error('[campaigns]', err)
		});
		return () => sub.unsubscribe();
	});

	// Track which campaign is active based on URL
	let activeCampaignId = $derived($page.params.campaignId ?? null);

	// Notes for the active campaign (expanded in sidebar)
	let activeNotes = $state<Note[]>([]);
	$effect(() => {
		if (!activeCampaignId) {
			activeNotes = [];
			return;
		}
		const observable = liveQuery(() =>
			db.notes
				.where('[campaign_id+title]')
				.between([activeCampaignId!, ''], [activeCampaignId!, '\uffff'])
				.limit(100)
				.toArray()
		);
		const sub = observable.subscribe({
			next: (result) => { activeNotes = result; },
			error: (err) => console.error('[notes]', err)
		});
		return () => sub.unsubscribe();
	});

	let showCreateForm = $state(false);
	let newCampaignName = $state('');

	async function handleCreate() {
		const name = newCampaignName.trim();
		if (!name) return;
		await createCampaign(name);
		newCampaignName = '';
		showCreateForm = false;
	}
</script>

<aside class="sidebar" class:collapsed aria-label="Campaign navigation">
	<div class="sidebar-header">
		<h2>Campaigns</h2>
		<button
			class="add-btn"
			onclick={() => { showCreateForm = !showCreateForm; newCampaignName = ''; }}
			aria-label="Create campaign"
			title="New campaign"
		>
			+
		</button>
	</div>

	{#if showCreateForm}
		<form class="create-form" onsubmit={async (e) => { e.preventDefault(); try { await handleCreate(); } catch (err) { console.error(err); } }} aria-label="New campaign">
			<label for="new-campaign-input" class="sr-only">Campaign name</label>
			<input
				id="new-campaign-input"
				type="text"
				bind:value={newCampaignName}
				placeholder="Campaign name..."
				required
				use:autofocus
			/>
			<div class="form-actions">
				<button type="submit" disabled={!newCampaignName.trim()}>Create</button>
				<button type="button" onclick={() => { showCreateForm = false; }}>Cancel</button>
			</div>
		</form>
	{/if}

	<nav class="sidebar-nav" aria-label="Campaigns">
		{#if campaigns.length === 0}
			<div class="empty-state">
				<p class="empty-text">No campaigns yet</p>
				<p class="empty-hint">Click <strong>+</strong> above to create one.</p>
			</div>
		{:else}
			<ul role="tree" aria-label="Campaign list">
				{#each campaigns as c (c.id)}
					<li role="treeitem" aria-expanded={activeCampaignId === c.id} aria-selected={activeCampaignId === c.id}>
						<a
							href="/campaign/{c.id}"
							class="campaign-link"
							class:active={activeCampaignId === c.id}
						>
							<span class="campaign-icon" aria-hidden="true">📁</span>
							<span class="campaign-name">{c.name}</span>
						</a>

						{#if activeCampaignId === c.id && activeNotes.length > 0}
							<ul class="note-list" role="group" aria-label="Notes in {c.name}">
								{#each activeNotes as n (n.id)}
									<li role="treeitem" aria-selected={$page.params.noteId === n.id}>
										<a
											href="/campaign/{c.id}/notes/{n.id}"
											class="note-link"
											class:active={$page.params.noteId === n.id}
										>
											<span class="note-icon" aria-hidden="true">📝</span>
											<span class="note-title">{n.title || 'Untitled'}</span>
										</a>
									</li>
								{/each}
							</ul>
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
	</nav>

	<div class="sidebar-footer">
		<a href="/settings" class="settings-link">
			<span class="settings-icon" aria-hidden="true">⚙</span>
			<span>Settings</span>
		</a>
	</div>
</aside>

<style>
	.sidebar {
		width: var(--sidebar-width);
		height: 100%;
		background: var(--color-surface);
		border-right: 1px solid var(--color-border);
		display: flex;
		flex-direction: column;
		overflow-y: auto;
		flex-shrink: 0;
		transition: width 0.2s ease;
	}

	.sidebar.collapsed {
		width: 0;
		overflow: hidden;
		border-right: none;
	}

	.sidebar-header {
		padding: 0.75rem 1rem;
		border-bottom: 1px solid var(--color-border);
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
	}

	.sidebar-header h2 {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.add-btn {
		background: none;
		border: 1px solid var(--color-border);
		color: var(--color-text-muted);
		width: 24px;
		height: 24px;
		border-radius: 4px;
		font-size: 1rem;
		line-height: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: background 0.15s;
		flex-shrink: 0;
	}

	.add-btn:hover {
		background: var(--color-border);
		color: var(--color-text);
	}

	.create-form {
		padding: 0.75rem 1rem;
		border-bottom: 1px solid var(--color-border);
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.create-form input {
		width: 100%;
		padding: 0.5rem 0.625rem;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 4px;
		color: var(--color-text);
		font-size: 0.875rem;
		font-family: inherit;
		outline: none;
	}

	.create-form input:focus {
		border-color: var(--color-primary);
	}

	.form-actions {
		display: flex;
		gap: 0.5rem;
	}

	.form-actions button {
		flex: 1;
		padding: 0.375rem 0.5rem;
		border: none;
		border-radius: 4px;
		font-size: 0.8125rem;
		font-weight: 600;
		cursor: pointer;
	}

	.form-actions button[type="submit"] {
		background: var(--color-primary);
		color: white;
	}

	.form-actions button[type="submit"]:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.form-actions button[type="button"] {
		background: var(--color-border);
		color: var(--color-text-muted);
	}

	.sidebar-nav {
		flex: 1;
		padding: 0.5rem;
	}

	.empty-state {
		padding: 1rem 0.75rem;
		text-align: center;
	}

	.empty-text {
		font-size: 0.875rem;
		color: var(--color-text-muted);
		opacity: 0.7;
		margin-bottom: 0.25rem;
	}

	.empty-hint {
		font-size: 0.75rem;
		color: var(--color-text-muted);
		opacity: 0.5;
	}

	ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.campaign-link {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.75rem;
		border-radius: 6px;
		color: var(--color-text);
		font-size: 0.875rem;
		text-decoration: none;
		transition: background 0.15s;
	}

	.campaign-link:hover {
		background: var(--color-surface-alt);
		text-decoration: none;
	}

	.campaign-link.active {
		background: var(--color-surface-alt);
		border-left: 3px solid var(--color-primary);
		padding-left: calc(0.75rem - 3px);
	}

	.campaign-icon,
	.note-icon {
		font-size: 0.875rem;
		flex-shrink: 0;
	}

	.campaign-name {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.note-list {
		margin-left: 0.25rem;
		padding-left: 1rem;
		border-left: 1px solid var(--color-border);
	}

	.note-link {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.375rem 0.75rem;
		border-radius: 4px;
		color: var(--color-text-muted);
		font-size: 0.8125rem;
		text-decoration: none;
		transition: background 0.15s;
	}

	.note-link:hover {
		background: var(--color-surface-alt);
		color: var(--color-text);
		text-decoration: none;
	}

	.note-link.active {
		color: var(--color-text);
		background: var(--color-surface-alt);
	}

	.note-title {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.sidebar-footer {
		padding: 0.5rem;
		border-top: 1px solid var(--color-border);
	}

	.settings-link {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.75rem;
		border-radius: 6px;
		color: var(--color-text-muted);
		font-size: 0.8125rem;
		text-decoration: none;
		transition: background 0.15s, color 0.15s;
	}

	.settings-link:hover {
		background: var(--color-surface-alt);
		color: var(--color-text);
	}

	.settings-icon {
		font-size: 1rem;
	}

	@media (max-width: 768px) {
		.sidebar {
			position: fixed;
			left: 0;
			top: var(--header-height);
			bottom: var(--comment-bar-height);
			z-index: 100;
			box-shadow: 2px 0 8px rgba(0, 0, 0, 0.3);
		}

		.sidebar:not(.collapsed) {
			width: var(--sidebar-width);
		}
	}
</style>
