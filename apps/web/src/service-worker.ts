/// <reference types="@sveltejs/kit" />
/// <reference no-default-lib="true"/>
/// <reference lib="esnext" />
/// <reference lib="webworker" />

import { build, files, version } from '$service-worker';

const sw = self as unknown as ServiceWorkerGlobalScope;

const CACHE_NAME = `koalanotes-${version}`;

const ASSETS = [
	...build,
	...files
];

// ---- Install: pre-cache all static assets ----

sw.addEventListener('install', (event) => {
	async function preCache() {
		const cache = await caches.open(CACHE_NAME);
		await cache.addAll(ASSETS);
	}
	event.waitUntil(preCache());
});

// ---- Activate: remove old caches ----

sw.addEventListener('activate', (event) => {
	async function deleteOldCaches() {
		for (const key of await caches.keys()) {
			if (key !== CACHE_NAME) {
				await caches.delete(key);
			}
		}
	}
	event.waitUntil(deleteOldCaches());
});

// ---- Fetch: cache-first for assets, network-first for API ----

sw.addEventListener('fetch', (event) => {
	const url = new URL(event.request.url);

	// Only handle same-origin requests
	if (url.origin !== location.origin) return;

	// Skip non-GET
	if (event.request.method !== 'GET') return;

	// API calls: network-first, fall back to cache
	if (url.pathname.startsWith('/api/')) {
		event.respondWith(networkFirst(event.request));
		return;
	}

	// Navigation requests (SPA fallback): network-first, fall back to cached index
	if (event.request.mode === 'navigate') {
		event.respondWith(networkFirst(event.request, '/'));
		return;
	}

	// Static assets: cache-first
	event.respondWith(cacheFirst(event.request));
});

async function cacheFirst(request: Request): Promise<Response> {
	const cached = await caches.match(request);
	if (cached) return cached;
	try {
		const response = await fetch(request);
		if (response.ok) {
			const cache = await caches.open(CACHE_NAME);
			cache.put(request, response.clone());
		}
		return response;
	} catch {
		return new Response('Offline', { status: 503 });
	}
}

async function networkFirst(request: Request, fallbackUrl?: string): Promise<Response> {
	try {
		const response = await fetch(request);
		if (response.ok) {
			const cache = await caches.open(CACHE_NAME);
			cache.put(request, response.clone());
		}
		return response;
	} catch {
		const cached = await caches.match(request);
		if (cached) return cached;
		if (fallbackUrl) {
			const fallback = await caches.match(fallbackUrl);
			if (fallback) return fallback;
		}
		return new Response('Offline', { status: 503 });
	}
}
