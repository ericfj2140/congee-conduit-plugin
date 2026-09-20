<script>
	import InfoLabel from '../lib/InfoLabel.svelte';
	import Switch from '../lib/Switch.svelte';
	import Tooltip from '../lib/Tooltip.svelte';

	let {
		settings = $bindable(),
		embedder = {},
		assets = {},
		testResult = '',
		testBusy = false,
		ontest,
		ondownload
	} = $props();

	let requiredDim = $derived(Number(settings.embed_dim) || embedder.required_dim || 384);

	function setProvider(value) {
		settings.embed_provider = value;
	}
</script>

<section class="space-y-6">
	<div>
		<div class="flex items-center gap-1.5">
			<h2 class="text-lg font-medium text-neutral-900 dark:text-neutral-100">Indexes</h2>
			<Tooltip
				label="About Indexes"
				tip="These switches control how the index is queried when Conduit intercepts a REQ. They do not change the client's filter kinds. Geo uses #g; vector rank uses embeddings when there is a search string."
			/>
		</div>
		<p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
			How Conduit filters and ranks listings when it intercepts a REQ.
		</p>
	</div>

	<div class="space-y-3">
		<Switch
			bind:checked={settings.active_filter}
			label="Active listings only"
			description="Only serve active listings on intercept respond. Inactive, draft, and deleted listings stay out of search hits."
			tip="When on, SQL only considers status=active rows. Sold NIP-99 listings, zero-quantity products, drafts (unless indexed as active), and deletions are omitted from ranked IDs."
		/>
		<Switch
			bind:checked={settings.geo_enabled}
			label="Geo index"
			description="Honor #g geohash prefix filters. Prefix length is configured on Search."
			tip="When a REQ includes #g, Conduit matches stored geohashes with a prefix LIKE, then sorts by haversine distance from the full geohash. Disable to ignore #g on intercept."
		/>
		<Switch
			bind:checked={settings.vector_enabled}
			label="Vector rank"
			description="Rank by embeddings when the REQ has a search string. Requires embeddings to be warm."
			tip="Search text is embedded with the configured model. Vector rank needs a packaged MiniLM ONNX runtime, a verified external provider that returns {requiredDim}-d vectors, or CONDUIT_EMBEDDER=fake for tests. Cosine similarity orders candidates."
		/>
	</div>

	<div class="space-y-3">
		<div class="flex items-center gap-1.5">
			<h3 class="text-sm font-medium text-neutral-900 dark:text-neutral-100">Embedding provider</h3>
			<Tooltip
				label="About Embedding provider"
				tip="On-device uses all-MiniLM-L6-v2 ({requiredDim}-d) from models/minilm.onnx plus the packaged onnxruntime library. External is an OpenAI-compatible POST /v1/embeddings endpoint. The response must be exactly {requiredDim} floats. Test must succeed before Conduit offloads MiniLM — Save after a passing Test."
			/>
		</div>
		<p class="text-sm text-neutral-500 dark:text-neutral-400">
			Set the vector width to match your model. Saving a new size rebuilds the index. On-device MiniLM
			is 384-d; use External HTTP for other widths.
		</p>

		<label class="block max-w-xs space-y-1">
			<InfoLabel
				text="Embedding dimensions"
				tip="Width of each stored vector. MiniLM is 384. An external provider must return this many floats (the probe sends dimensions=this value). Changing it marks old rows mismatched and rebuilds the index on Save."
			/>
			<input
				class="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none focus:ring-2 focus:ring-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-100"
				type="number"
				min="8"
				max="4096"
				bind:value={settings.embed_dim}
			/>
		</label>

		<div class="grid gap-3 sm:grid-cols-2">
			<button
				class={[
					'rounded-lg border p-4 text-left',
					settings.embed_provider !== 'http'
						? 'border-neutral-900 bg-neutral-900 text-white dark:border-neutral-100 dark:bg-neutral-100 dark:text-neutral-900'
						: 'border-neutral-200 bg-white text-neutral-800 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-200'
				]}
				type="button"
				onclick={() => setProvider('on_device')}
			>
				<div class="flex items-center gap-1.5 text-sm font-medium">
					On-device MiniLM
					<Tooltip
						label="About On-device MiniLM"
						tip="Packaged all-MiniLM-L6-v2 ONNX ({requiredDim}-d) plus libonnxruntime in the plugin tarball. No API key. Used until an external provider is tested and saved."
					/>
				</div>
				<p class="mt-1 text-xs opacity-80">
					Local model shipped with the plugin. No network call per search.
				</p>
			</button>
			<button
				class={[
					'rounded-lg border p-4 text-left',
					settings.embed_provider === 'http'
						? 'border-neutral-900 bg-neutral-900 text-white dark:border-neutral-100 dark:bg-neutral-100 dark:text-neutral-900'
						: 'border-neutral-200 bg-white text-neutral-800 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-200'
				]}
				type="button"
				onclick={() => setProvider('http')}
			>
				<div class="flex items-center gap-1.5 text-sm font-medium">
					External HTTP
					<Tooltip
						label="About External HTTP"
						tip="OpenAI-compatible embeddings API. URL can be a host, …/v1, or a full …/embeddings path. Model name is sent as-is. The JSON embedding array must have length {requiredDim} or Test fails. After a passing Test and Save, MiniLM is not loaded."
					/>
				</div>
				<p class="mt-1 text-xs opacity-80">
					Custom endpoint. Must return {requiredDim}-d vectors.
				</p>
			</button>
		</div>

		{#if settings.embed_provider !== 'http'}
			<div class="space-y-3">
				<label class="block space-y-1">
					<InfoLabel
						text="ONNX model URL"
						tip="Downloaded on plugin install and launch into the plugin data directory. Leave empty for the default Hugging Face all-MiniLM-L6-v2 ONNX. A failed download is reported in status; the plugin does not panic."
					/>
					<input
						class="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none focus:ring-2 focus:ring-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-100"
						bind:value={settings.embed_model_url}
						placeholder="default Hugging Face MiniLM ONNX"
						autocomplete="off"
					/>
				</label>
				<label class="block space-y-1">
					<InfoLabel
						text="onnxruntime archive URL"
						tip="Tarball of the ONNX Runtime C library for this OS/arch. Leave empty for the pinned Microsoft release. Install and launch hooks fetch this; Download now retries without restarting."
					/>
					<input
						class="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none focus:ring-2 focus:ring-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-100"
						bind:value={settings.embed_runtime_url}
						placeholder={'default onnxruntime {goos}_{goarch} tarball'}
						autocomplete="off"
					/>
				</label>
				<div class="flex flex-wrap items-center gap-3">
					<button
						class="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm font-medium text-neutral-800 disabled:opacity-50 dark:border-neutral-700 dark:bg-neutral-900 dark:text-neutral-100"
						type="button"
						disabled={testBusy}
						onclick={() => void ondownload()}
					>
						{testBusy ? 'Working…' : 'Download assets'}
					</button>
					{#if assets?.model_ok || assets?.runtime_ok}
						<span class="text-sm text-neutral-600 dark:text-neutral-300">
							Model {assets.model_ok ? 'ok' : 'missing'} · runtime {assets.runtime_ok ? 'ok' : 'missing'}
						</span>
					{/if}
					{#if assets?.error}
						<span class="text-sm text-amber-800 dark:text-amber-200">{assets.error}</span>
					{/if}
					{#if testResult}
						<span class="text-sm text-neutral-600 dark:text-neutral-300">{testResult}</span>
					{/if}
				</div>
			</div>
		{/if}

		{#if embedder?.http_verified}
			<div
				class="rounded-lg border border-neutral-200 bg-neutral-50 px-4 py-3 text-sm text-neutral-600 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-300"
			>
				External provider is active ({embedder.model_id || 'http'}). On-device MiniLM is not loaded.
			</div>
		{:else if settings.embed_provider === 'http'}
			<div
				class="rounded-lg border border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-900 dark:border-amber-800 dark:bg-amber-950/50 dark:text-amber-200"
			>
				Run Test against this URL and model, then Save. Until that fingerprint matches, Conduit
				keeps the on-device model (or disables vector rank if MiniLM is missing).
			</div>
		{/if}

		{#if settings.embed_provider === 'http'}
			<div class="space-y-3">
				<label class="block space-y-1">
					<InfoLabel
						text="Embeddings URL"
						tip="OpenAI-compatible POST target. A root like https://api.openai.com/v1 becomes /v1/embeddings. Ollama: http://127.0.0.1:11434/v1. The probe sends dimensions={requiredDim}."
					/>
					<input
						class="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none focus:ring-2 focus:ring-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-100"
						bind:value={settings.embed_http_url}
						placeholder="https://api.openai.com/v1"
						autocomplete="off"
					/>
				</label>
				<div class="grid gap-3 sm:grid-cols-2">
					<label class="block space-y-1">
						<InfoLabel
							text="Model"
							tip="Provider model id. It must emit {requiredDim} floats. OpenAI text-embedding-3-small is 1536 by default — only use it if the API honors dimensions={requiredDim}. Ollama all-minilm is {requiredDim}-d."
						/>
						<input
							class="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none focus:ring-2 focus:ring-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-100"
							bind:value={settings.embed_http_model}
							placeholder="all-minilm"
							autocomplete="off"
						/>
					</label>
					<label class="block space-y-1">
						<InfoLabel
							text="API key"
							tip="Stored in the plugin secrets file, not in settings JSON. Leave empty to keep the current key. Omit for local servers that do not authenticate."
						/>
						<input
							class="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none focus:ring-2 focus:ring-neutral-400 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-100"
							type="password"
							bind:value={settings.embed_http_api_key}
							placeholder="unchanged if empty"
							autocomplete="new-password"
						/>
					</label>
				</div>
				<div class="flex flex-wrap items-center gap-3">
					<button
						class="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm font-medium text-neutral-800 disabled:opacity-50 dark:border-neutral-700 dark:bg-neutral-900 dark:text-neutral-100"
						type="button"
						disabled={testBusy || !settings.embed_http_url}
						onclick={() => void ontest()}
					>
						{testBusy ? 'Testing…' : 'Test provider'}
					</button>
					{#if testResult}
						<span class="text-sm text-neutral-600 dark:text-neutral-300">{testResult}</span>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</section>
