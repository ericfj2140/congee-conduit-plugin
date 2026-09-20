<script>
	import ChipInput from '../lib/ChipInput.svelte';
	import Switch from '../lib/Switch.svelte';
	import Tooltip from '../lib/Tooltip.svelte';

	let { settings, resetKey = 0 } = $props();
</script>

<section class="space-y-6">
	<div>
		<div class="flex items-center gap-1.5">
			<h2 class="text-lg font-medium text-neutral-900 dark:text-neutral-100">Kinds</h2>
			<Tooltip
				label="About Kinds"
				tip="These numbers are the event kinds Conduit indexes on store and intercepts on REQ. Product kinds are ranked on search and #g. Stall kinds are stored as shop documents. Drafts are indexed only if Index drafts is on. Deletion kinds deactivate matching listings. Intercept does not inject kinds into the client's filter."
			/>
		</div>
		<p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
			Event kinds Conduit indexes on store and intercepts on REQ. Intercept ranks matching listings
			from the index; it does not rewrite the client's filter kinds.
		</p>
	</div>

	{#key resetKey}
		<ChipInput
			bind:values={settings.product_kinds}
			label="Product kinds"
			description="NIP-15 products and NIP-99 classified listings ranked on intercept."
			tip="Defaults: 30018 and 34560 (NIP-15 products), 30402 (NIP-99 classified). A REQ that already includes these kinds can be ranked from the index. Kind 34550 is a stall, not a product — that belongs under Stall kinds."
		/>
		<ChipInput
			bind:values={settings.stall_kinds}
			label="Stall kinds"
			description="Merchant stall documents. Indexed and shown in Listings / Embeddings; not the same as a classified listing."
			tip="Defaults: 30017 and 34550 (NIP-15 stall). These are shop records (name, currency, shipping), not product listings. They often have no title tag, so the table Title column may fall back to the d-tag."
		/>
		<ChipInput
			bind:values={settings.draft_kinds}
			label="Draft kinds"
			description="Draft listings. Indexed only when Index drafts is on."
			tip="Default 30403 (NIP-99 draft). Stored as inactive unless Index drafts is enabled."
		/>
		<ChipInput
			bind:values={settings.deletion_kinds}
			label="Deletion kinds"
			description="Deletion events that remove listings from the index."
			tip="Default kind 5 (NIP-09). Matching e/a tags mark indexed listings inactive so they drop out of search."
		/>
	{/key}

	<Switch
		bind:checked={settings.index_drafts}
		label="Index drafts"
		description="Include draft kinds in the store and intercept them on REQ."
		tip="Off: draft events are stored inactive and never returned on intercept. On: drafts are treated like other indexed kinds."
	/>
</section>
