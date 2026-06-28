<script lang="ts">
	import { auth, register, login, logout } from '$lib/stores/auth';
	import { db } from '$lib/db/database';
	import { deriveKey, generateSalt, deriveSaltFromPassword, generateCampaignKey, exportKeyAsBase64, uint8ArrayToBase64, base64ToUint8Array } from '$lib/crypto/keys';
	import { encrypt as cryptoEncrypt, decrypt as cryptoDecrypt, wrapKey, unwrapKey } from '$lib/crypto/encrypt';
	import { fullSync, pull, push, registerAutoSync, unregisterAutoSync } from '$lib/services/sync';
	import type { SyncStatus } from '$lib/types/models';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let keySetupError = $state('');

	// Master key held in memory (never persisted)
	let masterKey = $state<CryptoKey | null>(null);
	let keyReady = $state(false);

	// Sync state
	let syncStatus = $state<SyncStatus>('idle');
	let syncMessage = $state('');

	async function handleRegister(e: Event) {
		e.preventDefault();
		error = '';
		try {
			await register(email, password);
			await setupKey();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Registration failed';
		}
	}

	async function handleLogin(e: Event) {
		e.preventDefault();
		error = '';
		try {
			await login(email, password);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Login failed';
		}
	}

	function handleLogout() {
		masterKey = null;
		keyReady = false;
		unregisterAutoSync();
		logout();
	}

	const SALT_CAMPAIGN_ID = '__salt__';
	const SALT_RECORD_ID = '__salt_record__';

	async function setupKey() {
		keySetupError = '';
		try {
			const salt = generateSalt();
			const key = await deriveKey(password, salt);
			masterKey = key;
			password = ''; // clear from memory

			// Register auto-sync using the master key and per-campaign keys
			registerAutoSync(async () => {
				const mk = masterKey;
				if (!mk) return;
				await fullSync(
					async (plaintext) => {
						const data = JSON.parse(plaintext);
						const cid = data.campaign?.id;
						if (!cid) throw new Error('missing campaign_id');
						const rec = await db.crypto_keys.where('campaign_id').equals(cid).first();
						if (!rec) throw new Error('no key for campaign');
						const ck = await unwrapKey({ iv: rec.iv, ciphertext: rec.wrapped_campaign_key }, mk);
						return cryptoEncrypt(plaintext, ck);
					},
					async (payload) => {
						throw new Error('pull during auto-sync not supported');
					}
				);
			});

			const saltB64 = uint8ArrayToBase64(salt);

			// Always upsert the salt record (fixed ID ensures single row)
			await db.crypto_keys.put({
				id: SALT_RECORD_ID,
				salt: saltB64,
				wrapped_campaign_key: '',
				iv: '',
				campaign_id: SALT_CAMPAIGN_ID,
				created_at: new Date().toISOString()
			});

			// Generate a campaign key for each existing campaign (skip if exists)
			const campaigns = await db.campaigns.toArray();
			for (const campaign of campaigns) {
				const existing = await db.crypto_keys.where('campaign_id').equals(campaign.id).first();
				if (existing) continue;

				const campaignKey = await generateCampaignKey();
				const wrapped = await wrapKey(campaignKey, masterKey);

				await db.crypto_keys.put({
					id: crypto.randomUUID(),
					salt: '',
					wrapped_campaign_key: wrapped.ciphertext,
					iv: wrapped.iv,
					campaign_id: campaign.id,
					created_at: new Date().toISOString()
				});
			}

			keyReady = true;
		} catch (err) {
			keySetupError = err instanceof Error ? err.message : 'Key setup failed';
			masterKey = null;
		}
	}

	async function unlockKey() {
		keySetupError = '';
		try {
			// Try stored salt first (backward compat), fall back to deterministic salt
			const saltRecord = await db.crypto_keys.where('campaign_id').equals(SALT_CAMPAIGN_ID).first();
			let salt: Uint8Array;
			if (saltRecord?.salt) {
				salt = base64ToUint8Array(saltRecord.salt);
			} else {
				// Deterministic salt derived from password — enables recovery after data loss
				salt = await deriveSaltFromPassword(password);
				// Cache the derived salt for future use
				await db.crypto_keys.put({
					id: SALT_RECORD_ID,
					salt: uint8ArrayToBase64(salt),
					wrapped_campaign_key: '',
					iv: '',
					campaign_id: SALT_CAMPAIGN_ID,
					created_at: new Date().toISOString()
				});
			}
			const key = await deriveKey(password, salt);
			masterKey = key;
			password = ''; // clear from memory

			registerAutoSync(async () => {
				const mk = masterKey;
				if (!mk) return;
				await fullSync(
					async (plaintext) => {
						const data = JSON.parse(plaintext);
						const cid = data.campaign?.id;
						if (!cid) throw new Error('missing campaign_id');
						const rec = await db.crypto_keys.where('campaign_id').equals(cid).first();
						if (!rec) throw new Error('no key for campaign');
						const ck = await unwrapKey({ iv: rec.iv, ciphertext: rec.wrapped_campaign_key }, mk);
						return cryptoEncrypt(plaintext, ck);
					},
					async (payload) => {
						throw new Error('pull during auto-sync not supported');
					}
				);
			});

			keyReady = true;
		} catch (err) {
			keySetupError = err instanceof Error ? err.message : 'Unlock failed';
		}
	}

	async function handleFullSync() {
		const mk = masterKey;
		if (!mk) {
			syncStatus = 'error';
			syncMessage = 'Unlock your encryption key first.';
			return;
		}
		syncStatus = 'syncing';
		syncMessage = 'Syncing...';

		try {
			// Build encrypt/decrypt closures using per-campaign keys
			const encryptFn = async (plaintext: string) => {
				const data = JSON.parse(plaintext);
				const campaignId = data.campaign?.id;
				if (!campaignId) throw new Error('Missing campaign_id in sync payload');

				let keyRecord = await db.crypto_keys.where('campaign_id').equals(campaignId).first();
				if (!keyRecord) {
					const campaignKey = await generateCampaignKey();
					const wrapped = await wrapKey(campaignKey, mk);
					keyRecord = {
						id: crypto.randomUUID(),
						salt: '',
						wrapped_campaign_key: wrapped.ciphertext,
						iv: wrapped.iv,
						campaign_id: campaignId,
						created_at: new Date().toISOString()
					};
					await db.crypto_keys.put(keyRecord);
				}

				const campaignKey = await unwrapKey(
					{ iv: keyRecord.iv, ciphertext: keyRecord.wrapped_campaign_key },
					mk
				);
				return cryptoEncrypt(plaintext, campaignKey);
			};

			const contextualDecrypt = async (payload: { iv: string; ciphertext: string }, campaignId: string) => {
				const keyRecord = await db.crypto_keys.where('campaign_id').equals(campaignId).first();
				if (!keyRecord) throw new Error(`No key for campaign ${campaignId}`);
				const campaignKey = await unwrapKey(
					{ iv: keyRecord.iv, ciphertext: keyRecord.wrapped_campaign_key },
					mk
				);
				return cryptoDecrypt(payload, campaignKey);
			};

			// Pull manually since we need contextual decrypt
			const state = $auth;
			if (!state.token) throw new Error('Not authenticated');

			const pullRes = await fetch(`${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api/sync/pull`, {
				headers: { 'Authorization': `Bearer ${state.token}` }
			});

			if (!pullRes.ok) {
				if (pullRes.status === 401) {
					auth.set({ token: null, accountId: null, email: null });
				}
				const err = await pullRes.json().catch(() => ({ error: 'Pull failed' }));
				throw new Error(err.error || 'Pull failed');
			}

			const pullData = await pullRes.json();
			let pulled = 0;
			for (const blob of pullData.blobs) {
				try {
					const payload = JSON.parse(blob.encrypted_payload);
					const plaintext = await contextualDecrypt(
						{ iv: payload.iv, ciphertext: payload.ciphertext },
						blob.campaign_key_id
					);
					const campaignData = JSON.parse(plaintext);

					// Batch-write all entities for this blob in a single transaction
					await db.transaction('rw', [
						db.campaigns, db.notes, db.sessions,
						db.timeline_entries, db.campaign_members
					], async () => {
						if (campaignData.campaign) await db.campaigns.put(campaignData.campaign);
						if (campaignData.notes?.length) await db.notes.bulkPut(campaignData.notes);
						if (campaignData.sessions?.length) await db.sessions.bulkPut(campaignData.sessions);
						if (campaignData.timeline_entries?.length) await db.timeline_entries.bulkPut(campaignData.timeline_entries);
						if (campaignData.members?.length) await db.campaign_members.bulkPut(campaignData.members);
					});
					pulled++;
				} catch (err) {
					console.error('[sync] failed to process blob', blob.id, err);
				}
			}

			// Push
			const campaigns = await db.campaigns.toArray();
			const blobs = [];
			for (const campaign of campaigns) {
				const notes = await db.notes.where('campaign_id').equals(campaign.id).toArray();
				const sessions = await db.sessions.where('campaign_id').equals(campaign.id).toArray();
				const timeline_entries = await db.timeline_entries.where('campaign_id').equals(campaign.id).toArray();
				const members = await db.campaign_members.where('campaign_id').equals(campaign.id).toArray();

				const campaignData = { campaign, notes, sessions, timeline_entries, members };
				const plaintext = JSON.stringify(campaignData);
				const encrypted = await encryptFn(plaintext);

				blobs.push({
					id: campaign.id,
					campaign_key_id: campaign.id,
					encrypted_payload: JSON.stringify(encrypted),
					vector_clock: new Date().toISOString()
				});
			}

			let pushed = 0;
			if (blobs.length > 0) {
				const pushRes = await fetch(`${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api/sync/push`, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
						'Authorization': `Bearer ${state.token}`
					},
					body: JSON.stringify({ blobs })
				});

				if (!pushRes.ok) {
					if (pushRes.status === 401) {
						auth.set({ token: null, accountId: null, email: null });
					}
					const err = await pushRes.json().catch(() => ({ error: 'Push failed' }));
					throw new Error(err.error || 'Push failed');
				}

				const pushData = await pushRes.json();
				pushed = pushData.accepted ?? 0;
			}

			syncStatus = 'success';
			syncMessage = `Pulled ${pulled} blobs, pushed ${pushed} blobs`;
		} catch (err) {
			syncStatus = 'error';
			syncMessage = err instanceof Error ? err.message : 'Sync failed';
		}
	}
