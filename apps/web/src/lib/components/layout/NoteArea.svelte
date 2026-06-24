<script lang="ts">
	// Placeholder: will contain the Markdown editor + preview in Phase 1
</script>

<main class="note-area" aria-label="Note editor">
	<div class="note-placeholder">
		<div class="koala-logo" aria-hidden="true">🐨</div>
		<h1>KoalaNotes</h1>
		<p class="subtitle">Privacy-first, local-first TTRPG campaign notebook</p>
		<div class="status-badge">Early Foundation</div>
		<p class="description">
			Select a campaign from the sidebar or create a new one to begin.
			Your notes are stored locally in your browser.
			No account required.
		</p>
	</div>
</main>

<style>
	.note-area {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		overflow-y: auto;
		background: var(--color-bg);
	}

	.note-placeholder {
		text-align: center;
		max-width: 480px;
	}

	.koala-logo {
		font-size: 4rem;
		margin-bottom: 1rem;
		line-height: 1;
	}

	h1 {
		font-size: 2rem;
		font-weight: 700;
		margin-bottom: 0.25rem;
		color: var(--color-text);
	}

	.subtitle {
		font-size: 1rem;
		color: var(--color-text-muted);
		margin-bottom: 1.5rem;
	}

	.status-badge {
		display: inline-block;
		padding: 0.25rem 0.75rem;
		background: var(--color-warning);
		color: #1a1a2e;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 1.5rem;
	}

	.description {
		font-size: 0.9375rem;
		color: var(--color-text-muted);
		line-height: 1.6;
	}
</style>
