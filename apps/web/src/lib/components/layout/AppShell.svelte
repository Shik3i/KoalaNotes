<script lang="ts">
	import { liveQuery } from 'dexie';
	import { db } from '$lib/db/database';
	import { observeActiveSession } from '$lib/db/sessions';
	import { observeSessionEntries } from '$lib/db/timeline';
	import { viewingRole } from '$lib/stores/role';
	import { auth } from '$lib/stores/auth';
	import { page } from '$app/stores';
	import Sidebar from './Sidebar.svelte';
	import TimelinePanel from './TimelinePanel.svelte';
	import SessionTimer from './SessionTimer.svelte';
	import LiveCommentBar from './LiveCommentBar.svelte';
	import type { Role, Session, TimelineEntry, SyncStatus } from '$lib/types/models';

	interface Props {
		children?: import('svelte').Snippet;
	}

	let { children }: Props = $props();

	// Sidebar / timeline toggle state
	let sidebarOpen = $state(true);
	let timelineOpen = $state(false);

	// Active session state from Dexie
	let activeSession = $state<Session | null>(null);
	$effect(() => {
		let destroyed = false;
		let hadSession = false;
		const observable = observeActiveSession();
		const sub = observable.subscribe({
			next: (result) => {
				if (destroyed) return;
				const session = result ?? null;
				// Auto-open timeline only on null→session transition (session start)
				if (session && !hadSession) timelineOpen = true;
				hadSession = !!session;
				activeSession = session;
			},
			error: (err) => console.error('[active session]', err)
		});
		return () => { destroyed = true; sub.unsubscribe(); };
	});

	// Timer: compute elapsed seconds from started_at
	let elapsedSeconds = $state(0);
	let timerInterval: ReturnType<typeof setInterval> | undefined;

	$effect(() => {
		if (activeSession?.started_at) {
			// Initialize elapsed from session start
			const started = new Date(activeSession.started_at).getTime();
			if (Number.isNaN(started)) {
				elapsedSeconds = 0;
				return;
			}
			elapsedSeconds = Math.floor((Date.now() - started) / 1000);

			// Tick every second
			timerInterval = setInterval(() => {
				elapsedSeconds = Math.floor((Date.now() - started) / 1000);
			}, 1000);

			return () => {
				if (timerInterval) clearInterval(timerInterval);
			};
		} else {
			elapsedSeconds = 0;
			if (timerInterval) {
				clearInterval(timerInterval);
				timerInterval = undefined;
			}
		}
	});

	// Timeline entries for active session
	let timelineEntries = $state<TimelineEntry[]>([]);
	$effect(() => {
		if (!activeSession) {
			timelineEntries = [];
			return;
		}
		const observable = observeSessionEntries(activeSession.id);
		const sub = observable.subscribe({
			next: (result) => { timelineEntries = result; },
			error: (err) => console.error('[timeline]', err)
		});
		return () => sub.unsubscribe();
	});

	// Current note from URL for comment context
	let currentNoteId = $derived($page.params.noteId ?? undefined);

	// Sync status: derived from auth token
	let syncStatus = $derived($auth.token ? 'success' : 'idle');
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
			<select
				class="role-select"
				value={$viewingRole}
				onchange={(e) => viewingRole.set(e.currentTarget.value as Role)}
				aria-label="View as role"
			>
				<option value="gm">GM</option>
				<option value="player">Player</option>
				<option value="observer">Observer</option>
			</select>
			<SessionTimer
				active={activeSession !== null}
				elapsed={elapsedSeconds}
				sessionName={activeSession?.name ?? ''}
			/>
			<a href="/settings" class="settings-link" title="Settings" aria-label="Settings">
				⚙
			</a>
			{#if $auth.token && syncStatus === 'success'}
				<span class="sync-indicator sync-success" title="Connected (sync available)">✓</span>
			{/if}
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
		<TimelinePanel
			open={timelineOpen}
			entries={timelineEntries}
			active={activeSession !== null}
			campaignId={activeSession?.campaign_id ?? $page.params.campaignId}
		/>
	</div>

	<LiveCommentBar
		activeSession={activeSession}
		elapsedSeconds={elapsedSeconds}
		currentNoteId={currentNoteId}
	/>
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

	.role-select {
		font-size: 0.75rem;
		padding: 0.25rem 0.5rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 4px;
		color: var(--color-text-muted);
		font-family: inherit;
		cursor: pointer;
		transition: border-color 0.15s;
	}

	.role-select:hover {
		border-color: var(--color-primary);
	}

	.settings-link {
		text-decoration: none;
		color: var(--color-text-muted);
		font-size: 1.1rem;
		line-height: 1;
		transition: color 0.15s;
	}

	.settings-link:hover {
		color: var(--color-text);
	}

	.sync-indicator {
		font-size: 0.875rem;
		line-height: 1;
	}

	.sync-success { color: #38a169; }

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
