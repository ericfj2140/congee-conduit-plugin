<script>
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
		<h2 class="text-lg font-medium text-neutral-900 dark:text-neutral-100">Overview</h2>
		<p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
			Index health, backfill, and REQ intercept counters.
		</p>
	</div>

	<div class="grid gap-3 sm:grid-cols-2">
		<div class="rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900">
			<div class="text-xs tracking-wide text-neutral-500 uppercase dark:text-neutral-400">Backend</div>
			<div class="mt-1 text-sm font-medium text-neutral-900 dark:text-neutral-100">
				{status.backend || '—'}
			</div>
		</div>
		<div class="rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900">
			<div class="text-xs tracking-wide text-neutral-500 uppercase dark:text-neutral-400">Listings</div>
			<div class="mt-1 text-sm font-medium text-neutral-900 dark:text-neutral-100">
				{n(status.active)} active · {n(status.inactive)} inactive
			</div>
		</div>
		<div class="rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900">
			<div class="text-xs tracking-wide text-neutral-500 uppercase dark:text-neutral-400">Embeddings</div>
			<div class="mt-1 text-sm font-medium text-neutral-900 dark:text-neutral-100">
				{n(status.embeddings)}
			</div>
		</div>
		<div class="rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900">
			<div class="text-xs tracking-wide text-neutral-500 uppercase dark:text-neutral-400">Backfill</div>
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
		<div class="text-xs tracking-wide text-neutral-500 uppercase dark:text-neutral-400">Intercepts</div>
		<dl class="mt-3 grid grid-cols-2 gap-3 text-sm sm:grid-cols-4">
			<div>
				<dt class="text-neutral-500 dark:text-neutral-400">Seen</dt>
				<dd class="font-medium text-neutral-900 dark:text-neutral-100">{n(status.intercept_n)}</dd>
			</div>
			<div>
				<dt class="text-neutral-500 dark:text-neutral-400">Passthrough</dt>
				<dd class="font-medium text-neutral-900 dark:text-neutral-100">{n(status.passthrough_n)}</dd>
			</div>
			<div>
				<dt class="text-neutral-500 dark:text-neutral-400">Respond</dt>
				<dd class="font-medium text-neutral-900 dark:text-neutral-100">{n(status.respond_n)}</dd>
			</div>
			<div>
				<dt class="text-neutral-500 dark:text-neutral-400">Reshape</dt>
				<dd class="font-medium text-neutral-900 dark:text-neutral-100">{n(status.reshape_n)}</dd>
			</div>
		</dl>
	</div>

	<div class="flex items-center justify-between gap-3 rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900">
		<div>
			<div class="text-sm font-medium text-neutral-900 dark:text-neutral-100">Rebuild index</div>
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
