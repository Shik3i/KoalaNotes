<script lang="ts">
	import { liveQuery } from 'dexie';
	import { db } from '$lib/db/database';
	import { createNote } from '$lib/db/notes';
	import { seedTemplates, getAllTemplates } from '$lib/db/templates';
	import { exportNotesAsMarkdown } from '$lib/utils/export';
	import { observeSessionsByCampaign, startSession, stopSession, observeActiveSession, createSessionRecap } from '$lib/db/sessions';
	import { observeMembers, addMember, removeMember, updateMemberRole } from '$lib/db/members';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import type { Campaign, CampaignMember, Note, Role, Session, Template, TemplateType } from '$lib/types/models';

	let campaignId = $derived($page.params.campaignId ?? '');

	let campaign = $state<Campaign | undefined>();
	$effect(() => {
		const observable = liveQuery(() => db.campaigns.get(campaignId));
		const sub = observable.subscribe({
			next: (result) => { campaign = result; },
			error: (err) => console.error('[campaign]', err)
		});
		return () => sub.unsubscribe();
	});

	let notes = $state<Note[]>([]);
	$effect(() => {
		const observable = liveQuery(() =>
			db.notes.where('campaign_id').equals(campaignId).sortBy('title')
		);
		const sub = observable.subscribe({
			next: (result) => { notes = result; },
			error: (err) => console.error('[notes]', err)
		});
		return () => sub.unsubscribe();
	});

	let templates = $state<Template[]>([]);
	let showTemplatePicker = $state(false);
	let filterTag = $state<string | null>(null);
	let searchQuery = $state('');

	// Session management
	let sessions = $state<Session[]>([]);
	let activeSession = $state<Session | null>(null);
	let sessionBusy = $state(false);

	$effect(() => {
		const sObs = observeSessionsByCampaign(campaignId);
		const sub = sObs.subscribe({
			next: (result) => { sessions = result; },
			error: (err) => console.error('[sessions]', err)
		});
		return () => sub.unsubscribe();
	});

	$effect(() => {
		const aObs = observeActiveSession();
		const sub = aObs.subscribe({
			next: (result) => { activeSession = result ?? null; },
			error: (err) => console.error('[active session]', err)
		});
		return () => sub.unsubscribe();
	});

	let activeForThisCampaign = $derived(activeSession?.campaign_id === campaignId);

	async function handleStartSession(name?: string) {
		sessionBusy = true;
		try {
			await startSession(campaignId, name || undefined);
		} catch (err) {
			console.error('[start session]', err);
		} finally {
			sessionBusy = false;
		}
	}

	async function handleStopSession() {
		if (!activeSession) return;
		sessionBusy = true;
		try {
			await stopSession(activeSession.id);
		} catch (err) {
			console.error('[stop session]', err);
		} finally {
			sessionBusy = false;
		}
	}

	let recapBusy = $state(false);

	// Member management
	let members = $state<CampaignMember[]>([]);
	let showAddMember = $state(false);
	let newMemberName = $state('');
	let newMemberRole = $state<Role>('player');
	let memberBusy = $state(false);

	$effect(() => {
		const mObs = observeMembers(campaignId);
		const sub = mObs.subscribe({
			next: (result) => { members = result; },
			error: (err) => console.error('[members]', err)
		});
		return () => sub.unsubscribe();
	});

	async function handleAddMember() {
		const name = newMemberName.trim();
		if (!name || memberBusy) return;
		memberBusy = true;
		try {
			await addMember(campaignId, name, newMemberRole);
			newMemberName = '';
			showAddMember = false;
		} catch (err) {
			console.error('[add member]', err);
		} finally {
			memberBusy = false;
		}
	}

	async function handleRemoveMember(id: string) {
		if (memberBusy) return;
		memberBusy = true;
		try {
			await removeMember(id);
		} catch (err) {
			console.error('[remove member]', err);
		} finally {
			memberBusy = false;
		}
	}

	async function handleRoleChange(id: string, role: Role) {
		if (memberBusy) return;
		memberBusy = true;
		try {
			await updateMemberRole(id, role);
		} catch (err) {
			console.error('[update role]', err);
		} finally {
			memberBusy = false;
		}
	}

	async function handleCreateRecap(sessionId: string) {
		recapBusy = true;
		try {
			const noteId = await createSessionRecap(sessionId);
			await goto(`/campaign/${campaignId}/notes/${noteId}`);
		} catch (err) {
			console.error('[create recap]', err);
		} finally {
			recapBusy = false;
		}
	}

	function formatElapsed(seconds: number): string {
		const h = Math.floor(seconds / 3600);
		const m = Math.floor((seconds % 3600) / 60);
		const s = seconds % 60;
		return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
	}

	// Collect all unique tags from notes
	let allTags = $derived.by(() => {
		const set = new Set<string>();
		for (const n of notes) {
			for (const t of n.tags) set.add(t);
		}
		return [...set].sort();
	});

	// Filter notes by selected tag and/or search query
	let filteredNotes = $derived.by(() => {
		let result = filterTag ? notes.filter((n) => n.tags.includes(filterTag!)) : notes;
		if (searchQuery.trim()) {
			const q = searchQuery.trim().toLowerCase();
			result = result.filter((n) =>
				n.title.toLowerCase().includes(q) || n.content.toLowerCase().includes(q)
			);
		}
		return result;
	});

	function setFilter(tag: string | null) {
		filterTag = tag;
	}

	$effect(() => {
		let cancelled = false;
		seedTemplates().then(() => {
			if (!cancelled) getAllTemplates().then((t) => { templates = t; });
		});
		return () => { cancelled = true; };
	});

	async function handleCreateNote(type?: TemplateType) {
		const id = await createNote(campaignId, 'Untitled', '', type || 'blank');
		showTemplatePicker = false;
		await goto(`/campaign/${campaignId}/notes/${id}`);
	}

	async function handleQuickNote() {
		await handleCreateNote('blank');
	}
