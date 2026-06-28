<script lang="ts">
	import { liveQuery } from 'dexie';
	import { getNote, updateNote, deleteNote } from '$lib/db/notes';
	import { observeOutgoingLinks, observeBacklinks, getNoteTitleMap } from '$lib/db/wiki';
	import { viewingRole } from '$lib/stores/role';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import TagInput from '$lib/components/common/TagInput.svelte';
	import { exportNoteAsMarkdown } from '$lib/utils/export';
	import { triggerAutoSync } from '$lib/services/sync';
	import type { Note, NoteSection, Role, Visibility, WikiLink } from '$lib/types/models';

	let campaignId = $derived($page.params.campaignId ?? '');
	let noteId = $derived($page.params.noteId ?? '');

	let note = $state<Note | undefined>();
	let title = $state('');
	let content = $state('');
	let tags = $state<string[]>([]);
	let sections = $state<NoteSection[]>([]);
	let showPreview = $state(true);
	let showDeleteConfirm = $state(false);
	let currentRole = $state<Role>('gm');

	// Subscribe to the viewing role store
	$effect(() => {
		const unsub = viewingRole.subscribe((v) => { currentRole = v; });
		return () => unsub();
	});

	$effect(() => {
		const observable = liveQuery(() => getNote(noteId));
		const sub = observable.subscribe({
			next: (result) => {
				note = result;
				if (result) {
					title = result.title;
					content = result.content;
					tags = result.tags;
					sections = result.sections ?? [];
				}
			},
			error: (err) => console.error('[note]', err)
		});
		return () => sub.unsubscribe();
	});

	// Detect section headings from content
	let detectedHeadings = $derived.by(() => {
		const headings: string[] = [];
		const re = /^## (.+)$/gm;
		let m;
		while ((m = re.exec(content)) !== null) {
			headings.push(m[1]);
		}
		return headings;
	});

	// Build visibility map merging detected headings with stored sections
	let sectionVisibility = $derived.by(() => {
		const map = new Map<string, Visibility>();
		const stored = new Map(sections.map(s => [s.heading, s.visibility]));
		for (const h of detectedHeadings) {
			map.set(h, stored.get(h) ?? 'shared');
		}
		return map;
	});

	// Check if a section heading is visible for the current role
	function isSectionVisible(visibility: Visibility, role: Role): boolean {
		switch (visibility) {
			case 'shared': return true;
			case 'gm_only': return role === 'gm';
			case 'observer': return role === 'gm' || role === 'observer';
			case 'private': return role === 'gm';
		}
	}

	let sectionsChanged = $derived(
		JSON.stringify(sections) !== JSON.stringify(note?.sections ?? [])
	);

	let hasChanges = $derived(
		note !== undefined && (
			title !== note.title ||
			content !== note.content ||
			JSON.stringify(tags) !== JSON.stringify(note.tags) ||
			sectionsChanged
		)
	);

	async function handleSave() {
		if (!note || !hasChanges) return;
		await updateNote(note.id, { title: title.trim() || 'Untitled', content, tags, sections });
		triggerAutoSync();
	}

	function handleCycleVisibility(heading: string) {
		const cycle: Visibility[] = ['shared', 'gm_only', 'observer', 'private'];
		const current = sectionVisibility.get(heading) ?? 'shared';
		const idx = cycle.indexOf(current);
		const next = cycle[(idx + 1) % cycle.length];

		// Update or add section entry
		const existing = sections.findIndex(s => s.heading === heading);
		if (existing >= 0) {
			sections[existing] = { ...sections[existing], visibility: next };
			sections = [...sections];
		} else {
			sections = [...sections, {
				id: crypto.randomUUID(),
				heading,
				content: '',
				visibility: next,
				order: detectedHeadings.indexOf(heading)
			}];
		}
		// Auto-save sections immediately
		if (note) {
			updateNote(note.id, { sections }).catch(err => console.error('[visibility]', err));
		}
	}

	const VISIBILITY_LABELS: Record<Visibility, string> = {
		shared: 'Shared',
		gm_only: 'GM',
		observer: 'Obs.',
		private: 'Priv.'
	};

	async function handleDelete() {
		if (!note) return;
		await deleteNote(note.id);
		await goto(`/campaign/${campaignId}`);
	}

	// Outgoing wiki links resolved for this note
	let outgoingLinks = $state<WikiLink[]>([]);
	$effect(() => {
		if (!noteId) return;
		const observable = observeOutgoingLinks(noteId);
		const sub = observable.subscribe({
			next: (result) => { outgoingLinks = result; },
			error: (err) => console.error('[outgoing]', err)
		});
		return () => sub.unsubscribe();
	});

	// Backlinks (notes that link to this note)
	let backlinks = $state<WikiLink[]>([]);
	let backlinkTitles = $state<Map<string, string>>(new Map());
	$effect(() => {
		if (!noteId) return;
		let destroyed = false;
		const observable = observeBacklinks(noteId);
		const sub = observable.subscribe({
			next: async (result) => {
				if (destroyed) return;
				backlinks = result;
				const ids = result.map(l => l.source_note_id).filter(Boolean);
				const titles = await getNoteTitleMap(ids);
				if (destroyed) return;
				backlinkTitles = titles;
			},
			error: (err) => console.error('[backlinks]', err)
		});
		return () => {
			destroyed = true;
			sub.unsubscribe();
		};
	});

	let saveTimer: ReturnType<typeof setTimeout> | undefined;
	$effect(() => {
		return () => { if (saveTimer) clearTimeout(saveTimer); };
	});
	function onInput() {
		if (saveTimer) clearTimeout(saveTimer);
		saveTimer = setTimeout(() => { handleSave(); }, 1500);
	}

	/** Filter content by section visibility for the current role. */
	function filterContentByRole(md: string, role: Role): string {
		if (!md) return '';
		// Split on section boundaries: a `## ` heading at start of line (not indented)
		const parts = md.split(/\n(?=## )/);
		if (parts.length <= 1) return md; // no sections to filter

		return parts.filter((part) => {
			const headingMatch = part.match(/^## (.+)$/m);
			if (!headingMatch) return true; // keep preamble (content before first section)
			const visibility = sectionVisibility.get(headingMatch[1]) ?? 'shared';
			return isSectionVisible(visibility, role);
		}).join('\n');
	}

	function renderPreview(md: string, campId: string, role: Role): string {
		const filtered = filterContentByRole(md, role);
		if (!filtered) return '<p class="empty-preview">Nothing visible with current role.</p>';

		const linkHrefs = new Map<string, string>();
		for (const link of outgoingLinks) {
			linkHrefs.set(
				link.context,
				link.target_note_id ? `/campaign/${campId}/notes/${link.target_note_id}` : '#'
			);
		}

		let html = filtered
			.replace(/&/g, '&amp;')
			.replace(/</g, '&lt;')
			.replace(/>/g, '&gt;')
			.replace(/"/g, '&quot;')
			.replace(/^### (.+)$/gm, '<h3>$1</h3>')
			.replace(/^## (.+)$/gm, '<h2>$1</h2>')
			.replace(/^# (.+)$/gm, '<h1>$1</h1>')
			.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
			.replace(/\*(.+?)\*/g, '<em>$1</em>')
			.replace(/`(.+?)`/g, '<code>$1</code>')
			.replace(/~~(.+?)~~/g, '<del>$1</del>')
			.replace(/\[([^\]]+)\]\(([^)]+)\)/g,
				(_m, text, url) => {
					const safe = url.startsWith('javascript:') ? '#' : url;
					return '<a href="' + safe + '" target="_blank" rel="noopener">' + text + '</a>';
				})
			// Wiki links with resolved URLs
			.replace(/\[\[([^\]]+)\]\]/g, (_m, title) => {
				const href = linkHrefs.get(title) ?? '#';
				const cls = href === '#' ? 'wiki-link unresolved' : 'wiki-link';
				return '<a href="' + href + '" class="' + cls + '" data-wiki="' + title + '">' + title + '</a>';
			})
			.replace(/^---$/gm, '<hr />')
			.replace(/^- (.+)$/gm, '<li>$1</li>')
			.replace(/(<li>.*<\/li>\n?)+/g, (match) => `<ul>${match}</ul>`)
			.replace(/^\d+\. (.+)$/gm, '<li>$1</li>')
			.replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code>$2</code></pre>')
			.replace(/^> (.+)$/gm, '<blockquote>$1</blockquote>')
			.replace(/\n\n/g, '</p><p>')
			.replace(/\n/g, '<br />');

		return `<p>${html}</p>`;
	}

	// Debounce preview to avoid re-rendering on every keystroke
	let previewContent = $state('');
	$effect(() => {
		const c = content;
		const delay = previewContent === '' ? 0 : 300;
		const timer = setTimeout(() => { previewContent = c; }, delay);
		return () => clearTimeout(timer);
	});

	let previewHtml = $derived(renderPreview(previewContent, campaignId, currentRole));
</script>

<svelte:head>
	<title>{note?.title || 'Untitled'} — {note?.campaign_id ? `Campaign` : ''} KoalaNotes</title>
	<meta name="description" content="{note?.title || 'Untitled'} — Edit and view TTRPG campaign notes in KoalaNotes." />
	<meta property="og:title" content="{note?.title || 'Untitled'} — KoalaNotes" />
	<meta name="twitter:title" content="{note?.title || 'Untitled'} — KoalaNotes" />
</svelte:head>

<main class="note-editor">
	{#if note}
		<div class="editor-header">
			<div class="title-row">
				<input
					type="text"
					bind:value={title}
					oninput={onInput}
					class="title-input"
					placeholder="Note title..."
					aria-label="Note title"
				/>
				<div class="editor-actions">
					<button
						class="action-btn"
						onclick={() => { showPreview = !showPreview; }}
						aria-label={showPreview ? 'Editor only' : 'Split view'}
					>
						{showPreview ? 'Editor' : 'Split'}
					</button>
					<button
						class="action-btn"
						onclick={() => { if (note) exportNoteAsMarkdown(note); }}
						aria-label="Export note as Markdown"
					>
						Export
					</button>
					<button
						class="action-btn danger"
						onclick={() => { showDeleteConfirm = true; }}
						aria-label="Delete note"
					>
						Delete
					</button>
				</div>
			</div>
			<div class="meta-row">
				{#if note.template_type && note.template_type !== 'blank'}
					<span class="template-type">{note.template_type}</span>
				{/if}
				<span class="updated-at">Last edited {new Date(note.updated_at).toLocaleString()}</span>
			</div>
			<div class="tags-row">
				<TagInput bind:tags onchange={() => { onInput(); }} />
			</div>
		</div>

		{#if showDeleteConfirm}
			<div class="delete-confirm" role="alertdialog" aria-label="Confirm delete">
				<p>Delete "{note.title}"? This cannot be undone.</p>
				<div class="confirm-actions">
					<button class="action-btn danger" onclick={handleDelete}>Delete</button>
					<button class="action-btn" onclick={() => { showDeleteConfirm = false; }}>Cancel</button>
				</div>
			</div>
		{/if}

		<div class="editor-body" class:split={showPreview}>
			<textarea
				bind:value={content}
				oninput={onInput}
				class="editor-textarea"
				placeholder="Write your note in Markdown..."
				aria-label="Note content"
			></textarea>

			{#if showPreview}
				<div class="preview-pane">
					<div class="preview-header">Preview</div>
					<div class="preview-content">
						{@html previewHtml}
					</div>
				</div>
			{/if}
		</div>

		{#if detectedHeadings.length > 0}
			<div class="section-visibility-bar" aria-label="Section visibility controls">
				<span class="sv-label">Sections:</span>
				{#each detectedHeadings as heading (heading)}
					{@const vis = sectionVisibility.get(heading) ?? 'shared'}
					<button
						class="vis-chip"
						class:gm-only={vis === 'gm_only'}
						class:observer={vis === 'observer'}
						class:private={vis === 'private'}
						onclick={() => handleCycleVisibility(heading)}
						title="{heading}: {VISIBILITY_LABELS[vis]} (click to cycle)"
						aria-label="Section '{heading}' visibility: {VISIBILITY_LABELS[vis]}. Click to change."
					>
						<span class="vis-heading">{heading}</span>
						<span class="vis-badge">{VISIBILITY_LABELS[vis]}</span>
					</button>
				{/each}
			</div>
		{/if}

		<div class="editor-footer">
			<div class="footer-left">
				<span class="word-count">
					{content.trim() ? content.trim().split(/\s+/).length : 0} words
				</span>
				{#if hasChanges}
					<span class="unsaved">Unsaved changes</span>
				{:else}
					<span class="saved">All changes saved</span>
				{/if}
			</div>
			<div class="footer-right">
				{#if backlinks.length > 0}
					<div class="backlink-badge" title="Referenced by {backlinks.length} other note{backlinks.length === 1 ? '' : 's'}">
						🔗 {backlinks.length}
					</div>
				{/if}
			</div>
		</div>

		{#if backlinks.length > 0}
			<section class="backlinks-section" aria-label="Backlinks">
				<h3>Referenced by ({backlinks.length})</h3>
				<ul>
					{#each backlinks as bl (bl.id)}
						{@const sourceTitle = backlinkTitles.get(bl.source_note_id) ?? 'Untitled'}
						<li>
							<a href="/campaign/{campaignId}/notes/{bl.source_note_id}" class="backlink-item">
								{sourceTitle}
								<span class="backlink-context">via [[{bl.context}]]</span>
							</a>
						</li>
					{/each}
				</ul>
			</section>
		{/if}
	{:else}
		<div class="loading">
			<p>Note not found.</p>
		</div>
	{/if}
</main>

<style>
	.note-editor {
		flex: 1;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.editor-header {
		padding: 1rem 1.5rem;
		border-bottom: 1px solid var(--color-border);
		flex-shrink: 0;
	}

	.title-row {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 0.25rem;
	}

	.title-input {
		flex: 1;
		font-size: 1.375rem;
		font-weight: 700;
		color: var(--color-text);
		background: transparent;
		border: none;
		outline: none;
		padding: 0.25rem 0;
		font-family: inherit;
	}

	.title-input::placeholder {
		color: var(--color-text-muted);
		opacity: 0.4;
	}

	.editor-actions {
		display: flex;
		gap: 0.5rem;
		flex-shrink: 0;
	}

	.action-btn {
		padding: 0.375rem 0.75rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		color: var(--color-text-muted);
		border-radius: 4px;
		font-size: 0.8125rem;
		font-family: inherit;
		cursor: pointer;
		transition: background 0.15s;
	}

	.action-btn:hover {
		background: var(--color-surface-alt);
	}

	.action-btn.danger {
		color: var(--color-danger);
		border-color: var(--color-danger);
	}

	.meta-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 0.5rem;
	}

	.tags-row {
		margin-bottom: 0.25rem;
	}

	.template-type {
		font-size: 0.6875rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--color-primary);
		opacity: 0.8;
	}

	.updated-at {
		font-size: 0.75rem;
		color: var(--color-text-muted);
		opacity: 0.5;
	}

	.delete-confirm {
		padding: 0.75rem 1.5rem;
		background: var(--color-surface);
		border-bottom: 1px solid var(--color-danger);
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
	}

	.delete-confirm p {
		font-size: 0.875rem;
		color: var(--color-danger);
	}

	.confirm-actions {
		display: flex;
		gap: 0.5rem;
		flex-shrink: 0;
	}

	.editor-body {
		flex: 1;
		display: flex;
		overflow: hidden;
	}

	.editor-body.split .editor-textarea {
		border-right: 1px solid var(--color-border);
	}

	.editor-textarea {
		flex: 1;
		padding: 1.5rem;
		background: var(--color-bg);
		color: var(--color-text);
		font-family: var(--font-mono);
		font-size: 0.875rem;
		line-height: 1.7;
		border: none;
		outline: none;
		resize: none;
		overflow-y: auto;
	}

	.editor-textarea::placeholder {
		color: var(--color-text-muted);
		opacity: 0.4;
	}

	.preview-pane {
		flex: 1;
		display: flex;
		flex-direction: column;
		overflow-y: auto;
		background: var(--color-bg);
	}

	.preview-header {
		padding: 0.5rem 1.5rem;
		font-size: 0.6875rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--color-text-muted);
		opacity: 0.5;
		border-bottom: 1px solid var(--color-border);
		flex-shrink: 0;
	}

	.preview-content {
		padding: 1.5rem;
		font-size: 0.9375rem;
		line-height: 1.7;
		overflow-wrap: break-word;
	}

	:global(.empty-preview) {
		color: var(--color-text-muted);
		opacity: 0.4;
		font-style: italic;
	}

	:global(.wiki-link.unresolved) {
		color: var(--color-warning);
		opacity: 0.7;
		border-bottom: 1px dashed var(--color-warning);
		text-decoration: none;
	}

	.section-visibility-bar {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.5rem 1.5rem;
		border-top: 1px solid var(--color-border);
		background: var(--color-surface);
		flex-shrink: 0;
		flex-wrap: wrap;
	}

	.sv-label {
		font-size: 0.6875rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--color-text-muted);
		opacity: 0.5;
		margin-right: 0.25rem;
		flex-shrink: 0;
	}

	.vis-chip {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.125rem 0.5rem;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 4px;
		font-size: 0.75rem;
		font-family: inherit;
		cursor: pointer;
		transition: border-color 0.15s;
	}

	.vis-chip:hover {
		border-color: var(--color-primary);
	}

	.vis-heading {
		color: var(--color-text);
		opacity: 0.7;
	}

	.vis-badge {
		font-size: 0.625rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.03em;
		padding: 0.0625rem 0.3125rem;
		border-radius: 3px;
		background: var(--color-surface-alt);
		color: var(--color-text-muted);
	}

	.vis-chip.gm-only .vis-badge {
		background: var(--color-danger);
		color: white;
	}

	.vis-chip.observer .vis-badge {
		background: var(--color-warning);
		color: #1a1a2e;
	}

	.vis-chip.private .vis-badge {
		background: #533483;
		color: white;
	}

	.editor-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.5rem 1.5rem;
		border-top: 1px solid var(--color-border);
		font-size: 0.75rem;
		flex-shrink: 0;
	}

	.footer-left {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.footer-right {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.word-count {
		color: var(--color-text-muted);
		opacity: 0.5;
	}

	.saved {
		color: var(--color-success);
		opacity: 0.7;
	}

	.unsaved {
		color: var(--color-warning);
		opacity: 0.9;
	}

	.backlink-badge {
		padding: 0.125rem 0.5rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 4px;
		color: var(--color-text-muted);
		font-size: 0.75rem;
		cursor: help;
	}

	.backlinks-section {
		padding: 0.75rem 1.5rem;
		border-top: 1px solid var(--color-border);
		background: var(--color-surface);
		flex-shrink: 0;
		max-height: 120px;
		overflow-y: auto;
	}

	.backlinks-section h3 {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 0.5rem;
	}

	.backlinks-section ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.backlink-item {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.25rem 0.5rem;
		margin: 0.125rem 0.25rem 0.125rem 0;
		background: var(--color-surface-alt);
		border: 1px solid var(--color-border);
		border-radius: 4px;
		color: var(--color-text);
		font-size: 0.8125rem;
		text-decoration: none;
		transition: background 0.15s;
	}

	.backlink-item:hover {
		background: var(--color-border);
		text-decoration: none;
	}

	.backlink-context {
		font-size: 0.6875rem;
		color: var(--color-text-muted);
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
