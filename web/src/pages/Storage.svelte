<script>
	import InfoLabel from '../lib/InfoLabel.svelte';
	import Tooltip from '../lib/Tooltip.svelte';

	let { settings, relayType = '', testResult = '', testBusy = false, ontest } = $props();

	let mode = $derived(settings.postgres_url ? 'url' : 'auto');
	let splitBrain = $derived(relayType === 'postgres' && settings.index_backend === 'turso');

	function setBackend(value) {
		settings.index_backend = value;
	}

	function setMode(next) {
		if (next === 'auto') {
			settings.postgres_url = '';
		}
	}
</script>

<section class="space-y-6">
	<div>
		<div class="flex items-center gap-1.5">
			<h2 class="text-lg font-medium text-neutral-900 dark:text-neutral-100">Storage</h2>
			<Tooltip
				label="About Storage"
				tip="The Conduit index is a separate database from the relay event store. Turso/libSQL is a local file (conduit-index.db) in the plugin data directory. Postgres is for sharing the index across instances. Events still live in Congee; this store only keeps parsed listings, geo, and embeddings."
			/>
		</div>
		<p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
			Where Conduit keeps its marketplace index. This is separate from the relay event store.
		</p>
	</div>

	{#if splitBrain}
		<div
			class="rounded-lg border border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-900 dark:border-amber-800 dark:bg-amber-950/50 dark:text-amber-200"
		>
			<strong class="font-medium">WARN</strong>
			— Relay database is postgres while Conduit is on Turso. The index lives in a local file and
			will not follow a multi-instance relay. Switch Conduit to postgres or accept split-brain
			search.
		</div>
	{:else if relayType === 'postgres'}
		<div
			class="rounded-lg border border-neutral-200 bg-neutral-50 px-4 py-3 text-sm text-neutral-600 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-300"
		>
			Relay is postgres. Keep Conduit on postgres if you want the index next to the event store.
		</div>
	{/if}

	<div class="grid gap-3 sm:grid-cols-2">
		<button
			class={[
				'rounded-lg border p-4 text-left',
				settings.index_backend === 'turso'
					? 'border-neutral-900 bg-neutral-900 text-white dark:border-neutral-100 dark:bg-neutral-100 dark:text-neutral-900'
					: 'border-neutral-200 bg-white text-neutral-800 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-200'
			]}
			type="button"
			onclick={() => setBackend('turso')}
		>
			<div class="flex items-center gap-1.5 text-sm font-medium">
				Turso
				<Tooltip
					label="About Turso"
					tip="Local libSQL file conduit-index.db in the plugin data directory. Fine for a single instance. Events remain in Congee; only parsed listings, geo, and embeddings live here."
				/>
			</div>
			<p class="mt-1 text-xs opacity-80">Local libSQL file in the plugin data directory.</p>
		</button>
		<button
			class={[
				'rounded-lg border p-4 text-left',
				settings.index_backend === 'postgres'
					? 'border-neutral-900 bg-neutral-900 text-white dark:border-neutral-100 dark:bg-neutral-100 dark:text-neutral-900'
					: 'border-neutral-200 bg-white text-neutral-800 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-200'
			]}
			type="button"
			onclick={() => setBackend('postgres')}
		>
			<div class="flex items-center gap-1.5 text-sm font-medium">
				Postgres
				<Tooltip
					label="About Postgres"
					tip="Shared SQL index for multi-instance relays. Point at a URL or local user/password. Keep this on postgres if the relay itself is postgres, otherwise search and events can diverge."
				/>
			</div>
			<p class="mt-1 text-xs opacity-80">Shared SQL index. Use a URL or user / password.</p>
		</button>
	</div>

	{#if settings.index_backend === 'turso'}
		<div class="rounded-lg border border-neutral-200 bg-white p-4 text-sm text-neutral-600 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-300">
			Path is automatic: <code class="rounded bg-neutral-100 px-1 dark:bg-neutral-800">conduit-index.db</code>
			inside the plugin data directory. No URL to configure.
		</div>
	{:else}
		<div class="flex gap-2">
			<button
				class={[
					'rounded-lg border px-3 py-1.5 text-sm',
					mode === 'auto'
						? 'border-neutral-900 bg-neutral-900 text-white dark:border-neutral-100 dark:bg-neutral-100 dark:text-neutral-900'
						: 'border-neutral-200 bg-white dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-200'
				]}
				type="button"
				onclick={() => setMode('auto')}
			>
				Auto (local)
			</button>
			<button
				class={[
					'rounded-lg border px-3 py-1.5 text-sm',
					mode === 'url'
						? 'border-neutral-900 bg-neutral-900 text-white dark:border-neutral-100 dark:bg-neutral-100 dark:text-neutral-900'
						: 'border-neutral-200 bg-white dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-200'
				]}
				type="button"
				onclick={() => {
					if (!settings.postgres_url) settings.postgres_url = 'postgres://';
				}}
			>
				Connection URL
			</button>
		</div>

		{#if mode === 'url'}
			<label class="block space-y-1">
				<InfoLabel
					text="Postgres URL"
					tip="Full connection string for the Conduit index (not the relay DSN unless you intentionally share a database). Stored in plugin settings; password in the URL is persisted — prefer user/password fields if you want the secret in the plugin secrets file."
				/>
				<input
					class="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none focus:ring-2 focus:ring-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-100"
					bind:value={settings.postgres_url}
					placeholder="postgres://user:pass@host:5432/conduit"
					autocomplete="off"
				/>
			</label>
		{:else}
			<div class="grid gap-3 sm:grid-cols-2">
				<label class="block space-y-1">
					<InfoLabel
						text="User"
						tip="Postgres role used when URL is empty. Connects to 127.0.0.1:5432/conduit."
					/>
					<input
						class="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none focus:ring-2 focus:ring-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-100"
						bind:value={settings.postgres_user}
						placeholder="postgres"
						autocomplete="off"
					/>
				</label>
				<label class="block space-y-1">
					<InfoLabel
						text="Password"
						tip="Stored in the plugin secrets file, not in settings JSON. Leave empty to keep the current password."
					/>
					<input
						class="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none focus:ring-2 focus:ring-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-100"
						type="password"
						bind:value={settings.postgres_password}
						placeholder="unchanged if empty"
						autocomplete="new-password"
					/>
				</label>
			</div>
			<p class="text-xs text-neutral-500 dark:text-neutral-400">
				Connects to <code>127.0.0.1:5432/conduit</code> when URL is empty.
			</p>
		{/if}
	{/if}

	<div class="flex flex-wrap items-center gap-3">
		<button
			class="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm font-medium text-neutral-800 disabled:opacity-50 dark:border-neutral-700 dark:bg-neutral-900 dark:text-neutral-100"
			type="button"
			disabled={testBusy}
			onclick={() => void ontest()}
		>
			{testBusy ? 'Testing…' : 'Test connection'}
		</button>
		{#if testResult}
			<span class="text-sm text-neutral-600 dark:text-neutral-300">{testResult}</span>
		{/if}
	</div>
</section>
