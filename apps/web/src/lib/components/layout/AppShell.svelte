<script lang="ts">
	import Sidebar from './Sidebar.svelte';
	import TimelinePanel from './TimelinePanel.svelte';
	import SessionTimer from './SessionTimer.svelte';
	import LiveCommentBar from './LiveCommentBar.svelte';

	interface Props {
		sidebarOpen?: boolean;
		timelineOpen?: boolean;
		sessionActive?: boolean;
		sessionElapsed?: number;
		children?: import('svelte').Snippet;
	}

	let {
		sidebarOpen = true,
		timelineOpen = false,
		sessionActive = false,
		sessionElapsed = 0,
		children
	}: Props = $props();
</script>

<div class="app-shell">
	<header class="app-header">
		<div class="header-left">
			<button
				class="toggle-btn"
				onclick={() => sidebarOpen = !sidebarOpen}
				aria-label={sidebarOpen ? 'Close sidebar' : 'Open sidebar'}
			>
				☰
			</button>
			<span class="app-title">KoalaNotes</span>
		</div>
		<div class="header-right">
			<SessionTimer active={sessionActive} elapsed={sessionElapsed} />
			<button
				class="toggle-btn"
				onclick={() => timelineOpen = !timelineOpen}
				aria-label={timelineOpen ? 'Close timeline' : 'Open timeline'}
			>
				{timelineOpen ? '◀' : '▶'}
			</button>
		</div>
	</header>

	<div class="app-body">
		<Sidebar collapsed={!sidebarOpen} />
		<div class="main-content">
			{#if children}
				{@render children()}
			{/if}
		</div>
		<TimelinePanel open={timelineOpen} />
	</div>

	<LiveCommentBar sessionActive={sessionActive} />
</div>

<style>
	.app-shell {
		display: flex;
		flex-direction: column;
		height: 100vh;
		height: 100dvh;
		overflow: hidden;
	}

	.app-header {
		height: var(--header-height);
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 1rem;
		background: var(--color-surface-alt);
		border-bottom: 1px solid var(--color-border);
		flex-shrink: 0;
	}

	.header-left,
	.header-right {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.app-title {
		font-size: 1rem;
		font-weight: 700;
		color: var(--color-text);
	}

	.toggle-btn {
		background: none;
		border: 1px solid var(--color-border);
		color: var(--color-text-muted);
		font-size: 1rem;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		line-height: 1;
		transition: background 0.15s;
	}

	.toggle-btn:hover {
		background: var(--color-border);
	}

	.app-body {
		flex: 1;
		display: flex;
		overflow: hidden;
	}

	.main-content {
		flex: 1;
		display: flex;
		overflow: hidden;
	}
</style>
