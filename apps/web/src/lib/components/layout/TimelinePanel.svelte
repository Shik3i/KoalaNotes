<script lang="ts">
	import { getNoteTitleMap } from '$lib/db/wiki';
	import { formatElapsed } from '$lib/utils/export';
	import type { TimelineEntry } from '$lib/types/models';

	interface Props {
		open?: boolean;
		entries?: TimelineEntry[];
		active?: boolean;
		campaignId?: string;
	}

	let { open = false, entries = [], active = false, campaignId = '' }: Props = $props();

	let noteTitles = $state<Map<string, string>>(new Map());

	// Resolve note titles whenever entries change
	$effect(() => {
		const ids = [...new Set(entries.map(e => e.note_id).filter(Boolean) as string[])];
		if (ids.length === 0) {
			noteTitles = new Map();
			return;
		}
		getNoteTitleMap(ids).then(m => { noteTitles = m; });
	});

	function formatClock(iso: string): string {
		try {
			return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
		} catch {
			return '--:--:--';
		}
	}

</script>

{#if open}
	<aside class="timeline-panel" aria-label="Session timeline">
		<div class="timeline-header">
			<h2>Timeline</h2>
			{#if active}
				<span class="live-badge">Live</span>
			{/if}
		</div>
		<div class="timeline-body">
			{#if entries.length === 0}
				<div class="empty-state">
					<p class="empty-text">No entries yet.</p>
					<p class="empty-hint">Use the live comment bar at the bottom to capture moments.</p>
				</div>
			{:else}
				<ul class="entry-list" role="log" aria-label="Session timeline entries" aria-live="polite">
					{#each entries as entry (entry.id)}
						<li class="entry-item">
							<div class="entry-meta">
								<span class="entry-time">{formatClock(entry.clock_time)}</span>
								<span class="entry-elapsed" title="Session elapsed">{formatElapsed(entry.session_elapsed)}</span>
							</div>
							<p class="entry-content">{entry.content}</p>
							{#if entry.note_id && campaignId}
								{@const noteTitle = noteTitles.get(entry.note_id) ?? 'Untitled'}
								<a
									href="/campaign/{campaignId}/notes/{entry.note_id}"
									class="entry-context"
									aria-label="Open note: {noteTitle}"
								>
									via {noteTitle}
								</a>
							{/if}
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	</aside>
{/if}

<style>
	.timeline-panel {
		width: var(--timeline-width);
		height: 100%;
		background: var(--color-surface);
		border-left: 1px solid var(--color-border);
		display: flex;
		flex-direction: column;
		overflow-y: auto;
		flex-shrink: 0;
	}

	.timeline-header {
		padding: 0.75rem 1rem;
		border-bottom: 1px solid var(--color-border);
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
	}

	.timeline-header h2 {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.live-badge {
		font-size: 0.625rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 0.125rem 0.5rem;
		background: var(--color-success);
		color: #1a1a2e;
		border-radius: 9999px;
		font-weight: 700;
		animation: pulse 2s infinite;
	}

	@keyframes pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.6; }
	}

	.timeline-body {
		flex: 1;
		padding: 0.75rem;
		overflow-y: auto;
	}

	.empty-state {
		text-align: center;
		padding: 2rem 0.5rem;
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
		font-style: italic;
	}

	.entry-list {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.entry-item {
		padding: 0.625rem 0.75rem;
		border-radius: 6px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
	}

	.entry-meta {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.25rem;
	}

	.entry-time {
		font-size: 0.6875rem;
		font-family: var(--font-mono);
		color: var(--color-text-muted);
		opacity: 0.6;
	}

	.entry-elapsed {
		font-size: 0.6875rem;
		font-family: var(--font-mono);
		color: var(--color-text-muted);
		opacity: 0.4;
	}

	.entry-content {
		font-size: 0.8125rem;
		color: var(--color-text);
		line-height: 1.5;
		word-break: break-word;
	}

	.entry-context {
		display: inline-block;
		margin-top: 0.25rem;
		font-size: 0.625rem;
		color: var(--color-primary);
		opacity: 0.5;
		text-decoration: none;
		transition: opacity 0.15s;
	}

	.entry-context:hover {
		opacity: 1;
		text-decoration: underline;
	}

	@media (max-width: 1024px) {
		.timeline-panel {
			position: fixed;
			right: 0;
			top: var(--header-height);
			bottom: var(--comment-bar-height);
			z-index: 100;
			box-shadow: -2px 0 8px rgba(0, 0, 0, 0.3);
		}
	}
</style>
