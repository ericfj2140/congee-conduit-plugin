<script>
	import ChipInput from '../lib/ChipInput.svelte';
	import Switch from '../lib/Switch.svelte';
	import Tooltip from '../lib/Tooltip.svelte';
	import { defaultKindTip, kindsForRole } from '../lib/eventView.js';

	let { settings = $bindable(), resetKey = 0 } = $props();

	const productTip = defaultKindTip(
		'product',
		`Also ${kindsForRole('listing').join(', ') || '30402'} (classified listings). A REQ that already includes these kinds can be ranked from the index. Kind 34550 is a NIP-72 community definition, not a product.`
	);
	const stallTip = defaultKindTip(
		'stall',
		'These are NIP-15 shop records (name, currency, shipping), not classified listings and not NIP-72 communities. Kind 34550 is a community definition.'
	);
	const draftTip = defaultKindTip('listing_draft', 'Stored as inactive unless Index drafts is enabled.');
	const deletionTip = defaultKindTip('deletion', 'Matching e/a tags mark indexed listings inactive so they drop out of search.');
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
			tip={productTip}
		/>
		<ChipInput
			bind:values={settings.stall_kinds}
			label="Stall kinds"
			description="Merchant stall documents (NIP-15 kind 30017). Indexed and shown in Listings / Embeddings; not a classified listing or a NIP-72 community."
			tip={stallTip}
		/>
		<ChipInput
			bind:values={settings.draft_kinds}
			label="Draft kinds"
			description="Draft listings. Indexed only when Index drafts is on."
			tip={draftTip}
		/>
		<ChipInput
			bind:values={settings.deletion_kinds}
			label="Deletion kinds"
			description="Deletion events that remove listings from the index."
			tip={deletionTip}
		/>
	{/key}

	<Switch
		bind:checked={settings.index_drafts}
		label="Index drafts"
		description="Include draft kinds in the store and intercept them on REQ."
		tip="Off: draft events are stored inactive and never returned on intercept. On: drafts are treated like other indexed kinds."
	/>
</section>
