package embed

import (
	"context"
	"fmt"
	"os"
)

// NewONNX loads a MiniLM ONNX model if the file exists. v1 falls back when the runtime is unavailable.
func NewONNX(path string) (Embedder, error) {
	if path == "" {
		return nil, fmt.Errorf("onnx: empty path")
	}
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("onnx: model present at %s but onnxruntime is not linked in this build", path)
}

// Warm is a no-op for Fake; ONNX implementations should embed a dummy string.
func Warm(ctx context.Context, e Embedder) error {
	if e == nil {
		return fmt.Errorf("nil embedder")
	}
	_, err := e.Embed(ctx, "warmup")
	return err
}
