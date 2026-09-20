<script>
	import Tooltip from '../lib/Tooltip.svelte';
	import { goto } from '../api.js';

	let { status = {}, ready = false, busy = false, hint = '', onrebuild } = $props();

	let pendingRebuild = $state(false);

	function n(v) {
		return typeof v === 'number' ? v : 0;
	}

	function requestRebuild() {
		pendingRebuild = true;
	}

	async function confirmRebuild() {
		pendingRebuild = false;
		await onrebuild();
	}
</script>

<section class="space-y-6">
	<div>
		<div class="flex items-center gap-1.5">
			<h2 class="text-lg font-medium text-neutral-900 dark:text-neutral-100">Overview</h2>
			<Tooltip
				label="About Overview"
				tip="Counts come from the Conduit index database (conduit-index.db or Postgres), not from Congee Audit → Events. Audit is the relay activity log; with a kind filter it also lists stored relay events. Listings include NIP-15 stalls (30017), products (30018), and classifieds. Kind 34550 is a NIP-72 community, not a stall. Embeddings are vector rows — Model fake-bow-384 is the test embedder used only when CONDUIT_EMBEDDER=fake. http:<name> is a verified external provider (384-d). On-device MiniLM is all-MiniLM-L6-v2."
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
		<button
			class="rounded-lg border border-neutral-200 bg-white p-4 text-left transition-colors hover:border-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:hover:border-neutral-500"
			type="button"
			onclick={() => goto('#/listings')}
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
		</button>
		<button
			class="rounded-lg border border-neutral-200 bg-white p-4 text-left transition-colors hover:border-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:hover:border-neutral-500"
			type="button"
			onclick={() => goto('#/embeddings')}
		>
			<div class="flex items-center gap-1.5 text-xs tracking-wide text-neutral-500 uppercase dark:text-neutral-400">
				Embeddings
				<Tooltip
					label="About Embeddings count"
					tip="Number of stored search vectors. Leftover rows from an older index (for example kind 34550 community definitions that were mis-labeled as stalls) are removed on the next settings apply or rebuild."
				/>
			</div>
			<div class="mt-1 text-sm font-medium text-neutral-900 dark:text-neutral-100">
				{n(status.embeddings)}
			</div>
			<div class="mt-2 text-xs text-neutral-500 dark:text-neutral-400">View embedding rows</div>
		</button>
		<div class="rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900">
			<div class="flex items-center gap-1.5 text-xs tracking-wide text-neutral-500 uppercase dark:text-neutral-400">
				Backfill
				<Tooltip
					label="About Backfill"
					tip="Startup scan of existing relay events into the Conduit index. idle means the scan finished or has not started. Listings and embeddings can appear after backfill if the relay already had stall/product events. Kind 34550 communities are not marketplace documents."
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

	<div class="space-y-3 rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900">
		<div class="flex items-center justify-between gap-3">
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
			{#if busy}
				<button
					class="inline-flex items-center gap-2 rounded-lg border border-neutral-300 bg-neutral-900 px-3 py-2 text-sm font-medium text-white disabled:opacity-50 dark:border-neutral-700 dark:bg-neutral-100 dark:text-neutral-900"
					type="button"
					disabled
					aria-busy="true"
				>
					<span
						class="inline-block size-4 animate-spin rounded-full border-2 border-current border-t-transparent"
						aria-hidden="true"
					></span>
					Rebuilding…
				</button>
			{:else if pendingRebuild}
				<div class="flex shrink-0 gap-2">
					<button
						class="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm font-medium text-neutral-800 dark:border-neutral-700 dark:bg-neutral-900 dark:text-neutral-100"
						type="button"
						onclick={() => (pendingRebuild = false)}
					>
						Cancel
					</button>
					<button
						class="rounded-lg border border-neutral-300 bg-neutral-900 px-3 py-2 text-sm font-medium text-white dark:border-neutral-700 dark:bg-neutral-100 dark:text-neutral-900"
						type="button"
						onclick={() => void confirmRebuild()}
					>
						Confirm rebuild
					</button>
				</div>
			{:else}
				<button
					class="inline-flex items-center gap-2 rounded-lg border border-neutral-300 bg-neutral-900 px-3 py-2 text-sm font-medium text-white dark:border-neutral-700 dark:bg-neutral-100 dark:text-neutral-900"
					type="button"
					onclick={requestRebuild}
				>
					Rebuild
				</button>
			{/if}
		</div>
		{#if hint}
			<p class="text-sm text-neutral-600 dark:text-neutral-300">{hint}</p>
		{:else if status.backfill === 'running'}
			<p class="text-sm text-neutral-600 dark:text-neutral-300">
				Backfill running… {n(status.backfill_scanned)} scanned, {n(status.backfill_indexed)} indexed ·
				{n(status.active)} listings · {n(status.embeddings)} embeddings
			</p>
		{/if}
	</div>
</section>
