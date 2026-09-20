package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/michmich112/conduit-plugin/embed"
	"github.com/michmich112/conduit-plugin/handler"
	sdk "github.com/michmich112/congee/sdk/plugin"
)

func main() {
	dataDir := os.Getenv(sdk.EnvPluginDataDir)
	if dataDir == "" {
		dataDir = "."
	}
	model := filepath.Join(dataDir, "models", "minilm.onnx")
	sel := embed.Select(model)
	switch sel.Source {
	case embed.SourceExplicitFake:
		fmt.Fprintf(os.Stderr, "embedder fake-bow-384: CONDUIT_EMBEDDER=fake\n")
	case embed.SourceONNX:
		fmt.Fprintf(os.Stderr, "embedder onnx model_id=%s\n", sel.ModelID)
	default:
		fmt.Fprintf(os.Stderr, "embedder unavailable: %s\n", sel.Error)
		fmt.Fprintf(os.Stderr, "vector rank disabled until a real model loads or CONDUIT_EMBEDDER=fake\n")
	}
	h := handler.New(dataDir, sel)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	if err := sdk.Serve(ctx, h); err != nil && ctx.Err() == nil {
		os.Exit(1)
	}
}
