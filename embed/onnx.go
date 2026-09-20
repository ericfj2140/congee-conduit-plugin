package embed

import (
	"context"
	"fmt"
	"os"
	"runtime"
)

// NewONNX loads a MiniLM ONNX model if the packaged files exist.
// This build does not link onnxruntime; the model and shared library are still
// resolved from the plugin package so a linked release can run out of the box.
func NewONNX(path string) (Embedder, error) {
	if path == "" {
		return nil, fmt.Errorf("onnx: empty path")
	}
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("onnx: model not found at %s (%w); ship models/minilm.onnx in the plugin package", path, err)
	}
	lib, err := findRuntimeLib("")
	if err != nil {
		return nil, fmt.Errorf("onnx: model present at %s but onnxruntime library missing (%v); ship lib/%s_%s/libonnxruntime.* in the plugin package", path, err, runtime.GOOS, runtime.GOARCH)
	}
	return nil, fmt.Errorf("onnx: model present at %s and runtime at %s but onnxruntime is not linked in this build", path, lib)
}

// Warm embeds a dummy string so a remote or ONNX backend fails fast at apply.
func Warm(ctx context.Context, e Embedder) error {
	if e == nil {
		return fmt.Errorf("nil embedder")
	}
	return Probe(ctx, e)
}
