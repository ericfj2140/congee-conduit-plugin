<script>
	import Switch from '../lib/Switch.svelte';

	let { settings } = $props();
	let showAdvanced = $state(false);
</script>

<section class="space-y-6">
	<div>
		<h2 class="text-lg font-medium text-neutral-900 dark:text-neutral-100">Search</h2>
		<p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
			When Conduit intercepts a REQ and how large a result set it will return.
		</p>
	</div>

	<div class="space-y-3">
		<Switch
			bind:checked={settings.rank_all_product_reqs}
			label="Rank all product REQs"
			description="Apply Conduit ranking to product-kind REQs even when they have no search string."
		/>
		<Switch
			bind:checked={settings.inject_product_kinds_on_search}
			label="Inject product kinds on search"
			description="When a REQ has a search string, add configured product kinds so marketplace listings are included."
		/>
	</div>

	<div class="grid gap-3 sm:grid-cols-2">
		<label class="block space-y-1">
			<span class="text-sm font-medium text-neutral-800 dark:text-neutral-200">Max results</span>
			<input
				class="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none focus:ring-2 focus:ring-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-100"
				type="number"
				min="1"
				bind:value={settings.max_results}
			/>
		</label>
		<label class="block space-y-1">
			<span class="text-sm font-medium text-neutral-800 dark:text-neutral-200">Geo min prefix</span>
			<input
				class="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none focus:ring-2 focus:ring-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-100"
				type="number"
				min="1"
				max="12"
				bind:value={settings.geo_min_prefix_len}
			/>
			<span class="text-xs text-neutral-500 dark:text-neutral-400">Shortest #g prefix that still filters.</span>
		</label>
	</div>

	<div>
		<button
			class="text-sm font-medium text-neutral-600 underline-offset-2 hover:underline dark:text-neutral-300"
			type="button"
			onclick={() => (showAdvanced = !showAdvanced)}
		>
			{showAdvanced ? 'Hide advanced' : 'Advanced'}
		</button>
		{#if showAdvanced}
			<label class="mt-3 block space-y-1">
				<span class="text-sm font-medium text-neutral-800 dark:text-neutral-200">Search candidate cap</span>
				<input
					class="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none focus:ring-2 focus:ring-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-100"
					type="number"
					min="1"
					bind:value={settings.search_candidate_cap}
				/>
				<span class="text-xs text-neutral-500 dark:text-neutral-400">
					Upper bound on listings considered before ranking. Lower is cheaper; higher is more complete.
				</span>
			</label>
		{/if}
	</div>
</section>
