<script>
	import { PATHS, pluginApi } from '../api.js';
	import { isHttpUrl, kindLabel, kindRole, parseJSONObject, tagValues } from './eventView.js';

	let { eventId = '', coord = '', open = $bindable(false) } = $props();

	async function fetchEvent(id, c) {
		if (!id && !c) {
			return {
				ok: false,
				missing: true,
				error: 'This row has no event id or addressable coordinate.'
			};
		}
		const res = await pluginApi('POST', PATHS.getEvent, { id, coord: c });
		const j = res.json || {};
		if (j.missing || res.status === 404) {
			return {
				ok: false,
				missing: true,
				error:
					j.error ||
					'This row is in the Conduit index, but the event is not in the relay store. It may have been deleted, or the index is leftover from an earlier run.'
			};
		}
		if ((res.status && res.status >= 400) || j.ok === false) {
			return { ok: false, missing: false, error: j.error || `Failed (${res.status || 'error'})` };
		}
		if (!j.event) {
			return { ok: false, missing: true, error: 'Relay returned no event.' };
		}
		return { ok: true, event: j.event };
	}

	function roleOf(event) {
		return kindRole(event?.kind);
	}

	function stallOf(event) {
		return roleOf(event) === 'stall' ? parseJSONObject(event?.content) : null;
	}

	function productOf(event) {
		return roleOf(event) === 'product' ? parseJSONObject(event?.content) : null;
	}

	function listingTagsOf(event) {
		return {
			title: tagValues(event?.tags, 'title')[0] || '',
			summary: tagValues(event?.tags, 'summary')[0] || '',
			status: tagValues(event?.tags, 'status')[0] || '',
			price: tagValues(event?.tags, 'price'),
			location: tagValues(event?.tags, 'location')[0] || '',
			geo: tagValues(event?.tags, 'g')[0] || '',
			topics: tagValues(event?.tags, 't'),
			images: tagValues(event?.tags, 'image').filter(isHttpUrl)
		};
	}

	function headingOf(event) {
		const role = roleOf(event);
		const stall = stallOf(event);
		const product = productOf(event);
		const listingTags = listingTagsOf(event);
		if (role === 'stall') return stall?.name || tagValues(event?.tags, 'd')[0] || 'Stall';
		if (role === 'product') return product?.name || 'Product';
		if (role === 'listing' || role === 'draft') return listingTags.title || 'Classified listing';
		return kindLabel(event?.kind);
	}

	function shippingRows(stall) {
		return Array.isArray(stall?.shipping) ? stall.shipping : [];
	}

	function close() {
		open = false;
	}

	function onBackdrop(e) {
		if (e.target === e.currentTarget) close();
	}

	function onKey(e) {
		if (e.key === 'Escape' && open) close();
	}
</script>

<svelte:window onkeydown={onKey} />

