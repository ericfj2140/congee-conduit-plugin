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
	let modalOpen = $state(false);
	let selectedId = $state('');
	let selectedCoord = $state('');

	async function load(reset) {
		loading = true;
		err = '';
		const offset = reset ? 0 : items.length;
		try {
			const res = await pluginApi('POST', PATHS.listEmbeddings, { limit: PAGE, offset });
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
	<div>
		<p class="text-sm text-neutral-500 dark:text-neutral-400">
			<a class="underline-offset-2 hover:underline" href="#/">Overview</a>
			<span class="px-1">/</span>
			Embeddings
		</p>
		<div class="mt-1 flex items-center gap-1.5">
			<h2 class="text-lg font-medium text-neutral-900 dark:text-neutral-100">Embeddings</h2>
			<Tooltip
				label="About Embeddings"
				tip="Each row is a search vector in the Conduit index, not a Congee Audit → Events log line. Audit is connection activity; this table is marketplace documents Conduit embedded for ranking. Kind 34550 is a NIP-15 stall (a shop), so it will not show up if you browse classified listings (30402) in Audit. Click a row to fetch that event from the relay store and open a stall-, product-, or listing-specific view."
			/>
		</div>
		<p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
			Stored vectors used for search ranking. {total} rows. Click a row to open the event. Vectors
			themselves are not shown.
		</p>
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
							tip="Display name Conduit parsed when indexing. Classified listings (kind 30402) use the title tag. NIP-15 products and stalls use the JSON name field. Stalls often omit name, so this falls back to the d-tag (the stall slug, last segment of Coord) or an em dash. This is not a Congee audit title."
						/>
					</th>
					<th class="px-3 py-2 font-medium">
						<InfoLabel
							text="Type"
							tip="Stall = NIP-15 shop (30017 or parameterized 34550). Product = NIP-15 item (30018 / 34560). Listing = NIP-99 classified (30402). Kind 34550 is a stall, not a classified listing."
						/>
					</th>
					<th class="px-3 py-2 font-medium">
						<InfoLabel
							text="Status"
							tip="Index status of the document this vector belongs to. active can appear in search intercepts when Active listings only is on. inactive covers sold, deleted, draft, or zero-quantity."
						/>
					</th>
					<th class="px-3 py-2 font-medium">
						<InfoLabel
							text="Model"
							tip="Which embedder wrote this vector. fake-bow-384 is Conduit's built-in test embedder (CONDUIT_EMBEDDER=fake, or no ONNX file loaded). It hashes words into a 384-dimension bag-of-words vector for ranking tests — not a neural semantic model. Production ranking uses the ONNX ModelID after you load a real model and rebuild."
						/>
					</th>
					<th class="px-3 py-2 font-medium">
						<InfoLabel
							text="Dim"
							tip="Length of the stored vector. Must match the embedder. fake-bow-384 is always 384. After switching models, rebuild so old dimensions are replaced."
						/>
					</th>
					<th class="px-3 py-2 font-medium">
						<InfoLabel
							text="Coord"
							tip="Addressable event coordinate kind:pubkey:d-tag. The last segment is the stall or listing id. Click the row to load that event from the relay."
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
							<span class="font-mono text-xs text-neutral-500"> {row.kind || '—'}</span>
						</td>
						<td class="px-3 py-2 text-neutral-700 dark:text-neutral-300">{row.status || '—'}</td>
						<td class="px-3 py-2 font-mono text-neutral-700 dark:text-neutral-300">{row.model}</td>
						<td class="px-3 py-2 font-mono text-neutral-700 dark:text-neutral-300">{row.dim}</td>
						<td
							class="max-w-[14rem] truncate px-3 py-2 font-mono text-xs text-neutral-600 dark:text-neutral-400"
							title={row.coord}
						>
							{row.coord}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
		{#if !loading && items.length === 0}
			<p class="px-3 py-6 text-sm text-neutral-500 dark:text-neutral-400">No embeddings stored yet.</p>
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
