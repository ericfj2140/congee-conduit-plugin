<script>
	import { onMount } from 'svelte';
	import { PATHS, onHostMessage, pluginApi } from './api.js';
	import { cloneSettings, mergeSettings, settingsEqual } from './settings.js';
	import Indexes from './pages/Indexes.svelte';
	import Kinds from './pages/Kinds.svelte';
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
		'/kinds': 'kinds'
	};

	function routeFromHash(h) {
		const raw = (h || '#/').replace(/^#/, '') || '/';
		const path = raw.startsWith('/') ? raw : `/${raw}`;
		return routes[path] || 'overview';
	}

	let hash = $state(typeof location !== 'undefined' ? location.hash : '#/');
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
	let testBusy = $state(false);
	let testResult = $state('');

	function onHashChange() {
		hash = location.hash || '#/';
	}

	function apiError(json, fallback) {
		if (json && typeof json.error === 'string' && json.error) return json.error;
		return fallback;
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
				const j = statusRes.json || {};
				ready = j.ready === true;
				pluginStatus = j.status && typeof j.status === 'object' ? j.status : {};
				relayType = typeof j.relay_database_type === 'string' ? j.relay_database_type : '';
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
		try {
			const res = await pluginApi('POST', PATHS.rebuild);
			if (res.status && res.status >= 400) {
				saveError = apiError(res.json, `Rebuild failed (${res.status})`);
				return;
			}
			await loadAll();
		} catch (e) {
			saveError = e instanceof Error ? e.message : 'rebuild failed';
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

	onMount(() => {
		void loadAll();
	});
</script>

<svelte:window onhashchange={onHashChange} onmessage={onHostMessage} />

<div class="min-h-screen bg-neutral-50 text-neutral-900 dark:bg-neutral-950 dark:text-neutral-100">
	<header class="border-b border-neutral-200 bg-white/90 backdrop-blur dark:border-neutral-800 dark:bg-neutral-950/90">
		<div class="mx-auto flex max-w-3xl flex-col gap-3 px-4 py-4">
			<div>
				<h1 class="text-xl font-semibold tracking-tight">Conduit</h1>
				<p class="text-sm text-neutral-500 dark:text-neutral-400">Marketplace index</p>
			</div>
			<nav class="flex flex-wrap gap-1" aria-label="Sections">
				{#each nav as item (item.id)}
					<a
						href={item.href}
						aria-current={route === item.id ? 'page' : undefined}
						class={[
							'rounded-full px-3 py-1 text-sm',
							route === item.id
								? 'bg-neutral-900 text-white dark:bg-neutral-100 dark:text-neutral-900'
								: 'text-neutral-600 hover:bg-neutral-100 dark:text-neutral-300 dark:hover:bg-neutral-900'
						]}
					>
						{item.label}
					</a>
				{/each}
			</nav>
		</div>
	</header>

	<main class="mx-auto max-w-3xl space-y-4 px-4 py-6 {dirty ? 'pb-28' : 'pb-10'}">
		{#if ready === false}
			<div
				class="rounded-lg border border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-950 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-100"
			>
				REQs pass through until Ready
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
			<Overview status={pluginStatus} {ready} busy={actionBusy} onrebuild={rebuild} />
		{:else if route === 'storage'}
			<Storage {settings} {relayType} {testResult} {testBusy} ontest={testStore} />
		{:else if route === 'indexes'}
			<Indexes {settings} />
		{:else if route === 'search'}
			<Search {settings} />
		{:else if route === 'kinds'}
			<Kinds {settings} {resetKey} />
		{/if}
	</main>

	{#if dirty}
		<div
			class="fixed inset-x-0 bottom-0 border-t border-neutral-200 bg-white/95 px-4 py-3 backdrop-blur dark:border-neutral-800 dark:bg-neutral-950/95"
		>
			<div class="mx-auto flex max-w-3xl items-center justify-between gap-3">
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