</script>

<main class="campaign-page">
	{#if campaign}
		<header class="campaign-header">
			<div class="header-row">
				<h1>{campaign.name}</h1>
				<div class="header-actions">
			<button class="action-btn" onclick={() => { showTemplatePicker = !showTemplatePicker; }} aria-label="New note">
					+ New Note
				</button>
				{#if notes.length > 0}
					<button class="action-btn" onclick={() => exportNotesAsMarkdown(notes, campaign?.name ?? 'Campaign')} aria-label="Export all notes as Markdown">
						Export All
					</button>
				{/if}
				</div>
			</div>
			{#if campaign.description}
				<p class="campaign-desc">{campaign.description}</p>
			{/if}
			{#if campaign.system}
				<span class="system-badge">{campaign.system}</span>
			{/if}
		</header>

		<section class="session-section">
			<div class="session-section-row">
				<h2>Session</h2>
				{#if activeForThisCampaign}
					<button class="action-btn danger" onclick={handleStopSession} disabled={sessionBusy} aria-label="End current session">
						{sessionBusy ? 'Ending...' : 'End Session'}
					</button>
				{:else}
					<button class="action-btn" onclick={() => handleStartSession()} disabled={sessionBusy} aria-label="Start a new session">
						{sessionBusy ? 'Starting...' : 'Start Session'}
					</button>
				{/if}
			</div>
			{#if sessions.length > 0}
				<ul class="session-list" aria-label="Session history">
					{#each sessions as s (s.id)}
						<li class="session-item" class:active-session={s.id === activeSession?.id}>
							<span class="session-name">{s.name}</span>
							<span class="session-status">{s.status}</span>
							<span class="session-date">{s.started_at ? new Date(s.started_at).toLocaleDateString() : '—'}</span>
							<div class="session-actions">
								{#if s.recap_note_id}
									<a
										href="/campaign/{campaignId}/notes/{s.recap_note_id}"
										class="recap-link"
										aria-label="View recap for {s.name}"
									>
										Recap
									</a>
								{:else if s.status === 'completed'}
									<button
										class="recap-btn"
										onclick={() => handleCreateRecap(s.id)}
										disabled={recapBusy}
										aria-label="Create recap for {s.name}"
									>
										{recapBusy ? '...' : 'Create Recap'}
									</button>
								{/if}
							</div>
						</li>
					{/each}
				</ul>
			{:else}
				<p class="empty-sessions">No sessions yet. Start your first session to begin tracking!</p>
			{/if}
		</section>

		<section class="members-section">
			<div class="members-section-row">
				<h2>Members</h2>
				<button
					class="action-btn"
					onclick={() => { showAddMember = !showAddMember; }}
					aria-label="Add campaign member"
				>
					+ Add
				</button>
			</div>
			{#if showAddMember}
				<div class="add-member-form">
					<label for="member-name" class="sr-only">Member name</label>
					<input
						id="member-name"
						type="text"
						bind:value={newMemberName}
						placeholder="Display name..."
						aria-label="Member display name"
					/>
					<label for="member-role" class="sr-only">Role</label>
					<select
						id="member-role"
						bind:value={newMemberRole}
						aria-label="Member role"
					>
						<option value="player">Player</option>
						<option value="observer">Observer</option>
						<option value="gm">GM</option>
					</select>
					<button class="action-btn" onclick={handleAddMember} disabled={memberBusy || !newMemberName.trim()}>
						{memberBusy ? 'Adding...' : 'Add'}
					</button>
					<button class="cancel-btn" onclick={() => { showAddMember = false; }}>Cancel</button>
				</div>
			{/if}
			{#if members.length > 0}
				<ul class="member-list" aria-label="Campaign members">
					{#each members as m (m.id)}
						<li class="member-item">
							<span class="member-name">{m.display_name}</span>
							<select
								class="member-role-select"
								value={m.role}
								onchange={(e) => handleRoleChange(m.id, e.currentTarget.value as Role)}
								aria-label="Role for {m.display_name}"
							>
								<option value="gm">GM</option>
								<option value="player">Player</option>
								<option value="observer">Observer</option>
							</select>
							<button
								class="remove-btn"
								onclick={() => handleRemoveMember(m.id)}
								disabled={memberBusy}
								aria-label="Remove {m.display_name}"
							>
								&times;
							</button>
						</li>
					{/each}
				</ul>
			{:else}
				<p class="empty-members">No members added yet. Add players or observers to your campaign.</p>
			{/if}
		</section>

		{#if showTemplatePicker}
			<section class="template-picker" aria-label="Choose a template">
				<h2>Create from template</h2>
				<div class="template-grid">
					{#each templates as t (t.id)}
						<button class="template-card" onclick={() => handleCreateNote(t.type)}>
							<span class="template-name">{t.name}</span>
							<span class="template-desc">{t.description}</span>
						</button>
					{/each}
				</div>
				<button class="cancel-btn" onclick={() => { showTemplatePicker = false; }}>Cancel</button>
			</section>
		{/if}

		<div class="filter-bar">
			<label for="search-input" class="sr-only">Search notes</label>
			<input
				id="search-input"
				type="search"
				bind:value={searchQuery}
				class="search-input"
				placeholder="Search notes..."
				aria-label="Search notes by title or content"
			/>

			{#if allTags.length > 0}
				<div class="tag-filter" role="group" aria-label="Filter by tag">
					<button
						class="filter-chip"
						class:active={filterTag === null && !searchQuery.trim()}
						onclick={() => setFilter(null)}
					>
						All
					</button>
					{#each allTags as tag (tag)}
						<button
							class="filter-chip"
							class:active={filterTag === tag}
							onclick={() => setFilter(tag)}
						>
							{tag}
						</button>
					{/each}
					{#if filterTag}
						<button class="filter-clear" onclick={() => setFilter(null)} aria-label="Clear filter">
							&times; Clear
						</button>
					{/if}
				</div>
			{/if}
		</div>

		<section class="notes-list" aria-label="Notes">
			<h2>Notes ({filteredNotes.length})</h2>
			{#if filteredNotes.length === 0}
				<p class="empty-notes">{filterTag ? 'No notes with tag "' + filterTag + '".' : 'No notes yet. Create one to get started.'}</p>
			{:else}
				<ul>
					{#each filteredNotes as n (n.id)}
						<li>
							<a href="/campaign/{campaignId}/notes/{n.id}" class="note-item">
								<div class="note-info">
									<span class="note-title">{n.title}</span>
									{#if n.template_type && n.template_type !== 'blank'}
										<span class="template-type">{n.template_type}</span>
									{/if}
									{#if n.tags.length > 0}
										<div class="note-tags">
											{#each n.tags as tag (tag)}
												<button
													class="tag-clickable"
													onclick={(e) => { e.preventDefault(); e.stopPropagation(); setFilter(tag); }}
													aria-label="Filter by tag {tag}"
												>
													{tag}
												</button>
											{/each}
										</div>
									{/if}
								</div>
								<span class="note-date">{new Date(n.updated_at).toLocaleDateString()}</span>
							</a>
						</li>
					{/each}
				</ul>
			{/if}
		</section>
	{:else}
		<div class="loading">
			<p>Campaign not found.</p>
		</div>
	{/if}
</main>

<style>
	.campaign-page {
		flex: 1;
		display: flex;
		flex-direction: column;
		padding: 2rem;
		overflow-y: auto;
		max-width: 960px;
		width: 100%;
		margin: 0 auto;
	}

	.campaign-header {
		margin-bottom: 2rem;
	}

	.header-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 0.5rem;
	}

	h1 {
		font-size: 1.5rem;
		font-weight: 700;
	}

	.header-actions {
		display: flex;
		gap: 0.5rem;
		flex-shrink: 0;
	}

	.action-btn {
		padding: 0.5rem 1rem;
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: 6px;
		font-size: 0.875rem;
		font-weight: 600;
		white-space: nowrap;
		transition: background 0.15s;
	}

	.action-btn:hover {
		background: var(--color-primary-hover);
	}

	.campaign-desc {
		color: var(--color-text-muted);
		font-size: 0.9375rem;
		margin-bottom: 0.75rem;
	}

	.system-badge {
		display: inline-block;
		padding: 0.25rem 0.625rem;
		background: var(--color-surface-alt);
		border: 1px solid var(--color-border);
		border-radius: 4px;
		font-size: 0.75rem;
		color: var(--color-text-muted);
	}

	.session-section {
		margin-bottom: 2rem;
		padding: 1rem 1.25rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 8px;
	}

	.session-section-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.75rem;
	}

	.session-section-row h2 {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin: 0;
	}

	.action-btn.danger {
		background: var(--color-danger);
	}

	.action-btn.danger:hover:not(:disabled) {
		filter: brightness(1.15);
	}

	.session-list {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
	}

	.session-item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.5rem 0.75rem;
		flex-wrap: wrap;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 4px;
		font-size: 0.8125rem;
	}

	.session-item.active-session {
		border-color: var(--color-success);
	}

	.session-name {
		flex: 1;
		font-weight: 500;
		color: var(--color-text);
	}

	.session-status {
		font-size: 0.6875rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 0.125rem 0.5rem;
		border-radius: 9999px;
		background: var(--color-surface-alt);
		color: var(--color-text-muted);
	}

	.session-item.active-session .session-status {
		background: var(--color-success);
		color: #1a1a2e;
		font-weight: 700;
	}

	.session-date {
		font-size: 0.6875rem;
		color: var(--color-text-muted);
		opacity: 0.6;
		flex-shrink: 0;
	}

	.session-actions {
		flex-shrink: 0;
	}

	.recap-link {
		font-size: 0.6875rem;
		padding: 0.125rem 0.5rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 4px;
		color: var(--color-primary);
		text-decoration: none;
		transition: background 0.15s;
	}

	.recap-link:hover {
		background: var(--color-surface-alt);
		text-decoration: none;
	}

	.recap-btn {
		font-size: 0.6875rem;
		padding: 0.125rem 0.5rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 4px;
		color: var(--color-text-muted);
		font-family: inherit;
	}

	.recap-btn:hover:not(:disabled) {
		border-color: var(--color-primary);
		color: var(--color-text);
	}

	.recap-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.empty-sessions {
		font-size: 0.8125rem;
		color: var(--color-text-muted);
		opacity: 0.6;
		font-style: italic;
		margin: 0;
	}

	.members-section {
		margin-bottom: 2rem;
		padding: 1rem 1.25rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 8px;
	}

	.members-section-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.75rem;
	}

	.members-section-row h2 {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin: 0;
	}

	.add-member-form {
		display: flex;
		gap: 0.5rem;
		align-items: center;
		margin-bottom: 0.75rem;
		flex-wrap: wrap;
	}

	.add-member-form input,
	.add-member-form select {
		padding: 0.375rem 0.625rem;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 4px;
		color: var(--color-text);
		font-size: 0.8125rem;
		font-family: inherit;
	}

	.add-member-form input {
		min-width: 160px;
	}

	.add-member-form select {
		cursor: pointer;
	}

	.cancel-btn {
		background: none;
		border: 1px solid var(--color-border);
		color: var(--color-text-muted);
		padding: 0.375rem 0.75rem;
		border-radius: 4px;
		font-size: 0.8125rem;
		font-family: inherit;
	}

	.member-list {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
	}

	.member-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.375rem 0.75rem;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 4px;
	}

	.member-name {
		flex: 1;
		font-size: 0.8125rem;
		font-weight: 500;
		color: var(--color-text);
	}

	.member-role-select {
		font-size: 0.75rem;
		padding: 0.125rem 0.375rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 4px;
		color: var(--color-text-muted);
		font-family: inherit;
		cursor: pointer;
	}

	.remove-btn {
		background: none;
		border: none;
		color: var(--color-danger);
		font-size: 1.125rem;
		line-height: 1;
		padding: 0 0.25rem;
		opacity: 0.5;
		transition: opacity 0.15s;
		font-family: inherit;
	}

	.remove-btn:hover:not(:disabled) {
		opacity: 1;
	}

	.remove-btn:disabled {
		opacity: 0.2;
		cursor: not-allowed;
	}

	.empty-members {
		font-size: 0.8125rem;
		color: var(--color-text-muted);
		opacity: 0.6;
		font-style: italic;
		margin: 0;
	}

	.template-picker {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		padding: 1.5rem;
		margin-bottom: 2rem;
	}

	.template-picker h2 {
		font-size: 1rem;
		font-weight: 600;
		margin-bottom: 1rem;
		color: var(--color-text-muted);
	}

	.template-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		gap: 0.75rem;
		margin-bottom: 1rem;
	}

	.template-card {
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 6px;
		padding: 0.875rem;
		text-align: left;
		color: var(--color-text);
		transition: border-color 0.15s;
	}

	.template-card:hover {
		border-color: var(--color-primary);
	}

	.template-name {
		display: block;
		font-weight: 600;
		font-size: 0.875rem;
		margin-bottom: 0.25rem;
	}

	.template-desc {
		display: block;
		font-size: 0.75rem;
		color: var(--color-text-muted);
		opacity: 0.7;
	}

	.cancel-btn {
		background: none;
		border: 1px solid var(--color-border);
		color: var(--color-text-muted);
		padding: 0.375rem 0.75rem;
		border-radius: 4px;
		font-size: 0.8125rem;
	}

	.filter-bar {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	.search-input {
		width: 100%;
		padding: 0.5rem 0.75rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 6px;
		color: var(--color-text);
		font-size: 0.875rem;
		font-family: inherit;
		outline: none;
		transition: border-color 0.15s;
	}

	.search-input:focus {
		border-color: var(--color-primary);
	}

	.search-input::placeholder {
		color: var(--color-text-muted);
		opacity: 0.4;
	}

	.tag-filter {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.375rem;
		padding: 0.25rem 0;
	}

	.filter-chip {
		padding: 0.25rem 0.625rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 9999px;
		color: var(--color-text-muted);
		font-size: 0.75rem;
		font-family: inherit;
		cursor: pointer;
		transition: all 0.15s;
	}

	.filter-chip:hover {
		border-color: var(--color-primary);
		color: var(--color-text);
	}

	.filter-chip.active {
		background: var(--color-primary);
		border-color: var(--color-primary);
		color: white;
	}

	.filter-clear {
		background: none;
		border: none;
		color: var(--color-text-muted);
		font-size: 0.75rem;
		font-family: inherit;
		cursor: pointer;
		opacity: 0.6;
		text-decoration: underline;
	}

	.filter-clear:hover {
		opacity: 1;
		color: var(--color-danger);
	}

	.notes-list {
		flex: 1;
	}

	.notes-list h2 {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 1rem;
	}

	.empty-notes {
		color: var(--color-text-muted);
		font-size: 0.875rem;
		opacity: 0.6;
	}

	ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.note-item {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		padding: 0.75rem;
		border-radius: 6px;
		text-decoration: none;
		color: var(--color-text);
		transition: background 0.15s;
		gap: 1rem;
	}

	.note-item:hover {
		background: var(--color-surface);
		text-decoration: none;
	}

	.note-info {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.5rem;
	}

	.note-title {
		font-size: 0.9375rem;
		font-weight: 500;
	}

	.template-type {
		font-size: 0.6875rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--color-primary);
		opacity: 0.8;
	}

	.note-tags {
		display: flex;
		gap: 0.375rem;
		flex-wrap: wrap;
	}

	.tag-clickable {
		font-size: 0.6875rem;
		padding: 0.125rem 0.5rem;
		background: var(--color-surface-alt);
		border-radius: 9999px;
		color: var(--color-text-muted);
	}

	.tag-clickable {
		border: 1px solid transparent;
		cursor: pointer;
		font-family: inherit;
		transition: all 0.15s;
	}

	.tag-clickable:hover {
		border-color: var(--color-primary);
		color: var(--color-text);
		background: var(--color-surface);
	}

	.note-date {
		font-size: 0.75rem;
		color: var(--color-text-muted);
		flex-shrink: 0;
		opacity: 0.6;
	}

	.loading {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--color-text-muted);
	}
</style>
