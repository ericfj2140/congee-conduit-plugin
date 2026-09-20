# onnxruntime shared libraries

Put the ONNX Runtime C library for this platform here:

```
lib/<goos>_<goarch>/libonnxruntime.dylib   # darwin
lib/<goos>_<goarch>/libonnxruntime.so      # linux
lib/<goos>_<goarch>/onnxruntime.dll        # windows
```

`scripts/fetch-embed-assets.sh` downloads a pinned onnxruntime release into this tree. The plugin prepends this directory to the dynamic linker path at process start.

Do not commit the binaries; they belong in the per-arch plugin tarball.
