# Packaged MiniLM (384-d)

Conduit ranks listings with **all-MiniLM-L6-v2** vectors (`dim=384`). The ONNX weights and the onnxruntime shared library ship **next to the plugin binary**, not inside it (`go:embed` would add ~80MB to every build and still need a native runtime).

## Package layout

```
conduit/
  plugin.json
  bin/conduit-plugin
  models/minilm.onnx
  lib/<goos>_<goarch>/libonnxruntime.dylib   # or .so / onnxruntime.dll
  ui/
```

At startup the process looks for `models/minilm.onnx` in:

1. The plugin package root (parent of `bin/`)
2. Beside the executable
3. `$CONGEE_PLUGIN_DATA_DIR/models/` (override / operator drop-in)

The runtime library is resolved the same way under `lib/<goos>_<goarch>/` then `lib/`. The process prepends that directory to `DYLD_LIBRARY_PATH` / `LD_LIBRARY_PATH` / `PATH` so a linked binary can `dlopen` the packaged copy.

Install and upgrade copy the whole package directory and keep `data/` — so model and runtime update with the plugin; local secrets stay.

## Fetch pinned assets

From the plugin repo:

```bash
./scripts/fetch-embed-assets.sh
```

That writes `models/minilm.onnx` and `lib/$GOOS_$GOARCH/libonnxruntime.*` with sha256 checks. Release tarballs should include those files for the target OS/arch. Do not commit the blobs.

## Out of the box

A release that includes both files is what makes on-device ranking work without extra operator steps. This tree still needs an onnxruntime-linked plugin build to *run* the session; until that binary is produced, vector rank stays off unless `CONDUIT_EMBEDDER=fake` (tests) or a verified external HTTP provider is configured (Indexes).

External providers must return **384** floats per input (same width as MiniLM). After Test succeeds and settings are saved with provider “External”, the on-device model is not loaded.
