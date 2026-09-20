<script>
	import ChipInput from '../lib/ChipInput.svelte';
	import Switch from '../lib/Switch.svelte';

	let { settings, resetKey = 0 } = $props();
</script>

<section class="space-y-6">
	<div>
		<h2 class="text-lg font-medium text-neutral-900 dark:text-neutral-100">Kinds</h2>
		<p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
			Which event kinds Conduit indexes and intercepts. Comma or enter to add a chip.
		</p>
	</div>

	{#key resetKey}
		<ChipInput
			bind:values={settings.product_kinds}
			label="Product kinds"
			description="NIP-15 / NIP-99 listings served in search."
		/>
		<ChipInput
			bind:values={settings.stall_kinds}
			label="Stall kinds"
			description="Merchant stall documents linked from products."
		/>
		<ChipInput
			bind:values={settings.draft_kinds}
			label="Draft kinds"
			description="Draft listings. Indexed only when Index drafts is on."
		/>
		<ChipInput
			bind:values={settings.deletion_kinds}
			label="Deletion kinds"
			description="Deletion events that remove listings from the index."
		/>
	{/key}

	<Switch
		bind:checked={settings.index_drafts}
		label="Index drafts"
		description="Include draft kinds in the store and intercept them on REQ."
	/>
</section>
