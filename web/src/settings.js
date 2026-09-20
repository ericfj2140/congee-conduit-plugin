export function defaultSettings() {
	return {
		index_backend: 'turso',
		postgres_url: '',
		postgres_user: '',
		postgres_password: '',
		product_kinds: [30018, 34560, 30402],
		stall_kinds: [30017, 34550],
		draft_kinds: [30403],
		deletion_kinds: [5],
		index_drafts: false,
		geo_enabled: true,
		vector_enabled: true,
		active_filter: true,
		rank_all_product_reqs: false,
		inject_product_kinds_on_search: true,
		max_results: 50,
		geo_min_prefix_len: 2,
		search_candidate_cap: 2000
	};
}

export function mergeSettings(raw) {
	const base = defaultSettings();
	if (!raw || typeof raw !== 'object') return base;
	return {
		...base,
		...raw,
		product_kinds: Array.isArray(raw.product_kinds) ? raw.product_kinds : base.product_kinds,
		stall_kinds: Array.isArray(raw.stall_kinds) ? raw.stall_kinds : base.stall_kinds,
		draft_kinds: Array.isArray(raw.draft_kinds) ? raw.draft_kinds : base.draft_kinds,
		deletion_kinds: Array.isArray(raw.deletion_kinds) ? raw.deletion_kinds : base.deletion_kinds
	};
}

export function cloneSettings(s) {
	return JSON.parse(JSON.stringify(s));
}

export function settingsEqual(a, b) {
	return JSON.stringify(a) === JSON.stringify(b);
}
