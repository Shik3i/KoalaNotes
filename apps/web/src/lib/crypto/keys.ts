/**
 * Key management for client-side encryption.
 *
 * Key hierarchy:
 *   password → PBKDF2 → masterKey (kept in memory)
 *   campaignKey = random AES-256-GCM key (per campaign)
 *   wrappedCampaignKey = AES-GCM encrypt(campaignKey, masterKey) → stored in IndexedDB
 */

const PBKDF2_ITERATIONS = 600000;
const SALT_LENGTH = 16;
const KEY_LENGTH = 256;

/** Derive an AES-GCM key from a password and salt using PBKDF2. */
export async function deriveKey(password: string, salt: Uint8Array): Promise<CryptoKey> {
	const enc = new TextEncoder();
	const keyMaterial = await crypto.subtle.importKey(
		'raw',
		enc.encode(password),
		'PBKDF2',
		false,
		['deriveKey']
	);
	// Ensure salt is backed by ArrayBuffer for Web Crypto compatibility
	const saltBuffer = new Uint8Array(new ArrayBuffer(salt.length));
	saltBuffer.set(salt);
	return crypto.subtle.deriveKey(
		{
			name: 'PBKDF2',
			salt: saltBuffer,
			iterations: PBKDF2_ITERATIONS,
			hash: 'SHA-256'
		},
		keyMaterial,
		{ name: 'AES-GCM', length: KEY_LENGTH },
		false,
		['encrypt', 'decrypt', 'wrapKey', 'unwrapKey']
	);
}

/** Generate a random salt for key derivation. */
export function generateSalt(): Uint8Array {
	return crypto.getRandomValues(new Uint8Array(SALT_LENGTH));
}

/** Generate a new random AES-256-GCM key for a campaign. */
export async function generateCampaignKey(): Promise<CryptoKey> {
	return crypto.subtle.generateKey(
		{ name: 'AES-GCM', length: KEY_LENGTH },
		true,
		['encrypt', 'decrypt']
	);
}

/** Export a CryptoKey to a base64-encoded string. */
export async function exportKeyAsBase64(key: CryptoKey): Promise<string> {
	const raw = await crypto.subtle.exportKey('raw', key);
	const bytes = new Uint8Array(raw);
	let binary = '';
	for (let i = 0; i < bytes.length; i++) {
		binary += String.fromCharCode(bytes[i]);
	}
	return btoa(binary);
}

/** Import a base64-encoded raw key as an AES-GCM CryptoKey. */
export async function importKeyFromBase64(encoded: string): Promise<CryptoKey> {
	const binary = atob(encoded);
	const bytes = new Uint8Array(binary.length);
	for (let i = 0; i < binary.length; i++) {
		bytes[i] = binary.charCodeAt(i);
	}
	return crypto.subtle.importKey(
		'raw',
		bytes,
		{ name: 'AES-GCM', length: KEY_LENGTH },
		false,
		['encrypt', 'decrypt']
	);
}

/** Encode a Uint8Array as a base64 string. */
export function uint8ArrayToBase64(bytes: Uint8Array): string {
	let binary = '';
	for (let i = 0; i < bytes.length; i++) {
		binary += String.fromCharCode(bytes[i]);
	}
	return btoa(binary);
}

/** Decode a base64 string to a Uint8Array. */
export function base64ToUint8Array(encoded: string): Uint8Array<ArrayBuffer> {
	const binary = atob(encoded);
	const buffer = new ArrayBuffer(binary.length);
	const bytes = new Uint8Array(buffer);
	for (let i = 0; i < binary.length; i++) {
		bytes[i] = binary.charCodeAt(i);
	}
	return bytes;
}
