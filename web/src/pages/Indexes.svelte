<script>
	import Switch from '../lib/Switch.svelte';

	let { settings } = $props();
</script>

<section class="space-y-6">
	<div>
		<h2 class="text-lg font-medium text-neutral-900 dark:text-neutral-100">Indexes</h2>
		<p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
			How Conduit filters and ranks listings when it intercepts a REQ.
		</p>
	</div>

	<div class="space-y-3">
		<Switch
			bind:checked={settings.active_filter}
			label="Active listings only"
			description="Only serve active listings on intercept respond. Inactive, draft, and deleted listings stay out of search hits."
		/>
		<Switch
			bind:checked={settings.geo_enabled}
			label="Geo index"
			description="Honor #g geohash prefix filters. Prefix length is configured on Search."
		/>
		<Switch
			bind:checked={settings.vector_enabled}
			label="Vector rank"
			description="Rank by embeddings when the REQ has a search string. Requires embeddings to be warm."
		/>
	</div>
</section>
