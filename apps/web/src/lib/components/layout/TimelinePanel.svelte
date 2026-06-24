<script lang="ts">
	interface Props {
		open?: boolean;
	}

	let { open = false }: Props = $props();
</script>

{#if open}
	<aside class="timeline-panel" aria-label="Session timeline">
		<div class="timeline-header">
			<h2>Timeline</h2>
		</div>
		<div class="timeline-body">
			<!-- Placeholder: timeline entries will be populated in Phase 2 -->
			<div class="placeholder-entry">
				<span class="placeholder-time">--:--:--</span>
				<span class="placeholder-text">No session entries yet.</span>
			</div>
			<div class="placeholder-hint">
				Start a session and use the live comment bar to capture moments.
			</div>
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
		padding: 1rem;
		border-bottom: 1px solid var(--color-border);
	}

	.timeline-header h2 {
		font-size: 1rem;
		font-weight: 600;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.timeline-body {
		flex: 1;
		padding: 1rem;
	}

	.placeholder-entry {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		padding: 0.75rem;
		border-radius: 6px;
		background: var(--color-bg);
		margin-bottom: 0.5rem;
	}

	.placeholder-time {
		font-size: 0.75rem;
		font-family: var(--font-mono);
		color: var(--color-text-muted);
	}

	.placeholder-text {
		font-size: 0.875rem;
		color: var(--color-text-muted);
		opacity: 0.7;
	}

	.placeholder-hint {
		font-size: 0.75rem;
		color: var(--color-text-muted);
		opacity: 0.5;
		font-style: italic;
		margin-top: 0.5rem;
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
