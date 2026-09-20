let requestSeq = 0;

/** Correlation ids for postMessage. Never use crypto.randomUUID — opaque-origin
 * iframes (sandbox without allow-same-origin) often report it as a function and
 * still throw "crypto.randomUUID is not a function" when it is called. */
export function newRequestId() {
	requestSeq += 1;
	return `${requestSeq}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 12)}`;
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

export const GOTO_EVENT = 'conduit-goto';

export function goto(href) {
	window.dispatchEvent(new CustomEvent(GOTO_EVENT, { detail: href }));
}
