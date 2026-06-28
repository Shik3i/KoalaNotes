<script lang="ts">
	interface Props {
		active?: boolean;
		elapsed?: number;
		sessionName?: string;
	}

	let { active = false, elapsed = 0, sessionName = '' }: Props = $props();

	function formatTime(totalSeconds: number): string {
		const h = Math.floor(totalSeconds / 3600);
		const m = Math.floor((totalSeconds % 3600) / 60);
		const s = totalSeconds % 60;
		const pad = (n: number) => n.toString().padStart(2, '0');
		return `${pad(h)}:${pad(m)}:${pad(s)}`;
	}

	let display = $derived(formatTime(elapsed));
	let label = $derived(active ? (sessionName ? `${sessionName}: ${display}` : `Session: ${display}`) : 'No active session');
</script>

<div
	class="session-timer"
	class:active
	role="timer"
	aria-label={label}
	title={label}
>
	<span class="timer-dot" aria-hidden="true"></span>
	<span class="timer-display">{display}</span>
	{#if active}
		<span class="session-name" aria-hidden="true">{sessionName}</span>
	{/if}
</div>

<style>
	.session-timer {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.25rem 0.75rem;
		border-radius: 6px;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		font-size: 0.8125rem;
		font-family: var(--font-mono);
		color: var(--color-text-muted);
		cursor: default;
		user-select: none;
	}

	.session-timer.active {
		border-color: var(--color-success);
		color: var(--color-success);
	}

	.timer-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--color-text-muted);
		flex-shrink: 0;
	}

	.session-timer.active .timer-dot {
		background: var(--color-success);
		animation: pulse 2s infinite;
	}

	.session-name {
		font-family: var(--font-sans);
		font-size: 0.6875rem;
		opacity: 0.6;
		max-width: 120px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	@keyframes pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.4; }
	}
</style>
