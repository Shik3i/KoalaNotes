<script lang="ts">
	interface Props {
		active?: boolean;
		elapsed?: number; // seconds
	}

	let { active = false, elapsed = 0 }: Props = $props();

	function formatTime(totalSeconds: number): string {
		const hours = Math.floor(totalSeconds / 3600);
		const minutes = Math.floor((totalSeconds % 3600) / 60);
		const seconds = totalSeconds % 60;
		const pad = (n: number) => n.toString().padStart(2, '0');
		return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
	}

	let display = $derived(formatTime(elapsed));
</script>

<div class="session-timer" class:active role="timer" aria-label={`Session timer: ${display}`}>
	<span class="timer-dot" aria-hidden="true"></span>
	<span class="timer-display">{display}</span>
	<!-- Placeholder: start/stop button in Phase 2 -->
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
	}

	.session-timer.active .timer-dot {
		background: var(--color-success);
		animation: pulse 2s infinite;
	}

	@keyframes pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.4; }
	}
</style>
