let requestSeq = 0;

export function newRequestId() {
	requestSeq += 1;
	if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
		return `${requestSeq}-${crypto.randomUUID()}`;
	}
	const bytes = new Uint8Array(16);
	if (typeof crypto !== 'undefined' && typeof crypto.getRandomValues === 'function') {
		crypto.getRandomValues(bytes);
	} else {
		for (let i = 0; i < bytes.length; i++) {
			bytes[i] = Math.floor(Math.random() * 256);
		}
	}
	const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
	return `${requestSeq}-${hex}`;
}

export function pluginApi(method, path, body) {
	const id = newRequestId();
	return new Promise((resolve, reject) => {
		const timer = setTimeout(() => {
			window.removeEventListener('message', onMsg);
			reject(new Error('plugin API timed out'));
		}, 60_000);
		function onMsg(ev) {
			const d = ev.data;
			if (!d || d.type !== 'congee:plugin-api-result' || d.id !== id) return;
			clearTimeout(timer);
			window.removeEventListener('message', onMsg);
			resolve({ status: d.status, json: d.json });
		}
		window.addEventListener('message', onMsg);
		window.parent.postMessage({ type: 'congee:plugin-api', id, method, path, body }, '*');
	});
}

export const PATHS = {
	plugin: '/api/plugins/conduit',
	settings: '/api/plugins/conduit/settings',
	status: '/api/plugins/conduit/status',
	rebuild: '/api/plugins/conduit/actions/rebuild',
	testStore: '/api/plugins/conduit/actions/test_store',
	testEmbed: '/api/plugins/conduit/actions/test_embed',
	ensureAssets: '/api/plugins/conduit/actions/ensure_assets',
	listListings: '/api/plugins/conduit/actions/list_listings',
	listEmbeddings: '/api/plugins/conduit/actions/list_embeddings',
	getEvent: '/api/plugins/conduit/actions/get_event'
};

export function applyTheme(theme) {
	document.documentElement.classList.toggle('dark', theme !== 'light');
}

export function onHostMessage(ev) {
	const d = ev.data;
	if (!d || d.type !== 'congee:theme') return;
	if (d.theme === 'dark' || d.theme === 'light') applyTheme(d.theme);
}
