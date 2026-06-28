<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import AppShell from '$lib/components/layout/AppShell.svelte';
	import '../app.css';

	// Register service worker for offline support
	if (typeof navigator !== 'undefined' && 'serviceWorker' in navigator) {
		navigator.serviceWorker.register('/service-worker.js').catch((err) => console.error('[SW] registration failed', err));
	}

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<meta name="application-name" content="KoalaNotes" />

	<!-- Default SEO — overridden per-page -->
	<title>KoalaNotes — TTRPG Campaign Notebook</title>
	<meta name="description" content="Privacy-first, local-first TTRPG campaign notebook. Manage campaigns, notes, sessions, and timelines encrypted and under your control." />

	<!-- Open Graph -->
	<meta property="og:type" content="website" />
	<meta property="og:site_name" content="KoalaNotes" />
	<meta property="og:title" content="KoalaNotes — TTRPG Campaign Notebook" />
	<meta property="og:description" content="Privacy-first, local-first TTRPG campaign notebook. Manage campaigns, notes, sessions, and timelines encrypted and under your control." />
	<meta property="og:url" content="https://koalanotes.app/" />
	<meta property="og:locale" content="en_US" />

	<!-- Twitter Cards -->
	<meta name="twitter:card" content="summary" />
	<meta name="twitter:title" content="KoalaNotes — TTRPG Campaign Notebook" />
	<meta name="twitter:description" content="Privacy-first, local-first TTRPG campaign notebook." />
</svelte:head>

{@html '<script type="application/ld+json">' + JSON.stringify({"@context":"https://schema.org","@type":"WebApplication","name":"KoalaNotes","url":"https://koalanotes.app/","description":"Privacy-first, local-first TTRPG campaign notebook","applicationCategory":"Productivity","operatingSystem":"Web"}) + '</script>'}

<a href="#main-content" class="skip-link">Skip to main content</a>
<AppShell>
	{@render children()}
</AppShell>
