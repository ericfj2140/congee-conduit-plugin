# Conduit

Marketplace index plugin for [Congee](https://github.com/michmich112/congee). Indexes NIP-15 and NIP-99 listings already stored on the relay (geo + vector rank). It never writes events to the relay.

Depends only on `github.com/michmich112/congee/sdk/plugin`.

## Layout

- `listing/` — parse NIP-15 / NIP-99 / kind 5
- `embed/` — Fake (e2e) or ONNX MiniLM when a model file is present
- `index/` — Turso default, Postgres optional, ANN cache, Search
- `handler/` — gRPC Handler, intercept decisions, watermark backfill
- `web/` — Svelte 5 static UI compiled to `ui/`

## Run under Congee

Install from a local package directory (plugin.json + bin/ + ui/). Set `CONDUIT_EMBEDDER=fake` for tests. Raise `plugins.intercept_timeout_ms` to at least 200 (250 in `config.example.json`).
