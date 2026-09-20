<script>
	let { tip, label = 'More information' } = $props();

	let open = $state(false);
	const uid = $props.id();
	const tipId = `${uid}-tip`;

	function close() {
		open = false;
	}

	function toggle(e) {
		e.preventDefault();
		e.stopPropagation();
		open = !open;
	}

	function onWindowKey(e) {
		if (e.key === 'Escape') close();
	}
</script>

<svelte:window onkeydown={onWindowKey} />

<span
	class="relative inline-flex align-middle"
	role="group"
	onmouseenter={() => (open = true)}
	onmouseleave={close}
>
	<button
		class="inline-flex size-4 shrink-0 items-center justify-center rounded-full border border-neutral-300 text-[10px] font-semibold text-neutral-600 hover:bg-neutral-100 focus-visible:ring-2 focus-visible:ring-neutral-400 dark:border-neutral-600 dark:text-neutral-300 dark:hover:bg-neutral-800"
		type="button"
		aria-label={label}
		aria-expanded={open}
		aria-controls={tipId}
		onclick={toggle}
		onfocus={() => (open = true)}
		onblur={close}
	>
		?
	</button>
	{#if open}
		<span
			id={tipId}
			role="tooltip"
			class="absolute top-full left-0 z-50 mt-1 w-80 max-w-[min(20rem,calc(100vw-2rem))] rounded-md border border-neutral-200 bg-white p-2.5 text-left text-xs leading-relaxed font-normal tracking-normal text-neutral-700 whitespace-pre-wrap normal-case shadow-lg dark:border-neutral-700 dark:bg-neutral-900 dark:text-neutral-200"
		>
			{tip}
		</span>
	{/if}
</span>
