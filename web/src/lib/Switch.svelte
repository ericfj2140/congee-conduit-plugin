<script>
	let { checked = $bindable(false), label, description = '' } = $props();
</script>

<label class="flex cursor-pointer items-start gap-3 rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900">
	<input bind:checked class="peer sr-only" type="checkbox" />
	<span
		class="relative mt-0.5 inline-flex h-5 w-9 shrink-0 rounded-full bg-neutral-300 transition peer-focus-visible:ring-2 peer-focus-visible:ring-neutral-400 peer-checked:bg-neutral-800 after:absolute after:top-0.5 after:left-0.5 after:h-4 after:w-4 after:rounded-full after:bg-white after:transition-transform peer-checked:after:translate-x-4 dark:bg-neutral-700 dark:peer-checked:bg-neutral-200 dark:after:bg-neutral-950 dark:peer-checked:after:bg-neutral-900"
	></span>
	<span class="min-w-0">
		<span class="block text-sm font-medium text-neutral-900 dark:text-neutral-100">{label}</span>
		{#if description}
			<span class="mt-1 block text-sm text-neutral-500 dark:text-neutral-400">{description}</span>
		{/if}
	</span>
</label>
