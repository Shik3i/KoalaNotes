/**
 * AES-GCM encrypt/decrypt utilities for campaign payloads.
 */

export interface EncryptedPayload {
	iv: string;       // base64
	ciphertext: string; // base64
}

/** Encrypt a plaintext string with an AES-GCM key. Returns base64-encoded IV and ciphertext. */
export async function encrypt(plaintext: string, key: CryptoKey): Promise<EncryptedPayload> {
	const enc = new TextEncoder();
	const iv = crypto.getRandomValues(new Uint8Array(12)); // 96-bit IV for GCM

	const ciphertext = await crypto.subtle.encrypt(
		{ name: 'AES-GCM', iv },
		key,
		enc.encode(plaintext)
	);

	return {
		iv: bufferToBase64(iv),
		ciphertext: bufferToBase64(new Uint8Array(ciphertext))
	};
}

/** Decrypt a base64-encoded ciphertext with an AES-GCM key. Returns the plaintext string. */
export async function decrypt(
	payload: EncryptedPayload,
	key: CryptoKey
): Promise<string> {
	const iv = base64ToBuffer(payload.iv);
	const ciphertext = base64ToBuffer(payload.ciphertext);

	const plaintext = await crypto.subtle.decrypt(
		{ name: 'AES-GCM', iv },
		key,
		ciphertext
	);

	const dec = new TextDecoder();
	return dec.decode(plaintext);
}

/** Wrap (encrypt) a CryptoKey with a master key. Returns the wrapped key as base64 + IV. */
export async function wrapKey(
	keyToWrap: CryptoKey,
	masterKey: CryptoKey
): Promise<EncryptedPayload> {
	const iv = crypto.getRandomValues(new Uint8Array(12));
	const wrapped = await crypto.subtle.wrapKey('raw', keyToWrap, masterKey, {
		name: 'AES-GCM',
		iv
	});

	return {
		iv: bufferToBase64(iv),
		ciphertext: bufferToBase64(new Uint8Array(wrapped))
	};
}

/** Unwrap (decrypt) a wrapped key using the master key. */
export async function unwrapKey(
	wrapped: EncryptedPayload,
	masterKey: CryptoKey
): Promise<CryptoKey> {
	const iv = base64ToBuffer(wrapped.iv);
	const wrappedBytes = base64ToBuffer(wrapped.ciphertext);

	return crypto.subtle.unwrapKey(
		'raw',
		wrappedBytes,
		masterKey,
		{ name: 'AES-GCM', iv },
		{ name: 'AES-GCM', length: 256 },
		false,
		['encrypt', 'decrypt']
	);
}

function bufferToBase64(bytes: Uint8Array): string {
	let binary = '';
	for (let i = 0; i < bytes.length; i++) {
		binary += String.fromCharCode(bytes[i]);
	}
	return btoa(binary);
}

function base64ToBuffer(encoded: string): Uint8Array<ArrayBuffer> {
	const binary = atob(encoded);
	const buffer = new ArrayBuffer(binary.length);
	const bytes = new Uint8Array(buffer);
	for (let i = 0; i < binary.length; i++) {
		bytes[i] = binary.charCodeAt(i);
	}
	return bytes;
}
