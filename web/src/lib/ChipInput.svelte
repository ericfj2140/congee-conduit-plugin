<script>
	let { values = $bindable([]), label, description = '' } = $props();

	let draft = $state('');

	function parseParts(text) {
		return text
			.split(/[\s,]+/)
			.map((p) => p.trim())
			.filter(Boolean)
			.map((p) => Number(p))
			.filter((n) => Number.isInteger(n) && n >= 0);
	}

	function addFromDraft() {
		const next = [...values];
		for (const n of parseParts(draft)) {
			if (!next.includes(n)) next.push(n);
		}
		values = next;
		draft = '';
	}

	function remove(n) {
		values = values.filter((x) => x !== n);
	}

	function onKeydown(e) {
		if (e.key === 'Enter' || e.key === ',') {
			e.preventDefault();
			addFromDraft();
		} else if (e.key === 'Backspace' && draft === '' && values.length) {
			values = values.slice(0, -1);
		}
	}
</script>

<div class="space-y-2">
	<div>
		<div class="text-sm font-medium text-neutral-900 dark:text-neutral-100">{label}</div>
		{#if description}
			<p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">{description}</p>
		{/if}
	</div>
	<div
		class="flex min-h-11 flex-wrap items-center gap-1.5 rounded-lg border border-neutral-200 bg-white px-2 py-1.5 dark:border-neutral-800 dark:bg-neutral-900"
	>
		{#each values as n (n)}
			<button
				class="inline-flex items-center gap-1 rounded-full bg-neutral-100 px-2.5 py-0.5 text-xs font-medium text-neutral-800 dark:bg-neutral-800 dark:text-neutral-200"
				type="button"
				onclick={() => remove(n)}
			>
				{n}
				<span aria-hidden="true" class="text-neutral-400">×</span>
			</button>
		{/each}
		<input
			class="min-w-[8rem] flex-1 bg-transparent px-1 py-1 text-sm text-neutral-900 outline-none dark:text-neutral-100"
			bind:value={draft}
			onkeydown={onKeydown}
			onblur={addFromDraft}
			placeholder="kind, comma or enter"
			inputmode="numeric"
		/>
	</div>
</div>
