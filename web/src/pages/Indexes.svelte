<script>
	import Switch from '../lib/Switch.svelte';
	import Tooltip from '../lib/Tooltip.svelte';

	let { settings } = $props();
</script>

<section class="space-y-6">
	<div>
		<div class="flex items-center gap-1.5">
			<h2 class="text-lg font-medium text-neutral-900 dark:text-neutral-100">Indexes</h2>
			<Tooltip
				label="About Indexes"
				tip="These switches control how the index is queried when Conduit intercepts a REQ. They do not change the client's filter kinds. Geo uses #g; vector rank uses embeddings when there is a search string."
			/>
		</div>
		<p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
			How Conduit filters and ranks listings when it intercepts a REQ.
		</p>
	</div>

	<div class="space-y-3">
		<Switch
			bind:checked={settings.active_filter}
			label="Active listings only"
			description="Only serve active listings on intercept respond. Inactive, draft, and deleted listings stay out of search hits."
			tip="When on, SQL only considers status=active rows. Sold NIP-99 listings, zero-quantity products, drafts (unless indexed as active), and deletions are omitted from ranked IDs."
		/>
		<Switch
			bind:checked={settings.geo_enabled}
			label="Geo index"
			description="Honor #g geohash prefix filters. Prefix length is configured on Search."
			tip="When a REQ includes #g, Conduit matches stored geohashes with a prefix LIKE, then sorts by haversine distance from the full geohash. Disable to ignore #g on intercept."
		/>
		<Switch
			bind:checked={settings.vector_enabled}
			label="Vector rank"
			description="Rank by embeddings when the REQ has a search string. Requires embeddings to be warm."
			tip="Search text is embedded with the same model as listings (fake-bow-384 unless an ONNX model is loaded). Cosine similarity orders candidates. If embeddings are missing, Conduit falls back to newest-first among the SQL candidates."
		/>
	</div>
</section>
