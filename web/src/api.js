export function pluginApi(method, path, body) {
	const id = crypto.randomUUID();
	return new Promise((resolve) => {
		function onMsg(ev) {
			const d = ev.data;
			if (!d || d.type !== 'congee:plugin-api-result' || d.id !== id) return;
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
