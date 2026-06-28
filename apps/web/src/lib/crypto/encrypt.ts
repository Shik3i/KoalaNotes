/**
 * AES-GCM encrypt/decrypt utilities for campaign payloads.
 */

import { uint8ArrayToBase64, base64ToUint8Array } from './keys';

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
		iv: uint8ArrayToBase64(iv),
		ciphertext: uint8ArrayToBase64(new Uint8Array(ciphertext))
	};
}

/** Decrypt a base64-encoded ciphertext with an AES-GCM key. Returns the plaintext string. */
export async function decrypt(
	payload: EncryptedPayload,
	key: CryptoKey
): Promise<string> {
	const iv = base64ToUint8Array(payload.iv);
	const ciphertext = base64ToUint8Array(payload.ciphertext);

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
		iv: uint8ArrayToBase64(iv),
		ciphertext: uint8ArrayToBase64(new Uint8Array(wrapped))
	};
}

/** Unwrap (decrypt) a wrapped key using the master key. */
export async function unwrapKey(
	wrapped: EncryptedPayload,
	masterKey: CryptoKey
): Promise<CryptoKey> {
	const iv = base64ToUint8Array(wrapped.iv);
	const wrappedBytes = base64ToUint8Array(wrapped.ciphertext);

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