</script>

<h1>Settings</h1>

{#if !$auth.token}
	<!-- Not logged in: show login/register form -->
	<section class="auth-section">
		<h2>Account</h2>
		<form onsubmit={handleLogin}>
			<label for="email">Email</label>
			<input id="email" type="email" bind:value={email} required />

			<label for="password">Password</label>
			<input id="password" type="password" bind:value={password} required />

			{#if error}
				<p class="error">{error}</p>
			{/if}

			<div class="auth-actions">
				<button type="submit">Login</button>
				<button type="button" onclick={handleRegister}>Register</button>
			</div>
		</form>
	</section>
{:else if !keyReady}
	<!-- Logged in but key not unlocked -->
	<section class="auth-section">
		<h2>Encryption Key</h2>
		<p class="hint">Enter your password to unlock your encryption key for sync.</p>
		<label for="unlock-password">Password</label>
		<input id="unlock-password" type="password" bind:value={password} required />

		{#if keySetupError}
			<p class="error">{keySetupError}</p>
		{/if}

		<div class="auth-actions">
			<button type="button" onclick={unlockKey}>Unlock</button>
			<button type="button" onclick={setupKey}>Set Up Key</button>
		</div>

		<hr />

		<button type="button" class="logout-btn" onclick={handleLogout}>Logout</button>
	</section>
{:else}
	<!-- Logged in + key ready -->
	<section class="auth-section">
		<h2>Account</h2>
		<p>Logged in as <strong>{$auth.email}</strong></p>
		<button type="button" class="logout-btn" onclick={handleLogout}>Logout</button>
	</section>

	<section class="sync-section">
		<h2>Sync</h2>
		<button type="button" onclick={handleFullSync} disabled={syncStatus === 'syncing'}>
			{syncStatus === 'syncing' ? 'Syncing...' : 'Full Sync'}
		</button>

		{#if syncMessage}
			<p class="sync-status sync-{syncStatus}">{syncMessage}</p>
		{/if}
	</section>
{/if}

<style>
	h1 {
		font-size: 1.25rem;
		font-weight: 700;
		margin: 1rem;
	}

	h2 {
		font-size: 1rem;
		font-weight: 600;
		margin: 0 0 0.75rem 0;
	}

	.auth-section,
	.sync-section {
		margin: 1rem;
		padding: 1rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		max-width: 400px;
	}

	label {
		display: block;
		font-size: 0.8125rem;
		color: var(--color-text-muted);
		margin-bottom: 0.25rem;
	}

	input {
		width: 100%;
		padding: 0.5rem 0.625rem;
		margin-bottom: 0.75rem;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 4px;
		color: var(--color-text);
		font-size: 0.875rem;
		font-family: inherit;
		box-sizing: border-box;
	}

	input:focus {
		border-color: var(--color-primary);
		outline: none;
	}

	.auth-actions {
		display: flex;
		gap: 0.5rem;
		margin-top: 0.5rem;
	}

	button {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 4px;
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		background: var(--color-primary);
		color: white;
	}

	button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.logout-btn {
		background: var(--color-border);
		color: var(--color-text);
	}

	.error {
		color: var(--color-danger, #e53e3e);
		font-size: 0.8125rem;
		margin: 0.5rem 0;
	}

	.hint {
		font-size: 0.8125rem;
		color: var(--color-text-muted);
		margin-bottom: 0.75rem;
	}

	hr {
		border: none;
		border-top: 1px solid var(--color-border);
		margin: 1rem 0;
	}

	.sync-status {
		font-size: 0.8125rem;
		margin-top: 0.75rem;
	}

	.sync-idle { color: var(--color-text-muted); }
	.sync-syncing { color: var(--color-primary); }
	.sync-success { color: #38a169; }
	.sync-error { color: var(--color-danger, #e53e3e); }
</style>
