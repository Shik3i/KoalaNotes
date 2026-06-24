<script lang="ts">
	interface Props {
		sessionActive?: boolean;
	}

	let { sessionActive = false }: Props = $props();
	let comment = $state('');

	function handleSubmit() {
		// Placeholder: save timeline entry in Phase 2
		if (comment.trim()) {
			console.log('[Placeholder] Live comment:', comment);
			comment = '';
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSubmit();
		}
	}
</script>

<div class="live-comment-bar" aria-label="Live comment input">
	<label for="live-comment-input" class="sr-only">Quick note for current session</label>
	<input
		id="live-comment-input"
		type="text"
		bind:value={comment}
		onkeydown={handleKeydown}
		placeholder={sessionActive ? 'Capture a moment... (Enter to save)' : 'Start a session to capture live notes...'}
		disabled={!sessionActive}
		aria-label="Live comment input"
	/>
	<button
		class="save-btn"
		onclick={handleSubmit}
		disabled={!sessionActive || !comment.trim()}
		aria-label="Save comment"
	>
		Save
	</button>
</div>

<style>
	.live-comment-bar {
		height: var(--comment-bar-height);
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0 1rem;
		background: var(--color-surface);
		border-top: 2px solid var(--color-border);
		flex-shrink: 0;
	}

	input {
		flex: 1;
		height: 36px;
		padding: 0 0.75rem;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 6px;
		color: var(--color-text);
		font-size: 0.875rem;
		font-family: inherit;
		outline: none;
		transition: border-color 0.15s;
	}

	input:focus {
		border-color: var(--color-primary);
	}

	input:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	input::placeholder {
		color: var(--color-text-muted);
		opacity: 0.5;
	}

	.save-btn {
		height: 36px;
		padding: 0 1rem;
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: 6px;
		font-size: 0.8125rem;
		font-weight: 600;
		transition: background 0.15s;
	}

	.save-btn:hover:not(:disabled) {
		background: var(--color-primary-hover);
	}

	.save-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}
</style>
