<script>
	import { onMount } from 'svelte';
	import { PATHS, pluginApi } from '../api.js';
	import EventModal from '../lib/EventModal.svelte';
	import InfoLabel from '../lib/InfoLabel.svelte';
	import Tooltip from '../lib/Tooltip.svelte';
	import { displayTitle, kindLabel } from '../lib/eventView.js';

	const PAGE = 50;

	let items = $state.raw([]);
	let total = $state(0);
	let loading = $state(true);
	let err = $state('');
	let status = $state('');
	let modalOpen = $state(false);
	let selectedId = $state('');
	let selectedCoord = $state('');

	async function load(reset) {
		loading = true;
		err = '';
		const offset = reset ? 0 : items.length;
		try {
			const body = { limit: PAGE, offset };
			if (status) body.status = status;
			const res = await pluginApi('POST', PATHS.listListings, body);
			const j = res.json || {};
			if (res.status && res.status >= 400) {
				err = j.error || `Failed (${res.status})`;
				return;
			}
			if (j.ok === false) {
				err = j.error || 'Failed';
				return;
			}
			const next = Array.isArray(j.items) ? j.items : [];
			items = reset ? next : [...items, ...next];
			total = typeof j.total === 'number' ? j.total : items.length;
		} catch (e) {
			err = e instanceof Error ? e.message : 'load failed';
		} finally {
			loading = false;
		}
	}

	function onStatusChange() {
		void load(true);
	}

	function openRow(row) {
		selectedId = row.event_id || '';
		selectedCoord = row.coord || '';
		modalOpen = true;
	}

	function onRowKey(e, row) {
		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			openRow(row);
		}
	}

	onMount(() => {
		void load(true);
	});
</script>

<section class="space-y-6">
	<div class="flex flex-wrap items-end justify-between gap-3">
		<div>
			<p class="text-sm text-neutral-500 dark:text-neutral-400">
				<a class="underline-offset-2 hover:underline" href="#/">Overview</a>
				<span class="px-1">/</span>
				Listings
			</p>
			<div class="mt-1 flex items-center gap-1.5">
				<h2 class="text-lg font-medium text-neutral-900 dark:text-neutral-100">Listings</h2>
				<Tooltip
					label="About Listings"
					tip="Every marketplace document Conduit parsed into its own index (separate from the relay event store and from Congee Audit). Includes NIP-15 stalls (30017/34550), products (30018/34560), and NIP-99 classifieds (30402). Click a row to fetch the event from the relay and open a stall-, product-, or listing-specific view. Title is the parsed name; stalls without a JSON name show the d-tag slug."
				/>
			</div>
			<p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
				Indexed marketplace events. {total} stored. Click a row to inspect the event.
			</p>
		</div>
		<label class="block space-y-1">
			<InfoLabel
				text="Status"
				tip="Filter the index table, not the relay. All shows every stored document. active/inactive is Conduit's listing status (sold, deleted, draft, zero quantity)."
			/>
			<select
				class="rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none focus:ring-2 focus:ring-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-100"
				bind:value={status}
				onchange={onStatusChange}
			>
				<option value="">All</option>
				<option value="active">Active</option>
				<option value="inactive">Inactive</option>
			</select>
		</label>
	</div>

	{#if err}
		<div
			class="rounded-lg border border-red-300 bg-red-50 px-4 py-3 text-sm text-red-900 dark:border-red-900 dark:bg-red-950/40 dark:text-red-200"
		>
			{err}
		</div>
	{/if}

	<div class="overflow-x-auto rounded-lg border border-neutral-200 dark:border-neutral-800">
		<table class="min-w-full text-left text-sm">
			<thead class="bg-neutral-50 text-xs tracking-wide text-neutral-500 uppercase dark:bg-neutral-900 dark:text-neutral-400">
				<tr>
					<th class="px-3 py-2 font-medium">
						<InfoLabel
							text="Title"
							tip="Parsed display name: NIP-99 title tag, or NIP-15 JSON name. If missing, the d-tag (stall or product id, last segment of kind:pubkey:d) is shown."
						/>
					</th>
					<th class="px-3 py-2 font-medium">
						<InfoLabel
							text="Type"
							tip="Stall (30017/34550) is a shop record (currency, shipping). Product (30018/34560) is an item in a stall. Listing (30402) is a NIP-99 classified. They share this table because Conduit indexes all of them."
						/>
					</th>
					<th class="px-3 py-2 font-medium">
						<InfoLabel
							text="Status"
							tip="active: eligible for search intercepts when Active listings only is on. inactive: sold NIP-99 status, zero quantity, draft, or deleted via kind 5."
						/>
					</th>
					<th class="px-3 py-2 font-medium">
						<InfoLabel
							text="Geohash"
							tip="Stored #g geohash used by the geo index. Empty means this document has no location and will not match a REQ that filters on #g."
						/>
					</th>
					<th class="px-3 py-2 font-medium">
						<InfoLabel
							text="Embedding"
							tip="Whether a vector row exists for this document. no means vector rank cannot use it until a rebuild or re-index writes an embedding."
						/>
					</th>
				</tr>
			</thead>
			<tbody>
				{#each items as row (row.coord)}
					<tr
						class="cursor-pointer border-t border-neutral-200 hover:bg-neutral-50 dark:border-neutral-800 dark:hover:bg-neutral-900"
						role="button"
						tabindex="0"
						aria-label="Open {kindLabel(row.kind)} {displayTitle(row)}"
						onclick={() => openRow(row)}
						onkeydown={(e) => onRowKey(e, row)}
					>
						<td class="max-w-[16rem] truncate px-3 py-2 text-neutral-900 dark:text-neutral-100">
							{displayTitle(row)}
						</td>
						<td class="px-3 py-2 text-neutral-700 dark:text-neutral-300">
							{kindLabel(row.kind)}
							<span class="font-mono text-xs text-neutral-500"> {row.kind}</span>
						</td>
						<td class="px-3 py-2 text-neutral-700 dark:text-neutral-300">{row.status}</td>
						<td class="px-3 py-2 font-mono text-neutral-700 dark:text-neutral-300">{row.geohash || '—'}</td>
						<td class="px-3 py-2 text-neutral-700 dark:text-neutral-300">
							{row.has_embedding ? 'yes' : 'no'}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
		{#if !loading && items.length === 0}
			<p class="px-3 py-6 text-sm text-neutral-500 dark:text-neutral-400">No listings in the index yet.</p>
		{/if}
	</div>

	{#if items.length < total}
		<button
			class="rounded-lg border border-neutral-300 px-3 py-2 text-sm font-medium text-neutral-800 disabled:opacity-50 dark:border-neutral-700 dark:text-neutral-100"
			type="button"
			disabled={loading}
			onclick={() => void load(false)}
		>
			{loading ? 'Loading…' : 'Load more'}
		</button>
	{:else if loading}
		<p class="text-sm text-neutral-500 dark:text-neutral-400">Loading…</p>
	{/if}
</section>

<EventModal bind:open={modalOpen} eventId={selectedId} coord={selectedCoord} />
