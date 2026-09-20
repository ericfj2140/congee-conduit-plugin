#!/usr/bin/env node
import { readdirSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const uiAssets = join(dirname(fileURLToPath(import.meta.url)), '..', 'ui', 'assets');
let failed = false;
for (const name of readdirSync(uiAssets)) {
	if (!name.endsWith('.js')) continue;
	const src = readFileSync(join(uiAssets, name), 'utf8');
	if (src.includes('randomUUID')) {
		console.error(`${name} still contains randomUUID; opaque-origin iframes throw when it is called`);
		failed = true;
	}
}
if (failed) process.exit(1);
console.log('ok: built plugin UI does not reference randomUUID');
