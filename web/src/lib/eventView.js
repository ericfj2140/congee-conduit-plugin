import catalog from '../../../kinds.json';

const entries = Array.isArray(catalog?.kinds) ? catalog.kinds : [];
const byKind = new Map(entries.map((k) => [k.kind, k]));

function rolesOf(kind) {
	const e = byKind.get(kind);
	return Array.isArray(e?.roles) ? e.roles : [];
}

export function kindsForRole(role) {
	return entries.filter((k) => Array.isArray(k.roles) && k.roles.includes(role)).map((k) => k.kind);
}

export function defaultKindsForRoles(roles) {
	const seen = new Set();
	const out = [];
	for (const role of roles) {
		for (const n of kindsForRole(role)) {
			if (seen.has(n)) continue;
			seen.add(n);
			out.push(n);
		}
	}
	return out;
}

export function kindRole(kind) {
	const roles = rolesOf(kind);
	if (roles.includes('stall')) return 'stall';
	if (roles.includes('product')) return 'product';
	if (roles.includes('listing')) return 'listing';
	if (roles.includes('listing_draft')) return 'draft';
	if (roles.includes('community')) return 'community';
	if (roles.includes('deletion')) return 'deletion';
	return 'other';
}

export function kindLabel(kind) {
	const e = byKind.get(kind);
	if (e?.name) return e.name;
	return kind ? `Kind ${kind}` : 'Unknown';
}

export function kindDescription(kind) {
	const e = byKind.get(kind);
	return e?.description || '';
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

export function defaultKindTip(role, extra = '') {
	const ns = kindsForRole(role);
	const names = ns.map((n) => `${n} (${kindLabel(n)})`).join(', ');
	return names ? `Defaults from kinds.json: ${names}. ${extra}`.trim() : extra;
}
