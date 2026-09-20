import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';
import { dirname, extname, join, normalize } from 'node:path';
import { fileURLToPath } from 'node:url';

const webDir = dirname(fileURLToPath(import.meta.url));
const root = join(webDir, '..');
const uiDir = join(root, '..', 'ui');
const port = Number(process.env.SANDBOX_PORT || 4179);

const types = {
	'.html': 'text/html; charset=utf-8',
	'.js': 'text/javascript; charset=utf-8',
	'.css': 'text/css; charset=utf-8',
	'.svg': 'image/svg+xml',
	'.json': 'application/json'
};

function safeJoin(base, rel) {
	const p = normalize(join(base, rel));
	if (!p.startsWith(base)) throw new Error('path');
	return p;
}

const server = createServer(async (req, res) => {
	const origin = req.headers.origin;
	const cors = {
		'access-control-allow-origin': origin === 'null' ? 'null' : '*',
		'cross-origin-resource-policy': 'cross-origin',
		'access-control-allow-methods': 'GET, HEAD, OPTIONS',
		vary: 'Origin'
	};
	if (req.method === 'OPTIONS') {
		res.writeHead(204, cors);
		res.end();
		return;
	}
	try {
		const url = new URL(req.url || '/', `http://127.0.0.1:${port}`);
		let file;
		if (url.pathname === '/' || url.pathname === '/host.html') {
			file = join(webDir, 'sandbox-host.html');
		} else if (url.pathname === '/ui' || url.pathname === '/ui/') {
			file = join(uiDir, 'index.html');
		} else if (url.pathname.startsWith('/ui/')) {
			file = safeJoin(uiDir, url.pathname.slice('/ui/'.length) || 'index.html');
		} else {
			res.writeHead(404, cors);
			res.end('not found');
			return;
		}
		const body = await readFile(file);
		res.writeHead(200, {
			'content-type': types[extname(file)] || 'application/octet-stream',
			'cache-control': 'no-store',
			...cors
		});
		res.end(body);
	} catch {
		res.writeHead(404, cors);
		res.end('not found');
	}
});

server.listen(port, '127.0.0.1', () => {
	console.log(`sandbox host http://127.0.0.1:${port}/`);
});
