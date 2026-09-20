<script>
	import { onMount } from 'svelte';
	import { GOTO_EVENT, PATHS, onHostMessage, pluginApi } from './api.js';
	import { cloneSettings, mergeSettings, settingsEqual } from './settings.js';
	import Indexes from './pages/Indexes.svelte';
	import Kinds from './pages/Kinds.svelte';
	import Listings from './pages/Listings.svelte';
	import Embeddings from './pages/Embeddings.svelte';
	import Overview from './pages/Overview.svelte';
	import Search from './pages/Search.svelte';
	import Storage from './pages/Storage.svelte';

	const nav = [
		{ href: '#/', id: 'overview', label: 'Overview' },
		{ href: '#/storage', id: 'storage', label: 'Storage' },
		{ href: '#/indexes', id: 'indexes', label: 'Indexes' },
		{ href: '#/search', id: 'search', label: 'Search' },
		{ href: '#/kinds', id: 'kinds', label: 'Kinds' }
	];

	const routes = {
		'/': 'overview',
		'/storage': 'storage',
		'/indexes': 'indexes',
		'/search': 'search',
		'/kinds': 'kinds',
		'/listings': 'listings',
		'/embeddings': 'embeddings'
	};

	function routeFromHash(h) {
		const raw = (h || '#/').replace(/^#/, '') || '/';
		const path = raw.startsWith('/') ? raw : `/${raw}`;
		return routes[path] || 'overview';
	}

	let hash = $state(typeof location !== 'undefined' ? location.hash || '#/' : '#/');
	let route = $derived(routeFromHash(hash));

	let settings = $state(mergeSettings());
	let baseline = $state(mergeSettings());
	let resetKey = $state(0);
	let dirty = $derived(!settingsEqual(settings, baseline));

	let ready = $state(null);
	let pluginStatus = $state({});
	let relayType = $state('');
	let loadError = $state('');
	let saveError = $state('');
	let loading = $state(true);
	let saving = $state(false);
	let actionBusy = $state(false);
	let rebuildHint = $state('');
	let testBusy = $state(false);
	let testResult = $state('');
	let embedTestBusy = $state(false);
	let embedTestResult = $state('');

	function go(href) {
		hash = href.startsWith('#') ? href : `#/${href}`;
	}

	function onGotoEvent(e) {
		if (e instanceof CustomEvent && typeof e.detail === 'string') go(e.detail);
	}

	function apiError(json, fallback) {
		if (json && typeof json.error === 'string' && json.error) return json.error;
		return fallback;
	}

	function applyStatusPayload(j) {
		ready = j.ready === true;
		pluginStatus = j.status && typeof j.status === 'object' ? j.status : {};
		relayType = typeof j.relay_database_type === 'string' ? j.relay_database_type : '';
	}

	async function refreshStatus() {
		const statusRes = await pluginApi('GET', PATHS.status);
		if (statusRes.status && statusRes.status >= 400) {
			throw new Error(apiError(statusRes.json, `Status failed (${statusRes.status})`));
		}
		applyStatusPayload(statusRes.json || {});
	}

	function sleep(ms) {
		return new Promise((resolve) => setTimeout(resolve, ms));
	}

	function n(v) {
		return typeof v === 'number' && Number.isFinite(v) ? v : 0;
	}

	async function loadAll() {
		loading = true;
		loadError = '';
		try {
			const [settingsRes, statusRes] = await Promise.all([
				pluginApi('GET', PATHS.settings),
				pluginApi('GET', PATHS.status)
			]);
			if (settingsRes.status && settingsRes.status >= 400) {
				loadError = apiError(settingsRes.json, `Settings failed (${settingsRes.status})`);
			} else {
				settings = mergeSettings(settingsRes.json?.settings);
				baseline = cloneSettings(settings);
				resetKey += 1;
			}
			if (statusRes.status && statusRes.status >= 400) {
				loadError = loadError || apiError(statusRes.json, `Status failed (${statusRes.status})`);
			} else {
				applyStatusPayload(statusRes.json || {});
			}
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'load failed';
		} finally {
			loading = false;
		}
	}

	async function save() {
		saving = true;
		saveError = '';
		try {
			const body = cloneSettings(settings);
			const res = await pluginApi('PUT', PATHS.settings, body);
			if (res.status && res.status >= 400) {
				saveError = apiError(res.json, `Save failed (${res.status})`);
				return;
			}
			baseline = cloneSettings(settings);
			await loadAll();
		} catch (e) {
			saveError = e instanceof Error ? e.message : 'save failed';
		} finally {
			saving = false;
		}
	}

	function discard() {
		settings = cloneSettings(baseline);
		resetKey += 1;
		saveError = '';
	}

	async function rebuild() {
		actionBusy = true;
		saveError = '';
		rebuildHint = 'Starting rebuild…';
		try {
			const res = await pluginApi('POST', PATHS.rebuild);
			if (res.status && res.status >= 400) {
				saveError = apiError(res.json, `Rebuild failed (${res.status})`);
				rebuildHint = '';
				return;
			}
			const wantGen = n(res.json?.generation);
			const deadline = Date.now() + 10 * 60 * 1000;
			while (Date.now() < deadline) {
				await refreshStatus();
				const st = pluginStatus;
				const bf = typeof st.backfill === 'string' ? st.backfill : '';
				const gen = n(st.backfill_generation);
				const scanned = n(st.backfill_scanned);
				const indexed = n(st.backfill_indexed);
				if (wantGen > 0 && gen < wantGen) {
					rebuildHint = 'Waiting for rebuild to start…';
					await sleep(400);
					continue;
				}
				if (typeof bf === 'string' && bf.startsWith('error')) {
					saveError = bf;
					rebuildHint = '';
					return;
				}
				if (bf === 'running' || bf === '') {
					rebuildHint = `Scanning relay events… ${scanned} scanned, ${indexed} indexed · ${n(st.active)} listings · ${n(st.embeddings)} embeddings`;
					await sleep(1000);
					continue;
				}
				rebuildHint = `Rebuild ${bf || 'complete'}. ${n(st.active)} listings · ${n(st.embeddings)} embeddings.`;
				return;
			}
			rebuildHint = 'Rebuild is still running in the background. Counts will keep updating on Overview.';
		} catch (e) {
			saveError = e instanceof Error ? e.message : 'rebuild failed';
			rebuildHint = '';
		} finally {
			actionBusy = false;
		}
	}

	async function testStore() {
		testBusy = true;
		testResult = '';
		try {
			const res = await pluginApi('POST', PATHS.testStore);
			const j = res.json || {};
			if (res.status && res.status >= 400) {
				testResult = apiError(j, `Failed (${res.status})`);
			} else if (j.ok === false) {
				testResult = j.error || 'Connection failed';
			} else {
				testResult = 'Connection ok';
			}
		} catch (e) {
			testResult = e instanceof Error ? e.message : 'test failed';
		} finally {
			testBusy = false;
		}
	}

	async function testEmbed() {
		embedTestBusy = true;
		embedTestResult = '';
		try {
			const res = await pluginApi('POST', PATHS.testEmbed, {
				url: settings.embed_http_url,
				model: settings.embed_http_model,
				api_key: settings.embed_http_api_key,
				dim: settings.embed_dim
			});
			const j = res.json || {};
			if (res.status && res.status >= 400) {
				embedTestResult = apiError(j, `Failed (${res.status})`);
			} else if (j.ok === false) {
				embedTestResult = j.error || 'Embedding probe failed';
			} else {
				settings.embed_provider = 'http';
				embedTestResult = `Ok — ${j.dim || 384}-d (${j.model_id || 'http'}). Save to offload the on-device model.`;
			}
		} catch (e) {
			embedTestResult = e instanceof Error ? e.message : 'test failed';
		} finally {
			embedTestBusy = false;
		}
	}

	async function ensureAssets() {
		embedTestBusy = true;
		embedTestResult = '';
		try {
			const res = await pluginApi('POST', PATHS.ensureAssets, {
				model_url: settings.embed_model_url,
				runtime_url: settings.embed_runtime_url,
				force: true
			});
			const j = res.json || {};
			if (res.status && res.status >= 400) {
				embedTestResult = apiError(j, `Failed (${res.status})`);
			} else if (j.ok === false) {
				embedTestResult = j.error || 'Download failed';
			} else {
				embedTestResult = 'Assets downloaded. Save settings if URLs changed.';
				await loadAll();
			}
		} catch (e) {
			embedTestResult = e instanceof Error ? e.message : 'download failed';
		} finally {
			embedTestBusy = false;
		}
	}

	onMount(() => {
		window.addEventListener(GOTO_EVENT, onGotoEvent);
		void loadAll();
		return () => window.removeEventListener(GOTO_EVENT, onGotoEvent);
	});
</script>

<svelte:window onmessage={onHostMessage} />

<div class="flex min-h-screen flex-col bg-neutral-50 text-neutral-900 dark:bg-neutral-950 dark:text-neutral-100">
	<header class="border-b border-neutral-200 bg-white/90 backdrop-blur dark:border-neutral-800 dark:bg-neutral-950/90">
		<div class="flex w-full flex-col gap-3 px-6 py-4">
			<div>
				<h1 class="text-xl font-semibold tracking-tight">Conduit</h1>
				<p class="text-sm text-neutral-500 dark:text-neutral-400">Marketplace index</p>
			</div>
			<nav class="flex flex-wrap gap-1" aria-label="Sections">
				{#each nav as item (item.id)}
					<button
						type="button"
						aria-current={route === item.id ? 'page' : undefined}
						class={[
							'rounded-full px-3 py-1 text-sm',
							route === item.id
								? 'bg-neutral-900 text-white dark:bg-neutral-100 dark:text-neutral-900'
								: 'text-neutral-600 hover:bg-neutral-100 dark:text-neutral-300 dark:hover:bg-neutral-900'
						]}
						onclick={() => go(item.href)}
					>
						{item.label}
					</button>
				{/each}
			</nav>
		</div>
	</header>

	<main class="w-full flex-1 space-y-4 px-6 py-6 {dirty ? 'pb-28' : 'pb-10'}" data-route={route}>
		{#if ready === false}
			<div
				class="rounded-lg border border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-950 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-100"
			>
				REQs pass through until Ready
			</div>
		{/if}

		{#if pluginStatus.embedder?.error || pluginStatus.embedder?.warning}
			<div
				class="rounded-lg border border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-950 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-100"
			>
				<p class="font-medium">
					Embedder: {pluginStatus.embedder.model_id || 'none'}
					{#if pluginStatus.embedder.source}
						<span class="font-normal text-amber-800 dark:text-amber-200">
							({pluginStatus.embedder.source})</span
						>
					{/if}
				</p>
				{#if pluginStatus.embedder.error}
					<p class="mt-1">{pluginStatus.embedder.error}</p>
				{/if}
				{#if pluginStatus.embedder.warning}
					<p class="mt-1">{pluginStatus.embedder.warning}</p>
				{/if}
				<p class="mt-1 text-xs">
					Vector ranking is {pluginStatus.embedder.vector_ranking ? 'on' : 'off'}.
				</p>
			</div>
		{/if}

		{#if loadError}
			<div
				class="rounded-lg border border-red-300 bg-red-50 px-4 py-3 text-sm text-red-900 dark:border-red-900 dark:bg-red-950/40 dark:text-red-200"
			>
				{loadError}
			</div>
		{/if}

		{#if saveError}
			<div
				class="rounded-lg border border-red-300 bg-red-50 px-4 py-3 text-sm text-red-900 dark:border-red-900 dark:bg-red-950/40 dark:text-red-200"
			>
				{saveError}
			</div>
		{/if}

		{#if loading}
			<p class="text-sm text-neutral-500 dark:text-neutral-400">Loading…</p>
		{:else if route === 'overview'}
			<Overview status={pluginStatus} {ready} busy={actionBusy} hint={rebuildHint} onrebuild={rebuild} />
		{:else if route === 'storage'}
			<Storage bind:settings {relayType} {testResult} {testBusy} ontest={testStore} />
		{:else if route === 'indexes'}
			<Indexes
				bind:settings
				embedder={pluginStatus.embedder}
				assets={pluginStatus.assets}
				testResult={embedTestResult}
				testBusy={embedTestBusy}
				ontest={testEmbed}
				ondownload={ensureAssets}
			/>
		{:else if route === 'search'}
			<Search bind:settings />
		{:else if route === 'kinds'}
			<Kinds bind:settings {resetKey} />
		{:else if route === 'listings'}
			<Listings />
		{:else if route === 'embeddings'}
			<Embeddings />
		{/if}
	</main>

	{#if dirty}
		<div
			class="fixed inset-x-0 bottom-0 border-t border-neutral-200 bg-white/95 px-6 py-3 backdrop-blur dark:border-neutral-800 dark:bg-neutral-950/95"
		>
			<div class="flex w-full items-center justify-between gap-3">
				<p class="text-sm text-neutral-600 dark:text-neutral-300">Unsaved settings</p>
				<div class="flex gap-2">
					<button
						class="rounded-lg border border-neutral-300 px-3 py-2 text-sm font-medium text-neutral-800 dark:border-neutral-700 dark:text-neutral-100"
						type="button"
						onclick={discard}
					>
						Discard
					</button>
					<button
						class="rounded-lg bg-neutral-900 px-3 py-2 text-sm font-medium text-white disabled:opacity-50 dark:bg-neutral-100 dark:text-neutral-900"
						type="button"
						disabled={saving}
						onclick={() => void save()}
					>
						{saving ? 'Saving…' : 'Save'}
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>
