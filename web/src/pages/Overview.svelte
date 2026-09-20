<script>
	import Tooltip from '../lib/Tooltip.svelte';

	let { status = {}, ready = false, busy = false, onrebuild } = $props();

	function n(v) {
		return typeof v === 'number' ? v : 0;
	}

	async function confirmRebuild() {
		if (!confirm('Rebuild the Conduit index from stored relay events?')) return;
		await onrebuild();
	}
</script>

<section class="space-y-6">
	<div>
		<div class="flex items-center gap-1.5">
			<h2 class="text-lg font-medium text-neutral-900 dark:text-neutral-100">Overview</h2>
			<Tooltip
				label="About Overview"
				tip="Counts come from the Conduit index database (conduit-index.db or Postgres), not from Congee Audit → Events. Audit is the relay activity log. Listings include stalls (kind 34550/30017), products, and classifieds that were backfilled from the relay or stored live. Embeddings are vector rows used for search rank — Title is the parsed name (or d-tag when a stall has no JSON name), Model is the embedder id (fake-bow-384 is the test bag-of-words embedder). Click Listings or Embeddings, then a row, to load the event from the relay."
			/>
		</div>
		<p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
			Index health, backfill, and REQ intercept counters.
		</p>
	</div>

	<div class="grid gap-3 sm:grid-cols-2">
		<div class="rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900">
			<div class="flex items-center gap-1.5 text-xs tracking-wide text-neutral-500 uppercase dark:text-neutral-400">
				Backend
				<Tooltip
					label="About Backend"
					tip="Which database the plugin index is using (Turso/libSQL file or Postgres). This is not the relay event store. Events still live in Congee; this store only keeps parsed listings, geo, and embeddings."
				/>
			</div>
			<div class="mt-1 text-sm font-medium text-neutral-900 dark:text-neutral-100">
				{status.backend || '—'}
			</div>
		</div>
		<a
			href="#/listings"
			class="rounded-lg border border-neutral-200 bg-white p-4 transition-colors hover:border-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:hover:border-neutral-500"
		>
			<div class="flex items-center gap-1.5 text-xs tracking-wide text-neutral-500 uppercase dark:text-neutral-400">
				Listings
				<Tooltip
					label="About Listings count"
					tip="Indexed marketplace documents: stalls, products, and classifieds. Active vs inactive is Conduit's index status, not whether the event exists in Audit. Click to browse and open events."
				/>
			</div>
			<div class="mt-1 text-sm font-medium text-neutral-900 dark:text-neutral-100">
				{n(status.active)} active · {n(status.inactive)} inactive
			</div>
			<div class="mt-2 text-xs text-neutral-500 dark:text-neutral-400">View indexed listings</div>
		</a>
		<a
			href="#/embeddings"
			class="rounded-lg border border-neutral-200 bg-white p-4 transition-colors hover:border-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:hover:border-neutral-500"
		>
			<div class="flex items-center gap-1.5 text-xs tracking-wide text-neutral-500 uppercase dark:text-neutral-400">
				Embeddings
				<Tooltip
					label="About Embeddings count"
					tip="Number of stored search vectors. Two rows can be two stalls (kind 34550) that were backfilled from the relay. They will not appear under Audit unless you filter that kind. Click through to inspect each event."
				/>
			</div>
			<div class="mt-1 text-sm font-medium text-neutral-900 dark:text-neutral-100">
				{n(status.embeddings)}
			</div>
			<div class="mt-2 text-xs text-neutral-500 dark:text-neutral-400">View embedding rows</div>
		</a>
		<div class="rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900">
			<div class="flex items-center gap-1.5 text-xs tracking-wide text-neutral-500 uppercase dark:text-neutral-400">
				Backfill
				<Tooltip
					label="About Backfill"
					tip="Startup scan of existing relay events into the Conduit index. idle means the scan finished or has not started. Listings and embeddings can appear after backfill even on a 'fresh' plugin if the relay already had stall/product events."
				/>
			</div>
			<div class="mt-1 text-sm font-medium text-neutral-900 dark:text-neutral-100">
				{status.backfill || 'idle'}
			</div>
		</div>
	</div>

	{#if n(status.embedding_mismatch) > 0}
		<div
			class="rounded-lg border border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-900 dark:border-amber-800 dark:bg-amber-950/50 dark:text-amber-200"
		>
			{n(status.embedding_mismatch)} listing{n(status.embedding_mismatch) === 1 ? '' : 's'} missing
			or mismatched embeddings. Vector rank may be incomplete until a rebuild finishes.
		</div>
	{/if}

	<div class="rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900">
		<div class="flex items-center gap-1.5 text-xs tracking-wide text-neutral-500 uppercase dark:text-neutral-400">
			Intercepts
			<Tooltip
				label="About Intercepts"
				tip="Seen: REQs matched to Conduit. Passthrough: left to the relay. Respond: Conduit returned ranked event IDs. Reshape is unused (kinds are not rewritten)."
			/>
		</div>
		<dl class="mt-3 grid grid-cols-2 gap-3 text-sm sm:grid-cols-4">
			<div>
				<dt class="flex items-center gap-1 text-neutral-500 dark:text-neutral-400">
					Seen
					<Tooltip
						label="About Seen"
						tip="REQs whose filters matched Conduit's traffic subscription (product/stall kinds, search, or #g). Not every seen REQ is ranked."
					/>
				</dt>
				<dd class="font-medium text-neutral-900 dark:text-neutral-100">{n(status.intercept_n)}</dd>
			</div>
			<div>
				<dt class="flex items-center gap-1 text-neutral-500 dark:text-neutral-400">
					Passthrough
					<Tooltip
						label="About Passthrough"
						tip="Conduit left the REQ to the relay unchanged. Typical when there is no search/#g, Rank all product REQs is off, or the REQ kinds are not product kinds."
					/>
				</dt>
				<dd class="font-medium text-neutral-900 dark:text-neutral-100">{n(status.passthrough_n)}</dd>
			</div>
			<div>
				<dt class="flex items-center gap-1 text-neutral-500 dark:text-neutral-400">
					Respond
					<Tooltip
						label="About Respond"
						tip="Conduit answered with ranked listing event IDs from its index. The host hydrates those IDs from the relay store. Client filter kinds are not rewritten."
					/>
				</dt>
				<dd class="font-medium text-neutral-900 dark:text-neutral-100">{n(status.respond_n)}</dd>
			</div>
			<div>
				<dt class="flex items-center gap-1 text-neutral-500 dark:text-neutral-400">
					Reshape
					<Tooltip
						label="About Reshape"
						tip="Unused. Conduit does not rewrite the client's REQ kinds. This counter should stay at 0."
					/>
				</dt>
				<dd class="font-medium text-neutral-900 dark:text-neutral-100">{n(status.reshape_n)}</dd>
			</div>
		</dl>
	</div>

	<div class="flex items-center justify-between gap-3 rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900">
		<div>
			<div class="flex items-center gap-1.5">
				<div class="text-sm font-medium text-neutral-900 dark:text-neutral-100">Rebuild index</div>
				<Tooltip
					label="About Rebuild"
					tip="Re-scan relay events into the Conduit store and re-embed documents. Use after changing kinds, the embedder, or if embeddings look orphaned. Ready must be true before intercept ranking is live."
				/>
			</div>
			<p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
				Re-scan relay events into the Conduit store. Ready is {ready ? 'true' : 'false'}.
			</p>
		</div>
		<button
			class="rounded-lg border border-neutral-300 bg-neutral-900 px-3 py-2 text-sm font-medium text-white disabled:opacity-50 dark:border-neutral-700 dark:bg-neutral-100 dark:text-neutral-900"
			type="button"
			disabled={busy}
			onclick={() => void confirmRebuild()}
		>
			{busy ? 'Working…' : 'Rebuild'}
		</button>
	</div>
</section>
