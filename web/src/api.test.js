import assert from 'node:assert/strict';
import { after, before, test } from 'node:test';
import { newRequestId } from './api.js';

const originalCrypto = globalThis.crypto;

before(() => {
	Object.defineProperty(globalThis, 'crypto', {
		configurable: true,
		value: {
			randomUUID() {
				throw new TypeError('crypto.randomUUID is not a function');
			},
			getRandomValues() {
				throw new TypeError('crypto.getRandomValues is not a function');
			}
		}
	});
});

after(() => {
	Object.defineProperty(globalThis, 'crypto', {
		configurable: true,
		value: originalCrypto
	});
});

test('newRequestId does not call crypto.randomUUID even when it looks like a function', () => {
	const id = newRequestId();
	assert.match(id, /^\d+-[0-9a-z]+-[0-9a-z]+$/i);
	assert.notEqual(newRequestId(), id);
});
