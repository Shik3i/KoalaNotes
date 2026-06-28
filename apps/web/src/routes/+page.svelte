<script lang="ts">
	import { liveQuery } from 'dexie';
	import { db } from '$lib/db/database';
	import type { Campaign } from '$lib/types/models';

	let campaigns = $state<Campaign[]>([]);

	$effect(() => {
		const observable = liveQuery(() =>
			db.campaigns.where('archived').equals(0).sortBy('name')
		);
		const sub = observable.subscribe({
			next: (result) => { campaigns = result; },
			error: (err) => console.error('[campaigns]', err)
		});
		return () => sub.unsubscribe();
	});
</script>

<svelte:head>
	<title>KoalaNotes — Privacy-First TTRPG Campaign Notebook</title>
	<meta name="description" content="KoalaNotes is a privacy-first, local-first campaign notebook for TTRPGs. All data encrypted, never leaves your device. Manage campaigns, notes, sessions, and timelines." />
	<meta property="og:title" content="KoalaNotes — Privacy-First TTRPG Campaign Notebook" />
	<meta property="og:description" content="KoalaNotes is a privacy-first, local-first campaign notebook for TTRPGs. All data encrypted, never leaves your device." />
	<meta property="og:url" content="https://koalanotes.app/" />
	<meta name="twitter:title" content="KoalaNotes — Privacy-First TTRPG Campaign Notebook" />
	<meta name="twitter:description" content="KoalaNotes is a privacy-first, local-first campaign notebook for TTRPGs." />
</svelte:head>

<main class="campaign-list-page">
	<div class="welcome">
		<div class="koala-logo" aria-hidden="true">🐨</div>
		<h1>KoalaNotes</h1>
		<p class="subtitle">Privacy-first, local-first TTRPG campaign notebook</p>
	</div>

	{#if campaigns.length === 0}
		<section class="empty-state">
			<p>No campaigns yet.</p>
			<p class="hint">Create one from the sidebar to get started.</p>
		</section>
	{:else}
		<section class="campaign-grid" aria-label="Your campaigns">
			{#each campaigns as c (c.id)}
				<a href="/campaign/{c.id}" class="campaign-card">
					<h2 class="card-name">{c.name}</h2>
					{#if c.description}
						<p class="card-desc">{c.description}</p>
					{/if}
					<div class="card-meta">
						<span class="card-date">Created {new Date(c.created_at).toLocaleDateString()}</span>
						{#if c.system}
							<span class="card-system">{c.system}</span>
						{/if}
					</div>
				</a>
			{/each}
		</section>
	{/if}
</main>

<style>
	.campaign-list-page {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 2rem;
		overflow-y: auto;
	}

	.welcome {
		text-align: center;
		margin-bottom: 2rem;
	}

	.koala-logo {
		font-size: 3rem;
		margin-bottom: 0.5rem;
		line-height: 1;
	}

	h1 {
		font-size: 1.75rem;
		font-weight: 700;
		margin-bottom: 0.25rem;
	}

	.subtitle {
		font-size: 0.9375rem;
		color: var(--color-text-muted);
	}

	.empty-state {
		text-align: center;
		color: var(--color-text-muted);
		padding: 3rem 1rem;
	}

	.hint {
		font-size: 0.875rem;
		opacity: 0.6;
		margin-top: 0.5rem;
	}

	.campaign-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1rem;
		width: 100%;
		max-width: 960px;
	}

	.campaign-card {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		padding: 1.25rem;
		text-decoration: none;
		color: var(--color-text);
		transition: border-color 0.15s, background 0.15s;
	}

	.campaign-card:hover {
		border-color: var(--color-primary);
		background: var(--color-surface-alt);
		text-decoration: none;
	}

	.card-name {
		font-size: 1.125rem;
		font-weight: 600;
		margin-bottom: 0.5rem;
	}

	.card-desc {
		font-size: 0.8125rem;
		color: var(--color-text-muted);
		margin-bottom: 0.75rem;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		line-clamp: 2;
		overflow: hidden;
	}

	.card-meta {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		font-size: 0.75rem;
		color: var(--color-text-muted);
		opacity: 0.7;
	}

	.card-system {
		background: var(--color-surface-alt);
		padding: 0.125rem 0.5rem;
		border-radius: 4px;
	}
</style>
