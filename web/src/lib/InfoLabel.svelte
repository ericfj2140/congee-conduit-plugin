<script>
	import Tooltip from './Tooltip.svelte';

	let { text, tip = '', class: className = '' } = $props();
</script>

<span class={['inline-flex items-center gap-1', className]}>
	{text}
	{#if tip}
		<Tooltip {tip} label="About {text}" />
	{/if}
</span>
