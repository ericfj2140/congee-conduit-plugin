# Conduit

Marketplace index plugin for [Congee](https://github.com/michmich112/congee). Indexes NIP-15 and NIP-99 listings already stored on the relay (geo + vector rank). It never writes events to the relay.

Depends only on [`github.com/michmich112/congee/sdk/plugin`](https://github.com/michmich112/congee/tree/main/sdk/plugin):

```bash
go get github.com/michmich112/congee/sdk/plugin@v0.1.0
```

A sibling Congee checkout is not required. For local ABI iteration, `cp go.work.example go.work` (do not commit `go.work`).

Build with `CGO_ENABLED=1` (Turso/libSQL + onnxruntime).

## Layout

- `listing/` — parse NIP-15 / NIP-99 / kind 5 using roles from `kinds.json`
- `kinds.json` — kind numbers, NIPs, names, and roles (stall, product, listing, community, …)
- `embed/` — on-device MiniLM (onnxruntime) unless `CONDUIT_EMBEDDER=fake`
- `index/` — Turso default, Postgres optional, ANN cache, Search
- `handler/` — gRPC Handler, intercept decisions, watermark backfill
- `web/` — Svelte 5 static UI compiled to `ui/`

## Install

GitHub Releases ship per-platform tarballs (`plugin.json` + `bin/conduit-plugin` + `ui/`). In the Congee admin UI, **Plugins → Install** with the tarball URL and the **archive** SHA-256 from the release notes.

Raise `plugins.intercept_timeout_ms` to at least 200 (250 in Congee `config.example.json`).

On-device ranking downloads MiniLM ONNX, `tokenizer.json`, and onnxruntime into `$CONGEE_PLUGIN_DATA_DIR` (`plugins/conduit/data/`) via `--hook=install` / `--hook=launch`. Uninstall runs `--hook=uninstall` to delete those blobs; `wipe_data` also drops the index. Set `CONDUIT_EMBEDDER=fake` for tests only.

Vector width is `embed_dim` (default 384). Indexes can use a verified OpenAI-compatible embeddings URL that returns that many floats; after Test + Save, MiniLM is not loaded.
