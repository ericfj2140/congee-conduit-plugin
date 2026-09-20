export function kindRole(kind) {
	if (kind === 30017 || kind === 34550) return 'stall';
	if (kind === 30018 || kind === 34560) return 'product';
	if (kind === 30402) return 'listing';
	if (kind === 30403) return 'draft';
	return 'other';
}

export function kindLabel(kind) {
	switch (kind) {
		case 30017:
			return 'Stall';
		case 34550:
			return 'Stall';
		case 30018:
			return 'Product';
		case 34560:
			return 'Product';
		case 30402:
			return 'Listing';
		case 30403:
			return 'Draft listing';
		case 5:
			return 'Deletion';
		default:
			return kind ? `Kind ${kind}` : 'Unknown';
	}
}

export function displayTitle(row) {
	const title = typeof row?.title === 'string' ? row.title.trim() : '';
	if (title) return title;
	const d = typeof row?.d_tag === 'string' ? row.d_tag.trim() : '';
	if (d) return d;
	const coord = typeof row?.coord === 'string' ? row.coord.trim() : '';
	const parts = coord.split(':');
	if (parts.length >= 3 && parts.slice(2).join(':')) return parts.slice(2).join(':');
	return '—';
}

export function parseJSONObject(content) {
	if (typeof content !== 'string' || !content.trim()) return null;
	try {
		const v = JSON.parse(content);
		return v && typeof v === 'object' && !Array.isArray(v) ? v : null;
	} catch {
		return null;
	}
}

export function tagValues(tags, name) {
	const out = [];
	if (!Array.isArray(tags)) return out;
	for (const t of tags) {
		if (Array.isArray(t) && t.length >= 2 && t[0] === name && t[1]) out.push(t[1]);
	}
	return out;
}

export function isHttpUrl(s) {
	if (typeof s !== 'string') return false;
	try {
		const u = new URL(s);
		return u.protocol === 'http:' || u.protocol === 'https:';
	} catch {
		return false;
	}
}