{#if open}
	<div
		class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/50 p-4 sm:items-center"
		role="presentation"
		onclick={onBackdrop}
	>
		<div
			class="my-6 w-full max-w-2xl rounded-xl border border-neutral-200 bg-white shadow-xl dark:border-neutral-800 dark:bg-neutral-950"
			role="dialog"
			aria-modal="true"
			aria-labelledby="event-modal-title"
			tabindex="-1"
		>
			{#key `${eventId}\0${coord}`}
				{#await fetchEvent(eventId, coord)}
					<div
						class="flex items-start justify-between gap-3 border-b border-neutral-200 px-5 py-4 dark:border-neutral-800"
					>
						<h2 id="event-modal-title" class="text-lg font-semibold text-neutral-900 dark:text-neutral-100">
							Event
						</h2>
						<button
							class="rounded-lg px-2 py-1 text-sm text-neutral-500 hover:bg-neutral-100 dark:hover:bg-neutral-800"
							type="button"
							onclick={close}
						>
							Close
						</button>
					</div>
					<p class="px-5 py-4 text-sm text-neutral-500 dark:text-neutral-400">
						Loading event from the relay…
					</p>
				{:then result}
					{const event = result.ok ? result.event : null}
					{const role = roleOf(event)}
					{const stall = stallOf(event)}
					{const product = productOf(event)}
					{const listingTags = listingTagsOf(event)}
					{const productImages = Array.isArray(product?.images)
						? product.images.filter(isHttpUrl)
						: []}
					{const shipping = shippingRows(stall)}
					<div
						class="flex items-start justify-between gap-3 border-b border-neutral-200 px-5 py-4 dark:border-neutral-800"
					>
						<div class="min-w-0">
							<p class="text-xs tracking-wide text-neutral-500 uppercase dark:text-neutral-400">
								{kindLabel(event?.kind)}{#if event?.kind} · kind {event.kind}{/if}
							</p>
							<h2
								id="event-modal-title"
								class="mt-1 text-lg font-semibold text-neutral-900 dark:text-neutral-100"
							>
								{event ? headingOf(event) : 'Event'}
							</h2>
						</div>
						<button
							class="rounded-lg px-2 py-1 text-sm text-neutral-500 hover:bg-neutral-100 dark:hover:bg-neutral-800"
							type="button"
							onclick={close}
						>
							Close
						</button>
					</div>
					<div class="max-h-[75vh] space-y-4 overflow-y-auto px-5 py-4">
						{#if !result.ok}
							<p class="text-sm text-red-700 dark:text-red-300">{result.error}</p>
							{#if result.missing}
								<p class="text-sm text-neutral-500 dark:text-neutral-400">
									Congee Audit → Events is the relay audit log, not a marketplace catalog. These rows
									are kind 30017 / 34550 stalls (shops) or 30018 / 30402 products and classifieds.
									Filter Audit by that kind, or click a Conduit row to load the event here.
								</p>
							{/if}
						{:else if role === 'stall'}
							{#if stall?.description}
								<p class="text-sm whitespace-pre-wrap text-neutral-700 dark:text-neutral-300">
									{stall.description}
								</p>
							{/if}
							<dl class="grid gap-3 text-sm sm:grid-cols-2">
								<div>
									<dt class="text-xs text-neutral-500 uppercase">Stall id</dt>
									<dd class="mt-0.5 break-all font-mono text-neutral-800 dark:text-neutral-200">
										{stall?.id || tagValues(event.tags, 'd')[0] || '—'}
									</dd>
								</div>
								{#if stall?.currency}
									<div>
										<dt class="text-xs text-neutral-500 uppercase">Currency</dt>
										<dd class="mt-0.5">{stall.currency}</dd>
									</div>
								{/if}
							</dl>
							{#if shipping.length}
								<div>
									<p class="text-xs text-neutral-500 uppercase">Shipping</p>
									<ul class="mt-1 space-y-1 text-sm text-neutral-700 dark:text-neutral-300">
										{#each shipping as zone, i (zone?.id || i)}
											<li>
												{zone?.name || zone?.id || 'Zone'}{#if zone?.cost != null}
													· {zone.cost}{stall?.currency ? ` ${stall.currency}` : ''}{/if}
												{#if Array.isArray(zone?.countries) && zone.countries.length}
													· {zone.countries.join(', ')}{/if}
											</li>
										{/each}
									</ul>
								</div>
							{/if}
						{:else if role === 'product'}
							{#if productImages.length}
								<div class="flex flex-wrap gap-2">
									{#each productImages as src (src)}
										<img class="h-24 w-24 rounded-md object-cover" {src} alt="" />
									{/each}
								</div>
							{/if}
							{#if product?.description}
								<p class="text-sm whitespace-pre-wrap text-neutral-700 dark:text-neutral-300">
									{product.description}
								</p>
							{/if}
							<dl class="grid gap-3 text-sm sm:grid-cols-2">
								{#if product?.price != null}
									<div>
										<dt class="text-xs text-neutral-500 uppercase">Price</dt>
										<dd class="mt-0.5">
											{product.price}{product.currency ? ` ${product.currency}` : ''}
										</dd>
									</div>
								{/if}
								{#if product?.quantity != null}
									<div>
										<dt class="text-xs text-neutral-500 uppercase">Quantity</dt>
										<dd class="mt-0.5">{product.quantity}</dd>
									</div>
								{/if}
								{#if product?.stall_id}
									<div class="sm:col-span-2">
										<dt class="text-xs text-neutral-500 uppercase">Stall</dt>
										<dd class="mt-0.5 break-all font-mono">{product.stall_id}</dd>
									</div>
								{/if}
							</dl>
						{:else if role === 'listing' || role === 'draft'}
							{#if listingTags.images.length}
								<div class="flex flex-wrap gap-2">
									{#each listingTags.images as src (src)}
										<img class="h-24 w-24 rounded-md object-cover" {src} alt="" />
									{/each}
								</div>
							{/if}
							{#if event.content}
								<p class="text-sm whitespace-pre-wrap text-neutral-700 dark:text-neutral-300">
									{event.content}
								</p>
							{:else if listingTags.summary}
								<p class="text-sm text-neutral-700 dark:text-neutral-300">{listingTags.summary}</p>
							{/if}
							<dl class="grid gap-3 text-sm sm:grid-cols-2">
								{#if listingTags.status}
									<div>
										<dt class="text-xs text-neutral-500 uppercase">Status</dt>
										<dd class="mt-0.5">{listingTags.status}</dd>
									</div>
								{/if}
								{#if listingTags.price.length}
									<div>
										<dt class="text-xs text-neutral-500 uppercase">Price</dt>
										<dd class="mt-0.5">{listingTags.price.join(' ')}</dd>
									</div>
								{/if}
								{#if listingTags.location}
									<div>
										<dt class="text-xs text-neutral-500 uppercase">Location</dt>
										<dd class="mt-0.5">{listingTags.location}</dd>
									</div>
								{/if}
								{#if listingTags.geo}
									<div>
										<dt class="text-xs text-neutral-500 uppercase">Geohash</dt>
										<dd class="mt-0.5 font-mono">{listingTags.geo}</dd>
									</div>
								{/if}
								{#if listingTags.topics.length}
									<div class="sm:col-span-2">
										<dt class="text-xs text-neutral-500 uppercase">Tags</dt>
										<dd class="mt-0.5">{listingTags.topics.join(', ')}</dd>
									</div>
								{/if}
							</dl>
						{:else}
							<p class="text-sm text-neutral-600 dark:text-neutral-300">
								No marketplace layout for kind {event.kind}. Raw event below.
							</p>
						{/if}

						{#if event}
							<dl class="grid gap-3 border-t border-neutral-200 pt-4 text-sm dark:border-neutral-800">
								<div>
									<dt class="text-xs text-neutral-500 uppercase">Event id</dt>
									<dd class="mt-0.5 break-all font-mono text-xs">{event.id}</dd>
								</div>
								<div>
									<dt class="text-xs text-neutral-500 uppercase">Pubkey</dt>
									<dd class="mt-0.5 break-all font-mono text-xs">{event.pubkey}</dd>
								</div>
							</dl>
							<details class="rounded-lg border border-neutral-200 dark:border-neutral-800">
								<summary class="cursor-pointer px-3 py-2 text-sm text-neutral-600 dark:text-neutral-300">
									Raw event JSON
								</summary>
								<pre
									class="max-h-64 overflow-auto border-t border-neutral-200 p-3 font-mono text-xs break-all whitespace-pre-wrap dark:border-neutral-800"
								>{JSON.stringify(event, null, 2)}</pre>
							</details>
						{/if}
					</div>
				{:catch e}
					<div class="px-5 py-4">
						<p class="text-sm text-red-700 dark:text-red-300">
							{e instanceof Error ? e.message : 'load failed'}
						</p>
						<button class="mt-3 text-sm underline" type="button" onclick={close}>Close</button>
					</div>
				{/await}
			{/key}
		</div>
	</div>
{/if}
